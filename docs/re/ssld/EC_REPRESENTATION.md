# ssld PKEX EC representation (cracked) + why byte-exact reconstruction stalled

## Representation
libucrypto stores EC coordinates/BigNums as **Montgomery form, little-endian bytes**.
To recover the affine integer from a captured 32-byte limb dump `h`:
`affine = int(reverse_bytes(h)) * Rinv mod p`, where `R = 2^256 mod p` (P-256).
Verified: the constant masking point's captured limbs
`cb2c53bb8680d462b67b5d44e6c19a6e52109714eeab88c7bb6e92780f8da122`
decode to a valid on-curve affine x:
**P.x = 418e71a41c451893efa10f10c958f36cbe68857cd3698f82a6387db9a8c24f9f**
(constant across pairings for password "LabPass123!" -> the password point / a fixed base).
Wire masked points (id11/id0, `[32B big-endian x][1B y-parity]`) are standard affine.

## Live flow (plugin capture, docs/re/ssld/pkex_full_pairing.log)
Post-TLS (after the TLS-1.2 Finished PRF): ~3 `WCurve.mul` + 2 `WCurve.add` in ssld's
inlined handler. A full pairing completes: id6 discovery -> id10 -> id0 descriptor ->
id11 commit (username + 33B masked point) -> id0/id4 (32B) responses -> id10 `E 00`
SUCCESS -> `com.mikrotik.cmr.config` push. Wrong password => stuck `waiting-for-pairing`.

## Why it stalled
- The PKEX is **inlined inside the 73KB TLS state-machine** (`FUN_0806004a` / `FUN_08053d82`,
  Ghidra "type propagation not settling") — no clean separable function to decompile.
- EC internals are **Montgomery form**; the captured add/mul operands do not satisfy the
  obvious SPAKE2 relations (`A±P != M_i`, `y*G != A`), so the exact op wiring
  (which operand is ephemeral vs recovered-peer vs password-scaled, and where the
  Montgomery<->affine conversions sit) can't be pinned from captures alone.
- No PKEX-specific HMAC was observed (all HMAC was TLS PRF), so the confirm is likely
  **bound to TLS key material** (recurring 48-byte HMAC key == TLS master-secret size),
  i.e. the handshake is welded to ssld's TLS internals rather than a clean m3-layer PKEX.

Reaching byte-exact would require per-instruction tracing of the inlined handler (register
+ Montgomery-context state at each EC op) or source — a very large, uncertain effort.
