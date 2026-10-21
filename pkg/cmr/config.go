package cmr

import (
	"encoding/binary"

	"github.com/athenanetworks/cmr-server/pkg/m3"
)

// ClassConfigValue is the "config value" class that terminates the entry part
// of each config object and introduces its value part. It pairs with
// ClassConfig ("config entry"); see object.go.
const ClassConfigValue uint32 = 0x3fc001c9 // "c901c03f" little-endian

// DefaultBridgeName is the managed bridge a RouterOS controller builds for a
// CMR vlan config.
const DefaultBridgeName = "cmr-bridge"

// VLANConfig is a single-vlan CMR configuration. The managed device builds a
// vlan-filtering bridge named BridgeName and tags its uplink into VLANID — the
// effect of `/cmr/vlan/add vlan-ids=<VLANID>` on a RouterOS controller.
type VLANConfig struct {
	VLANID     int    // 802.1Q vlan id (1..4094)
	BridgeName string // managed bridge name; empty means DefaultBridgeName
}

// BuildVLANConfig returns the id30 config object graph for c — the payload a
// controller writes into the device's own config scope (com.mikrotik.cmr.config)
// to push configuration. Hand the result to the write path in cmd/cmr-ctrl.
//
// The graph is a nested m3 chunk stream (see docs/re/figman/NV_SYNC_FINDINGS.md):
//
//	02 01 00                        header
//	<object 0>                      global config entry
//	02 00 00 01 <idx> <object idx>  each further object, 0-based index
//	03 0121 00d40900f007 02 01 01   trailer
//
// Each object is a chunk whose body is itself a chunk stream of positional,
// schema-typed fields: a ClassConfig id, some fields, a ClassConfigValue id,
// then more fields. Most field positions are schema-defined and not yet
// reversed; they are carried verbatim from a captured `/cmr/vlan/add` push. The
// positions confirmed live — by authoring changed graphs that a stock RouterOS
// 7.26 device applies (vlan ids 100/200/300, verified on the wire) — are the
// vlan id and the managed-bridge name, parameterised here.
func BuildVLANConfig(c VLANConfig) []byte {
	name := c.BridgeName
	if name == "" {
		name = DefaultBridgeName
	}
	return cat(
		chunk(0x02), chunk(0x01), chunk(0x00), // header
		chunk(rootObject()...),
		cfgDelim(0),
		chunk(bridgeObject(name)...),
		cfgDelim(1),
		chunk(vlanObject(c.VLANID)...),
		chunk(0x03), chunk(0x01, 0x21), chunk(0x00, 0xd4, 0x09, 0x00, 0xf0, 0x07), // trailer
		chunk(0x02), chunk(0x01), chunk(0x01),
	)
}

// rootObject is the global config entry (object 0). The two integers are a
// controller default set (1000000, and 86400 = one day in seconds); the
// remaining fields are opaque schema positions carried verbatim.
func rootObject() []byte {
	return cat(
		classChunk(ClassConfig),
		chunk(0x8a, 0x21),       // opaque
		chunk(0x09, 0x00, 0x36), // opaque
		chunk(0x40, 0x42, 0x0f), // 1000000
		chunk(0x85),             // opaque
		chunk(0x80, 0x51, 0x01), // 86400
		classChunk(ClassConfigValue),
		chunk(0x0e, 0x00, 0xfe), // opaque
	)
}

// bridgeObject is the managed-bridge entry (object 1); it carries the bridge
// name as a trailing string. The remaining fields are opaque.
func bridgeObject(name string) []byte {
	return cat(
		classChunk(ClassConfig),
		chunk(0x14, 0x00),       // opaque
		chunk(0x81, 0x1a),       // opaque
		chunk(0x01),             // opaque
		chunk(0x49, 0x00, 0x40), // opaque
		chunk(0x0c),             // opaque
		classChunk(ClassConfigValue),
		chunk(0x05, 0x00, 0xfe), // opaque
		chunk(0xa1, 0x01, 0x40), // opaque (precedes the name)
		chunk([]byte(name)...),  // managed bridge name
	)
}

// vlanObject is the vlan entry (object 2). The vlan id is a minimal
// little-endian chunk; the remaining fields are opaque.
func vlanObject(vlanID int) []byte {
	return cat(
		classChunk(ClassConfig),
		chunk(0x10, 0x02),               // opaque
		chunk(0xc9),                     // opaque
		chunk(0x02),                     // opaque
		chunk(0xc9, 0x04),               // opaque
		chunk(minLE(uint64(vlanID))...), // vlan id
		chunk(0x85),                     // opaque
		chunk(0x02),                     // opaque
		classChunk(ClassConfigValue),
		chunk(0x05, 0x00, 0xfe), // opaque
	)
}

// cfgDelim is the 5-chunk delimiter before object idx (idx>=0).
func cfgDelim(idx int) []byte {
	return cat(chunk(0x02), chunk(0x00), chunk(0x00), chunk(0x01), chunk(byte(idx)))
}

// chunk encodes data as one m3 chunk.
func chunk(data ...byte) []byte { return m3.AppendChunk(nil, data) }

// classChunk encodes a 4-byte little-endian class id as its own chunk.
func classChunk(class uint32) []byte {
	var w [4]byte
	binary.LittleEndian.PutUint32(w[:], class)
	return chunk(w[:]...)
}

// cat concatenates chunk slices.
func cat(parts ...[]byte) []byte {
	var out []byte
	for _, p := range parts {
		out = append(out, p...)
	}
	return out
}

// minLE renders v as minimal little-endian bytes (always at least one byte).
func minLE(v uint64) []byte {
	out := []byte{byte(v)}
	for v >>= 8; v > 0; v >>= 8 {
		out = append(out, byte(v))
	}
	return out
}
