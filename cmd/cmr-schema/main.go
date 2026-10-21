// cmr-schema extracts the CMR config dictionary (nv attribute id -> name -> type)
// from a RouterOS compiled console schema (.mem), e.g. the cmr package's
// nova/lib/console/*.mem. The schema is a flat array of property records:
//
//	[type-descriptor ptr : 4][kind = 0x00040008 : 4][nv id : 4][name]
//
// where name is either an inline NUL-terminated string or a 0x50800000-based
// pointer into the string pool. The nv ids are the same ids the cmr/libumsg code
// addresses with message::get/insert<nv::TYPE_id>(msg, id), so this dictionary
// decodes and authors config object graphs. It re-runs on any firmware's .mem,
// so the dictionary grows with new RouterOS versions.
//
// Static analysis of a shipped data file; the .mem is never executed.
//
//	cmr-schema <schema.mem>            # human-readable table
//	cmr-schema -json <schema.mem>      # JSON array of {id,name,type}
package main

import (
	"encoding/binary"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"sort"
)

const (
	memBase  = 0x50800000 // the schema is mapped here; pointers are absolute
	kindWord = 0x00040008 // property-record marker preceding (id, name)
)

// typeName maps the observed type-descriptor pointers to a best-effort label.
// These are the distinct value types the console schema uses; labels are refined
// by cross-referencing the typed accessors in the cmr/libumsg code.
var typeName = map[uint32]string{
	0x50820e50: "u32",
	0x50820498: "flag",
	0x5082064c: "string",
	0x508204e4: "u32ro",
	0x50820c34: "mac",
	0x508210e8: "misc",
}

type entry struct {
	ID   uint32 `json:"id"`
	Name string `json:"name"`
	Type string `json:"type"`
	raw  uint32
}

// inlineStr reads a short printable NUL-terminated ASCII string at off.
func inlineStr(b []byte, off int) string {
	if off < 0 || off >= len(b) {
		return ""
	}
	end := off
	for end < len(b) && b[end] != 0 && end-off <= 48 {
		end++
	}
	if end >= len(b) || b[end] != 0 || end == off {
		return ""
	}
	s := b[off:end]
	hasAlpha := false
	for _, c := range s {
		switch {
		case c >= 'a' && c <= 'z', c >= 'A' && c <= 'Z':
			hasAlpha = true
		case c >= '0' && c <= '9', c == '-', c == '.', c == '_':
			// identifier char
		default:
			return "" // not a bare param name (spaces, punctuation = help text)
		}
	}
	if !hasAlpha {
		return ""
	}
	return string(s)
}

// name resolves the name after an id: inline, else a string-pool pointer.
func name(b []byte, at int) string {
	if s := inlineStr(b, at); s != "" {
		return s
	}
	if at+4 <= len(b) {
		p := binary.LittleEndian.Uint32(b[at : at+4])
		if p >= memBase && p < memBase+uint32(len(b)) {
			return inlineStr(b, int(p-memBase))
		}
	}
	return ""
}

// plausibleID screens for an nv attribute id: a 2-byte id (0x200..0xffff) or a
// framework id in the 0xfe00xx/0xff00xx space. Filters out small console ids.
func plausibleID(v uint32) bool {
	if v == 0 || v >= 0x01000000 {
		return false
	}
	if hi := v >> 16; hi == 0xfe || hi == 0xff {
		lo := v & 0xffff
		return lo >= 0x100 && lo < 0x10000
	}
	return v >= 0x200 && v < 0x10000
}

func u32(b []byte, i int) uint32 {
	if i < 0 || i+4 > len(b) {
		return 0
	}
	return binary.LittleEndian.Uint32(b[i:])
}

// poolStr resolves a pointer that may point into the string pool.
func poolStr(b []byte, p uint32) string {
	if p >= memBase && p < memBase+uint32(len(b)) {
		return inlineStr(b, int(p-memBase))
	}
	return ""
}

func parse(b []byte) []entry {
	seen := map[[2]uint64]entry{}
	add := func(id uint32, nm string, tptr uint32) {
		if nm == "" {
			return
		}
		t := typeName[tptr]
		if t == "" {
			t = "unknown"
		}
		key := [2]uint64{uint64(id), hash(nm)}
		if _, ok := seen[key]; !ok {
			seen[key] = entry{ID: id, Name: nm, Type: t, raw: tptr}
		}
	}
	for i := 0; i+16 <= len(b); i += 4 {
		// format 1 (property/enum record): [type-ptr][0x00040008][id][name]
		if u32(b, i) == kindWord {
			if id := u32(b, i+4); plausibleID(id) {
				add(id, name(b, i+8), u32(b, i-4))
			}
		}
		// format 2 (menu-arg record): name-ptr at id-12, a pointer right after id
		if id := u32(b, i); plausibleID(id) && i >= 12 {
			if nm := poolStr(b, u32(b, i-12)); nm != "" && (u32(b, i+4)>>24) == 0x50 {
				add(id, nm, 0)
			}
		}
	}
	out := make([]entry, 0, len(seen))
	for _, e := range seen {
		out = append(out, e)
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i].ID != out[j].ID {
			return out[i].ID < out[j].ID
		}
		return out[i].Name < out[j].Name
	})
	return out
}

func hash(s string) uint64 {
	var h uint64 = 1469598103934665603
	for i := 0; i < len(s); i++ {
		h ^= uint64(s[i])
		h *= 1099511628211
	}
	return h
}

func main() {
	asJSON := flag.Bool("json", false, "emit JSON")
	flag.Parse()
	if flag.NArg() < 1 {
		fmt.Fprintln(os.Stderr, "usage: cmr-schema [-json] <schema.mem> [schema.mem ...]")
		os.Exit(2)
	}
	seen := map[[2]uint64]bool{}
	var entries []entry
	for _, path := range flag.Args() {
		b, err := os.ReadFile(path)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		for _, e := range parse(b) {
			key := [2]uint64{uint64(e.ID), hash(e.Name)}
			if !seen[key] {
				seen[key] = true
				entries = append(entries, e)
			}
		}
	}
	sort.Slice(entries, func(i, j int) bool {
		if entries[i].ID != entries[j].ID {
			return entries[i].ID < entries[j].ID
		}
		return entries[i].Name < entries[j].Name
	})
	if *asJSON {
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		enc.Encode(entries)
		return
	}
	fmt.Printf("# %d config attributes (nv id -> name [type]) from %d file(s)\n", len(entries), flag.NArg())
	for _, e := range entries {
		fmt.Printf("  %#08x  %-30s %s\n", e.ID, e.Name, e.Type)
	}
}
