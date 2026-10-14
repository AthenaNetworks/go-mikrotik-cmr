package m3

import (
	"bytes"
	"testing"
)

func TestRoundTrip(t *testing.T) {
	msg := New().
		PutU32(fID(1), 7).
		PutBytes(fID(0x11), []byte{1, 2, 3, 4}).
		PutVector(fID(0x15), [][]byte{[]byte("10.0.0.1"), []byte("10.0.0.2")}).
		PutNested(fID(0x20), New().PutU32(fID(2), 42))

	enc := msg.Encode()
	got, err := Decode(enc)
	if err != nil {
		t.Fatalf("decode: %v", err)
	}
	if v, ok := got.U32(1); !ok || v != 7 {
		t.Fatalf("u32 = %d,%v", v, ok)
	}
	if b, ok := got.GetBytes(0x11); !ok || !bytes.Equal(b, []byte{1, 2, 3, 4}) {
		t.Fatalf("bytes = %v,%v", b, ok)
	}
	vec, ok := got.GetVector(0x15)
	if !ok || len(vec) != 2 || string(vec[1]) != "10.0.0.2" {
		t.Fatalf("vector = %v,%v", vec, ok)
	}
	sub, ok := got.GetNested(0x20)
	if !ok {
		t.Fatal("missing nested")
	}
	if v, _ := sub.U32(2); v != 42 {
		t.Fatalf("nested u32 = %d", v)
	}
}

func TestDecodeTruncated(t *testing.T) {
	enc := New().PutBytes(fID(1), []byte("hello")).Encode()
	if _, err := Decode(enc[:len(enc)-2]); err == nil {
		t.Fatal("expected error decoding truncated message")
	}
}

func fID(v uint32) uint32 { return v }
