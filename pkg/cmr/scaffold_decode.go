package cmr

import (
	"fmt"
	"strings"
)

// This file names the structure of the controller config-scope objects that
// scaffold.go emits — it turns the opaque chunk catalog (scaffold_catalog.go)
// into a self-describing model. The grammar below is fully reversed from the
// capture; what remains numeric are the leaf ids (see "Leaf ids" note), which
// are a runtime-assigned nv type/column space with no static names available.
//
// Each config-scope row has a KEY and an optional VALUE.
//
// KEY grammar:
//
//	ClassConfig · attr(u16) · <descriptor> · <query>
//
// where <descriptor> selects one of four forms and <query> is the trailing
// marker that selects which getall variant the row answers:
//
//	simple : 0503                                           (most attributes)
//	nested : 490040 · sub(u8) · class 0x3f800309 · ext · 85fd7f   (attr 0x0014)
//	ext    : 8dfa7f · 07 · 85fd7f                           (attr 0x048a)
//	bare   : (none — the row is just ClassConfig · attr)    (attr 0x0a8a)
//
// VALUE grammar (a type signature for the object's columns):
//
//	class(0x07f00015) · tHead · <middle> · tTail0 · tTail1 · tTail2
//
// 26 of the 27 value-bearing rows carry exactly this envelope; <middle> is the
// per-object column list, a mix of type references (0x07f0-family) and inline
// column-id references. One row (attr 0x0a8a) is a different object class
// (0x07f00009) and carries no envelope.
//
// Leaf ids: the 2-byte attribute id in the key, the inline column-id references
// in the value middle, and the 0x07f0-family type handles are a runtime-assigned
// nv id space. Prior Ghidra RE confirmed the class/column ids are assigned by the
// nv base class at startup and are NOT stored as constants in cmr/libumsg/.mem,
// and no capture carries a name string beside them. They are therefore stable by
// construction (which is why byte-identical replay works) but cannot be mapped to
// human names (vlan-id, band, …) from static material alone — that needs a live
// getall that pairs each id with its name. See docs/re/CMR_CONFIG_SCHEMA.md.

// Config-scope object classes (the value-object wrapper classes) and the nested
// key class. Like ClassConfig these are runtime-assigned nv class ids, named here
// by the role they play in the scaffold.
const (
	// ClassScopeObject wraps a standard config-scope value (its type signature).
	ClassScopeObject uint32 = 0x07f00015
	// ClassScopeObjectB wraps the one variant object (attr 0x0a8a) and also
	// appears as a member type inside value middles.
	ClassScopeObjectB uint32 = 0x07f00009
	// classScopeNested is the inner class embedded in the nested key form.
	classScopeNested uint32 = 0x3f800309
)

// Type references seen in the config-scope value signatures. These are nv
// type-table handles (runtime-assigned, not static symbols), named here by the
// structural position they hold in the capture. The same handle can appear in
// more than one position, so these names describe the dominant role, not an
// intrinsic meaning; the precise nv type each denotes is not yet reversed.
const (
	tHead  uint32 = 0x07f80101 // envelope head, immediately after the class
	tTail0 uint32 = 0x07f80109 // envelope tail, before the closing pair
	tTail1 uint32 = 0x07f00019 // envelope closing pair (first)
	tTail2 uint32 = 0x07f000ad // envelope closing pair (second)
	// middle type references (role not yet reversed)
	tMid4c  uint32 = 0x07f0004c
	tMid50  uint32 = 0x07f00050
	tMid84  uint32 = 0x07f00084
	tMidE4  uint32 = 0x07f000e4
	tMid125 uint32 = 0x07f00125
	tMid7f9 uint32 = 0x07f7f901
)

// Descriptor tokens in the key, between the attribute id and the query marker.
// Their exact meaning (a path / type / column hint) is not yet reversed; they
// are recognised here to classify the key form.
var (
	descSimple      = []byte{0x05, 0x03}       // simple form
	descNestedIntro = []byte{0x49, 0x00, 0x40} // nested form, before sub + class
	descNestedMid   = []byte{0x85, 0xfd, 0x7f} // nested/ext form, before query
	descExtIntro    = []byte{0x8d, 0xfa, 0x7f} // ext form
)

// KeyForm classifies a config-scope key's descriptor.
type KeyForm int

const (
	KeyUnknown KeyForm = iota
	KeySimple          // attr · 0503 · query
	KeyNested          // attr · 490040 · sub · class · ext · 85fd7f · query
	KeyExt             // attr · 8dfa7f · 07 · 85fd7f · query
	KeyBare            // attr (no descriptor, no query)
)

func (f KeyForm) String() string {
	switch f {
	case KeySimple:
		return "simple"
	case KeyNested:
		return "nested"
	case KeyExt:
		return "ext"
	case KeyBare:
		return "bare"
	default:
		return "unknown"
	}
}

// ScopeKey is the decoded key of a config-scope row.
type ScopeKey struct {
	Class uint32  // object class (ClassConfig)
	Attr  uint16  // attribute id (runtime nv id space — see note above)
	Form  KeyForm // descriptor form
	Query []byte  // trailing query-variant marker (nil for KeyBare)

	// Nested-form extras (zero for other forms):
	NestedSub   byte   // the 1-byte selector after 490040 (0x01/0x44/0x0c)
	NestedClass uint32 // the embedded class (classScopeNested)
	NestedExt   []byte // the chunk after the nested class (07 or 070004)
}

// QueryLabel names the query-variant marker where its effect is known.
func (k ScopeKey) QueryLabel() string {
	switch {
	case k.Form == KeyBare:
		return "bare"
	case len(k.Query) == 1 && k.Query[0] == 0x01:
		return "key-only" // rows with this marker carry no value in the capture
	case len(k.Query) == 1 && k.Query[0] == 0x02:
		return "v2"
	case len(k.Query) == 3 && k.Query[0] == 0x04 && k.Query[1] == 0x00 && k.Query[2] == 0xfe:
		return "full"
	default:
		return fmt.Sprintf("%x", k.Query)
	}
}

// ScopeTerm is one element of a value middle: either a type reference or an
// inline column-id reference.
type ScopeTerm struct {
	Ref   bool   // true: Type is a type reference; false: Bytes is an inline value
	Type  uint32 // the type reference (when Ref)
	Bytes []byte // the raw inline bytes (when !Ref)
}

// ScopeValue is the decoded value (type signature) of a config-scope row.
type ScopeValue struct {
	Present   bool
	Class     uint32      // value-object class (ClassScopeObject or ...B)
	Enveloped bool        // true when the standard tHead/…/tTail envelope is present
	Middle    []ScopeTerm // the per-object column list (envelope stripped)
}

// ScopeRow is a fully decoded config-scope object: its handle, index and the
// decoded key and value.
type ScopeRow struct {
	Handle byte
	Index  byte
	Key    ScopeKey
	Value  ScopeValue
}

// isRef reports whether a chunk is a 0x07-family type reference.
func isRef(c Chunk) bool { return len(c) == 4 && c[3] == 0x07 }

func chunkEq(c Chunk, b []byte) bool {
	if len(c) != len(b) {
		return false
	}
	for i := range b {
		if c[i] != b[i] {
			return false
		}
	}
	return true
}

// decodeScopeKey parses a key chunk list into a ScopeKey.
func decodeScopeKey(key []Chunk) ScopeKey {
	var k ScopeKey
	if len(key) == 0 {
		return k
	}
	k.Class = uint32(key[0].Uint())
	if len(key) >= 2 {
		k.Attr = uint16(key[1].Uint())
	}
	switch {
	case len(key) == 2:
		k.Form = KeyBare
	case len(key) >= 4 && chunkEq(key[2], descSimple):
		k.Form = KeySimple
		k.Query = key[len(key)-1]
	case len(key) == 8 && chunkEq(key[2], descNestedIntro) && chunkEq(key[6], descNestedMid):
		k.Form = KeyNested
		k.NestedSub = byte(key[3].Uint())
		k.NestedClass = uint32(key[4].Uint())
		k.NestedExt = key[5]
		k.Query = key[7]
	case len(key) == 6 && chunkEq(key[2], descExtIntro) && chunkEq(key[4], descNestedMid):
		k.Form = KeyExt
		k.NestedExt = key[3]
		k.Query = key[5]
	default:
		k.Form = KeyUnknown
		k.Query = key[len(key)-1]
	}
	return k
}

// scopeEnvelope is the standard value envelope: after the class chunk comes
// tHead, and the value ends with tTail0, tTail1, tTail2.
func hasEnvelope(v []Chunk) bool {
	if len(v) < 5 {
		return false
	}
	if !isRef(v[1]) || uint32(v[1].Uint()) != tHead {
		return false
	}
	n := len(v)
	return isRef(v[n-3]) && uint32(v[n-3].Uint()) == tTail0 &&
		isRef(v[n-2]) && uint32(v[n-2].Uint()) == tTail1 &&
		isRef(v[n-1]) && uint32(v[n-1].Uint()) == tTail2
}

// decodeScopeValue parses a value chunk list into a ScopeValue.
func decodeScopeValue(val []Chunk) ScopeValue {
	var v ScopeValue
	if len(val) == 0 {
		return v
	}
	v.Present = true
	v.Class = uint32(val[0].Uint())
	var middle []Chunk
	if hasEnvelope(val) {
		v.Enveloped = true
		middle = val[2 : len(val)-3]
	} else {
		middle = val[1:]
	}
	for _, c := range middle {
		if isRef(c) {
			v.Middle = append(v.Middle, ScopeTerm{Ref: true, Type: uint32(c.Uint())})
		} else {
			v.Middle = append(v.Middle, ScopeTerm{Bytes: append([]byte(nil), c...)})
		}
	}
	return v
}

// DecodeConfigScope returns the embedded config-scope catalog as decoded rows.
// It is the structured, named view of what EncodeConfigScope emits.
func DecodeConfigScope() []ScopeRow {
	rows := make([]ScopeRow, len(configScopeCatalog))
	for i, o := range configScopeCatalog {
		rows[i] = ScopeRow{
			Handle: scopeHandles[i],
			Index:  o.index,
			Key:    decodeScopeKey(o.key),
			Value:  decodeScopeValue(o.value),
		}
	}
	return rows
}

// u16chunk encodes an attribute id as its 2-byte little-endian chunk.
func u16chunk(v uint16) Chunk { return Chunk{byte(v), byte(v >> 8)} }

// chunks reconstructs the key's chunk list from the decoded model. It is the
// inverse of decodeScopeKey for every form the catalog uses (KeyUnknown, which
// never occurs in the catalog, is not reconstructed), and proves the model is a
// complete, lossless naming of the key bytes (see scaffold_decode_test.go).
func (k ScopeKey) chunks() []Chunk {
	out := []Chunk{RefChunk(k.Class), u16chunk(k.Attr)}
	switch k.Form {
	case KeyBare:
		return out
	case KeySimple:
		out = append(out, Chunk(descSimple), k.Query)
	case KeyNested:
		out = append(out, Chunk(descNestedIntro), Chunk{k.NestedSub}, RefChunk(k.NestedClass), k.NestedExt, Chunk(descNestedMid), k.Query)
	case KeyExt:
		out = append(out, Chunk(descExtIntro), k.NestedExt, Chunk(descNestedMid), k.Query)
	}
	return out
}

// chunks reconstructs the value's chunk list from the decoded model — the
// inverse of decodeScopeValue.
func (v ScopeValue) chunks() []Chunk {
	if !v.Present {
		return nil
	}
	out := []Chunk{RefChunk(v.Class)}
	if v.Enveloped {
		out = append(out, RefChunk(tHead))
	}
	for _, t := range v.Middle {
		if t.Ref {
			out = append(out, RefChunk(t.Type))
		} else {
			out = append(out, Chunk(t.Bytes))
		}
	}
	if v.Enveloped {
		out = append(out, RefChunk(tTail0), RefChunk(tTail1), RefChunk(tTail2))
	}
	return out
}

// typeRefName renders a type reference for annotation.
func typeRefName(t uint32) string {
	switch t {
	case ClassScopeObject:
		return "obj"
	case ClassScopeObjectB:
		return "obj-b"
	case tHead:
		return "head"
	case tTail0:
		return "tail0"
	case tTail1:
		return "tail1"
	case tTail2:
		return "tail2"
	default:
		return fmt.Sprintf("t:%06x", t&0xffffff)
	}
}

// String renders a one-line human-readable summary of the row.
func (r ScopeRow) String() string {
	var b strings.Builder
	fmt.Fprintf(&b, "h=%02x idx=%x %s attr=%#06x q=%s", r.Handle, r.Index, r.Key.Form, r.Key.Attr, r.Key.QueryLabel())
	if r.Key.Form == KeyNested {
		fmt.Fprintf(&b, " sub=%#x", r.Key.NestedSub)
	}
	if r.Value.Present {
		fmt.Fprintf(&b, " | val class=%s", typeRefName(r.Value.Class))
		if r.Value.Enveloped {
			b.WriteString(" env[")
		} else {
			b.WriteString(" raw[")
		}
		var parts []string
		for _, t := range r.Value.Middle {
			if t.Ref {
				parts = append(parts, typeRefName(t.Type))
			} else {
				parts = append(parts, fmt.Sprintf("c:%x", t.Bytes))
			}
		}
		b.WriteString(strings.Join(parts, " "))
		b.WriteString("]")
	}
	return b.String()
}

// AnnotateConfigScope returns a human-readable dump of the whole config-scope
// catalog, one decoded row per line.
func AnnotateConfigScope() string {
	var b strings.Builder
	for _, r := range DecodeConfigScope() {
		b.WriteString(r.String())
		b.WriteByte('\n')
	}
	return b.String()
}
