package cmr

import (
	"bytes"
	"strings"
	"testing"
)

func chunksEqual(a, b []Chunk) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if !bytes.Equal(a[i], b[i]) {
			return false
		}
	}
	return true
}

// TestDecodeConfigScopeLossless proves the named model (scaffold_decode.go) is a
// complete, lossless decode of the catalog: reconstructing the key/value chunks
// from the decoded ScopeRow reproduces the catalog bytes exactly. If a future
// regeneration emits a shape the grammar does not model, this fails.
func TestDecodeConfigScopeLossless(t *testing.T) {
	rows := DecodeConfigScope()
	if len(rows) != len(configScopeCatalog) || len(rows) != len(scopeHandles) {
		t.Fatalf("row count mismatch: rows=%d catalog=%d handles=%d", len(rows), len(configScopeCatalog), len(scopeHandles))
	}
	for i, r := range rows {
		o := configScopeCatalog[i]
		if r.Key.Form == KeyUnknown {
			t.Errorf("row %d (handle %#x): key did not match any known form: %v", i, r.Handle, o.key)
			continue
		}
		if got := r.Key.chunks(); !chunksEqual(got, o.key) {
			t.Errorf("row %d (handle %#x): key reconstruct mismatch\n got %x\nwant %x", i, r.Handle, flattenChunks(got), flattenChunks(o.key))
		}
		if got := r.Value.chunks(); !chunksEqual(got, o.value) {
			t.Errorf("row %d (handle %#x): value reconstruct mismatch\n got %x\nwant %x", i, r.Handle, flattenChunks(got), flattenChunks(o.value))
		}
	}
}

// TestConfigScopeGrammar asserts the structural invariants the grammar claims.
func TestConfigScopeGrammar(t *testing.T) {
	rows := DecodeConfigScope()
	forms := map[KeyForm]int{}
	queries := map[string]int{}
	var enveloped, rawVal, present int
	for i, r := range rows {
		if r.Key.Class != ClassConfig {
			t.Errorf("row %d: key class %#x, want ClassConfig %#x", i, r.Key.Class, ClassConfig)
		}
		forms[r.Key.Form]++
		queries[r.Key.QueryLabel()]++
		if r.Key.Form == KeyNested && r.Key.NestedClass != classScopeNested {
			t.Errorf("row %d: nested class %#x, want %#x", i, r.Key.NestedClass, classScopeNested)
		}
		if r.Value.Present {
			present++
			switch r.Value.Class {
			case ClassScopeObject, ClassScopeObjectB:
			default:
				t.Errorf("row %d: unexpected value class %#x", i, r.Value.Class)
			}
			if r.Value.Enveloped {
				enveloped++
			} else {
				rawVal++
			}
		}
	}
	// Every query label must be one we recognise (no raw-hex fallthrough).
	for q := range queries {
		switch q {
		case "key-only", "v2", "full", "bare":
		default:
			t.Errorf("unrecognised query label %q", q)
		}
	}
	// Known shape of the 7.26 capture: 33 rows, 27 with a value (26 enveloped,
	// 1 raw variant object), the rest key-only.
	if len(rows) != 33 {
		t.Errorf("rows = %d, want 33", len(rows))
	}
	if present != 27 || enveloped != 26 || rawVal != 1 {
		t.Errorf("value shape: present=%d enveloped=%d raw=%d, want 27/26/1", present, enveloped, rawVal)
	}
	if forms[KeyBare] != 1 || forms[KeyNested] != 6 || forms[KeyExt] != 2 {
		t.Errorf("forms: bare=%d nested=%d ext=%d, want 1/6/2", forms[KeyBare], forms[KeyNested], forms[KeyExt])
	}
}

func TestAnnotateConfigScope(t *testing.T) {
	out := AnnotateConfigScope()
	if n := strings.Count(out, "\n"); n != len(configScopeCatalog) {
		t.Errorf("annotate produced %d lines, want %d", n, len(configScopeCatalog))
	}
	// Spot-check a known row renders with named parts.
	if !strings.Contains(out, "nested attr=0x000014") {
		t.Errorf("annotate missing nested vlan-scope row; got:\n%s", out)
	}
}
