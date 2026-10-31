package cmr

import (
	"bytes"
	"encoding/hex"
	"testing"
)

// TestDecodeGraphRoundTrip decodes the captured vlan-100 graph and re-encodes it,
// which must reproduce the input byte-for-byte.
func TestDecodeGraphRoundTrip(t *testing.T) {
	want, err := hex.DecodeString(goldenVLAN100)
	if err != nil {
		t.Fatal(err)
	}
	g, err := DecodeGraph(want)
	if err != nil {
		t.Fatalf("DecodeGraph: %v", err)
	}
	if len(g.Objects) != 3 {
		t.Fatalf("got %d objects, want 3", len(g.Objects))
	}
	if got := g.Encode(); !bytes.Equal(got, want) {
		t.Fatalf("re-encode mismatch\n got %x\nwant %x", got, want)
	}
}

// TestDecodeGraphStructure checks the decoded objects carry the expected slots:
// the bridge name as the bridge object's trailing value chunk, and the vlan id
// as object 2's vlan-id entry slot.
func TestDecodeGraphStructure(t *testing.T) {
	b := BuildVLANConfig(VLANConfig{VLANID: 250, BridgeName: "lab-br"})
	g, err := DecodeGraph(b)
	if err != nil {
		t.Fatal(err)
	}
	bridge := g.Objects[1]
	if name := string(bridge.Value[len(bridge.Value)-1]); name != "lab-br" {
		t.Fatalf("bridge name = %q, want lab-br", name)
	}
	if id := g.Objects[2].Entry[vlanIDSlot].Uint(); id != 250 {
		t.Fatalf("vlan id = %d, want 250", id)
	}
}

// TestDecodeGraphRejectsGarbage ensures DecodeGraph does not accept malformed
// input as a valid graph.
func TestDecodeGraphRejectsGarbage(t *testing.T) {
	for _, h := range []string{"", "00", "020101", "0201009999"} {
		b, _ := hex.DecodeString(h)
		if _, err := DecodeGraph(b); err == nil {
			t.Errorf("DecodeGraph(%q) accepted malformed input", h)
		}
	}
}

func TestEncodeEmptyGraph(t *testing.T) {
	// A graph with no objects is still well-formed: header + trailer.
	g := &Graph{}
	b := g.Encode()
	rt, err := DecodeGraph(b)
	if err != nil {
		t.Fatalf("empty graph did not round-trip: %v", err)
	}
	if len(rt.Objects) != 0 {
		t.Fatalf("empty graph decoded %d objects", len(rt.Objects))
	}
}

func TestTypedChunks(t *testing.T) {
	if got := hex.EncodeToString(U32Chunk(1000000)); got != "40420f" {
		t.Errorf("U32Chunk(1000000) = %s, want 40420f", got)
	}
	if got := hex.EncodeToString(VLANIDChunk(300)); got != "2c01" {
		t.Errorf("VLANIDChunk(300) = %s, want 2c01", got)
	}
	if U32Chunk(86400).Uint() != 86400 {
		t.Errorf("U32Chunk/Uint round-trip failed")
	}
	if string(StrChunk("cmr-bridge")) != "cmr-bridge" {
		t.Errorf("StrChunk mismatch")
	}
}
