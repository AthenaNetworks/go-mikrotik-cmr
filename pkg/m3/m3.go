// Package m3 implements MikroTik's nv/m3 message wire format, as used on the
// CMR control channel (the post-TLS "figman" protocol on tcp/54321).
//
// This is byte-exact with RouterOS 7.26: the codec was reverse-engineered from
// libumsg.so (m3::message::{fold,unfold,encnum32,fromraw32,read::each} and the
// getraw chunk reader) and validated against live CHR handshake frames (see
// docs/PROTOCOL.md and pkg/m3/m3_test.go). The format:
//
//	message = field*
//	field   = tag-chunk value-chunk
//	tag     = fold(type,id,prevID); type = tag&3, and the id is either absolute
//	          (tag>>3, when bit2==0) or delta-coded from the previous field's id
//	          (±(tag>>4), sign in bit3, when bit2==1)
//	chunk   = intro + data, where intro<=0xcf is a 1-byte inline value (the byte
//	          itself), intro in 0xd0..0xef is a short run of (intro-0xd0) bytes,
//	          and intro 0xfe/0xfd/0xfc/0xfb/0xfa introduces a length read from the
//	          next 1/2/4/8/16 little-endian bytes. 0xff is padding and is skipped.
//
// Numbers are little-endian and carried as raw chunk bytes (the receiver knows
// each id's type from its schema), so there are only a few physical wire types.
package m3

import (
	"encoding/binary"
	"errors"
	"fmt"
	"strings"
)

// Field ids observed in the cmr binary (queries and the dev object).
const (
	IDSel     uint32 = 0xff0001 // selector/vector in getall queries
	IDTopic   uint32 = 0xff0007 // topic id
	IDPkgs    uint32 = 0xfe0004 // installed-packages topic
	IDDevRow  uint32 = 0x2001   // dev object payload (Srv::onPKEXResult)
	IDDevFlag uint32 = 0x200b   // dev object flag (Srv::onPKEXResult)
)

// DefaultType is the physical wire type used for ordinary fields (the only one
// seen on the CMR handshake). The high-level type is resolved from the id by
// the peer's schema; the 2-bit wire type mainly distinguishes control fields.
const DefaultType uint8 = 1

// Field is one id -> value entry. Value holds the raw chunk bytes; numeric
// fields store a little-endian integer, nested messages store a sub-encoding.
type Field struct {
	ID    uint32
	Type  uint8
	Value []byte
}

// Message is an ordered collection of fields.
type Message struct {
	Fields []Field
}

// New returns an empty message.
func New() *Message { return &Message{} }

// Put appends a raw field with the default wire type.
func (m *Message) Put(id uint32, v []byte) *Message { return m.PutType(id, DefaultType, v) }

// PutType appends a raw field with an explicit wire type (0..3).
func (m *Message) PutType(id uint32, typ uint8, v []byte) *Message {
	m.Fields = append(m.Fields, Field{ID: id, Type: typ, Value: append([]byte(nil), v...)})
	return m
}

// PutU32 appends a field carrying v as a minimal little-endian integer.
func (m *Message) PutU32(id, v uint32) *Message { return m.Put(id, minimalLE(uint64(v))) }

// PutU64 appends a field carrying v as a minimal little-endian integer.
func (m *Message) PutU64(id uint32, v uint64) *Message { return m.Put(id, minimalLE(v)) }

// PutBytes appends a byte-string field.
func (m *Message) PutBytes(id uint32, b []byte) *Message { return m.Put(id, b) }

// PutNested appends a field whose value is the encoding of n.
func (m *Message) PutNested(id uint32, n *Message) *Message { return m.Put(id, n.Encode()) }

// PutVector appends a field whose value is a sub-message of elements, each
// carried as a repeated id=1 entry (a self-consistent container encoding).
func (m *Message) PutVector(id uint32, vs [][]byte) *Message {
	child := New()
	for _, v := range vs {
		child.PutBytes(1, v)
	}
	return m.Put(id, child.Encode())
}

// get returns the first field with id.
func (m *Message) get(id uint32) (Field, bool) {
	for _, f := range m.Fields {
		if f.ID == id {
			return f, true
		}
	}
	return Field{}, false
}

// U32 returns field id interpreted as a little-endian integer.
func (m *Message) U32(id uint32) (uint32, bool) {
	f, ok := m.get(id)
	if !ok {
		return 0, false
	}
	return uint32(leUint(f.Value)), true
}

// GetBytes returns the raw value of field id.
func (m *Message) GetBytes(id uint32) ([]byte, bool) {
	f, ok := m.get(id)
	if !ok {
		return nil, false
	}
	return f.Value, true
}

// GetNested decodes field id as a sub-message.
func (m *Message) GetNested(id uint32) (*Message, bool) {
	f, ok := m.get(id)
	if !ok {
		return nil, false
	}
	sub, err := Decode(f.Value)
	if err != nil {
		return nil, false
	}
	return sub, true
}

// GetVector returns the elements of a field encoded by PutVector.
func (m *Message) GetVector(id uint32) ([][]byte, bool) {
	sub, ok := m.GetNested(id)
	if !ok {
		return nil, false
	}
	out := make([][]byte, 0, len(sub.Fields))
	for _, f := range sub.Fields {
		out = append(out, f.Value)
	}
	return out, true
}

// Encode serializes the message to the m3 wire format.
func (m *Message) Encode() []byte {
	var b []byte
	var base uint32
	for _, f := range m.Fields {
		tag := fold(f.Type, f.ID, base)
		b = appendChunk(b, minimalLE(uint64(tag)))
		b = appendChunk(b, f.Value)
		base = f.ID
	}
	return b
}

var errTruncated = errors.New("m3: truncated message")

// Decode parses a message from b.
func Decode(b []byte) (*Message, error) {
	m := &Message{}
	var base uint32
	p := 0
	for p < len(b) {
		tagRaw, np, err := readChunk(b, p)
		if err != nil {
			return nil, err
		}
		p = np
		typ, id := unfold(uint32(leUint(tagRaw)), base)
		val, np, err := readChunk(b, p)
		if err != nil {
			return nil, err
		}
		p = np
		m.Fields = append(m.Fields, Field{ID: id, Type: typ, Value: val})
		base = id
	}
	return m, nil
}

// fold combines a wire type and field id into a tag, delta-coding the id
// against the previous field's id when that is shorter (matching m3::fold).
func fold(typ uint8, id, base uint32) uint32 {
	t := uint32(typ) & 3
	if base == 0 {
		return id*8 + t
	}
	var d uint32
	if id >= base {
		d = (id-base)*16 + 4
	} else {
		d = (base-id)*16 + 12
	}
	if id*8 < d {
		d = id * 8
	}
	return d + t
}

// unfold reverses fold given the previous field's id (m3::message::unfold).
func unfold(folded, base uint32) (uint8, uint32) {
	t := uint8(folded & 3)
	if folded&4 == 0 {
		return t, folded >> 3
	}
	delta := folded >> 4
	if folded&8 != 0 {
		return t, base - delta
	}
	return t, base + delta
}

// appendChunk writes data as an m3 chunk (getraw's inverse).
func appendChunk(b, data []byte) []byte {
	if len(data) == 1 && data[0] <= 0xcf {
		return append(b, data[0]) // inline single byte
	}
	l := len(data)
	switch {
	case l <= 0x1f:
		b = append(b, byte(0xd0+l))
	case l <= 0xff:
		b = append(b, 0xfe, byte(l))
	case l <= 0xffff:
		b = append(b, 0xfd, byte(l), byte(l>>8))
	case int64(l) <= 0xffffffff:
		var w [4]byte
		binary.LittleEndian.PutUint32(w[:], uint32(l))
		b = append(b, 0xfc)
		b = append(b, w[:]...)
	default:
		var w [8]byte
		binary.LittleEndian.PutUint64(w[:], uint64(l))
		b = append(b, 0xfb)
		b = append(b, w[:]...)
	}
	return append(b, data...)
}

// readChunk reads one m3 chunk starting at p, returning its data and the new
// offset.
func readChunk(b []byte, p int) ([]byte, int, error) {
	for p < len(b) && b[p] == 0xff { // skip padding
		p++
	}
	if p >= len(b) {
		return nil, p, errTruncated
	}
	intro := b[p]
	p++
	switch {
	case intro <= 0xcf:
		return []byte{intro}, p, nil
	case intro <= 0xef:
		n := int(intro - 0xd0)
		if p+n > len(b) {
			return nil, p, errTruncated
		}
		return append([]byte(nil), b[p:p+n]...), p + n, nil
	}
	var w int
	switch intro {
	case 0xfe:
		w = 1
	case 0xfd:
		w = 2
	case 0xfc:
		w = 4
	case 0xfb:
		w = 8
	case 0xfa:
		w = 16
	default:
		return nil, p, errTruncated
	}
	if p+w > len(b) {
		return nil, p, errTruncated
	}
	n := int(leUint(b[p : p+w]))
	p += w
	if n < 0 || p+n > len(b) {
		return nil, p, errTruncated
	}
	return append([]byte(nil), b[p:p+n]...), p + n, nil
}

// minimalLE returns v as the shortest little-endian byte slice (min 1 byte).
func minimalLE(v uint64) []byte {
	if v == 0 {
		return []byte{0}
	}
	var b []byte
	for v > 0 {
		b = append(b, byte(v))
		v >>= 8
	}
	return b
}

// leUint reads up to 8 bytes as a little-endian integer.
func leUint(b []byte) uint64 {
	var v uint64
	for i := 0; i < len(b) && i < 8; i++ {
		v |= uint64(b[i]) << (8 * i)
	}
	return v
}

// Dump returns a human-readable, recursively decoded view of an m3 message for
// debugging and reverse-engineering. Nested fields are expanded only when the
// value round-trips through the codec (a heuristic: opaque blobs that happen to
// decode are still shown, but malformed nesting is left as raw bytes). depthMax
// bounds recursion.
func Dump(b []byte, depthMax int) string {
	var sb strings.Builder
	dumpInto(&sb, b, 0, depthMax)
	return sb.String()
}

func dumpInto(sb *strings.Builder, b []byte, depth, depthMax int) {
	msg, err := Decode(b)
	if err != nil || msg == nil {
		return
	}
	ind := strings.Repeat("  ", depth+1)
	for _, f := range msg.Fields {
		v := f.Value
		asc := make([]byte, 0, len(v))
		for _, c := range v {
			if c >= 0x20 && c < 0x7f {
				asc = append(asc, c)
			} else {
				asc = append(asc, '.')
			}
		}
		shown := v
		trail := ""
		if len(shown) > 48 {
			shown = shown[:48]
			trail = "…"
			asc = asc[:min(len(asc), 48)]
		}
		fmt.Fprintf(sb, "%sid=%-4d len=%-4d %x%s |%s|\n", ind, f.ID, len(v), shown, trail, asc)
		if len(v) >= 2 && depth < depthMax {
			if sub, e := Decode(v); e == nil && sub != nil && len(sub.Fields) > 0 && bytesEqual(sub.Encode(), v) {
				dumpInto(sb, v, depth+1, depthMax)
			}
		}
	}
}

func bytesEqual(a, b []byte) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// FirstChunk reads the first m3 chunk from b (the getraw primitive), returning
// its data and the number of input bytes consumed. It is exported for decoders
// of fields whose value is itself chunk-framed (wire type 2 string fields such
// as the descriptor's os field carry their value as a single length-chunk).
func FirstChunk(b []byte) (data []byte, n int, err error) {
	data, p, err := readChunk(b, 0)
	return data, p, err
}
