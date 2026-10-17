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

## Open item: exact wire framing of the command

Captured messages decode (continuous-base) as a nested outer opcode (id7 getall,
id9 object, id11 subscribe, id1 reply). Decoded base-0 the getall is
`{ id4{ id21 reqid, id23 scope, id29 flag, id27 }, id22 nonce, id21 reqid }` with
no visible id3 — so the command is carried in the m3 message header
(`m3::MidType` + u32, per figman's `<lambda(m3::MidType,uint32_t,m3::In)>`), which
our field-only `pkg/m3` folds into the outer field id. Pinning the
MidType→command map (and whether a valid ADDOBJ must set `0xff0003` explicitly) is
the one remaining item before we can synthesize a valid create.

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
