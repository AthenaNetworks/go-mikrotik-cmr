package cmr

import (
	"encoding/hex"
	"testing"

	"github.com/athenanetworks/cmr-server/pkg/m3"
)

// The device descriptor body (value of the top-level id0 field) from the live
// open-pairing capture, frame #10 — the device reporting its own identity.
const descriptorHex = "d20902d84b60be6a01621a0715d378383636d9d8726f757465726f73" +
	"2dd343485215d085d34348521dd3de450535d015d40066180715fe70" +
	"d26a01fe4fe0d2c1010149d40000000885d4ffffffffdc49d400000d07" +
	"85d4ffffff07e0d2c1010149d40000320685d4ffffff06dc49d4086631" +
	"0685d4ff663106e2d2c10101d20901d4ffffffff8dd4d1f83864450569" +
	"e8d2890200d23108e0912ecf98fb1599cd81d7b47df3e47990" +
	"a9d81400000000000004"

func TestParseDescriptor(t *testing.T) {
	body, err := hex.DecodeString(descriptorHex)
	if err != nil {
		t.Fatal(err)
	}
	// It must decode as a well-formed m3 object and look like a descriptor.
	obj, err := m3.Decode(body)
	if err != nil {
		t.Fatalf("decode: %v", err)
	}
	if !IsDescriptor(obj) {
		t.Fatalf("IsDescriptor = false, want true")
	}

	d, err := ParseDescriptor(body)
	if err != nil {
		t.Fatalf("ParseDescriptor: %v", err)
	}
	if d.Identity != "CHR" {
		t.Errorf("Identity = %q, want CHR", d.Identity)
	}
	if d.Board != "CHR" {
		t.Errorf("Board = %q, want CHR", d.Board)
	}
	if d.Arch != "x86" {
		t.Errorf("Arch = %q, want x86", d.Arch)
	}
	if d.OS != "routeros" {
		t.Errorf("OS = %q, want routeros (type-2 chunk unwrap)", d.OS)
	}
	if got := hex.EncodeToString(d.Build); got != "4b60be6a01621a07" {
		t.Errorf("Build = %s, want 4b60be6a01621a07", got)
	}
	if len(d.Blob) != 112 {
		t.Errorf("Blob len = %d, want 112", len(d.Blob))
	}
	if got := hex.EncodeToString(d.Handle); got != "1400000000000004" {
		t.Errorf("Handle = %s, want 1400000000000004", got)
	}
	t.Logf("parsed: %s (blob %d bytes, handle %x)", d, len(d.Blob), d.Handle)
}
