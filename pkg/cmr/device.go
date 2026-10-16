// Package cmr models the CMR control/management plane: the structured objects a
// controller reads from (and later writes to) a paired MikroTik device over the
// post-TLS nv/m3 channel. This is the "reading" layer built on top of pkg/m3.
package cmr

import (
	"encoding/hex"

	"github.com/athenanetworks/cmr-server/pkg/m3"
)

// Descriptor field ids inside the device's id0 self-report object, reversed from
// a live open-pairing capture (docs/OPEN_PAIRING.md, frame #10).
const (
	fBuild = 65 // build-time + version (8 bytes, matches the version frame's id25)
	fArch  = 66 // CPU architecture string, e.g. "x86"
	fIdent = 67 // device identity / system name, e.g. "CHR"
	fOS    = 69 // OS string, e.g. "routeros" (wire type 2: value is one chunk)
	fBoard = 76 // board/model string, e.g. "CHR"
	fVer   = 75 // packed version word (3 bytes)
	fCaps  = 79 // capability/version word (4 bytes)
	fBlob  = 80 // radio/capability descriptor blob (opaque for now)
	fHdl   = 21 // object handle (8 bytes) the controller addresses this object by
)

// DeviceInfo is the identity and capability summary a device reports about
// itself. Fields are best-effort: absent ones are left zero.
type DeviceInfo struct {
	Identity string // id67 — system name
	Board    string // id76 — board/model
	Arch     string // id66 — CPU arch
	OS       string // id69 — operating system
	Build    []byte // id65 — build-time+version bytes
	Caps     []byte // id79 — capability word
	Blob     []byte // id80 — raw radio/capability descriptor (not yet decoded)
	Handle   []byte // id21 — object handle
}

// IsDescriptor reports whether a decoded id0 object looks like a device
// descriptor (as opposed to an error object or a config row): it carries the
// build/arch identity fields rather than a status+message.
func IsDescriptor(obj *m3.Message) bool {
	if obj == nil {
		return false
	}
	_, hasBuild := obj.GetBytes(fBuild)
	_, hasArch := obj.GetBytes(fArch)
	return hasBuild && hasArch
}

// ParseDescriptor decodes the inner body of a device's id0 descriptor object
// (the value of the top-level id0 field) into a DeviceInfo.
func ParseDescriptor(body []byte) (DeviceInfo, error) {
	msg, err := m3.Decode(body)
	if err != nil {
		return DeviceInfo{}, err
	}
	var d DeviceInfo
	for _, f := range msg.Fields {
		switch f.ID {
		case fIdent:
			d.Identity = string(f.Value)
		case fBoard:
			d.Board = string(f.Value)
		case fArch:
			d.Arch = string(f.Value)
		case fOS:
			d.OS = string(stringValue(f))
		case fBuild:
			d.Build = f.Value
		case fCaps:
			d.Caps = f.Value
		case fBlob:
			d.Blob = f.Value
		case fHdl:
			d.Handle = f.Value
		}
	}
	return d, nil
}

// stringValue returns a field's string value, unwrapping the single length-chunk
// that wire-type-2 fields wrap their value in (e.g. the os field "routeros").
func stringValue(f m3.Field) []byte {
	if f.Type == 2 {
		if inner, n, err := m3.FirstChunk(f.Value); err == nil && n == len(f.Value) {
			return inner
		}
	}
	return f.Value
}

// String renders a one-line summary for logs.
func (d DeviceInfo) String() string {
	s := "device"
	if d.Identity != "" {
		s += " ident=" + d.Identity
	}
	if d.Board != "" {
		s += " board=" + d.Board
	}
	if d.Arch != "" {
		s += " arch=" + d.Arch
	}
	if d.OS != "" {
		s += " os=" + d.OS
	}
	if len(d.Build) > 0 {
		s += " build=" + hex.EncodeToString(d.Build)
	}
	return s
}
