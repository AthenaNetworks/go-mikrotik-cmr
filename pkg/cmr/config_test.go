package cmr

import (
	"bytes"
	"encoding/hex"
	"testing"

	"github.com/athenanetworks/cmr-server/pkg/m3"
)

// goldenVLAN100 is the id30 config graph a RouterOS 7.26 controller emits for
// `/cmr/vlan/add vlan-ids=100` (docs/samples/write-config-vlan100.hex), captured
// live. BuildVLANConfig must reproduce it byte-for-byte.
const goldenVLAN100 = "020100eed44a00c03fd28a21d3090036d340420f85d3805101d4c901c03fd30e00fe" +
	"0200000100fe29d44a00c03fd21400d2811a01d34900400cd4c901c03fd30500fed3a10140da636d722d627269646765" +
	"0200000101e9d44a00c03fd21002c902d2c904648502d4c901c03fd30500fe03d20121d600d40900f007020101"

func TestBuildVLANConfigGolden(t *testing.T) {
	want, err := hex.DecodeString(goldenVLAN100)
	if err != nil {
		t.Fatal(err)
	}
	got := BuildVLANConfig(VLANConfig{VLANID: 100})
	if !bytes.Equal(got, want) {
		t.Fatalf("vlan-100 graph mismatch\n got %x\nwant %x", got, want)
	}
	// The default bridge name must match RouterOS's.
	if !bytes.Equal(got, BuildVLANConfig(VLANConfig{VLANID: 100, BridgeName: DefaultBridgeName})) {
		t.Fatal("empty BridgeName must default to DefaultBridgeName")
	}
}

func TestBuildVLANConfigVLANID(t *testing.T) {
	base := BuildVLANConfig(VLANConfig{VLANID: 100})

	// A single-byte id (<=207) changes exactly one byte: the vlan-id chunk.
	v200 := BuildVLANConfig(VLANConfig{VLANID: 200})
	if len(v200) != len(base) {
		t.Fatalf("vlan-200 length %d != vlan-100 length %d", len(v200), len(base))
	}
	diff := 0
	for i := range v200 {
		if v200[i] != base[i] {
			diff++
			if base[i] != 0x64 || v200[i] != 0xc8 { // 100 -> 200 inline
				t.Fatalf("unexpected diff at %d: %#x -> %#x", i, base[i], v200[i])
			}
		}
	}
	if diff != 1 {
		t.Fatalf("vlan-200 should differ from vlan-100 in exactly one byte, got %d", diff)
	}

	// A multi-byte id (>207) grows the graph: a run-of-2 chunk (intro + 2 data)
	// replaces the inline id byte, +2 bytes.
	v300 := BuildVLANConfig(VLANConfig{VLANID: 300})
	if len(v300) != len(base)+2 {
		t.Fatalf("vlan-300 length %d, want %d", len(v300), len(base)+2)
	}
	if !bytes.Contains(v300, []byte{0xd2, 0x2c, 0x01}) { // run-of-2: 300 = 0x012c LE
		t.Fatalf("vlan-300 missing 2-byte id chunk d2 2c 01: %x", v300)
	}
}

func TestBuildVLANConfigBridgeName(t *testing.T) {
	g := BuildVLANConfig(VLANConfig{VLANID: 100, BridgeName: "office-net"})
	if !bytes.Contains(g, []byte("office-net")) {
		t.Fatalf("custom bridge name not present in graph: %x", g)
	}
}

// TestBuildVLANConfigWellFormed walks the graph as an m3 chunk stream and
// checks it consumes cleanly with no trailing bytes — the structural invariant
// the device relies on.
func TestBuildVLANConfigWellFormed(t *testing.T) {
	for _, id := range []int{1, 100, 200, 300, 4094} {
		g := BuildVLANConfig(VLANConfig{VLANID: id})
		b := g
		for len(b) > 0 {
			_, n, err := m3.FirstChunk(b)
			if err != nil || n == 0 {
				t.Fatalf("vlan-%d: chunk walk failed with %d bytes left: %v", id, len(b), err)
			}
			b = b[n:]
		}
	}
}

func TestMinLE(t *testing.T) {
	cases := map[uint64]string{0: "00", 100: "64", 200: "c8", 300: "2c01", 65535: "ffff"}
	for v, want := range cases {
		if got := hex.EncodeToString(minLE(v)); got != want {
			t.Errorf("minLE(%d) = %s, want %s", v, got, want)
		}
	}
}
