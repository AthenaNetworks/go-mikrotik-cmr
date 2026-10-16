package m3

import (
	"encoding/hex"
	"testing"
)

// The running base is what distinguishes stream decoding from per-message
// decoding: a field with delta tag 0x2d means "base-2", so after an absolute
// id9 it decodes to id7 — exactly the getall opcode that mis-decodes as
// 0xFFFFFFFE when the base is wrongly reset per TLS write.
func TestStreamRunningBase(t *testing.T) {
	// id9(absolute, value 0x01) then tag 0x2d (=> id = 9-2 = 7) value empty(0xd0)
	buf, _ := hex.DecodeString("49012dd0")
	fields := StreamAll(buf)
	if len(fields) != 2 {
		t.Fatalf("got %d fields, want 2", len(fields))
	}
	if fields[0].ID != 9 {
		t.Errorf("field0 id=%d, want 9", fields[0].ID)
	}
	if fields[1].ID != 7 {
		t.Errorf("field1 id=%d, want 7 (base-2 from 9); got the per-message-reset bug if this is 0xFFFFFFFE (%d)", fields[1].ID, fields[1].ID)
	}
}

// Feeding in arbitrary splits (as TLS reads would chunk the stream) yields the
// same fields as feeding the whole buffer at once.
func TestStreamSplitFeeds(t *testing.T) {
	whole, _ := hex.DecodeString("49012dd0")
	want := StreamAll(whole)

	d := NewStreamDecoder()
	var got []Field
	for i := 0; i < len(whole); i++ { // feed one byte at a time
		d.Feed(whole[i : i+1])
		for {
			f, ok := d.Next()
			if !ok {
				break
			}
			got = append(got, f)
		}
	}
	if len(got) != len(want) {
		t.Fatalf("split feed got %d fields, want %d", len(got), len(want))
	}
	for i := range got {
		if got[i].ID != want[i].ID || string(got[i].Value) != string(want[i].Value) {
			t.Errorf("field %d mismatch: got id=%d %x, want id=%d %x", i, got[i].ID, got[i].Value, want[i].ID, want[i].Value)
		}
	}
}
