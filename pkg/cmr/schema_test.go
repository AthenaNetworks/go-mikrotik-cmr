package cmr

import "testing"

func TestSchemaLoads(t *testing.T) {
	attrs := Attrs()
	if len(attrs) < 400 {
		t.Fatalf("schema has %d attrs, want >= 400", len(attrs))
	}
	// Ordered by id, non-decreasing.
	for i := 1; i < len(attrs); i++ {
		if attrs[i].ID < attrs[i-1].ID {
			t.Fatalf("schema not ordered by id at %d: %#x < %#x", i, attrs[i].ID, attrs[i-1].ID)
		}
	}
}

func TestSchemaLookups(t *testing.T) {
	cases := []struct {
		name string
		id   uint32
		typ  string
	}{
		{"vlan-id", 0xd701, "unknown"},
		{"band", 0x2c02, "u32"},
		{"ssid", 0x101b, "unknown"},
	}
	for _, c := range cases {
		a, ok := AttrByName(c.name)
		if !ok {
			t.Errorf("AttrByName(%q) not found", c.name)
			continue
		}
		if a.ID != c.id {
			t.Errorf("%s id = %#x, want %#x", c.name, a.ID, c.id)
		}
		if a.Type != c.typ {
			t.Errorf("%s type = %q, want %q", c.name, a.Type, c.typ)
		}
		byID, ok := AttrByID(c.id)
		if !ok || byID.Name != c.name {
			t.Errorf("AttrByID(%#x) = %+v, want name %q", c.id, byID, c.name)
		}
	}
}

func TestSchemaUnknownLookup(t *testing.T) {
	if _, ok := AttrByName("definitely-not-an-attribute"); ok {
		t.Error("AttrByName returned ok for a missing name")
	}
	if _, ok := AttrByID(0xffffffff); ok {
		t.Error("AttrByID returned ok for a missing id")
	}
}
