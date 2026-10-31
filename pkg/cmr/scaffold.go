package cmr

//go:generate sh -c "go run ../../cmd/cmr-scaffold ../../docs/samples/controller-config.hex | gofmt > scaffold_catalog.go"

import "github.com/athenanetworks/cmr-server/pkg/m3"

// This file synthesises the controller's own scaffold objects — the id9 rows a
// RouterOS controller streams to a freshly paired device so the device begins,
// and then completes, its state dump:
//
//   - the ROOTS: two self-advertisement rows (one bare, one carrying the board
//     name) sent riding the id10 SUCCESS frame. They make the device start its
//     dump.
//   - the CONFIG SCOPE burst: 33 id9 objects describing the controller's config
//     scope (com.mikrotik.cmr), streamed in answer to the device's getall. The
//     device needs them to reciprocate with its own full state.
//
// Both were previously replayed from captured hex blobs (docs/samples). Here they
// are rebuilt from structure instead, the same way BuildVLANConfig authors the
// config-write graph: the framing, field wire-types, index delimiters, flag and
// the generated handle sequence are explicit code, and each object's decoded
// key/value content is a chunk list (scaffold_catalog.go). EncodeConfigScope and
// EncodeRoots reproduce the captured blobs byte-for-byte (see scaffold_test.go),
// so the synthesis is proven against the bytes a real controller sent and the
// server no longer depends on the captured files at runtime.
//
// What is NOT yet reversed: the per-object semantics. The keys are an object
// class (ClassConfig) + a 2-byte attribute id + a marker; the values are lists of
// 0x07f0-family type references (RefChunk) and inline values. These are carried
// faithfully but mostly unnamed — naming them is the next RE step and would let
// the catalog be generated from a high-level controller config rather than a
// captured snapshot.

// DefaultControllerBoard is the board string the controller advertises in its
// roots (a real RouterOS CHR controller reports "CHR").
const DefaultControllerBoard = "CHR"

// Controller-side m3 field ids used in the scaffold id9 rows.
const (
	scKey    = 74 // id74 object key (wire type 1)
	scValue  = 75 // id75 object value (wire type 2)
	scFlag   = 29 // id29 query flag — the device dictates it; rows must echo it
	scIndex  = 64 // id64 object index
	scHandle = 21 // id21 object handle
	scRow    = 9  // id9 controller->device object row
)

// valueWireType is the non-default m3 wire type the value field (id75) carries;
// every other scaffold field uses the default type. Reproducing it is required
// for byte-identity with a real controller's stream.
const valueWireType = 2

// RefChunk encodes a 4-byte little-endian type reference (the 0x07f0-family ids
// that fill the config-scope object values). It is U32-like but always four bytes
// so a reference such as 0x07f00015 round-trips exactly.
func RefChunk(v uint32) Chunk {
	return Chunk{byte(v), byte(v >> 8), byte(v >> 16), byte(v >> 24)}
}

// scopeObject is one config-scope id9 row before its envelope (flag, handle) is
// applied: the decoded key/value chunk lists and the object index. The handle is
// assigned positionally by the encoder from scopeHandles.
type scopeObject struct {
	key   []Chunk
	value []Chunk // nil when the row carries no id75 value
	index byte
}

// flattenChunks reframes a chunk list back into a field value (each chunk as one
// m3 chunk, concatenated) — the inverse of the chunk split used to build it.
func flattenChunks(cs []Chunk) []byte {
	var b []byte
	for _, c := range cs {
		b = m3.AppendChunk(b, c)
	}
	return b
}

// scopeHandleBytes expands a handle counter into the 8-byte id21 handle a real
// controller uses: <counter> 49 04 00 00 00 00 04.
func scopeHandleBytes(counter byte) []byte {
	return []byte{counter, 0x49, 0x04, 0x00, 0x00, 0x00, 0x00, 0x04}
}

// encode renders one scope row's body with the given query flag and handle.
func (o scopeObject) encode(flag, handle byte) []byte {
	m := m3.New().PutBytes(scKey, flattenChunks(o.key))
	if o.value != nil {
		m.PutType(scValue, valueWireType, flattenChunks(o.value))
	}
	m.PutBytes(scFlag, []byte{flag}).
		PutBytes(scIndex, []byte{o.index}).
		PutBytes(scHandle, scopeHandleBytes(handle))
	return m.Encode()
}

// EncodeConfigScope builds the controller's config-scope object burst — the 33
// id9 rows streamed in answer to the device's getall for com.mikrotik.cmr — with
// every row's query flag set to flag (the value the device's query dictates).
// With flag 0x10 it reproduces docs/samples/controller-config.hex byte-for-byte.
func EncodeConfigScope(flag byte) []byte {
	m := m3.New()
	for i, o := range configScopeCatalog {
		m.PutBytes(scRow, o.encode(flag, scopeHandles[i]))
	}
	return m.Encode()
}

// EncodeRoots builds the controller's root self-advertisement rows — two id9 rows
// (a bare one and one carrying board) that ride the SUCCESS frame and trigger the
// device's state dump — with the given query flag and board name. With flag 0x10
// and board "CHR" it reproduces docs/samples/controller-roots.hex byte-for-byte.
func EncodeRoots(flag byte, board string) []byte {
	if board == "" {
		board = DefaultControllerBoard
	}
	r0 := m3.New().
		PutBytes(scFlag, []byte{flag}).
		PutBytes(scIndex, []byte{0x01}).
		PutBytes(scHandle, scopeHandleBytes(0x24))
	r1 := m3.New().
		PutBytes(76, []byte(board)).
		PutBytes(scFlag, []byte{flag}).
		PutBytes(scIndex, []byte{0x05}).
		PutBytes(scHandle, scopeHandleBytes(0x25))
	return m3.New().PutBytes(scRow, r0.Encode()).PutBytes(scRow, r1.Encode()).Encode()
}
