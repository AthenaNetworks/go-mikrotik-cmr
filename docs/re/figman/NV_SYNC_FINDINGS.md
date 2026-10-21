# nv object-sync engine — RE findings (handle allocation / enumeration wall)

Target binaries (extracted, read-only): `nova/bin/figman` (ELF32 i386, stripped,
ET_EXEC @ 0x8048000) and `lib/libumsg.so` (ELF32 i386, **rich C++ symbols**).
figman drives the sync; libumsg implements it. All addresses below are libumsg
file offsets unless prefixed `fig:`.

## The model: a local↔remote object-id pairing store

figman keeps two on-disk stores and a per-peer id pairing between them:

- `/nova/store/figman/local`  — the device's own objects, by **local-id**.
- `/nova/store/figman/remote` — a peer (controller) view, by **remote-id**.

Strings in figman: `local id same as remote`, `invalid local-id`,
`id and local-id pair must be unique`, ` not found`, `figman local`,
`figman remote`; functions `Peer::doRemotes`, `Peer::remoteUpd(Remote*)`,
`Peer::handleAdd(const m3::msg_view&, Local**)`, `Local::peerHandleAdd(...)`.

An incoming object reference carries a **remote-id**; figman resolves it to a
local object through the pairing. No pairing ⇒ `Store::get(id)` miss ⇒ the
`" not found"` / `id and local-id` error. **The pairing is per controller
peer**, which is exactly the observed "handle mapping is per-controller-identity"
wall: a fresh controller has no pairings, so every remote-id it names resolves to
nothing.

`nv::Store::get(unsigned int)` @ 0x3fb74: `Allocator::isAllocated(id)` →
`StoreCord::getEntry(a,b)`; on miss it streams `"<name> not found id=<id>"` and
returns an empty message. `Store::set(id,msg)` @ 0x3fe5a, `Store::remove(id)`
@ 0x3fdf4. The store is keyed by a **u32 id**; the 8-byte wire handle
`IDX EP 00..00 TT` is (topic/type in the high bits) + (u32 id in the low), e.g.
`3a 01 00 00 00 00 00 01` → id `0x13a`, topic byte `0x01`.

## The command set (the decisive table)

Commands are `0xfe0000 + index`; `nv::Handler::handleCmd(msg,cmd)` @ 0x3759c does
`cmd - 0xfe0000`, range-checks `<= 0x1a`, and dispatches through a 27-entry PIC
jump table. The command **name** table (libumsg @ 0x6d054) gives, in index order:

| idx | cmd | | idx | cmd |
|----|------|--|----|------|
| 0 | NOOP | | 0x0d | SET |
| 1 | ERROR_UNKNOWN | | 0x0e | START |
| 2 | GETOBJ | | 0x0f | POLL |
| 3 | **SETOBJ** (update) | | 0x10 | CANCEL |
| 4 | **GETALL** | | 0x11 | SUBSCRIBE |
| 5 | STOP_GETALL | | 0x12 | UNSUBSCRIBE |
| 6 | **ADDOBJ** (create) | | 0x13 | DISCONNECTED |
| 7 | REMOVEOBJ | | 0x14 | NOTIFYACK |
| 8 | MOVEOBJ | | 0x15 | GETCOUNT |
| 9 | SETFORM | | 0x16 | RESET |
| 0x0a | NOTIFY | | 0x17 | (ERROR_OK) |
| 0x0b | SHUTDOWN | | 0x18 | GETPOLICIES |
| 0x0c | GET | | | |

The central result: **ADDOBJ (create) and SETOBJ (update) are distinct
commands.** ADDOBJ is what establishes a local↔remote pairing and lets the device
allocate a local-id; SETOBJ/GETOBJ operate on an id that must already be paired,
and miss with "not found" otherwise. Our controller's object pushes so far behave
as SETOBJ-style references to unpaired ids — hence the rejections.

## Message fields and the command carrier

nv field ids are `0xff0000 | <small wire id>`:

- `0xff0003` (wire id3) = **type/command** (value `0xfe00XX`). `nv::setType` @
  0x341e9 writes it; `nv::getType` @ 0x35313 = `msg.get<u32>(0xff0003, default 1)`
  (default = ERROR_UNKNOWN).
- `0xff0004` (id4) = message kind; `nv::isError` @ 0x3532d checks `id4 == 2` and
  reads the error code from `0xff0008` (id8), default `0xfe0001`.
- `0xff0007` (id7) = IDTopic (scope/topic), `0xff0001` (id1) = IDSel (selector),
  `0xff0005` (id5) = a bool. (`nv::sendCmd` @ 0x34280 assembles these.)

`Store::get`-miss errors come back as the `id1{ id1{ id27=0x83, id28="not found" },
id21=<handle> }` frames we see live, echoing the unresolved handle.

## Two layers, and where the command lives

There are two protocols here, which earlier notes conflated:

1. **figman peer wire** (controller↔device, over TLS): messages are m3 frames
   whose outer opcode is the small field id — id2 hello, id6 discovery, id7
   getall, id9 object, id10 status, id11 subscribe, id1 reply. The getall is
   `id7{ id4{ id21 reqid, id23 scope, id29 flag, id27 }, id22 nonce, id21 reqid }`
   — our `cmr-ctrl` getAll already matches this shape (the wrapper is not the bug).
2. **nv store commands** (figman↔nova store, internal IPC): the `0xfe00XX`
   command set. `nv::isCtrl(msg, u32* out)` @ 0x3543a is the extractor — a message
   is a control/command iff it carries field **`0xff000d` (wire id13)**, whose
   value is the `0xfe00XX` command. (`m3::MidType` is just the 2-bit m3 tag type
   via `m3::message::fold`, not the command.)

So on the peer wire the opcode is the outer field id; figman then decides, per
incoming object, whether to ADD (create pairing, allocate local-id) or SET
(update existing pairing) in its nv store. Our object pushes resolve to SET of an
unpaired remote-id → `Store::get` miss → "not found".

## Open item: the cold-start ADD / full-dump trigger

mitmF is a **warm resume** (srv's pairings persisted on the device — config reset
does not clear `/nova/store/figman`), so srv used SET/cursor-based reget, never a
cold-start ADD. We have no capture of a true first sync, so the exact wire form
that makes the device (a) accept the controller's first objects as ADDs and (b)
stream its own `.config` objects is still unconfirmed. Our getall(`.config`) is
structurally valid but draws only an ack (no cursor ⇒ nothing to resume; or the
device awaits a completed mutual sync of `com.mikrotik.cmr` first).

Live-confirmed: our getall for `com.mikrotik.cmr.config` draws only a bland ack
with flag `0x33` AND `0x07` (the device's own getall flag) — the flag is not the
trigger. The device will not serve `.config` until the `com.mikrotik.cmr` sync is
completed, which needs a valid ADD from us.

## figman function map (ET_EXEC @ 0x8048000)

- `0x806af04` — emits `"local not found, id="`. The surrounding function resolves
  an incoming remote-id to a local-id (a lookup returning `-1` at `0x806ae90` →
  error). This is the resolve/update path; it does NOT create. Our object pushes
  land here and miss.
- `0x806def8` — emits `"peer not found, id="` (sibling error).
- `~0x8057c71` — `Peer::handleAdd`: the CREATE path. Validates local-id
  (`"invalid local-id"` @0x8057d7b-region, `"id and local-id pair must be unique"`
  @0x80581bc), reads a bool flag field `0xfe000a`, allocates and inserts the
  local object + pairing. This is where a first-time object must be routed.

So an incoming object is routed to resolve-or-create by its wire structure/flags;
ours is treated as resolve (→ miss). The remaining RE step is the exact field
layout that routes to `handleAdd` and the local-id it allocates. The field ids in
`handleAdd` are register-loaded (figman internal calls are mis-symbolized by
objdump), so pinning them is instruction-level work.

## Two ways to finish

1. **Finish the figman RE**: read `Peer::handleAdd` / the object-routing dispatch
   instruction-by-instruction to get the ADD field layout and local-id
   allocation, then synthesize a valid ADD in `cmr-ctrl`.
2. **Cold-start capture**: reconstruct a reference controller (srv as an
   open-pairing controller, or wipe cli's `/nova/store/figman` for a stranger
   device) and MITM a true first sync to read the ADD off the wire.

Either way, driving the full mutual sync to the point the device dumps is a
multi-cycle implementation, not a one-shot.

## Cold-start capture — the epoch mechanism (2026-10-17)

Captured a clean from-scratch sync by MITMing cli↔srv (`cmr-mitm`, fresh proxy
identity both legs; cli `forget`+repair). srv requires PKEX ("password required"),
`password=x` is accepted. Capture: `scratchpad/coldstart.log`. **Zero "not
found"** — srv's `id9` push is fully accepted and the device dumps its `id1`
state. Decoded with `capdump -timeline -objects`.

Decisive comparison with mitmF (same object *content*, different handles):

| | mitmF | cold-start |
|---|---|---|
| srv root handles | `3c 01 ..01` | `90 7a ..04` |
| device dump handles | `3e 01 ..01` | `94 7a ..04` |

So the 8-byte handle is `INDEX · EPOCH · 00 00 00 00 00 · TYPE`:
- **EPOCH** is a per-session generation (`01`, `7a`, `4b` across sessions). Both
  directions share it. It is announced in the **device's descriptor handle** (the
  `id1` 177-byte descriptor ends `…a9 d8 90 7a 01 00 00 00 00 04`), which
  `cmr.ParseDescriptor` already extracts.
- **INDEX** is allocated from a shared space based at the descriptor index (`0x90`
  here); srv takes `90,91,92…`, the device takes `93,94,95…`, interleaved, no
  collision.
- **TYPE** `0x01`/`0x04` (topic/class).

## Decisive: object-sync requires PKEX pairing

Tested the epoch hypothesis live: our open-pairing controller pushed the exact
cold-start root objects but with a freshly-chosen epoch (`0xAA`). The device
**rejected** them — `not found` for `90 aa ..04` / `91 aa ..04`, the handles we
chose. So the controller cannot freely choose the epoch; the device validates it.

Yet srv *does* choose the epoch: in the cold-start, connection-1 carries no `7a`,
and srv introduces `7a` at its first roots with no prior device-provided value —
and the device accepts it. The only difference between srv and our controller is
that **srv completed the PKEX password handshake** and we did open pairing.

Every object sync in every capture (`mitm*.log`, `coldstart.log`) is preceded by
the PKEX "password required" exchange; there is no capture of a sync without it,
and our open-paired controller never gets a dump. Conclusion:

> The nv object-sync is gated on PKEX authentication. A device accepts a
> controller's object ADDs (and its chosen session epoch) only from a
> PKEX-authenticated controller. Open pairing reaches `paired,connected` but the
> device will not sync config or dump state to an unauthenticated controller —
> it answers the controller's objects with `not found`.

(The one untested alternative is a RouterOS controller configured for open
pairing; srv in this lab is PKEX-only, so it can't disprove the above. But the
security logic — only an authenticated controller may push config — and the
uniform evidence make PKEX the real requirement.)

### Consequence for the project

Live enumeration needs the **PKEX password PAKE**, not more object-layer work.
That is the EC+HMAC handshake already mapped in `FINDINGS.md`, whose KDF/MAC byte
calibration is the remaining piece (`pkg/pairing` has the mechanics;
`PKEX::start`/`authDeriveZ`/`reveal1` are decompiled). Once PKEX completes, the
object decoder (`pkg/cmr`) already reads the full dump, and srv's object content
in `coldstart.log` is the template for the controller's own config push.

## Implications for live enumeration

1. A fresh controller must **ADDOBJ** its scope objects (device allocates the
   local-ids and records the pairing) — not announce self-chosen handles via a
   SETOBJ-style push.
2. To pull the device's state, issue a command the device dispatches as GETALL or
   SUBSCRIBE **with the correct command carrier** (our prior id7/id11 frames
   lacked it, so getType fell through to ERROR_UNKNOWN → bland ack / no dump).
3. Validation path given KVM gdbstub limits: the in-guest ptrace tracer
   (`docs/re/figman/figtrace.c`) retargeted to figman's PLT stubs for
   `getType`, `handleCmd`, `Store::get/set`, and `cmdAddObj` — log command + id
   per message during a live sync.

## CORRECTION (2026-10-18): object-sync does NOT require PKEX

The "requires PKEX" conclusion above is **WRONG** and is superseded here. It was
inferred because every prior capture happened to be PKEX; the one untested case —
a real RouterOS controller in OPEN pairing mode — was never run. It has now been
run and it drives a full sync.

**Experiment:** set srv's controller to open pairing (`/cmr/print` →
`pairing-requirement: none`), MITM the device↔srv (`cmr-mitm`, defaults
listen 10.99.0.1:54321 → srv 10.99.0.113:54321), `cli /cmr/client/{forget,set
controller-addresses=10.99.0.1,pair password=x username=admin}`. Capture:
`docs/samples/open-mode-sync.log` (decoded `.decoded.txt`).

**Result:** srv (OPEN, no PKEX) drove a complete controller-driven sync — **467
`id9` ADDOBJ** pushes accepted, device dumped its full state as **`id1`** rows,
exactly ONE soft reject (`st=0x83 "not talking to you!"` on the board object,
non-fatal), **zero "not found"**. So an authenticated/PKEX controller is NOT
required to enumerate; a correctly-behaving open controller works.

**Why our controller failed before — the real diffs.** Byte-compare srv's
working id9 roots vs our `ctrl_id9_push.hex`:

```
srv  (WORKS): 49e1 d1e9 10 d2010201 a9d8 2449040000000004   id29=0x10  handle ...TYPE=04  epoch=0x49 idx=0x24
ours (FAILS): 49e1 d1e9 33 d2010201 a9d8 3a01000000000001   id29=0x33  handle ...TYPE=01  epoch=0x01 idx=0x3a
```

Two concrete encoding bugs in our push:
1. **id29 flag**: srv uses **0x10**, we used **0x33**.
2. **handle TYPE byte**: srv uses **0x04**, we used **0x01**.

The epoch/index are controller-chosen and free (srv picked 0x49/0x24 for this
fresh pairing; the device adopted them — it echoes handle `24…` back as its own
descriptor seq, frame 21). So the earlier "device validates/rejects our epoch"
reading was a symptom of the malformed flag/type, not an epoch rule.

**Working choreography (open mode), from the capture:**
`id2/id1(ver)/id6(disc)` handshake → device emits `id3 st=0x85 "configuration
changed"` and reconnects → handshake again → srv `id10 SUCCESS (scope
com.mikrotik.cmr, st=0)` → srv pushes 2 roots (`id9` handle 24=root/no-key,
25=board) → device acks + sends `id7` getall(com.mikrotik.cmr) → srv streams its
config as `id9` (class=config, flag 0x10, type 04, handles 26,27,…) → device
dumps its own state+config as `id1` under the matching handles. No PKEX anywhere.

**Consequence:** live enumeration from our Go controller needs the correct `id9`
encoding + choreography, NOT the PKEX password PAKE. The PKEX RE
(FINDINGS.md / pkg/pairing) is shelved as not-on-the-critical-path.

## SOLVED (2026-10-18): cmr-ctrl drives full live enumeration (open pairing)

`cmd/cmr-ctrl` now makes a stock device sync and dump its complete state, with no
PKEX and no RouterOS controller in the path. Verified live: **0 not-found**, the
device returns its 178-byte descriptor then dumps ether1 (52:54:00:cc:01:02),
identity/board/version (MAC 52:54:00:cc:01:01), and the full config/state graph
(commit 0578315).

The three things our controller was getting wrong, each pinned against srv's
open-mode sync (`docs/samples/open-mode-sync.log`):

1. **Object flag = the device's query flag.** The device's id6 discovery (and id7
   getall) carries an `id29` flag inside its nested `id4` query — 0x10/0x14/0x16,
   different every session ("password required", st=0x02, rides the same query).
   Every object the controller pushes MUST carry that same `id29` flag. We were
   sending a stale/fixed flag, so the device could not match our objects to its
   query and answered every one `not found`. Fix: capture the query flag
   (`conn.devFlag`) and stamp it into each pushed object.

2. **SUCCESS + roots are one continuous m3 stream.** srv sends `id10` SUCCESS and
   the two root `id9` objects in a single write, delta-chained (10→9→9). Sending
   them as separate `Encode()` frames breaks the chain. Fix:
   `successWithRootsFrame()` builds one message.

3. **Answer the device's id7 getall with the config burst.** srv replies to the
   device's getall by streaming its `com.mikrotik.cmr` config as an `id9` burst;
   that is what makes the device reciprocate with its own state dump. We were
   absorbing the getall. Fix: `-resyncfile` streams the config on id7.

Handle **idx and epoch are controller-chosen and the device adopts them** (it
echoes the root handle back as its descriptor seq). The earlier "device validates
our epoch" and "verAck inner seq == epoch" readings were both wrong — symptoms of
the flag mismatch. The objects pushed are currently srv's captured valid config
(`docs/samples/controller-{roots,config}.hex`).

## SOLVED (2026-10-20): cmr-ctrl DRIVES the WRITE direction — the device applies controller config

Our controller now pushes a config object graph that a stock device applies as
MANAGED config, live, with no RouterOS controller in the path. Verified: with
`-writefile`, the device builds a **MANAGED `cmr-bridge`** (`vlan-filtering=yes`,
flag `Y`, `RUNNING`) with `ether1` at `pvid=100` — byte-for-byte the effect a real
controller produces from `/cmr/vlan/add vlan-ids=100`. Capture of our controller
driving it: `docs/samples/ctrl-drives-write-vlan.log`.

How it was found: CMR config is authored on the controller (`/cmr/vlan`,
`/cmr/wifi`, `/cmr/device`, `/cmr/layout`). We set `/cmr/vlan/add vlan-ids=100` on
srv and MITM'd a device sync (`docs/samples/write-sync-vlan.log`). srv isolates
its own single uplink when the vlan applies (it builds the same managed bridge),
so the push was captured by adding the config mid-session over the already-open
MITM connection. The device built the managed bridge; we then replicated the
exact push from cmr-ctrl.

**The write is an `id7` into the device's OWN config scope**, NOT an `id9` in the
`com.mikrotik.cmr` controller scope (that scope stayed pure skeleton — no bridge/
vlan objects). Shape, exactly as the device's own getall is shaped:

```
id7{ id4{ id21 seq, id23 "com.mikrotik.cmr.config", id29 flag, id30 <graph> },
     id22 nonce, id21 reqid }
```

The `id22` nonce and `id21` reqid live INSIDE the `id7` envelope (siblings of
`id4`), not outside it — getting that nesting wrong (the old experimental
`getAll()` put them outside) makes the device treat the frame as a plain read and
answer by dumping its config instead of applying the graph. `id30` carries the
desired object graph:

```
02 01 00                                   header
4a00c03f d28a21 d3090036 d340420f(=1000000) 85 d3805101 d4c901c03f d30e00 fe   root cfg obj
02 00 00 01 00                             delimiter (index 0)
4a00c03f d21400 d2811a01 d3490040 0c d4c901c03f d30500 fe d3a10140 da"cmr-bridge"   bridge obj
02 00 00 01 01                             delimiter (index 1)
4a00c03f d21002 c902 d2c904 64(=vlan 100) 85 02 d4c901c03f d30500 fe   vlan obj
03 0121 00d40900f007 02 01 01              trailer
```

Objects are class `4a00c03f` (config) with sub-type `c901c03f` (= `3fc001c9` LE).
The device reconciles its config scope to this graph.

Other details that matter:
- **Timing.** The write must land when the device is in STEADY state (initial
  `com.mikrotik.cmr` sync done, state dumped). Sent mid initial-sync the device
  reads instead of applies. `cmr-ctrl` fires it on a timer (`-writedelay`, 6s)
  after SUCCESS, off a goroutine with a write mutex (the controller→device
  direction is one continuous stream, so sends must not interleave mid-frame).
- **Flag.** The write's `id29` flag does NOT have to equal the device's query
  flag (unlike the read-side object flag). srv used `0x1c`; it works verbatim even
  when the device's own query flag that session is `0x1a`. `-writeflag` (default
  `0x1c`).

`cmr-ctrl` flags for the write: `-writefile <id30 graph hex>` `-writescope`
(default `com.mikrotik.cmr.config`) `-writeflag` `-writedelay`.

## SOLVED (2026-10-21): author the vlan config graph from structured input

The `id30` graph is no longer a replayed capture — `pkg/cmr.BuildVLANConfig`
constructs it from a `VLANConfig{VLANID, BridgeName}`, and a stock device applies
authored graphs it has never seen. Verified live by driving `cmr-ctrl` → device
(no RouterOS, no hex file): the device builds a managed bridge with the vlan id
and bridge name we choose.

- vlan 100 — `BuildVLANConfig` reproduces srv's capture **byte-for-byte**
  (golden round-trip test, `pkg/cmr/config_test.go`).
- vlan 200 — authored (1-byte id), device → `ether1 pvid=200`, bridge vlan 200.
- vlan 300 — authored (2-byte id `d2 2c01`), device → `ether1 pvid=300`.
- vlan 150 + bridge `athena-cmr` — authored in-process via `cmr-ctrl -vlan 150
  -bridgename athena-cmr`, device → managed bridge `athena-cmr`, `ether1 pvid=150`.

**The graph is a nested m3 chunk stream** (decode with `cmd/cfgdump`):

```
02 01 00                        header
<object 0>                      global config entry
02 00 00 01 <idx>  <object>     each further object, 0-based index
03 0121 00d40900f007 02 01 01   trailer
```

Each top-level item is one m3 chunk (`pkg/m3.AppendChunk` / `FirstChunk`): a
single byte ≤0xcf is inline, ≤31 bytes is a run (`0xd0+len`), longer is
length-prefixed (`0xfe len8` …). Objects are chunks whose body is itself a chunk
stream; the `02`/`00`/`01` framing bytes are 1-byte inline chunks. The vlan-100
graph decomposes as three objects:

```
root : 4a00c03f | 8a21 | 090036 | 40420f(=1000000) | 85 | 805101(=86400) | c901c03f | 0e00fe
bridge: 4a00c03f | 1400 | 811a | 01 | 490040 | 0c | c901c03f | 0500fe | a10140 | "cmr-bridge"
vlan : 4a00c03f | 1002 | c9 | 02 | c904 | 64(=vlan id) | 85 | 02 | c901c03f | 0500fe
```

Each object is `ClassConfig` (`4a00c03f`) entry fields, then `ClassConfigValue`
(`c901c03f`, LE `0x3fc001c9`) and value fields. **Confirmed positions:** the vlan
id (a minimal-LE chunk, so `64` for 100, `d2 2c01` for 300) and the bridge name
(a trailing string chunk in the bridge object). The remaining per-object chunks
(`8a21`, `090036`, `490040`, `a10140`, `1002`, `c904`, the `85`/`02` markers …)
are schema positions not yet mapped; `BuildVLANConfig` carries them verbatim, and
the device accepts the result. Mapping them (e.g. multiple vlans, other object
classes such as wifi/device, and the two root integers) is the method's next
turn: drive srv with varied `/cmr/…` config and diff the captured graphs.
