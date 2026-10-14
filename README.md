# cmr-server

A standalone, embeddable **CMR controller** written in Go.

CMR ("CMR server") is the centralized Wi-Fi/mesh network controller that ships
as a package in MikroTik RouterOS 7.26+. It manages a fleet of devices
(topology, Wi-Fi radios/networks, VLANs, coordinated firmware upgrades) and
enrolls them with a **PKEX** handshake (the Public Key Exchange from Wi-Fi
DPP / Easy Connect) over **NIST P-256**.

This project is a **clean-room reimplementation of the controller side**,
derived from static analysis (Ghidra) of the stock `cmr` binary. It has no
RouterOS/Nova dependency: it runs as a single Go binary (or embeds as a
library) and exposes hooks so other projects can plug in their own storage,
policy, transport, and crypto.

> **Status — read this.** The protocol *structure* (frame types, state machine,
> enrollment flow, data model) is recovered from the binary and implemented
> here. The exact DPP-PKEX key schedule and the `m3` wire bytes are **not yet
> byte-verified**, because the EC math and message serialization live in
> RouterOS shared libraries, not in `cmr`. The bundled crypto suite and UDP
> transport therefore interoperate **with this project's own client**, not
> (yet) with real MikroTik devices. See [docs/PROTOCOL.md](docs/PROTOCOL.md)
> for exactly what is verified vs. pending. This is an interoperability /
> research project; it is not affiliated with or endorsed by MikroTik.

## Layout

```
cmd/cmr-server/   daemon: UDP transport + controller + optional HTTP status
pkg/pkex/         discovery + 4-message PKEX state machine (client & server)
pkg/m3/           typed-TLV message codec (structure of MikroTik m3::message)
pkg/model/        managed objects (dev, wifi, vlan, layout, upgrade...) + Store
pkg/transport/    Transport interface + reference UDP transport
pkg/server/       Server (controller) + Enroller (device side) + Hooks
examples/extend/  embedding example: custom hooks + store.Watch
docs/PROTOCOL.md  reverse-engineered protocol notes (provenance, open items)
```

## Quick start

```sh
go build ./...
go test ./...

# run a controller (reference UDP transport)
go run ./cmd/cmr-server -code my-shared-code -http 127.0.0.1:8080 -v

# in another terminal, watch enrolled devices
curl -s http://127.0.0.1:8080/devices | jq

# see an end-to-end enrollment in one process (controller + device + hooks)
go run ./examples/extend
```

## Embedding

The controller is a library first. Minimal embed:

```go
store := model.NewStore()
store.Watch(func(e model.Event) { /* feed your DB / UI / metrics */ })

tr, _ := transport.NewUDP(transport.SynthMAC(1), "0.0.0.0:14700", "")
srv := server.New(server.Config{
    Code:     []byte("shared-code"),
    Identity: "my-controller",
    Hooks: server.Hooks{
        Authorize: func(mac net.HardwareAddr) bool { return allowlist[mac.String()] },
        CodeFor:   func(mac net.HardwareAddr) []byte { return perDeviceCode(mac) },
        OnEnroll:  func(d *model.Device, r pkex.Result) { provision(d) },
    },
}, store, tr)
srv.Run(ctx)
```

### Extensibility seams

| Seam | Interface | Use it to |
|------|-----------|-----------|
| Storage / events | `model.Store.Watch` | mirror managed objects into your own DB/UI |
| Policy | `server.Hooks` | authorize peers, assign per-peer codes, react to enrollment |
| Transport | `transport.Transport` | carry frames over 802.11 action frames, raw L2, or UDP |
| Crypto | `pkex.Suite` | drop in the exact MikroTik key schedule once validated |

## Toward real-device interop

The seam that matters most is `transport.Transport`: real CAP↔controller
enrollment rides link-layer frames emitted by the wifi stack, not UDP.
Implement an 802.11 public-action-frame transport, capture a real enrollment,
then align `pkg/pkex` (key schedule) and `pkg/m3` (wire bytes) to the capture.
The frame-type map and state machine in `pkg/pkex` already match the binary.

## License

MIT — see [LICENSE](LICENSE).
