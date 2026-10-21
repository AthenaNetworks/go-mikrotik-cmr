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

## Remaining work

1. **Type coverage:** 256 of 410 attributes still show `unknown` — the format-2
   menu-arg records carry their own type descriptor (at `id+4`) that isn't yet
   mapped; map those and cross-reference more decompiled accessors to type them.
   A handful of params in deeper menu-tree layouts (some dotted `security.*`) are
   not yet captured by the two record shapes.
2. **Schema-driven encoder:** fold the dictionary into `pkg/cmr` so
   `BuildVLANConfig` generalises to `{type, {attr: value}}` → object graph, using
   the id+type per attribute — the payoff that turns the dictionary into
   free-form config authoring.

Lab reality: multi-vlan/wifi can't be *validated* on CHR (no 2nd switch port
semantics, no radio), but the schema-driven encoder can be built and unit-tested
from the dictionary, and validated against real switch/AP hardware later.
