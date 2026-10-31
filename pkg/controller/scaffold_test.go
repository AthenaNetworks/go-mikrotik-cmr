package controller

import (
	"bytes"
	"encoding/hex"
	"net"
	"os"
	"strings"
	"testing"

	"github.com/athenanetworks/cmr-server/pkg/m3"
)

func loadHex(t *testing.T, name string) []byte {
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

// TestSynthesizedScaffoldDefaults confirms the server's default scaffold (no
// captured blobs configured) is byte-identical to the captured controller blobs
// that pair a stock device live. This is what lets the server run without the
// docs/samples hex files; it guards the rootsBlob/configBurst fallback wiring.
func TestSynthesizedScaffoldDefaults(t *testing.T) {
	var empty Config
	if got, want := empty.rootsBlob(), loadHex(t, "controller-roots.hex"); !bytes.Equal(got, want) {
		t.Errorf("synthesised roots != capture (%d vs %d bytes)", len(got), len(want))
	}
	if got, want := empty.configBurst(), loadHex(t, "controller-config.hex"); !bytes.Equal(got, want) {
		t.Errorf("synthesised config burst != capture (%d vs %d bytes)", len(got), len(want))
	}

	// A configured capture still takes precedence over synthesis.
	cap := []byte{0x01, 0x02, 0x03}
	if got := (&Config{Roots: cap}).rootsBlob(); !bytes.Equal(got, cap) {
		t.Errorf("configured Roots not used: %x", got)
	}
	if got := (&Config{ConfigBurst: cap}).configBurst(); !bytes.Equal(got, cap) {
		t.Errorf("configured ConfigBurst not used: %x", got)
	}
}

// TestSynthesizedRootsOnWire drives a full handshake against a server with no
// captured blobs and asserts the SYNTHESISED roots ride the SUCCESS stream with
// the device's query flag applied — the rows that make a device start its dump.
func TestSynthesizedRootsOnWire(t *testing.T) {
	cert, err := Identity("")
	if err != nil {
		t.Fatal(err)
	}
	srv, err := New(Config{Certificate: cert}) // no Roots/ConfigBurst -> synthesised
	if err != nil {
		t.Fatal(err)
	}
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer ln.Close()
	go srv.Serve(ln)

	conn, readFrame, write := mockDevice(t, ln.Addr().String())
	defer conn.Close()
	bringUp(t, readFrame, write) // device query flag 0x14; reads through the SUCCESS (id10)

	// Two id9 roots rows ride the SUCCESS stream right after the id10.
	r0, r1 := readFrame(), readFrame()
	if r0.ID != 9 || r1.ID != 9 {
		t.Fatalf("expected two id9 roots after SUCCESS, got %d then %d", r0.ID, r1.ID)
	}
	b0, _ := m3.Decode(r0.Value)
	b1, _ := m3.Decode(r1.Value)
	if fl, _ := b0.GetBytes(29); len(fl) != 1 || fl[0] != 0x14 {
		t.Errorf("roots flag not patched to device flag: %x", fl)
	}
	if board, _ := b1.GetBytes(76); string(board) != "CHR" {
		t.Errorf("board row = %q, want CHR", board)
	}
}
