# CMR protocol notes (reverse-engineered)

This document records what the CMR enrollment/control protocol looks like, as
recovered **statically** from MikroTik's `cmr` binary (RouterOS 7.26beta1,
i386) using Ghidra. It separates what was **observed** in the disassembly from
what is **inferred** or **still unverified**. The Go code in this repo
implements the observed structure; everything marked *pending* needs a packet
capture or on-device test to confirm byte-level interoperability.

## Process shape (observed)

`cmr` is a RouterOS "Nova" daemon. `main` (`FUN_08052c55`) builds an
`nv::Looper` event loop and registers managed-object stores on the Nova message
bus: `dev`, `wifiradio`, `wifinet`, `vlan`, `layout`/`laynode`/`laylink`,
`upgrade`/`upgrade/job`, `alert`, `inst`. Other RouterOS processes (console,
WebFig/figman, www) drive it by reading/writing those stores. This project
replaces the bus with an in-process `model.Store` + observer hooks.

## Enrollment handshake (observed)

Built on **PKEX** (the Public Key Exchange from Wi-Fi DPP / Easy Connect), over
**NIST P-256** (`getECParamsForP256`). Classes: `PKEX`, `PKEXImpl`,
`PKEXSrvImpl` (responder/controller), `PKEXClntImpl` (initiator/enrollee),
transports `PKEXPushButton` / `PKEXConsole`.

Frame dispatch is keyed by a message-type byte. Both `PKEXSrvImpl::handleFrame`
(`FUN_0808cf6a`) and `PKEXClntImpl::handleFrame` (`FUN_0808dcb4`) were
decompiled; the type map is **observed** (not guessed):

| Type | Name            | Direction        | Handled by            |
|------|-----------------|------------------|-----------------------|
| 0x40 | DiscoveryReq    | client -> server | `PKEXSrvImpl`         |
| 0x41 | DiscoveryResp   | server -> client | `PKEXClntImpl`        |
| 0x42 | Msg1            | server -> client | `PKEXClntImpl`        |
| 0x43 | Msg2            | client -> server | `PKEXSrvImpl`         |
| 0x44 | Msg3            | server -> client | `PKEXClntImpl`        |
| 0x45 | Msg4            | client -> server | `PKEXSrvImpl`         |

Server state machine (`this+0xc8`, values observed in the decompile):

```
0 Init        -> on DiscoveryReq: log "session started", go 1, send DiscoveryResp
1 Discovery1  -> (discovery handshake continues)
2 Discovery2  -> log "discovery finished"; begin PKEX (send Msg1)
3 Discovery3  -> on Msg2: auth ("auth failed (2)"), send Msg3, go 4
4 Auth        -> on Msg4: decrypt, read "peer id" (32 bytes) + "received addrs",
                 finish -> Srv::onPKEXResult -> add device to /dev store
5 Done/killed -> e.g. "session overlap detected"
```

Per-message crypto primitives (observed as `PKEXImpl` methods, calling the
external `BigNum`/`WCurve`/`ECPoint` library): `encrypt`/`decrypt`,
`getPoint`, `reveal1`/`reveal2`, `authDeriveZ`, `auth1`/`auth2`. Error gates
`auth failed (1..4)` and `protocol version missmatch` (a version field is
checked `== 1` in `handleMsg1`).

On success, `Srv::onPKEXResult` (`FUN_0805dac0`) copies the 32-byte peer id and
inserts a row into the `dev` store (observed `m3` field ids `0x2001`, `0x200b`).

## What is NOT in the binary (substituted here)

- **EC math** (`BigNum`/`WCurve`/`ECPoint`) and **`m3::message` serialization**
  are external RouterOS shared-library symbols, not statically present. This
  repo uses Go's `crypto/ecdh` P-256 and its own `m3` codec.

## Pending validation (do not assume interop yet)

1. **Exact PKEX key schedule**: the DPP hash-to-curve for the password element,
   the KDF labels, and the `auth1`/`auth2` tag inputs are not byte-confirmed.
   The crypto is therefore behind a `pkex.Suite` interface; the bundled
   `ReferenceSuite` is a correct P-256 PKEX-shaped handshake that interoperates
   **with itself** (client role <-> server role in this repo), not yet with
   MikroTik devices.
2. **`m3` wire bytes**: our codec is structurally faithful (typed TLV: u32,
   bytes, vector, nested) but not byte-identical to MikroTik's `m3`.
3. **Transport**: real devices exchange these frames as link-layer frames via
   the wifi stack (`PKEXPushButton::getIfs`/`send(...MacAddr)`). The bundled
   transport is UDP for testing; an 802.11 action-frame transport is the seam
   to implement for real interop.

Confirming (1)–(3) needs a capture of a real CAP<->controller enrollment, or
dynamic analysis on a device. Until then this is a complete, self-consistent
controller, not a verified MikroTik-compatible one.

## Validated against a lab device (2026-10-14)

Tested against MikroTik CHR 7.26beta1 running the `cmr` client, in an isolated
two-VM lab (controller 10.99.0.113, client 10.99.0.114 on an L2 bridge).

- **Transport confirmed:** the CMR control plane is **mutual TLS 1.2 on tcp/54321**.
  The controller listens; the client connects from an ephemeral port. The client
  is aimed with `/cmr/client/set controller-addresses=<ip> enabled=yes`; the
  pairing password is set via `/cmr/client/pair password=...` (this is the PKEX
  code).
- **Identity/trust confirmed:** self-signed **Ed25519** certs. The lab client
  cert: `CN="figman-1fef"` (CN = `figman-` + first 2 bytes of the id, hex),
  validity 1970-01-01 .. 9999-12-31, `CA:TRUE`, self-referential SKI/AKI. The
  CMR **device id is the SHA-256 of the identity** (this device: `1fef...`),
  matching the 32-byte peer id PKEX exchanges.
- **Interop gotcha (important):** MikroTik certs encode the `signatureAlgorithm`
  as a **zero-length OBJECT IDENTIFIER** (empty OID) + NULL, in both
  `tbsCertificate.signature` and the outer signature. Go's `crypto/x509` rejects
  this with `x509: malformed OID`; OpenSSL tolerates it. A Go controller must
  parse these leniently (rewrite the empty sig-alg OID to Ed25519 `1.3.101.112`
  before `x509.ParseCertificate`) or terminate TLS via OpenSSL/socat. The device
  **accepts a standard Ed25519 server cert** from us (our half of the handshake
  works).
- **First inner (post-TLS) client frame captured** (`docs/samples/first-inner-client-frame.hex`):
  `11 dc d2 11 01 d8 70 7e ae 26 5b 81 f4 c7` — the client sends this 14-byte
  hello and waits for a server reply. Our `m3` codec did not match it (expected);
  decoding it plus the required server reply is the next milestone and needs a
  bidirectional responder, not the one-way capture used here.

A captured client certificate is saved at
`docs/samples/mikrotik-cmr-client-cert.der` for parser work.

### Reproduce
- `go run ./cmd/cmr-learn -listen <ip>:54321` — a Go mutual-TLS endpoint; it
  demonstrates the empty-OID blocker (fails to parse the client cert) until
  lenient parsing is added.
- `socat -u OPENSSL-LISTEN:54321,bind=<ip>,cert=ed.crt,key=ed.key,verify=0,fork OPEN:plain.bin,creat,append`
  terminates TLS via OpenSSL and captures the inner client plaintext.

## Discovery mechanism (answered 2026-10-15)

Determined from the binary and live lab captures:

- **Connection = direct IP + TCP.** The client is given `/cmr/client/set
  controller-addresses=<ip>` and opens **tcp/54321** (mutual TLS). Lab captures
  show only TCP/54321 for CMR.
- **Not DNS SRV.** Zero DNS traffic in captures; no DNS/SRV resolver code or
  strings in the `cmr` binary.
- **Not MNDP.** The `cmr` binary contains no MNDP code (no `5678`/MNDP strings).
  MNDP frames on the wire are RouterOS's *general* neighbor-discovery service
  (TLVs: Identity, Version, Software-ID, Board, interfaces) that populates
  `/ip/neighbor` independently of CMR — CMR does not use it to find controllers.
- **CMR's own zero-config pairing is `push-button`**: link-layer PKEX discovery
  (`DiscoveryReq 0x40` / `DiscoveryResp 0x41`, broadcast per interface via
  `PKEXPushButton::getIfs`/`send`), not MNDP or DNS. After pairing the controller
  address is stored and used directly over TCP/TLS.

## Inner control handshake — pure-Go observation (2026-10-15)

Terminating the TLS in pure Go (`cmd/cmr-learn`, no socat/OpenSSL):

- The client completes **TLS 1.2 without presenting a client cert** when the
  server does not request one (`ClientAuth: NoClientCert`), cipher
  `ECDHE-ECDSA-AES128-GCM-SHA256`. So a pure-Go controller needs neither socat
  nor lenient x509 to read the inner protocol; mutual-auth client-cert handling
  (the empty-OID issue) only matters if/when we verify device identity at the
  TLS layer. CMR's trust is the app-layer pinned-fingerprint pairing.
- Frame sequence (revealed by mirroring each frame back — a probe, not real
  controller content; one app frame per TLS record, no explicit length prefix):
  1. client 14B hello: `11 dc d2 11 | 01(ver) | d8 70 7e ae 26 5b 81 f4 c7` (token constant across connections)
  2. client 23B: `09 e5 b1 d4 | <4 rand> | 35 d8 | 4b 60 be 6a` (build time 0x6abe604b) `| 01 62 1a 07` (version bytes, same as NPK pkg-info) `| 1d d3 | "CHR"` (identity)
  3. client 10B: `31 d8 b1 d4 d0 89 3e fe 1d 07`
  4. client 10B: `09 d8 b1 d4 15 08 f0 fc 1d 07`
- Mirroring advances the client through these steps. Real pairing needs correct
  per-frame replies — next milestone: decode each frame's semantics and craft
  controller responses.

`cmd/cmr-learn` flags: `-listen`, `-reply mirror|mirror-all|serverhello|none`,
`-mutual` (request client cert; fails on MikroTik's empty-OID until lenient
parsing is added).

## m3 wire format — fully decoded (2026-10-15)

The post-TLS frames are **nv/m3 messages**. The codec is NOT in the `cmr`
binary; it lives in the base-RouterOS `libumsg.so`, which `cmr` links. The
base system ships as an NPK (`/var/pdb/system/image` inside the CHR disk:
magic `0xbad0f11e`, squashfs at 0x1000) — extract it and `libumsg.so` is under
`/lib`. The relevant symbols (i386, exported in `.dynsym`):
`m3::message::{fold,unfold,measure}`, `m3::{encnum32,encnum64,fromraw32,getraw,
putraw}`, `m3::message::read::{each,consume}`. `pkg/m3` reimplements them
byte-for-byte (validated by `pkg/m3/m3_test.go` against the live frames).

**Structure** (`message = field*`, `field = tag-chunk value-chunk`):

- **chunk** = `intro` byte + data:
  - `intro <= 0xcf` → the value is that single byte (inline integer / 1-byte blob).
  - `0xd0..0xef` → a raw run of `intro-0xd0` bytes (lengths 0..31).
  - `0xfe/0xfd/0xfc/0xfb/0xfa` → length = next `1/2/4/8/16` little-endian bytes, then that many data bytes.
  - `0xff` → padding, skipped.
- **tag** → `(type, id) = unfold(tag, prevID)`: `type = tag & 3` (2 physical wire
  types; the logical type is resolved from the id's schema). The id is
  **absolute** `tag>>3` when bit2==0, else **delta-coded** from the previous
  field's id: `id = prevID ± (tag>>4)`, sign in bit3. `fold` picks whichever of
  the absolute/delta forms is numerically smaller.
- Numbers are little-endian raw bytes; nested messages are a field whose value
  is itself an encoded message (schema-driven, not flagged on the wire).

**Decoded handshake frames** (client -> server; all `type=1`):

| Frame | Opcode (outer id) | Contents |
|-------|-------------------|----------|
| F1 | id=2  | 12-byte token `d2 11 01 d8 70 7e ae 26 5b 81 f4 c7` (constant) — hello |
| F2 | id=1  | sub-msg { id=22: 4B nonce, id=25: 8B = build-time `0x6abe604b` + version `0x071a6201`, id=24: "CHR" } |
| F3 | id=6  | sub-msg { id=22: 4B nonce, id=21: 1B sequence } |
| F4 | id=1  | sub-msg { id=22: 4B nonce, id=21: 1B sequence } |

Field ids: **id=22** = per-message 4-byte nonce; **id=21** = 1-byte session
sequence (7, then 8 on the next connection); **id=25** = build-time+version
(matches the NPK header); **id=24** = board/identity string. Outer ids 1/2/6 are
the FigConn **respid**/opcode dispatch (`FigConn::recv: unhandled respid`).

No EC/DH key material appears in the client's first four frames, so the
password **key handshake** (`key handshake timeout (wrong pass)`) is driven by
the **server's** replies, which the mirror probe cannot produce. Capturing the
real server replies (a cert-tolerant TLS client to a live controller, since the
empty-OID cert blocks Go *and* OpenSSL on the client side) is the next step.

## Full bidirectional handshake — captured from a live controller (2026-10-15)

`pkg/mtls` is a minimal, pure-Go TLS 1.2-ECDHE client that completes the
handshake **without parsing the server certificate** (the only cipher is
`0xc02b`; the Ed25519 cert is not needed to derive the ECDHE keys), sidestepping
the empty-OID blocker. `cmd/cmr-dial` uses it to connect to a real controller
as an enrollee and decode the replies.

Findings against CHR 7.26beta1 controller (10.99.0.113):

- The controller **requires a client certificate** (mutual TLS). With no client
  cert it completes TLS then immediately closes. `cmr-dial` presents a
  self-signed Ed25519 identity (+ a CertificateVerify, Ed25519 signing the
  transcript per RFC 8422) and the controller then speaks.
- Server leaf cert captured: `CN=figman-dfbe`, Ed25519 SPKI pubkey
  `dfbeec82…3b4b58cf`, self-signed CA, validity 1970..9999, and the empty
  `signatureAlgorithm` OID (`3004 0600 0500`). Saved to
  `docs/samples/controller-cert.der`.
- The control-channel bring-up is a symmetric 4-exchange (C = client/enrollee,
  S = controller):

  | Step | Dir | Frame | Meaning |
  |------|-----|-------|---------|
  | 1 | C→S | `{id2: d2 11 01 d8 ‖ nonce_C(8)}` | client hello |
  | 2 | S→C | `{id2: d2 11 01 d8 ‖ nonce_S(8)}` | controller hello |
  | 3 | S→C | `{id1:{id22 nonce, id25 build+ver, id24 "CHR"}}` | controller version/identity |
  | 4 | C→S | `{id1:{id22 nonce, id25 build+ver, id24 "CHR"}}` | client version/identity |
  | 5 | S→C | `{id6:{id22 nonce, id21 seq}}` | controller |
  | 6 | C→S | `{id6:{id22 nonce, id21 seq}}` | client |
  | 7 | S→C | `{id1:{id22 nonce, id21 seq}}` | controller |
  | 8 | C→S | `{id1:{id22 nonce, id21 seq}}` | client |

  The hello token is a constant 4-byte magic `d2 11 01 d8` + an 8-byte nonce.
  After step 8 the controller holds the connection open (no EOF): the bring-up
  completed. The replayed (stale) client nonces were accepted, so the id22
  nonces are liveness tokens, not a validated challenge/response.
- **Managed-device promotion still needs pairing.** The controller's
  `/cmr/device/print` lists already-paired peers as CONNECTED with
  `ids=<SHA-256(peer identity)>/<controller Ed25519 pubkey>`; a fresh unpaired
  identity completes steps 1–8 but is **not** added to the device table. With
  `/cmr/set pairing-requirement=password`, a new device must additionally pass
  the password **key handshake** (the PKEX/`handleMsg1..4` path). That is the
  only crypto-gated step and is required only to enroll a *new* device; a
  controller with open pairing, or an already-paired identity, does not use it.

So a byte-parity **controller** needs: mutual TLS with an Ed25519 identity,
the 4-exchange bring-up above (bytes now known for both roles), optional
password pairing for new devices, and then the management plane.

## Password pairing captured in plaintext (2026-10-16)

`/cmr/client/reset` unpairs/stops the client (it is the unpair command — not
`reset-pairing`/`unpair`, which do not exist). With the client reset, a
`cmd/cmr-mitm` transparent proxy (TLS-terminate the client with its own
identity, relay to the real controller via `pkg/mtls` with its own client
identity, preserving frame boundaries) captured a full `/cmr/client/pair
password=… username=admin` in plaintext. **The pairing completed through the
MITM**, which proves the PAKE binds the password + exchanged elements + the
scope string `com.mikrotik.cmr`, NOT the TLS certificates — so it is
reproducible by a third party that knows the password.

Sequence after the 4-exchange bring-up (C = client/enrollee, S = controller):

| Dir | m3 | Meaning |
|-----|----|---------|
| C→S | `id6{ id4[scope, "password required"], nonce, seq }` | client states scope + pairing mode |
| S→C | `id10{ id23 "com.mikrotik.cmr", id27=02, id28 "password required", nonce }` | controller pairing offer |
| C→S | `id0{ device descriptor: id65 build+ver, id66 "x86", id69 "routeros", id67/76 "CHR", id80 112-byte blob, … }` | client device info |
| C→S | `id11{ id23 scope, id32 = "admin\0" + 33-byte element, seq }` | **pairing request: username + PAKE commit** |
| C→S | `id7{ id4[scope, 65 02 2d 00], nonce, seq }` | |
| S→C | `id0{ id32 = 49-byte element, seq }` | **server PAKE response** |
| C→S | `id4{ id23 scope, id33 = 32-byte value, seq }` | **client confirm** |
| S→C | `id0{ id33 = 32-byte value, seq }` | **server confirm** |
| S→C | `id10{ id23 scope, id27=00, nonce }` | **pairing success** (id27: 02=required → 00=ok) |
| S→C | 1474-byte `id0{…}` and onward | **management plane**: plaintext m3 config push |

Full decoded capture: `docs/samples/pairing-handshake.txt`.

**Post-pairing management is plaintext m3 over the TLS channel** (no additional
encryption layer) — the PAKE only authenticates the device and pins its cert.

### The PAKE — algorithm stack and crypto map (reverse-engineered)

The password handshake is an **EKE-style PKEX**: each side encrypts its
ephemeral P-256 public point under a key derived from the password, exchanges
it, both run ECDH to a shared secret, and confirm with HMAC. The primitives
(all in `libucrypto.so`, which `cmr` links) are:

- **P-256** — `getECParamsForP256`, `WCurve::{mul,add,isOnCurve,toEC,fromEC}`,
  `ECPoint::{toBin(len,compressed),fromOctetString}`, `BigNum`, `randomBigNum`.
  The ephemeral point serializes via `ECPoint::toBin(0x20, compressed)` → the
  33-byte compressed point seen on the wire.
- **SHA-256** (`SHA256`, `createHash`), **HMAC-SHA256** (`HMAC`).
- **HKPDF** — an HKDF-style KDF, `HKPDF(Hash&, salt, ikm, info, len)`.
- **AES-SIV** (RFC 5297), `AESSIV::{encrypt,decrypt}(plaintext, vector<AD>, out)`
  — deterministic AEAD; its 16-byte SIV is why the server response is
  `49 = 16 + 33` (SIV ‖ encrypted 33-byte point).

`PKEXImpl` methods: `encrypt`/`decrypt` (AES-SIV wrap of a point),
`getPoint` (parse+validate), `reveal1`/`reveal2` (send/consume the encrypted
point), `authDeriveZ(BigNum,BigNum)` (shared-secret → Z), `auth1`/`auth2`
(HMAC confirm compute/verify). Errors gate as `auth failed (1..4)`.

**Crypto call sites in `cmr` (for continued RE / byte-exact extraction):**

| What | Address(es) | Notes |
|------|-------------|-------|
| password → key (`HKPDF`) | `0x8055a9d`, `0x8055c73` | identity/password key derivation |
| AES-SIV encrypt | `0x80867c7` | wrap ephemeral point |
| AES-SIV decrypt | `0x80868f1` | unwrap peer point |
| derive (`createHash`+`HKPDF`) | `0x8089948`, `0x808997a` | `authDeriveZ` / Z |
| HMAC confirm | `0x8089bc6`, `0x8089ee3` | `auth1`/`auth2` (32-byte tags) |
| ephemeral keygen | `0x808a1f5`+`0x808a3d4` | `getECParamsForP256`+`randomBigNum` |
| server frame handler | `FUN_0808cf6a` | `PKEXSrvImpl::handleFrame` (Msg2/Msg4) |
| client frame handler | `FUN_0808dcb4` | `PKEXClntImpl::handleFrame` (Msg1/Msg3) |

Snapshot for REA: `/home/josh/cmr/extract/cmr.snapshot`. The wire bytes to
match are in `docs/samples/pairing-handshake.txt`; `cmd/cmr-mitm` re-captures a
live pairing on demand.

**Remaining to implement a controller that enrols a new device:** extract the
exact `HKPDF` inputs (salt/ikm/info — password, username `admin`, scope
`com.mikrotik.cmr`, role labels), the AES-SIV associated-data vector, and the
HMAC confirm transcript from the functions above, then implement the server
role in Go (`crypto/ecdh` P-256 + `crypto/hmac` + `golang.org/x/crypto/hkdf` +
an AES-SIV lib) and validate against the lab cli (`/cmr/client/reset`, point at
our controller, confirm it reaches the controller's device table).
