package m3

import (
	"bytes"
	"encoding/hex"
	"testing"
)

// Frames captured from a live MikroTik CHR 7.26beta1 cmr client on the CMR
// control channel (post-TLS, tcp/54321). The codec must round-trip them
// byte-for-byte. See docs/PROTOCOL.md.
var liveFrames = []string{
	"11dcd21101d8707eae265b81f4c7",                             // F1 hello + 12-byte token
	"09e5b1d41bce15ad35d84b60be6a01621a071dd3434852",           // F2 version/identity
	"31d8b1d4d0893efe1d07",                                     // F3 {nonce, seq}
	"09d8b1d41508f0fc1d07",                                     // F4 {nonce, seq}
	"09e5b1d4557ebe5f35d84b60be6a01621a071dd3434852",           // F2 (conn2)
	"31d8b1d4c0c10c9a1d08",                                     // F3 (conn2)
	"09d8b1d4b584f4631d08",                                     // F4 (conn2)
}

func TestRoundTripLiveFrames(t *testing.T) {
	for _, h := range liveFrames {
		raw, err := hex.DecodeString(h)
		if err != nil {
			t.Fatalf("bad hex %q: %v", h, err)
		}
		msg, err := Decode(raw)
		if err != nil {
			t.Fatalf("decode %s: %v", h, err)
		}
		got := msg.Encode()
		if !bytes.Equal(got, raw) {
			t.Errorf("re-encode mismatch\n frame: %s\n   got: %s", h, hex.EncodeToString(got))
		}
	}
}

// TestDecodeVersionFrame checks the semantic fields recovered from F2: the
// outer opcode wraps a sub-message carrying a 4-byte nonce, the 8-byte
// build-time+version, and the board identity "CHR".
func TestDecodeVersionFrame(t *testing.T) {
	raw, _ := hex.DecodeString("09e5b1d41bce15ad35d84b60be6a01621a071dd3434852")
	msg, err := Decode(raw)
	if err != nil {
		t.Fatal(err)
	}
	if len(msg.Fields) != 1 || msg.Fields[0].ID != 1 {
		t.Fatalf("want single id=1 field, got %+v", msg.Fields)
	}
	inner, ok := msg.GetNested(1)
	if !ok {
		t.Fatal("id=1 is not a sub-message")
	}
	nonce, ok := inner.GetBytes(22)
	if !ok || len(nonce) != 4 {
		t.Fatalf("id=22 nonce: ok=%v %x", ok, nonce)
	}
	bv, ok := inner.GetBytes(25)
	if !ok || len(bv) != 8 {
		t.Fatalf("id=25 build/version: ok=%v %x", ok, bv)
	}
	if got := leUint(bv[:4]); got != 0x6abe604b {
		t.Errorf("build-time = %#x, want 0x6abe604b", got)
	}
	if got := leUint(bv[4:]); got != 0x071a6201 {
		t.Errorf("version = %#x, want 0x071a6201", got)
	}
	ident, ok := inner.GetBytes(24)
	if !ok || string(ident) != "CHR" {
		t.Errorf("id=24 identity = %q, want CHR", ident)
	}
}

func TestChunkEdgeCases(t *testing.T) {
	cases := [][]byte{
		{},
		{0x00},
		{0xcf},
		{0xd0},             // single byte that must escape (not inline)
		{0xff},             // padding byte as a value
		bytes.Repeat([]byte{0xab}, 31),
		bytes.Repeat([]byte{0xcd}, 32),  // crosses into marker length
		bytes.Repeat([]byte{0xee}, 300), // 2-byte length
	}
	for _, c := range cases {
		m := New().PutBytes(7, c)
		dec, err := Decode(m.Encode())
		if err != nil {
			t.Fatalf("decode len=%d: %v", len(c), err)
		}
		got, _ := dec.GetBytes(7)
		if !bytes.Equal(got, c) {
			t.Errorf("chunk round-trip len=%d: got %x want %x", len(c), got, c)
		}
	}
}
