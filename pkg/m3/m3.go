// Package m3 is a compact, self-delimiting message codec used for the body of
// CMR protocol frames.
//
// The real MikroTik `m3::message` / `nv::message` wire format lives in RouterOS
// shared libraries and is NOT present in the cmr binary, so this is a faithful
// re-creation of its *structure* (a typed TLV of id -> value, with u32, bytes,
// vector and nested-message values) rather than a byte-identical copy. The
// field-id constants below are the ones observed in the binary. See
// docs/PROTOCOL.md; byte-level interop is a pending validation item.
package m3

import (
	"encoding/binary"
	"errors"
	"fmt"
)

// Field ids observed in the cmr binary (used in queries and the dev object).
const (
	IDSel     uint32 = 0xff0001 // selector/vector in getall queries
	IDTopic   uint32 = 0xff0007 // topic id
	IDPkgs    uint32 = 0xfe0004 // installed-packages topic
	IDDevRow  uint32 = 0x2001   // dev object payload (Srv::onPKEXResult)
	IDDevFlag uint32 = 0x200b   // dev object flag (Srv::onPKEXResult)
)

// Type tags for field values.
type Type uint8

const (
	TU32 Type = iota + 1
	TBytes
	TVector
	TNested
)

// Field is one id -> typed value entry.
type Field struct {
	ID     uint32
	Type   Type
	U32    uint32
	Bytes  []byte
	Vector [][]byte
	Nested *Message
}

// Message is an ordered collection of fields.
type Message struct {
	Fields []Field
}

// New returns an empty message.
func New() *Message { return &Message{} }

// PutU32 appends a u32 field.
func (m *Message) PutU32(id, v uint32) *Message {
	m.Fields = append(m.Fields, Field{ID: id, Type: TU32, U32: v})
	return m
}

// PutBytes appends a byte-string field.
func (m *Message) PutBytes(id uint32, b []byte) *Message {
	m.Fields = append(m.Fields, Field{ID: id, Type: TBytes, Bytes: append([]byte(nil), b...)})
	return m
}

// PutVector appends a vector-of-byte-strings field.
func (m *Message) PutVector(id uint32, vs [][]byte) *Message {
	cp := make([][]byte, len(vs))
	for i, v := range vs {
		cp[i] = append([]byte(nil), v...)
	}
	m.Fields = append(m.Fields, Field{ID: id, Type: TVector, Vector: cp})
	return m
}

// PutNested appends a nested-message field.
func (m *Message) PutNested(id uint32, n *Message) *Message {
	m.Fields = append(m.Fields, Field{ID: id, Type: TNested, Nested: n})
	return m
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

// U32 returns the u32 value of field id.
func (m *Message) U32(id uint32) (uint32, bool) {
	f, ok := m.get(id)
	if !ok || f.Type != TU32 {
		return 0, false
	}
	return f.U32, true
}

// GetBytes returns the byte value of field id.
func (m *Message) GetBytes(id uint32) ([]byte, bool) {
	f, ok := m.get(id)
	if !ok || f.Type != TBytes {
		return nil, false
	}
	return f.Bytes, true
}

// GetVector returns the vector value of field id.
func (m *Message) GetVector(id uint32) ([][]byte, bool) {
	f, ok := m.get(id)
	if !ok || f.Type != TVector {
		return nil, false
	}
	return f.Vector, true
}

// GetNested returns the nested message of field id.
func (m *Message) GetNested(id uint32) (*Message, bool) {
	f, ok := m.get(id)
	if !ok || f.Type != TNested {
		return nil, false
	}
	return f.Nested, true
}

// Encode serializes the message.
func (m *Message) Encode() []byte {
	var b []byte
	b = appendUvarint(b, uint64(len(m.Fields)))
	for _, f := range m.Fields {
		b = appendUvarint(b, uint64(f.ID))
		b = append(b, byte(f.Type))
		switch f.Type {
		case TU32:
			var u [4]byte
			binary.BigEndian.PutUint32(u[:], f.U32)
			b = append(b, u[:]...)
		case TBytes:
			b = appendUvarint(b, uint64(len(f.Bytes)))
			b = append(b, f.Bytes...)
		case TVector:
			b = appendUvarint(b, uint64(len(f.Vector)))
			for _, v := range f.Vector {
				b = appendUvarint(b, uint64(len(v)))
				b = append(b, v...)
			}
		case TNested:
			n := f.Nested.Encode()
			b = appendUvarint(b, uint64(len(n)))
			b = append(b, n...)
		}
	}
	return b
}

var errTruncated = errors.New("m3: truncated message")

// Decode parses a message from b.
func Decode(b []byte) (*Message, error) {
	m := &Message{}
	p := 0
	nfields, n := binary.Uvarint(b[p:])
	if n <= 0 {
		return nil, errTruncated
	}
	p += n
	for i := uint64(0); i < nfields; i++ {
		id, n := binary.Uvarint(b[p:])
		if n <= 0 {
			return nil, errTruncated
		}
		p += n
		if p >= len(b) {
			return nil, errTruncated
		}
		t := Type(b[p])
		p++
		f := Field{ID: uint32(id), Type: t}
		switch t {
		case TU32:
			if p+4 > len(b) {
				return nil, errTruncated
			}
			f.U32 = binary.BigEndian.Uint32(b[p : p+4])
			p += 4
		case TBytes:
			v, np, err := readChunk(b, p)
			if err != nil {
				return nil, err
			}
			f.Bytes, p = v, np
		case TVector:
			cnt, n := binary.Uvarint(b[p:])
			if n <= 0 {
				return nil, errTruncated
			}
			p += n
			for j := uint64(0); j < cnt; j++ {
				v, np, err := readChunk(b, p)
				if err != nil {
					return nil, err
				}
				f.Vector = append(f.Vector, v)
				p = np
			}
		case TNested:
			v, np, err := readChunk(b, p)
			if err != nil {
				return nil, err
			}
			sub, err := Decode(v)
			if err != nil {
				return nil, err
			}
			f.Nested, p = sub, np
		default:
			return nil, fmt.Errorf("m3: unknown field type %d", t)
		}
		m.Fields = append(m.Fields, f)
	}
	return m, nil
}

func readChunk(b []byte, p int) ([]byte, int, error) {
	l, n := binary.Uvarint(b[p:])
	if n <= 0 {
		return nil, 0, errTruncated
	}
	p += n
	if p+int(l) > len(b) {
		return nil, 0, errTruncated
	}
	v := append([]byte(nil), b[p:p+int(l)]...)
	return v, p + int(l), nil
}

func appendUvarint(b []byte, v uint64) []byte {
	var tmp [binary.MaxVarintLen64]byte
	n := binary.PutUvarint(tmp[:], v)
	return append(b, tmp[:n]...)
}
