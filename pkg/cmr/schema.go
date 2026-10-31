package cmr

import (
	_ "embed"
	"encoding/json"
	"sync"
)

// schemaJSON is the CMR config attribute dictionary, extracted from the RouterOS
// console schema (.mem) by cmd/cmr-schema and regenerated per firmware version.
// Regenerate with:
//
//	go run ./cmd/cmr-schema -json <cmr.mem> <wifi.mem> > pkg/cmr/schema.json
//
//go:embed schema.json
var schemaJSON []byte

// Attr is one config attribute: its nv id, name and value type. These are the
// attribute ids of the control-plane object model — the ids the RouterOS console
// and the device use to address fields in the nv getall object rows
// (com.mikrotik.cmr / com.mikrotik.cmr.config). They are NOT the positional
// field layout of the id30 config-push graph (see Graph), which is a separate
// encoding under the generic ClassConfig/ClassConfigValue wrapper. Type is one
// of u32/u32ro/flag/string/mac/misc or "unknown" when the descriptor is not yet
// mapped.
type Attr struct {
	ID   uint32 `json:"id"`
	Name string `json:"name"`
	Type string `json:"type"`
}

var (
	schemaOnce   sync.Once
	schemaAttrs  []Attr
	schemaByName map[string]Attr
	schemaByID   map[uint32]Attr
)

func loadSchema() {
	schemaOnce.Do(func() {
		_ = json.Unmarshal(schemaJSON, &schemaAttrs)
		schemaByName = make(map[string]Attr, len(schemaAttrs))
		schemaByID = make(map[uint32]Attr, len(schemaAttrs))
		for _, a := range schemaAttrs {
			if _, ok := schemaByName[a.Name]; !ok {
				schemaByName[a.Name] = a
			}
			if _, ok := schemaByID[a.ID]; !ok {
				schemaByID[a.ID] = a
			}
		}
	})
}

// Attrs returns the full attribute dictionary, ordered by id.
func Attrs() []Attr {
	loadSchema()
	return schemaAttrs
}

// AttrByName returns the attribute with the given name (the first, if a name is
// reused across menus).
func AttrByName(name string) (Attr, bool) {
	loadSchema()
	a, ok := schemaByName[name]
	return a, ok
}

// AttrByID returns the attribute with the given nv id.
func AttrByID(id uint32) (Attr, bool) {
	loadSchema()
	a, ok := schemaByID[id]
	return a, ok
}
