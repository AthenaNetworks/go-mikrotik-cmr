package cmr

import (
	"encoding/binary"
	"encoding/hex"
	"fmt"

	"github.com/athenanetworks/cmr-server/pkg/m3"
)

// ClassConfigValue is the "config value" class that terminates the entry part of
// each config object and introduces its value part. It pairs with ClassConfig
// ("config entry", see object.go); both are generic wrappers shared by every
// config object, so the object's actual type and fields live in the chunk
// positions between and after them, not in the class id.
const ClassConfigValue uint32 = 0x3fc001c9 // "c901c03f" little-endian

// defaultTrailer closes a config object graph. These are the raw framing bytes
// that follow the last object in a RouterOS 7.26 push graph; they are carried
// verbatim (03 marker, a 0121 chunk, a 00d40900f007 chunk, then 02 01 01).
var defaultTrailer = []byte{0x03, 0xd2, 0x01, 0x21, 0xd6, 0x00, 0xd4, 0x09, 0x00, 0xf0, 0x07, 0x02, 0x01, 0x01}

// A Chunk is one config field's data bytes (unframed). Encode frames it as a
// single m3 chunk; DecodeGraph returns the unframed data. Build chunks with the
// typed constructors (U32Chunk / VLANIDChunk / StrChunk / HexChunk) so intent is
// visible, or supply raw bytes directly.
type Chunk []byte

// U32Chunk encodes v as a minimal little-endian integer, the form RouterOS uses
// for numeric config fields (e.g. 1000000 -> 40 42 0f).
func U32Chunk(v uint32) Chunk { return Chunk(minLE(uint64(v))) }

// VLANIDChunk encodes an 802.1Q vlan id as a minimal little-endian integer
// (100 -> 64, 300 -> 2c 01). It is U32Chunk with an intent-revealing name.
func VLANIDChunk(id int) Chunk { return Chunk(minLE(uint64(id))) }

// StrChunk encodes s as its raw bytes (e.g. a bridge name).
func StrChunk(s string) Chunk { return Chunk([]byte(s)) }

// HexChunk decodes a hex string into a chunk. It is used for schema positions
// that are carried verbatim from a captured graph and not yet interpreted.
func HexChunk(h string) Chunk {
	b, err := hex.DecodeString(h)
	if err != nil {
		panic("cmr: bad HexChunk " + h) // only ever called with compile-time constants
	}
	return Chunk(b)
}

// Uint returns the chunk's bytes as a little-endian integer.
func (c Chunk) Uint() uint64 { return leU64(c) }

// ConfigObject is one object in a config graph. Its on-wire body is a chunk
// stream: ClassConfig, the Entry fields, ClassConfigValue, then the Value
// fields. Field positions are schema-defined; the typed slots we have confirmed
// live are filled by the per-type builders (see config.go) and the rest are
// carried verbatim.
type ConfigObject struct {
	Entry []Chunk // fields between ClassConfig and ClassConfigValue
	Value []Chunk // fields after ClassConfigValue
}

// body renders the object's on-wire body (the inner chunk stream that Encode
// wraps as a single top-level chunk).
func (o ConfigObject) body() []byte {
	b := classChunk(ClassConfig)
	for _, c := range o.Entry {
		b = append(b, chunk(c...)...)
	}
	b = append(b, classChunk(ClassConfigValue)...)
	for _, c := range o.Value {
		b = append(b, chunk(c...)...)
	}
	return b
}

// Graph is the decoded form of an id30 config object graph — the payload a
// controller writes into a device's own config scope (com.mikrotik.cmr.config)
// to push configuration. It is a fixed header, an ordered list of config
// objects separated by index delimiters, and a trailer:
//
//	02 01 00                        header
//	<object 0>                      a top-level chunk wrapping the object body
//	02 00 00 01 00  <object 1>      each further object, preceded by its 0-based delimiter
//	02 00 00 01 01  <object 2>
//	03 .. 02 01 01                  trailer
//
// DecodeGraph(b).Encode() reproduces b byte-for-byte, so a captured graph can be
// parsed, edited field by field and re-emitted. The object body is positional,
// not nv-id-tagged: the attribute dictionary in schema.go describes the separate
// control-plane object model, not these field positions.
type Graph struct {
	Objects []ConfigObject
	Trailer []byte // raw trailer bytes; empty means defaultTrailer
}

// Encode serialises the graph to the id30 wire form.
func (g *Graph) Encode() []byte {
	out := cat(chunk(0x02), chunk(0x01), chunk(0x00)) // header
	for i, obj := range g.Objects {
		if i > 0 {
			out = append(out, cfgDelim(i-1)...)
		}
		out = append(out, chunk(obj.body()...)...)
	}
	if len(g.Trailer) > 0 {
		return append(out, g.Trailer...)
	}
	return append(out, defaultTrailer...)
}

// DecodeGraph parses an id30 config object graph into its structured form.
func DecodeGraph(b []byte) (*Graph, error) {
	g := &Graph{}
	p := 0
	read := func() ([]byte, error) {
		d, n, err := m3.FirstChunk(b[p:])
		if err != nil || n == 0 {
			return nil, fmt.Errorf("graph: truncated at %d", p)
		}
		p += n
		return d, nil
	}
	for _, want := range []byte{0x02, 0x01, 0x00} {
		d, err := read()
		if err != nil {
			return nil, err
		}
		if len(d) != 1 || d[0] != want {
			return nil, fmt.Errorf("graph: bad header near %d", p)
		}
	}
	for {
		if p >= len(b) {
			return nil, fmt.Errorf("graph: missing trailer")
		}
		if b[p] == 0x03 { // trailer begins
			g.Trailer = append([]byte(nil), b[p:]...)
			return g, nil
		}
		body, err := read()
		if err != nil {
			return nil, err
		}
		obj, err := parseConfigObject(body)
		if err != nil {
			return nil, err
		}
		g.Objects = append(g.Objects, obj)
		if p >= len(b) {
			return nil, fmt.Errorf("graph: missing trailer")
		}
		if b[p] == 0x02 { // index delimiter: 02 00 00 01 <idx>
			want := cfgDelim(len(g.Objects) - 1)
			if p+len(want) > len(b) || !equalBytes(b[p:p+len(want)], want) {
				return nil, fmt.Errorf("graph: bad delimiter before object %d", len(g.Objects))
			}
			p += len(want)
		}
	}
}

// parseConfigObject splits an object body into its Entry and Value fields on the
// ClassConfig / ClassConfigValue markers.
func parseConfigObject(body []byte) (ConfigObject, error) {
	var o ConfigObject
	p := 0
	read := func() ([]byte, bool) {
		d, n, err := m3.FirstChunk(body[p:])
		if err != nil || n == 0 {
			return nil, false
		}
		p += n
		return d, true
	}
	first, ok := read()
	if !ok || len(first) != 4 || binary.LittleEndian.Uint32(first) != ClassConfig {
		return o, fmt.Errorf("graph: object does not start with ClassConfig")
	}
	inValue := false
	for p < len(body) {
		d, ok := read()
		if !ok {
			return o, fmt.Errorf("graph: truncated object body")
		}
		if len(d) == 4 && binary.LittleEndian.Uint32(d) == ClassConfigValue {
			inValue = true
			continue
		}
		c := Chunk(append([]byte(nil), d...))
		if inValue {
			o.Value = append(o.Value, c)
		} else {
			o.Entry = append(o.Entry, c)
		}
	}
	if !inValue {
		return o, fmt.Errorf("graph: object has no ClassConfigValue")
	}
	return o, nil
}

// cfgDelim is the 5-chunk delimiter 02 00 00 01 <idx> that precedes object idx+1.
func cfgDelim(idx int) []byte {
	return cat(chunk(0x02), chunk(0x00), chunk(0x00), chunk(0x01), chunk(byte(idx)))
}

// chunk encodes data as one m3 chunk.
func chunk(data ...byte) []byte { return m3.AppendChunk(nil, data) }

// classChunk encodes a 4-byte little-endian class id as its own chunk.
func classChunk(class uint32) []byte {
	var w [4]byte
	binary.LittleEndian.PutUint32(w[:], class)
	return chunk(w[:]...)
}

// cat concatenates byte slices.
func cat(parts ...[]byte) []byte {
	var out []byte
	for _, p := range parts {
		out = append(out, p...)
	}
	return out
}

// minLE renders v as minimal little-endian bytes (always at least one byte).
func minLE(v uint64) []byte {
	out := []byte{byte(v)}
	for v >>= 8; v > 0; v >>= 8 {
		out = append(out, byte(v))
	}
	return out
}

// leU64 reads up to 8 bytes as a little-endian integer.
func leU64(b []byte) uint64 {
	var v uint64
	for i := 0; i < len(b) && i < 8; i++ {
		v |= uint64(b[i]) << (8 * i)
	}
	return v
}

func equalBytes(a, b []byte) bool {
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
