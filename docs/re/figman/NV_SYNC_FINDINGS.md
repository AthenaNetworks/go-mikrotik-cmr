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
