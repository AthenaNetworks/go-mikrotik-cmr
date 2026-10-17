# go-mikrotik-cmr

A clean-room, pure-Go implementation of the controller side of MikroTik's **CMR**
protocol — the centralized Wi-Fi/mesh controller that ships in RouterOS 7.26+.
A stock, unmodified MikroTik device pairs and connects to this controller with
no RouterOS or Nova controller in the path.

The protocol was reverse-engineered from the RouterOS 7.26 binaries (`figman`,
`ssld`, `libumsg.so`) and validated live against a RouterOS 7.26 CHR acting as
the device. There is no affiliation with MikroTik; this is an interoperability
and research project.

## Status

| Capability | State |
|---|---|
| Device pairs + connects (open pairing, `pairing-requirement=none`) | **Working, verified live** — device reports `status: paired,connected` |
| TLS 1.2 control transport (ECDHE-ECDSA-AES128-GCM, Ed25519 identities) | Working |
| `nv`/`m3` wire codec | Byte-exact vs. captured frames (unit tests) |
| Control-plane reads decoded (device identity, board, version, interfaces, MACs) | Decoder verified against live captures |
| Full device inventory decoded from a real sync | Working — `capdump -objects` decodes the whole `id1`/`id9` object dump |
| Password pairing (the PKEX PAKE) | Not implemented — runs in `ssld`, inlined/Montgomery-form; open pairing avoids it |
| Object-sync dump driven live by our controller | Blocked — the device derives object handles per controller identity; see [Protocol notes](#protocol-notes) |

The controller currently implements the **open pairing** path. Password (PKEX)
pairing is a custom, TLS-channel-bound EC+HMAC PAKE inside `ssld` and is out of
scope; open pairing skips it entirely (the controller sends `id10` SUCCESS with
status 0, which promotes the device to `paired`).

## How it works

```
device  ──TLS 1.2 (cipher 0xc02b, self-signed Ed25519)──▶  cmr-ctrl :54321
        ◀── nv/m3 control stream (continuous per-direction field stream) ──▶
```

1. **Transport.** TLS 1.2 on tcp/54321. MikroTik's client certificate encodes an
   empty `signatureAlgorithm` OID that `crypto/x509` rejects and that cannot be
   rewritten on the wire without breaking the Finished MAC, so the server does
   not request a client cert at the TLS layer; trust is the app-layer pinned
   fingerprint.
2. **Framing.** Each direction is a single continuous `m3` field stream — the
   field-tag delta base carries across the whole direction and TLS writes chunk
   it arbitrarily (`pkg/m3.StreamDecoder`). This is the key to decoding anything
   past the bring-up.
3. **Bring-up** (strictly alternating): `hello → version → discovery → verAck +
   id10 SUCCESS`. The device is then paired and holds the connection open.
4. **Management plane.** The `nv` object-sync protocol (`libumsg.so`): scoped,
   handle-addressed objects (`pkg/cmr.ParseObject` decodes them into identity,
   board, version, interfaces and MACs).

See [docs/OPEN_PAIRING.md](docs/OPEN_PAIRING.md) and
[docs/PROTOCOL.md](docs/PROTOCOL.md) for frame-level detail.

## Layout

```
cmd/cmr-ctrl/    the controller a real device pairs+connects to (primary)
pkg/m3/          byte-exact nv/m3 codec + continuous StreamDecoder
pkg/cmr/         control-plane object model (descriptor + object decoding)
pkg/mtls/        minimal cert-tolerant TLS 1.2 client

cmd/cmr-mitm/    TLS-terminating control-channel proxy (capture real traffic)
cmd/capdump/     decode an nv/m3 capture (-timeline, -stream C|S)
cmd/m3dump/      recursively decode hex m3 frames
cmd/cmr-dial/    connect to a real controller as a client and decode replies
cmd/cmr-learn/   earlier controller bring-up probe

docs/            protocol notes
```

`pkg/pkex`, `pkg/pairing`, `pkg/model`, `pkg/transport`, `pkg/server` and
`cmd/cmr-server` are earlier exploratory code built against a UDP/PKEX model
that predates the live TLS findings; they do not drive the real protocol and
are retained only for reference.

## Build

```sh
go build ./...
go test ./...
```

## Run

```sh
# stable identity so the device-pinned fingerprint survives restarts
go run ./cmd/cmr-ctrl -listen 0.0.0.0:54321 -keyfile ctrl.ed25519.seed
```

On the device:

```
/cmr/client/set enabled=yes controller-addresses=<controller-ip>
/cmr/client/pair password=x username=admin
/cmr/client/print          ;# -> status: paired,connected
```

The controller is open-pairing, so the password is ignored; the command still
requires one. `-v` logs a full recursive decode of every frame; `-idle 0`
(default) holds idle connections open.

## Protocol notes

Pairing is complete and the object decoder is complete: a captured sync decodes
into the full device inventory (interfaces, identity, vendor, system-id,
version, board, MACs) via `pkg/cmr.ParseObject` / `capdump -timeline -objects`,
so passive MITM (`cmd/cmr-mitm`) is a working read path today.

After SUCCESS the plane is a symmetric `nv` getall over two scopes:
`com.mikrotik.cmr` (the device pulls the controller's config as `id9` rows) and
`com.mikrotik.cmr.config` (the controller pulls the device's state as `id1`
rows). See [docs/PROTOCOL.md](docs/PROTOCOL.md#management-plane--nv-object-sync-2026-10-17).

The remaining work is driving that dump from our own controller. The device
**owns** the object-handle space and derives handles **per controller
identity**, so a fresh controller cannot announce objects with self-chosen
handles — it gets `not found` echoing the exact handle it sent, and the device
will not stream `.config` until the base scope is established with valid objects
at device-derived handles. Verified live three ways (verbatim replay, bare
getall, `id11` subscribe). Unblocking it means reversing the handle allocation
in `libumsg.so` (`nv::Handler::cmdGetAll` / `subscribe` / `sendAsyncGetall` /
`AsyncReget`) and the figman local-store lookup.
