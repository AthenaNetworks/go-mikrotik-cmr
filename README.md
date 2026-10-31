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
| Object-sync dump driven live by **our** controller | **Working, verified live** — a stock device syncs and dumps its full state to `cmd/cmr-ctrl`, 0 not-found, no PKEX; see [Protocol notes](#protocol-notes) |
| Config **write** driven live by **our** controller | **Working, verified live** — `cmd/cmr-ctrl -writefile` makes a stock device build a MANAGED `cmr-bridge` (vlan-filtering, `ether1` pvid=100), no RouterOS controller in the path; see [Protocol notes](#protocol-notes) |
| CMR vlan config **authored from structured input** (not replayed) | **Working, verified live** — `pkg/cmr.BuildVLANConfig` constructs the object graph; `cmd/cmr-ctrl -vlan N -bridgename X` pushes it and the device builds the bridge with the chosen vlan id + name (100/150/200/300 verified on the wire) |
| Config graph **codec** (author, decode, round-trip) + attribute dictionary | **Working** — `pkg/cmr.Graph`/`DecodeGraph`/`Encode` round-trip a graph byte-for-byte; `ParseVLANConfig` is the inverse of `BuildVLANConfig`; the 410-attribute dictionary (`cmd/cmr-schema`) is embedded and queryable via `Attrs`/`AttrByName`/`AttrByID`; `cfgdump -tree` decodes structurally |
| **Embeddable controller library** (`pkg/controller`) | **Working** — `controller.New(Config).ListenAndServe()` runs the whole pairing + object-sync + config-write state machine with `OnConnect`/`OnPair`/`OnObject`/`OnIdentify` callbacks and `Session.WriteConfig`; `cmd/cmr-ctrl` is a thin wrapper over it; bring-up covered by an in-process integration test |
| **Multi-device registry + persistence** | **Working** — a `Store` (JSON `FileStore` provided) makes the server identify each device by a stable key (MAC), track it (`Devices`/`Device`), persist enrolments + per-device desired config across restarts, and re-apply a device's config on reconnect; `SetDesiredConfig(id, graph)` drives one device; covered by an end-to-end reconcile test |
| **HTTP admin API** | **Working** — `srv.AdminHandler()` exposes the registry as a JSON API (health/identity, list/inspect/forget devices, set/clear per-device config, list live sessions); `cmd/cmr-ctrl -admin <addr>` serves it, optional bearer token (`-admintoken` / `-admintoken-file` / `CMR_ADMIN_TOKEN`); covered by handler tests |
| **Controller scaffold synthesised from structure** | **Working** — `pkg/cmr.EncodeRoots` / `EncodeConfigScope` rebuild the controller's own root + config-scope id9 objects from a decoded catalog, reproducing the captured blobs byte-for-byte; the server defaults to these, so it drives a device's full state dump with no captured hex. Per-object grammar reversed + named (`DecodeConfigScope` / `cfgdump -scope`, lossless round-trip); leaf nv ids stay numeric (runtime-assigned, no static names); round-trip + on-wire tests |

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
pkg/controller/  embeddable CMR controller: a stock device pairs+connects to it (the server library)
cmd/cmr-ctrl/    thin CLI over pkg/controller (the self-hosted daemon)
examples/embed/  minimal program embedding the controller in another project
pkg/m3/          byte-exact nv/m3 codec + continuous StreamDecoder
pkg/cmr/         control-plane object model, config graph codec (Graph/BuildVLANConfig/DecodeGraph) + embedded attribute dictionary
pkg/mtls/        minimal cert-tolerant TLS 1.2 client

cmd/cmr-mitm/    TLS-terminating control-channel proxy (capture real traffic)
cmd/capdump/     decode an nv/m3 capture (-timeline, -stream C|S)
cmd/m3dump/      recursively decode hex m3 frames
cmd/cfgdump/     decode a config object graph (id30 write payload): chunk list, or -tree for the structured object/field view
cmd/cmr-schema/  extract the config dictionary (nv id -> name -> type) from a RouterOS .mem
cmd/cmr-dial/    connect to a real controller as a client and decode replies
cmd/cmr-learn/   earlier controller bring-up probe

docs/            protocol notes
```

`pkg/pkex` and `pkg/pairing` are not used by the live open-pairing flow (which
is TLS + an id10 `status=0x00` SUCCESS, no password). They are the
reverse-engineered PKEX/DPP password handshake — `pkg/pairing` the P-256 + SIV
crypto, `pkg/pkex` the session/frame layer from the earlier UDP enrollment
model — retained as the head start for MikroTik's password-protected pairing
(the `status=0x02` "password required" path the open flow never enters).

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
requires one. `-v` raises logging to Debug (every frame, recursively decoded);
`-idle 0` (default) holds idle connections open.

## Embedding

`cmd/cmr-ctrl` is a thin CLI over `pkg/controller`; any Go program can run the
same controller in-process. The whole integration surface is a `Config` and a
few callbacks:

```go
cert, _ := controller.Identity("ctrl.ed25519.seed") // persistent pinned identity
srv, _ := controller.New(controller.Config{
    Certificate: cert,
    Logger:      slog.Default(),
    AutoWrite:   cmr.BuildVLANConfig(cmr.VLANConfig{VLANID: 100}), // pushed to each device once it pairs
    OnObject: func(s *controller.Session, o controller.Object) {   // device inventory as it arrives
        log.Printf("%s: %s", s.Peer(), o.Key)
    },
})
srv.ListenAndServe("0.0.0.0:54321")
```

`Session.WriteConfig(graph)` pushes config to a device at any time; `srv.Sessions()`
lists the connected devices. Add a `Store` (a JSON `FileStore` is provided) to make
the server multi-device: it identifies each device by a stable key, tracks it
(`srv.Devices()` / `srv.Device(id)`), persists enrolments and per-device desired
config across restarts, and re-applies a device's config whenever it reconnects;
`srv.SetDesiredConfig(id, graph)` drives one device by id. See
[examples/embed](examples/embed/main.go).

`srv.AdminHandler()` returns an `http.Handler` that exposes the registry as a JSON
API, so the same control is reachable over HTTP — mount it on your own mux (behind
your own auth) or let the daemon serve it with `-admin`:

```
cmr-ctrl -registry devices.json -admin 127.0.0.1:8080 -admintoken s3cret
#   token also via -admintoken-file <path> or the CMR_ADMIN_TOKEN env var

GET    /healthz                 liveness + device count + version
GET    /identity                controller fingerprint, listen address and version
GET    /devices                 known devices (incl. last_remote_addr, connected state)
GET    /devices/{id}            one device
PUT    /devices/{id}/config     set config — {"vlan":100,"bridge":"cmr-bridge"} or {"graph":"<hex>"}
DELETE /devices/{id}/config     clear config
DELETE /devices/{id}            forget a device (remove its registry record)
GET    /sessions                live device sessions
```

The registry (`-registry`) is a JSON file whose format is additive across
versions — a newer `cmr-ctrl` reads an older file (new fields default empty) and
an older one ignores fields it does not know — so an upgrade is just a binary
swap. `last_remote_addr` is the device's source IP (a NAT address for a device
behind NAT, so treat it as a hint for the poll target, not authoritative).

Setting a device's config pre-enrols it: the record is created now and applied when
that device next connects. Wrap the handler with `controller.BearerAuth(token, h)`
for a shared-token guard.

The controller's own scaffold objects — the root rows that ride the SUCCESS frame
and the config-scope burst that answers the device's getall, both needed to drive
the device's full state dump — are **synthesised from structure** by default
(`pkg/cmr.EncodeRoots` / `EncodeConfigScope`), so the server owes nothing to a
captured RouterOS controller at runtime. They are rebuilt from a decoded object
catalog (`pkg/cmr/scaffold_catalog.go`, generated by `cmd/cmr-scaffold` from the
capture) and reproduce the captured blobs byte-for-byte. Set `Config.Roots` /
`Config.ConfigBurst` to replay a specific capture instead. The per-object
**grammar** is reversed and named (`pkg/cmr.DecodeConfigScope` / `cfgdump -scope`):
each row decodes into a named key (class + attribute id + descriptor form + query
marker) and a named value type-signature, and a round-trip test proves the naming
is lossless. What stays numeric is the leaf nv id space (attribute ids, column-id
references, type handles) — runtime-assigned and carried in no static table or
capture, so mapping it to human names needs a live name-carrying getall (see
`docs/re/CMR_CONFIG_SCHEMA.md`).

## Protocol notes

Pairing is complete and the object decoder is complete: a captured sync decodes
into the full device inventory (interfaces, identity, vendor, system-id,
version, board, MACs) via `pkg/cmr.ParseObject` / `capdump -timeline -objects`,
so passive MITM (`cmd/cmr-mitm`) is a working read path today.

After SUCCESS the plane is a symmetric `nv` getall over two scopes:
`com.mikrotik.cmr` (the device pulls the controller's config as `id9` rows) and
`com.mikrotik.cmr.config` (the controller pulls the device's state as `id1`
rows). See [docs/PROTOCOL.md](docs/PROTOCOL.md#management-plane--nv-object-sync-2026-10-17).

`cmd/cmr-ctrl` now **drives that dump from our own controller** — a stock device
syncs and dumps its full state to us, 0 not-found, no PKEX and no RouterOS
controller in the path. The earlier "the device owns the handle space / derives
handles per identity" reading was wrong; the real requirements, reversed from a
RouterOS controller running in open-pairing mode, are three:

1. **Echo the device's query flag.** The device's `id6` discovery / `id7` getall
   carries an `id29` flag (different each session) inside its nested `id4`. Every
   object the controller pushes must carry that same flag, or the device answers
   `not found`. `cmr-ctrl` captures it (`conn.devFlag`) and stamps it in.
2. **SUCCESS + roots in one stream.** `id10` SUCCESS and the root `id9` objects
   must be one continuous delta-chained `m3` message, not separate frames.
3. **Answer the getall.** Reply to the device's `id7` by streaming the
   controller's config `id9` burst — that triggers the device's own state dump.

Handle index/epoch are controller-chosen and the device adopts them. Run it with
`-pushfile docs/samples/controller-roots.hex -resyncfile docs/samples/controller-config.hex`
(a captured valid controller-config scaffold).

`cmd/cmr-ctrl` also **drives the write/management direction**: with `-writefile`
it pushes a config object graph into the device's own config scope and the device
applies it as MANAGED config — verified live by making a stock device build a
`cmr-bridge` with VLAN filtering and `ether1` at pvid 100, exactly as a real
controller does from `/cmr/vlan/add`. The write is an `id7` into
`com.mikrotik.cmr.config` whose `id4` carries the desired object graph in `id30`
(with the nonce/reqid nested *inside* the `id7`), pushed once the device reaches
steady state (`-writedelay`). See
[docs/re/figman/NV_SYNC_FINDINGS.md](docs/re/figman/NV_SYNC_FINDINGS.md). Run with
`-writefile docs/samples/write-config-vlan100.hex`.

That graph is **no longer a replay**: `pkg/cmr.BuildVLANConfig` authors it from a
`VLANConfig{VLANID, BridgeName}` on top of a declarative graph model. The graph is
a nested `m3` chunk stream (02 01 00 header, class-wrapped objects
`4a00c03f`/`c901c03f`, `02 00 00 01 <idx>` delimiters, a trailer) exposed as
`Graph`/`ConfigObject`; `Encode` and `DecodeGraph` round-trip it byte-for-byte, so
a captured graph can be parsed, edited field by field and re-emitted
(`ParseVLANConfig` is `BuildVLANConfig`'s inverse; `cfgdump -tree` shows the
structured view). `cmd/cmr-ctrl -vlan 150 -bridgename athena-cmr` builds and writes
it in-process, and a stock device builds the managed bridge accordingly (vlan ids
100/150/200/300 verified live).

The object body is **positional**, not nv-id-tagged: the confirmed slots (vlan id,
bridge name) sit at fixed field positions under the generic wrapper, and the
remaining schema positions are carried verbatim. The 410-attribute dictionary that
`cmd/cmr-schema` extracts from the RouterOS console schema (`.mem`) describes the
separate **control-plane** object model — it is embedded in `pkg/cmr`
(`Attrs`/`AttrByName`/`AttrByID`) for decoding and validation, not as the push
graph's field ids. See
[docs/re/CMR_CONFIG_SCHEMA.md](docs/re/CMR_CONFIG_SCHEMA.md). Adding wifi/device/
multi-vlan types means capturing a reference graph on hardware that can manifest it
and declaring its `Graph` layout — the vlan type is the worked example.
