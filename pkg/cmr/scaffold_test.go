package cmr

import (
	"bytes"
	"encoding/hex"
	"os"
	"strings"
	"testing"
)

// loadSample reads a hex sample blob from docs/samples.
func loadSample(t *testing.T, name string) []byte {
	t.Helper()
	raw, err := os.ReadFile("../../docs/samples/" + name)
	if err != nil {
		t.Fatalf("read %s: %v", name, err)
	}
	b, err := hex.DecodeString(strings.TrimSpace(string(raw)))
	if err != nil {
		t.Fatalf("hex %s: %v", name, err)
	}
	return b
}

// TestEncodeConfigScopeMatchesCapture proves the synthesised config-scope burst
// reproduces the captured controller blob byte-for-byte at the capture's flag
// (0x10). This is the correctness anchor: the synthesis equals the bytes a real
// RouterOS controller sent and that a stock device accepted live.
func TestEncodeConfigScopeMatchesCapture(t *testing.T) {
	want := loadSample(t, "controller-config.hex")
	got := EncodeConfigScope(0x10)
	if !bytes.Equal(got, want) {
		t.Fatalf("config scope mismatch:\n got %d bytes %x\nwant %d bytes %x",
			len(got), firstDiff(got, want), len(want), want)
	}
}

// TestEncodeRootsMatchesCapture proves the synthesised roots reproduce the
// captured blob byte-for-byte at the capture's flag (0x10) and board ("CHR").
func TestEncodeRootsMatchesCapture(t *testing.T) {
	want := loadSample(t, "controller-roots.hex")
	got := EncodeRoots(0x10, "CHR")
	if !bytes.Equal(got, want) {
		t.Fatalf("roots mismatch:\n got %d bytes %x\nwant %d bytes %x", len(got), got, len(want), want)
	}
	if !bytes.Equal(EncodeRoots(0x10, ""), want) {
		t.Fatal("empty board should default to CHR")
	}
}

// TestConfigScopeFlagPatched confirms the query flag is applied everywhere: at a
// different flag the output differs from the capture only in the flag bytes, so a
// synthesised burst for a device's own flag is equivalent to patching the capture.
func TestConfigScopeFlagPatched(t *testing.T) {
	cap := loadSample(t, "controller-config.hex")
	got := EncodeConfigScope(0x1c)
	if len(got) != len(cap) {
		t.Fatalf("length changed with flag: %d vs %d", len(got), len(cap))
	}
	diffs := 0
	for i := range got {
		if got[i] != cap[i] {
			diffs++
			if cap[i] != 0x10 || got[i] != 0x1c {
				t.Fatalf("byte %d changed %#x->%#x, not a flag patch", i, cap[i], got[i])
			}
		}
	}
	if diffs != len(configScopeCatalog) {
		t.Fatalf("patched %d flag bytes, want one per object (%d)", diffs, len(configScopeCatalog))
	}
}

// TestRefChunk checks the 4-byte little-endian reference encoding.
func TestRefChunk(t *testing.T) {
	if got := RefChunk(0x07f00015); !bytes.Equal(got, []byte{0x15, 0x00, 0xf0, 0x07}) {
		t.Fatalf("RefChunk = %x", got)
	}
}

// firstDiff returns got truncated around the first byte that differs from want,
// for a readable failure message.
func firstDiff(got, want []byte) []byte {
	n := len(got)
	if len(want) < n {
		n = len(want)
	}
	for i := 0; i < n; i++ {
		if got[i] != want[i] {
			lo, hi := i-4, i+8
			if lo < 0 {
				lo = 0
			}
			if hi > len(got) {
				hi = len(got)
			}
			return got[lo:hi]
		}
	}
	return got
}
