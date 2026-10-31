package cmr

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

// vlanGraph builds the config object graph for c. The graph has three objects —
// a global config entry, the managed bridge, and the vlan — expressed as typed
// field slots. The slots confirmed live (by authoring changed graphs that a
// stock RouterOS 7.26 device applies: vlan ids 100/200/300 and custom bridge
// names, verified on the wire) are the vlan id and the bridge name. The
// remaining positions are schema-defined and carried verbatim as HexChunks.
func vlanGraph(c VLANConfig) *Graph {
	name := c.BridgeName
	if name == "" {
		name = DefaultBridgeName
	}
	return &Graph{Objects: []ConfigObject{
		{ // object 0: global config entry (controller defaults: 1000000, 86400=1 day)
			Entry: []Chunk{HexChunk("8a21"), HexChunk("090036"), U32Chunk(1000000), HexChunk("85"), U32Chunk(86400)},
			Value: []Chunk{HexChunk("0e00fe")},
		},
		{ // object 1: the managed bridge, carrying its name
			Entry: []Chunk{HexChunk("1400"), HexChunk("811a"), HexChunk("01"), HexChunk("490040"), HexChunk("0c")},
			Value: []Chunk{HexChunk("0500fe"), HexChunk("a10140"), StrChunk(name)},
		},
		{ // object 2: the vlan, carrying its id
			Entry: []Chunk{HexChunk("1002"), HexChunk("c9"), HexChunk("02"), HexChunk("c904"), VLANIDChunk(c.VLANID), HexChunk("85"), HexChunk("02")},
			Value: []Chunk{HexChunk("0500fe")},
		},
	}}
}

// vlanIDSlot and bridgeNameSlot name the field positions ParseVLANConfig reads
// back — the inverse of the slots vlanGraph fills.
const (
	vlanIDSlot = 4 // object 2, Entry index of the vlan-id chunk
)

// BuildVLANConfig returns the id30 config object graph for c — hand the result
// to the write path in cmd/cmr-ctrl. It round-trips byte-for-byte with the graph
// a RouterOS 7.26 controller emits for `/cmr/vlan/add vlan-ids=<VLANID>`.
func BuildVLANConfig(c VLANConfig) []byte { return vlanGraph(c).Encode() }

// ParseVLANConfig decodes a vlan config graph (the inverse of BuildVLANConfig)
// into its VLANConfig, reading the vlan-id and bridge-name slots. It reports
// false if b is not a well-formed three-object vlan graph.
func ParseVLANConfig(b []byte) (VLANConfig, bool) {
	g, err := DecodeGraph(b)
	if err != nil || len(g.Objects) < 3 {
		return VLANConfig{}, false
	}
	bridge := g.Objects[1]
	if len(bridge.Value) == 0 {
		return VLANConfig{}, false
	}
	name := string(bridge.Value[len(bridge.Value)-1]) // the trailing name chunk
	vlan := g.Objects[2]
	if len(vlan.Entry) <= vlanIDSlot {
		return VLANConfig{}, false
	}
	return VLANConfig{VLANID: int(vlan.Entry[vlanIDSlot].Uint()), BridgeName: name}, true
}
