# Open pairing: a MikroTik CMR device joins our Go controller

`cmd/cmr-ctrl` is a pure-Go CMR **controller** that a real, unmodified MikroTik
CMR device pairs and connects to — with **no MikroTik controller involved**.
Verified live against a RouterOS 7.26beta1 CHR device (lab `cli`, 10.99.0.114):
`/cmr/client/print` reports `status: paired,connected`, `pairing-status: ok/ok`,
on a single stable TLS socket to our server.

## Why this works without PKEX

CMR has two pairing modes. `pairing-requirement=password` runs a custom,
TLS-channel-bound, Montgomery-form EC + HMAC PAKE inside the base-RouterOS `ssld`
daemon (see `docs/re/ssld/FINDINGS.md`) — impractical to reimplement byte-exact.
`pairing-requirement=none` ("open") skips the PAKE entirely: the controller just
sends an `id10` SUCCESS with status `0x00` and the device is paired. The device
still *offers* its password in discovery (`id4 … "password required"`); an open
controller ignores it. The device does **not** require the controller to prove
knowledge of the password, which is exactly why this path is tractable.

## The wire flow (post-TLS nv/m3 control plane, tcp/54321)

TLS 1.2, cipher `0xc02b`, self-signed Ed25519 identities, no client cert at the
TLS layer (MikroTik's empty-OID certs break `crypto/x509`; trust is the
app-layer pinned fingerprint). The bring-up is **strictly alternating** — the
device waits for each reply before sending the next frame:

| # | client → controller            | controller → client                              |
|---|--------------------------------|--------------------------------------------------|
| 1 | `id2` hello                    | `id2` hello (`d21101d8` + 8 random)              |
| 2 | `id1` version (has `id25`)     | `id1` version (`id22` nonce, `id25` build, `id24` board) |
| 3 | `id6` discovery (has `id4`)    | `id6` ack (`id22` nonce, `id21` seq)             |
| 4 | `id1` nonce/seq                | `id1` verAck **+** `id10` SUCCESS                |

- **verAck** = `id1{ id1{ id27:0x01, id21:innerSeq }, id22:nonce, id21:outerSeq }`
  where `outerSeq` echoes the client's discovery `id21` and
  `innerSeq = [outerSeq[0]-1, outerSeq[1:]]` (the controller's own seq, one below).
- **SUCCESS** = `id10{ id23:"com.mikrotik.cmr", id27:0x00, id22:nonce }`. The
  `0x00` status is the promotion to *paired* (vs `0x02` "password required" in
  PKEX mode).

Every controller frame above is reproduced byte-for-byte from a live capture by
`pkg/m3`; `cmd/cmr-ctrl/main_test.go` asserts the exact bytes.

### Two things that break it (learned the hard way)

1. **Do not send fabricated `id9` device-table rows after SUCCESS.** The live
   controller sends rows for devices that actually exist; rows referencing
   unknown devices make the client reply `id0{ id27:0x83 "not found" }` for each
   row and then drop the session into a reconnect loop. SUCCESS alone pairs it.
2. **Hold the socket open.** The device stays silent when idle and holds
   "connected" only while the TLS socket is up. Use `-idle 0` (default) so the
   server never times out a quiet connection; TCP keepalive detects dead peers.

A discovery/resync (`id6`/`id7`) carries `id22`/`id21` **inside** the nested
value, not at the top level — decode the field value to read the seq.

## Run it

```sh
go build -o cmr-ctrl ./cmd/cmr-ctrl
./cmr-ctrl -listen 10.99.0.1:54321 -keyfile ctrl.ed25519.seed
```

`-keyfile` persists the Ed25519 identity so the fingerprint the device pins
survives restarts (otherwise reconnects from an already-paired device fail).

On the device:

```
/cmr/client/set enabled=yes controller-addresses=<our-ip>
/cmr/client/pair password=<anything> username=admin
/cmr/client/print          ;# -> status: paired,connected
```

(The password is ignored by an open controller but the command requires one.)
An already-paired device auto-reconnects; `/cmr/client/reset` forces a fresh pair.
