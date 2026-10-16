# The CMR password PKEX actually runs in `ssld` — captured live via a QEMU plugin

**Correction #2 (2026-10-16):** the password handshake runs in neither `cmr` nor
`figman`. It runs in the base-RouterOS TLS terminator **`ssld`** (`/nova/bin/ssld`,
the daemon on tcp/54321). `cmr` and `figman` both carry PKEX *source* but it is dead
code for this flow — `ssld` authenticates the channel (TLS **and** the password
handshake) and only then hands plaintext m3 to `cmr`.

## How this was captured (hypervisor-level, no guest shell)

The RouterOS CHR guest has no shell and signed NPKs, and the QEMU gdbstub is
unusable here (KVM: sw breakpoints never fire, hw breakpoints clobbered by the guest
kernel; TCG+gdb too slow → session-overlap). So I wrote a **QEMU TCG plugin**
(`tools/figtrace/figplugin.c`, host-side, no guest modification) that watches crypto
addresses and dumps guest buffers. Boot srv under TCG with `-plugin file=figplugin.so`.

- **libucrypto.so load base = `0xf7edd000`** (ASLR is OFF — consistent across cmr,
  figman, ssld and reboots). Derived from the exe GOT (cmr HKPDF GOT `0x809a4c8` /
  figman `0x807c4b0` → libucrypto HKPDF; base = that − `0x12106`).
- Watch **libucrypto function entries directly** (base + offset) to catch every
  caller, exe or library: HKPDF `+0x12106`, HMAC ctor `+0x10b8c` / upd `+0x10c6e` /
  dig `+0x10cac`, AESSIV enc `+0x19f6e` / dec `+0x1a078`, SHA256 `+0x10af2`,
  WCurve.mul `+0x16680` / add `+0x16290`, ECPoint.toBin `+0x15d60` /
  fromOctetString `+0x15b70`, getECParamsForP256 `+0x1f554`.
- Process identity per hit: read `com.mikrotik.cmr` sig at `0x807592d` (figman) /
  `0x809070a` (cmr); neither ⇒ the PKEX process (`ssld`).

## What the live pairing actually does (plugin capture, docs/re/ssld/pkex_pairing_capture.log)

During a full MITM-forced pairing (fresh identity, srv device table cleared):

- **ssld** (`?`): `WCurve.mul` ×12 and `WCurve.add` ×593 (the double-and-add
  internals), plus `HMAC` (ctor/upd/dig). **No HKPDF, no AES-SIV, no SHA256 anywhere.**
  So the earlier "HKDF(HKPDF) + AES-SIV" model (from the cmr/figman dead code) is wrong
  for the live flow: the KDFs/confirm are **HMAC-based**, the point exchange is EC.
- cmr/figman: **zero** crypto during the pairing (only at boot, for identity).
- The other `?` HMACs are ssld's TLS-1.2 PRF (labels "master secret", "key expansion",
  "client/server finished") — not PKEX.

### Captured PKEX material (scalars are `WCurve.mul` BigNum derefs; 32-byte LE/limbs)

- **Qpw scalar (constant for the password "LabPass123!")**, from `WCurve.mul` called at
  ssld `.text` `0x80593ac`:
  `32a33283e5cdeca0106f6915672d41de1ef48a61f9731e9997e754bd46d254c3`
  (same every call ⇒ password-derived; not via HKPDF/SHA256 ⇒ an HMAC-based map).
- Per-session ephemeral / ECDH scalars (vary each run): `6e0029fb…b55e97ff`,
  `8dabf012…a7f6057f`, used from ssld `0x8055321` / `0x8056e31` / `0x8056f45`.
- **HMAC confirm/KDF** (from ssld `0x8058182`): 32-byte key
  `0e20c58129e40a811b432d9bc696c06c85ad0b6b6488bf784ed97969b0a93b88` and a recurring
  48-byte session key `ff69cfa4544e1011b1b53db19dd1154f96927624310e2eb16e031d477103a472
  1395bcb395b9b41e3133d411c73ec8c6`; the transcript is fed as 32-byte elements
  (point coordinates) via `HMAC.update`.

### ssld PKEX code map (caller sites of the EC/HMAC ops, ssld ET_EXEC @0x8048000)

`WCurve.mul` @ `0x80593ac` (Qpw=s·G), `0x8055321`, `0x8056e31`, `0x8056f45`;
`HMAC` ctor @ `0x8058182`; PRF HMAC @ `0x8057de9/0x8057e04/0x8057e13`. Decompiling
`ssld`'s functions around these is the next step to pin the exact scalar/KDF/transcript
construction.

## Status / remaining work

The capture pipeline is reliable and reusable (reboot srv under TCG+plugin, run a
MITM-forced pairing, read `/tmp/figplugin.log`). Remaining to a byte-exact controller:
decompile `ssld`'s PKEX (password→scalar map, the HMAC key schedule, the exact confirm
transcript + EC flow), reimplement it, and wire it under the TLS layer (`pkg/mtls`)
ahead of the m3 management plane. The prior `pkg/pairing` (AES-SIV + HKDF P-256) was
built against the cmr/figman dead code and does **not** match ssld's live scheme.
