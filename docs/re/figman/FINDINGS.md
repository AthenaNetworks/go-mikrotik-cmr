# figman is the CMR pairing/PKEX engine (not cmr)

**Key correction (2026-10-16):** the password PKEX handshake does **not** run in the
`cmr` package binary. It runs in the base-RouterOS daemon **`figman`**
(`/nova/bin/figman`, "FigConn manager"). The `cmr` binary contains a *copy* of the
PKEX source (same strings/addresses we decompiled earlier) but that copy is **not
executed** during pairing — `cmr` only receives the final `cmr::ClntCfg::onPKEXResult`
callback. All prior RE that targeted `cmr`'s PKEX addresses was looking at dead code.

Verified: running `figman` is byte-identical (md5 `e1cb669f8bdaef9b7576407aacf0a738`)
to `sysextract/rootfs/nova/bin/figman`, confirmed against the live CHR image at
`var/pdb/system/image`. Only `figman` (of all base `/nova/bin/*`) imports the full
PKEX crypto set (`AESSIV::encrypt`, `HKPDF`, `getECParamsForP256`) **and** carries the
`PKEXSrvImpl`/`PKEXClntImpl`/`pkex/pkex.cpp` strings (`mproxy`/`ssld` import only 1
crypto fn each).

## figman crypto PLT stubs (ET_EXEC, fixed @0x8048000)

| function | PLT |
|---|---|
| `HKPDF(Hash&, vec, vec, vec, uint)` | `0x80502e0` |
| `HMAC::HMAC(Hash*, key, keylen)` | `0x80502d0` |
| `HMAC::update(data, len)` | `0x804f120` |
| `HMAC::digest(out)` | `0x80500d0` |
| `AESSIV::encrypt(span pt, vec<span> AD, out)` | `0x804fc90` |
| `AESSIV::decrypt(span ct, vec<span> AD, out)` | `0x8050540` |
| `WCurve::mul(BigNum, Point)` | `0x804f490` |
| `WCurve::add(Point&, Point&, RedNum*)` | `0x8050220` |
| `ECPoint::toBin(len, compressed)` | `0x8050320` |
| `getECParamsForP256(BigNum&, BigNum&, ECPoint&)` | `0x8050640` |
| `BigNum::bin(len)` | `0x804f3d0` |
| signature string `com.mikrotik.cmr` | `0x807592d` (CR3/process identity probe) |

## figman PKEX functions (decompiled → *.c here)

- `0x806f408` **`PKEX::start()`** — role dispatch (`*role==1` initiator alloc 0xe4,
  `==2` responder alloc 0xd8); builds the PKEX impl, zeroes the context, then
  `FUN_0806f38c` / `FUN_0806ca0e` / `FUN_0806d06e` do the per-role element setup.
  The Qpw/"code" field (`ctx+0xb8`) is written at `0x806f6c9`/`0x806f889`.
- `0x806f1e4` — **identity** derivation (rosMode + eddsa pubkey + HKPDF, salt = the
  static SHA256 `Hash` object at `0x807bcb0`). NOT the password scalar.
- `0x806e57e` **`PKEXImpl::authDeriveZ(const BigNum&, const BigNum&)`** —
  `K = ecPeerProt · x_own` (WCurve::mul→toEC), then
  `Z = HKPDF(hash=createHash(5)=SHA256, ikm=K.x [BigNum::bin], salt=empty, info=…, len=0x20)`.
  Context fields referenced: `+0xa0` kProt, `+0xb0` ecPeerProt, `+0xb8` code.
- `0x806e820` **`PKEXImpl::reveal1(vector&)`** — the confirm tag:
  `J = ecPeerProt · ...` (WCurve::mul→toEC), `key = J.x` (BigNum::bin),
  `HMAC(hash=SHA256@0x807bcb0, key=J.x)`, then **update** with three `BigNum::bin`
  values from `ctx+0x60`, `ctx+0xb0` (ecPeerProt), `ctx+0xa0+4` (kProt), then `digest`
  → `u/v`. This is the DPP-style `u = HMAC(J.x, …)` confirm.
- `0x806b886` **`PKEXImpl::encrypt(m3::msg_view, vector&)`** — `AESSIV::encrypt` of a
  message under `encryptionKey` (`ctx+0xbc..0xc0`), 16-byte SIV. Used for the config
  attribute exchange (`com.mikrotik.cmr.config`), **not** the point exchange.
- `0x806b974` **`PKEXImpl::decrypt(span, vector&)`** — matching `AESSIV::decrypt`.
- `0x806fe4a` — masking `M = X + Qpw` via `WCurve::add`.

`0x807bcb0` is the static **SHA256 `Hash` vtable/object**, passed as the `Hash&` arg to
both `HKPDF` and `HMAC` (so both use SHA-256), not a label string.

## Wire flow (plaintext m3, captured via cmr-mitm — see docs/samples/)

`C→S id6{scope,"password required",nonce}` · `S→C id10{scope,02,"password required",nonce}`
· `C→S id0{device descriptor: build,x86,routeros,CHR,112B id80}`
· **`C→S id11{scope, id32="admin\0"+33B masked point M_i}`**
· `S→C id0{id32 = 49B = 16 SIV + 33B}` · `C→S id4{id33 = 32B confirm}`
· `S→C id0{id33 = 32B confirm}` · `S→C id10{27=00 SUCCESS}`.
Masked point encoding = `[32B big-endian X][1B y-parity]`. Wrong password ⇒ pairing
stays `waiting-for-pairing` (password is genuinely checked).

## Why the exact labels are still not byte-pinned

The remaining unknowns are the exact KDF/MAC vector contents: the password→Qpw
derivation (nested under `PKEX::start`), the `Z` `info` vector, the precise confirm
transcript ordering, and `BigNum::bin` width. Two ways to close this:

1. **Runtime capture** — blocked in this lab's QEMU: under **KVM** gdbstub software
   breakpoints never fire (int3 not intercepted) and hardware breakpoints (max 4,
   DR-register) get clobbered by the guest kernel across the PKEX's blocking I/O, so
   the one-shot PKEX crypto is missed; under **TCG** gdb breakpoints work (verified:
   figman caught via getJiffies) but gdb+TCG is ~20–50× slower, so the handshake hits
   session-overlap timeouts ("not talking to you!") before reaching the commit. The
   gdbstub also wedges on reconnect (one session per VM boot).
2. **Native in-guest hook** (reliable, not yet done): LD_PRELOAD a musl-i386 `.so`
   that interposes `HKPDF`/`HMAC*`/`AESSIV*` and logs exact I/O during a normal
   KVM-speed pairing — needs a musl-i386 build + writing `/etc/ld.so.preload` into
   `srv.qcow2`. Or a surgical binary patch forcing a deterministic ephemeral so the
   existing captures become crackable offline. Both modify the guest image.
