# CMR config schema — building a version-portable dictionary from the binaries

Goal: instead of capturing each config type's object graph live (blocked by lab
hardware — single-NIC/no-radio CHR can't manifest multi-vlan or wifi), recover
the config **schema** directly from the shipped RouterOS binaries, so our tool
can decode/author any config object and re-derive the dictionary on new firmware.

## Where the schema lives (RouterOS 7.26, cmr package)

The CMR *controller* config logic is in the **cmr package** (`cmr-7.26beta1.npk`),
not the base system:

- `nova/bin/cmr` — the controller daemon that builds the config object graph
  (`/cmr/vlan`, `/cmr/wifi`, `/cmr/layout`, `/cmr/device`, `/cmr/upgrade`).
  Strings confirm the nv store paths: `/nova/store/cmr/{layout,upgrade,wifinet,
  wifiradio}`, and object types "cmr layout {node,link}", "cmr wifi {,radio}",
  "cmr upgrade {job,rule}", "cmr-bridge".
- `nova/lib/console/*.mem` — the **compiled console schema** (RouterOS CLI syntax
  database): every menu, submenu, parameter, type and help string. This is the
  authoritative, machine-readable source of the config model and its names.
- `nova/etc/figman-extension/cmr.x3` — a 45-byte extension-registration stub
  (`167,2`), not the schema itself.

## The nv / m3 type system (`lib/libumsg.so`)

Decoded symbols show how objects are typed and encoded:

- Fields are `m3::message::put(m3::MidType, unsigned int id, span<data>)` — each
  field carries a `MidType` (wire type) and a numeric `id` (attribute id);
  `m3::message::fold(MidType, id, prevId)` is the delta-tag coder (already in
  `pkg/m3`).
- `nv::IdTraits<...>` enumerates attribute categories: `message_id`, `raw_id`,
  `string_id`, `u32_array_id`, `u64_array_id`, `bool_array_id`, `addr6_array_id`,
  `string_array_id`, `message_array_id` — the type vocabulary for attributes.
- `siphash` / `halfsiphash` are defined in libumsg but have **no callers** in
  the decompiled cmr/libumsg — they are NOT the type-id mechanism (likely PKEX /
  MAC crypto). Correcting an earlier guess.

### How type/attribute ids actually work (from the Ghidra decompile)

Full decompiles: `scratchpad/{cmr,umsg}_decomp.txt` (2287 + ~3000 functions, via
`analyzeHeadless` + a Java `DecompAll` script — Ghidra 12.x needs Java scripts,
not Python/PyGhidra).

- **The graph framing is generic.** All three objects in the vlan graph (root,
  bridge, vlan) share class `0x3fc0004a` ("cfg entry") + `0x3fc001c9`
  ("cfg value"). So these are not per-type ids — they are a generic config
  entry/value wrapper. The actual type and attributes live in the object's
  **nv columns**.
- **Stores register by name with a column count** (`FUN_08052c55` in cmr):
  `AMap::AMap("/nova/store/cmr/<type>", <ncols>, 1, -2, …)` —
  `vlan`=0x14, `dev`=0x13, `alert`=0x15, `layout`=0x16, `wifinet`=0x17,
  `wifiradio`=0x19, `laynode`/`laylink`=0x17; `upgrade`/`upgrade/job` are `AOmap`
  (ordered maps). The class id and column ids are assigned by the nv base class
  (`AMap`/`Store` in libumsg, `nv::getType`), **at runtime, not stored as
  constants** (confirmed: the ids don't appear as constants in cmr, libumsg, or
  the `.mem`, and the `.mem` is pure CLI syntax — names/types/help, no nv ids).
- So the per-type **dictionary is the nv column schema**: for each store, the
  ordered columns (name → type → the value encoding in the object body). The
  object body is a positional/column encoding under the generic wrapper, not a
  tagged (id,value) stream, which is why single-vlan decoded as opaque positional
  chunks.

## Two schema layers

1. **Config model** (names, types, structure, help) — in the console `.mem`.
   Drives a human-facing config API (`VLANConfig`-style structs per type).
2. **Graph serialization** (class id + positional field layout of the id30
   object graph) — in `cmr`'s object-build code. This is what `pkg/cmr`'s
   builder needs to emit bytes. For single-vlan it is already reversed
   (see NV_SYNC_FINDINGS.md); the rest is recoverable by decompiling the
   per-type serialize functions in `nova/bin/cmr`.

## First-pass vocabulary (extracted from the console .mem)

- **vlan**: `vlan-id`, `vlan-range` (+ comment, disabled).
- **wifi security.** (34): authentication-types, encryption, passphrase, sae-pwe,
  group-encryption, management-protection, ft*, eap-*, dpp-*, wps, …
- **wifi channel.** (23): band, frequency, width, secondary-frequency,
  skip-dfs-channels, reselect-interval, afc-* (automated frequency coordination).
- **wifi configuration.** (9): country, tx-power, tx-chains, chains, antenna-gain,
  distance, installation, atpc, atpc-rssi-threshold.
- **layout**: node, link (topology — maps devices/ports; empty by default, which
  is why srv assigns every device port to the first vlan rather than per-port).
- **device**: identity, board, version, channel, upgrade-rule.
- **upgrade**: rule, job, version, check-interval.

## SOLVED: name ↔ id ↔ type is the compiled console schema

The nv attribute ids are **not hashed and not column-serialized from scratch** —
they are stored verbatim in the compiled console schema (`.mem`), right next to
the attribute name and a type descriptor. The console needs them to issue nv
commands, and they are the *same* ids the cmr/libumsg code addresses with
`message::get/insert<nv::TYPE_id>(msg, id)` (verified: `0x2007`=`security.ft`,
`0x2103`=`channel`, `0x2600`=`starts-in`, `0x2203..06`=critical/high/medium/low).

The `.mem` is a flat array of property records:

```
[type-descriptor ptr : 4][kind = 0x00040008 : 4][nv id : 4][name]
```

`name` is inline NUL-terminated or a `0x50800000`-based pointer into the string
pool. There are two record shapes:

- **format 1** (property/enum): `[type-ptr][0x00040008][id][inline name]` — the
  type-ptr is one of six descriptors (u32 / flag / string / u32ro / mac / misc).
- **format 2** (menu arg): the name is a string-pool pointer at `id-12`, with a
  pointer (type/help) right after the id.

`cmd/cmr-schema` parses both shapes across one or more `.mem` files and emits the
dictionary. Run on the cmr package + wifi-bundle console schemas it yields **410
config attributes** (`cmr-schema-7.26beta1.txt`): vlan-id, ssid, band,
channel.frequency/width, bssid, bhaul-ssid, wifi-band, hide-ssid, target-layout,
node1/2, links, upgrade-rule, action.log/script, … It re-runs on any firmware's
`.mem`, so **the dictionary grows with new versions by design.** Types are filled
from the `.mem` descriptor and, where the code addresses the id, confirmed from
the decompile (`*` in the dump) — 154 of 410 typed so far.

Full decompiles used to confirm this: `scratchpad/{cmr,umsg}_decomp.txt`
(`analyzeHeadless` + Java `DecompAll` — Ghidra 12.x needs Java scripts, not
PyGhidra). The attribute accessors there are `message::get<nv::TYPE_id>(msg, ID)`,
so the object body is an m3-folded (id, value) stream under the generic
`0x3fc0004a`/`0x3fc001c9` wrapper, not opaque positional data.

## Two encodings — and why the dictionary does not key the push graph

An important correction to the earlier plan ("fold the dictionary in by id").
The nv attribute ids in this dictionary (`vlan-id`=0xd701, `band`=0x2c02, …) are
the **control-plane** object model — the ids the console and the device use in
the `nv` getall rows (`com.mikrotik.cmr` / `com.mikrotik.cmr.config`). They are
**not** the field layout of the id30 config-**push** graph.

Verified empirically against the captured vlan-100 graph:

- The dictionary ids do **not** appear literally in the graph (`d701`, `6000`,
  `101b`, `2c02` are all absent).
- The graph's object bodies are **positional** under the generic
  `ClassConfig` (0x3fc0004a) / `ClassConfigValue` (0x3fc001c9) wrapper. The
  vlan id sits at a fixed slot (object 2, entry 4), the bridge name at the
  bridge object's trailing value chunk — both confirmed live by authoring
  changed graphs a stock device applies.

So authoring the push graph is driven by **per-type field positions**, not by the
attribute dictionary. The dictionary remains the authority for the control-plane
model (decoding what the device reports, and the attribute vocabulary/validation).

## Config graph codec (`pkg/cmr`) — DONE

The structured encoder/decoder is built:

- `Graph` / `ConfigObject` with `Encode()` and `DecodeGraph()` — the id30 graph
  as a header, ordered class-wrapped objects, index delimiters and a trailer.
  `DecodeGraph(b).Encode()` reproduces `b` byte-for-byte, so a captured graph can
  be parsed, edited field by field and re-emitted.
- Typed field constructors `U32Chunk` / `VLANIDChunk` / `StrChunk` / `HexChunk`
  (the last carries still-opaque schema positions verbatim).
- `BuildVLANConfig` is re-expressed on top of the model (byte-identical to the
  capture) and `ParseVLANConfig` is its inverse (graph → struct).
- The dictionary is embedded as `pkg/cmr/schema.json` and exposed via
  `Attrs()` / `AttrByName()` / `AttrByID()`.
- `cfgdump -tree` decodes a graph structurally from the CLI.

New config types are added by declaring their object/field layout as a `Graph`
(once a reference graph is captured) — the framing, round-trip and typed slots
are handled by the codec.

## Controller scaffold synthesis (`pkg/cmr`) — DONE

Separate from the config-**write** graph above is the controller's own **scaffold**:
the id9 objects a RouterOS controller streams *to* a freshly paired device so it
begins, then completes, its state dump. There are two:

- **Roots** (`docs/samples/controller-roots.hex`, 45 B) — two id9 self-advertisement
  rows that ride the id10 SUCCESS frame. Row 0 is bare; row 1 carries the board
  string (`CHR`). They make the device start dumping.
- **Config-scope burst** (`controller-config.hex`, 2897 B) — 33 id9 objects
  describing the controller's `com.mikrotik.cmr` scope, streamed in answer to the
  device's getall. The device needs them to reciprocate with its full state.

These were replayed from captured hex. They are now **synthesised from structure**:

- Each id9 row is `{id74 key, [id75 value], id29 flag, id64 index, id21 handle}`.
  The key is `classChunk(ClassConfig)` + a 2-byte attribute id + a marker; the value
  (wire **type 2**, unlike every other field's type 1 — reproducing that is required
  for byte-identity) is a list of `0x07f0`-family type references and inline values.
- The envelope is generated: the query **flag** (`0x10` in the capture, rewritten to
  the device's flag by `patchObjectFlag`), the per-object **index**, and the **handle**
  counter sequence (`0x26..0x36` then `0x47..0x56` — a captured discontinuity).
- `EncodeRoots(flag, board)` and `EncodeConfigScope(flag)` rebuild the streams;
  both reproduce the captures **byte-for-byte** (`scaffold_test.go`). The object
  catalog (`scaffold_catalog.go`) is generated from the capture by `cmd/cmr-scaffold`
  (`go generate ./pkg/cmr`). The server defaults to these, so it drives a device's
  dump with **no captured hex at runtime**; `Config.Roots`/`ConfigBurst` override.

### Config-scope object grammar (`pkg/cmr/scaffold_decode.go`) — DONE

The per-object **grammar** is now fully reversed and named. `DecodeConfigScope()`
turns the opaque chunk catalog into a `[]ScopeRow`; `AnnotateConfigScope()` /
`cfgdump -scope` print it; and `TestDecodeConfigScopeLossless` proves the named
model reconstructs the catalog **byte-for-byte** (so the naming loses nothing).

Each row has a KEY and an optional VALUE.

**Key** = `ClassConfig · attr(u16) · <descriptor> · <query>`, where `<descriptor>`
is one of four forms and `<query>` is the trailing marker selecting which getall
variant the row answers:

| form     | descriptor                                             | used by     |
|----------|--------------------------------------------------------|-------------|
| `simple` | `0503`                                                 | most attrs  |
| `nested` | `490040 · sub(u8) · class 0x3f800309 · ext · 85fd7f`   | attr 0x0014 |
| `ext`    | `8dfa7f · 07 · 85fd7f`                                 | attr 0x048a |
| `bare`   | *(none — just `ClassConfig · attr`)*                   | attr 0x0a8a |

Query markers: `01` = **key-only** (these rows carry no value), `02` = a second
value variant, `0400fe` = **full** (the full type signature). The same attribute
can appear under several markers (e.g. 0x2658 appears key-only, v2 and full), and
the whole set is enumerated **twice** (handles 0x26..0x36 then 0x47..0x56) — a
getall followed by a re-enumeration.

**Value** = a type signature. 26 of the 27 value-bearing rows carry the envelope

```
class(0x07f00015) · 0x07f80101 · <middle> · 0x07f80109 · 0x07f00019 · 0x07f000ad
```

where `<middle>` is the object's column list — a mix of 0x07f0-family **type
references** and inline **column-id references** (e.g. `340008`, `4923`). One row
(attr 0x0a8a) is a different object class (`0x07f00009`) with no envelope. The 12
distinct type references and the two value classes are named as constants in
`scaffold_decode.go` by their structural role.

**The hard blocker — leaf ids are a runtime nv space.** What the grammar leaves
numeric is the leaf id space: the 2-byte attribute ids, the inline column-id
references, and the 0x07f0-family type handles. These are **not** the console
schema ids (none of the attr ids resolve against `schema.json` in any byte order),
and prior Ghidra RE already found the class/column ids are assigned by the nv base
class at startup and stored nowhere static (not in `cmr`, `libumsg`, or the
`.mem`). No capture carries a name string beside them (the controller advertises
only the *shape* of its scope; the device answers by handle, not by name). So they
are stable **by construction** — which is why byte-identical replay works — but
cannot be mapped to human names from static material alone.

Two paths could name the leaves, both needing hardware (lab is down):

1. A **live getall that carries names** — capture a RouterOS console/API dump of
   `com.mikrotik.cmr` that pairs each runtime id with its attribute name, giving a
   direct id→name table for this firmware.
2. **Type-signature fingerprinting** — match each row's column type signature
   against the console schema's per-store column sets (the stores register with
   known column counts: vlan=0x14, dev=0x13, alert=0x15, layout=0x16, wifinet=
   0x17, wifiradio=0x19), to name rows at the store/type level even without a live
   dump. The nested `attr 0x0014` row is the strongest candidate (it alone uses
   the embedded sub-class and the richest column list).

Because the ids are runtime-assigned and version-specific, the **engineering**
answer for a new firmware is unchanged: re-capture and re-run `cmd/cmr-scaffold`.
Naming the leaves is a comprehension/validation win, not a new authoring path.

## Remaining work

1. **Type coverage:** 256 of 410 attributes still show `unknown` — the format-2
   menu-arg records carry their own type descriptor (at `id+4`) that isn't yet
   mapped; map those and cross-reference more decompiled accessors to type them.
   A handful of params in deeper menu-tree layouts (some dotted `security.*`) are
   not yet captured by the two record shapes.
2. **More per-type templates:** wifi, device, layout and multi-vlan graphs —
   each needs a reference graph captured on hardware that can manifest it, then a
   `Graph` layout declared for it. The vlan type is the worked example.
3. **Positional → named fields:** the graph object body may be a second
   (column-id, value) space assigned at runtime (the vlan-id slot unfolds to a
   column id of 0x99, distinct from the nv id 0xd701). Mapping that space would
   let fields be addressed by name rather than position.

Lab reality: multi-vlan/wifi can't be *validated* on CHR (no 2nd switch port
semantics, no radio). The codec and the vlan type are built and unit-tested; the
other types need reference captures and validation on real switch/AP hardware.
