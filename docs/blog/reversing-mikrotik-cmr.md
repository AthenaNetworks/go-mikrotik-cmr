# Talking to a MikroTik controller that doesn't exist yet

RouterOS 7.26 shipped a feature called CMR — a centralized controller for
Wi-Fi and mesh devices. You point a device at a controller's address, pair it,
and from then on the controller owns its configuration and sees its state. The
catch: the controller is also RouterOS. If you want to manage a fleet of
MikroTik devices from your own software, there's no documented protocol to do
it. The wire format is undocumented, the pairing handshake is a custom PAKE
buried in a stripped binary, and the whole thing runs over TLS so you can't
just watch it go by.

We're building the controller side in Go, clean-room, so a stock unmodified
MikroTik device pairs and connects to it with no RouterOS controller anywhere
in the path. This is an interoperability project — no affiliation with
MikroTik, no modified firmware on the device. Everything below was recovered
from the shipped 7.26 binaries and validated against a real device in a lab.

## Why bother

If you run MikroTik hardware at any scale, "the controller must also be
MikroTik" is a constraint, not a feature. A controller you own is a controller
you can integrate: push config from your own source of truth, read device state
into your own monitoring, and stop treating the fleet as a pile of individual
boxes you SSH into. That's the outcome we're after — a controller that pairs,
reads, and eventually writes, built on a protocol we actually understand rather
than one we hope stays stable.

## The process

Reverse engineering a live protocol is mostly a loop: capture, decode, form a
theory, test it against the real thing, repeat. A few of the turns were worth
the trip.

**Get the bytes.** The control channel is TLS 1.2 on tcp/54321. The device
presents a client certificate with an empty signature-algorithm OID that every
sane TLS stack rejects — and you can't rewrite it on the wire without breaking
the handshake MAC. So the controller simply doesn't ask for a client cert at
the TLS layer; trust is a pinned fingerprint at the application layer instead.
With that, a terminating proxy gives us the plaintext.

**Decode the wire format.** Underneath TLS is MikroTik's `m3`/`nv` codec:
every message is a sequence of fields, each a tag chunk and a value chunk, with
the tags delta-coded against the previous one. We wrote a byte-exact
encoder/decoder and unit-tested it against captured frames until it round-trips
cleanly.

**The framing breakthrough.** The delta coding was the thing that cost the most
and taught the most. Each *direction* of the connection is one continuous field
stream — the delta base carries across the entire direction, and the TLS layer
chops it into records at arbitrary points. Decode each TLS read as its own
message and the whole burst mis-frames: the enumeration opcode decodes as
`0xFFFFFFFE` instead of `id7` and nothing downstream makes sense. Decode with a
running base that survives record boundaries and both directions parse as
thousands of fields with zero errors. This one fact is the key to everything
past the handshake.

**Pairing.** There are two paths. Password pairing is a custom
elliptic-curve-plus-HMAC PAKE that lives inside a different binary; open pairing
(the controller set to require nothing) skips it entirely — the controller
sends a success status and the device promotes itself to *paired*. We did open
pairing first, and a real device now reports `paired, connected` against our Go
controller, verified repeatedly.

**Reading state.** After pairing, the channel runs an object-sync protocol.
Both sides exchange structured objects: the device reports its interfaces,
identity, board, version and MAC addresses as a graph of typed records. Our
decoder turns a captured sync into a full device inventory — `ether1`, the
board string `CHR QEMU Standard PC (i440FX + PIIX, 1996)`, the firmware version,
the MACs — all pulled straight off the wire.

## Where it stands

- Pairing works live, end to end.
- The wire codec is byte-exact and tested.
- A captured sync decodes into the complete device inventory.
- **Our own controller now drives a live device to sync and dump its full
  state** — no RouterOS controller in the path, no password handshake. A stock
  device pairs to our Go process, and we read `ether1` and its MAC, the identity,
  the board string, the firmware version and the whole config/state graph
  straight off a device we're driving ourselves.
- **And it now writes.** Our controller pushes a configuration into the device
  and the device applies it: we make a stock box build a managed VLAN-filtering
  bridge, with its uplink tagged into the VLAN, driven entirely by our Go process.

Getting there meant un-learning a wrong conclusion. For a while the device
answered "not found" to every object our controller announced, and the evidence
pointed at the password handshake — every working sync we'd captured happened to
be password-paired, so we assumed authentication was the gate. It wasn't. We put
a real RouterOS controller into open-pairing mode and watched it drive a full
sync with no password at all. That one experiment moved the whole problem: the
blocker was never auth, it was three details of how a controller must speak.

The device's own query carries a one-byte flag that changes every session, and
every object the controller pushes has to echo it back — stamp the wrong flag and
the device simply doesn't recognise the objects as answers to its request. The
success message and the first objects have to travel as one continuous delta-coded
stream, not separate frames. And the device only dumps its own state once the
controller answers the device's "get everything" request by streaming its
configuration first. Miss any one of these and you get a polite "not found";
get all three right and the device hands you everything.

The write took one more turn of the same crank: capture a real controller doing
it, then reproduce it. We defined a VLAN on a RouterOS controller and watched it
push the config to the device over the management channel — and watched the
device build a managed bridge and tag its uplink into the VLAN. The push turned
out to be a single message into the device's own config scope, carrying the
desired object graph, with one easy-to-miss detail: a couple of envelope fields
have to sit *inside* the message, not beside it, or the device reads instead of
writes. Get that right, send it once the device has settled, and the device
reconfigures itself to match. Our controller now does exactly that — no RouterOS
in the path.

## What "done" looks like

A controller process you can run anywhere that a fleet of stock MikroTik devices
will pair with, that reads their full state and writes their configuration, built
on a protocol documented down to the byte. Pairing, live reading, and live
writing are now in hand — a device we drive dumps its inventory and then builds
the bridge and VLAN we tell it to.

And the config we push is no longer a recording. We cracked the shape of the
object graph: a stream of length-tagged chunks, three objects — a global entry, a
bridge, a VLAN — each a class id followed by positional fields. Most of those
fields are still opaque, but two we pinned down by changing them and watching the
device obey: the VLAN id and the bridge name. So we now build the graph from a
plain `{VLANID, BridgeName}` and the device applies it — we asked for VLAN 150 on
a bridge named `athena-cmr`, from a struct, and that is exactly what it built,
with nothing replayed. What remains is breadth: map the rest of the fields, and
the other object types (Wi-Fi, device), the same way — change one thing, diff the
bytes, watch the device.
