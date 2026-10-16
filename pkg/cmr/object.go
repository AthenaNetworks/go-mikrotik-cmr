package cmr

import (
	"encoding/binary"
	"fmt"
	"strings"

	"github.com/athenanetworks/cmr-server/pkg/m3"
)

// Object is a decoded CMR management object (the body of an id9/id1 object
// field: its id74 key or id75 value). Objects use a flat, positional sequence
// of m3 chunks whose first chunk is a class id; the per-position semantics are
// schema-defined (not yet fully reversed), so this decoder exposes the generic
// structure plus the human-legible content (strings and MAC addresses) that the
// device reports about itself.
type Object struct {
	Class   uint32   // first chunk as a little-endian u32 (object class)
	Chunks  [][]byte // every chunk in order
	Strings []string // chunks that are entirely printable ASCII (>=2 bytes)
	MACs    []string // 6-byte chunks, rendered as MAC addresses (candidates)
}

// ObjectClass values seen live. aa00803f carries runtime state (interfaces,
// system identity); 4a00c03f carries config/relational objects.
const (
	ClassState  uint32 = 0x3f8000aa // "aa00803f" little-endian
	ClassConfig uint32 = 0x3fc0004a // "4a00c03f" little-endian
)

// ParseObject decodes an object body (an id74/id75 blob) into its chunk
// structure and the legible content within it. Object bodies are a tree:
// value chunks (inline/run) are leaves, while length-prefixed chunks (0xfa..fe)
// are containers whose bytes are themselves a chunk sequence, so content is
// collected recursively.
func ParseObject(body []byte) Object {
	var o Object
	b := body
	first := true
	for len(b) > 0 {
		c, n, container := readChunk(b)
		if n == 0 {
			break
		}
		b = b[n:]
		if first {
			if len(c) >= 4 {
				o.Class = binary.LittleEndian.Uint32(c[:4])
			}
			first = false
			o.Chunks = append(o.Chunks, c)
			continue // the class chunk is not content
		}
		o.Chunks = append(o.Chunks, c)
		o.collect(c, container, 0)
	}
	return o
}

// collect gathers legible content from a chunk. A printable chunk is a leaf
// string and a 6-byte chunk a MAC candidate (checked first, since a long string
// is length-prefixed just like a container); any other length-prefixed chunk is
// a container whose bytes are a nested chunk sequence and is walked recursively.
func (o *Object) collect(c []byte, container bool, depth int) {
	if len(c) >= 2 && printable(c) {
		o.Strings = append(o.Strings, string(c))
		return
	}
	if len(c) == 6 {
		o.MACs = append(o.MACs, mac(c))
		return
	}
	if container && depth < 6 {
		b := c
		for len(b) > 0 {
			sub, n, subContainer := readChunk(b)
			if n == 0 {
				break
			}
			b = b[n:]
			o.collect(sub, subContainer, depth+1)
		}
	}
}

// readChunk reads one m3 chunk, returning its data, the bytes consumed, and
// whether it was a length-prefixed container (0xfa..0xfe) rather than a value.
func readChunk(b []byte) (data []byte, n int, container bool) {
	c, nn, err := m3.FirstChunk(b)
	if err != nil || nn == 0 {
		return nil, 0, false
	}
	// A length-prefixed chunk consumes more than one framing byte: intro 0xfe
	// uses 2 (intro+len8), 0xfd 3, etc.; inline uses 1 and run uses 1.
	intro := b[0]
	return c, nn, intro >= 0xfa && intro <= 0xfe
}

func printable(b []byte) bool {
	for _, c := range b {
		if c < 0x20 || c > 0x7e {
			return false
		}
	}
	return true
}

func mac(b []byte) string {
	p := make([]string, len(b))
	for i, c := range b {
		p[i] = fmt.Sprintf("%02x", c)
	}
	return strings.Join(p, ":")
}

// ClassName returns a short label for a known object class.
func ClassName(c uint32) string {
	switch c {
	case ClassState:
		return "state"
	case ClassConfig:
		return "config"
	default:
		return fmt.Sprintf("%#x", c)
	}
}

// String renders a one-line summary for logs.
func (o Object) String() string {
	s := fmt.Sprintf("obj class=%s chunks=%d", ClassName(o.Class), len(o.Chunks))
	if len(o.Strings) > 0 {
		s += " strings=" + strings.Join(o.Strings, ",")
	}
	if len(o.MACs) > 0 {
		s += " macs=" + strings.Join(o.MACs, ",")
	}
	return s
}
