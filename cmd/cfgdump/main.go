// cfgdump decodes a CMR config object graph (the id30 write payload) as a linear
// m3 chunk stream: for each chunk it prints the offset, intro kind (inline / run
// / length-prefixed), raw bytes, little-endian integer and ASCII, labelling the
// known class ids. Objects are themselves chunk streams, so pass an object body
// to see its fields. This is the inverse view of pkg/cmr.BuildVLANConfig and the
// tool for diffing two captured graphs field by field.
//
//	cfgdump <hexfile|hexstring>
package main

import (
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"os"
	"strings"

	"github.com/athenanetworks/cmr-server/pkg/cmr"
	"github.com/athenanetworks/cmr-server/pkg/m3"
)

func asc(b []byte) string {
	var s strings.Builder
	for _, c := range b {
		if c >= 0x20 && c < 0x7f {
			s.WriteByte(c)
		} else {
			s.WriteByte('.')
		}
	}
	return s.String()
}

func le(b []byte) string {
	if len(b) == 0 || len(b) > 8 {
		return ""
	}
	var w [8]byte
	copy(w[:], b)
	return fmt.Sprintf("=%d", binary.LittleEndian.Uint64(w[:]))
}

func classLabel(b []byte) string {
	if len(b) != 4 {
		return ""
	}
	switch binary.LittleEndian.Uint32(b) {
	case cmr.ClassConfig:
		return " <CLASS cfg-entry>"
	case cmr.ClassConfigValue:
		return " <CLASS cfg-value>"
	case cmr.ClassState:
		return " <CLASS state>"
	}
	return ""
}

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintln(os.Stderr, "usage: cfgdump <hexfile|hexstring>")
		os.Exit(2)
	}
	src := os.Args[1]
	if raw, err := os.ReadFile(src); err == nil {
		src = string(raw)
	}
	b, err := hex.DecodeString(strings.TrimSpace(src))
	if err != nil {
		fmt.Fprintln(os.Stderr, "bad hex:", err)
		os.Exit(1)
	}
	fmt.Printf("# %d bytes\n", len(b))
	for p := 0; p < len(b); {
		intro := b[p]
		data, n, err := m3.FirstChunk(b[p:])
		if err != nil || n == 0 {
			fmt.Printf("@%-4d ?%02x (stop)\n", p, intro)
			break
		}
		var kind string
		switch {
		case intro <= 0xcf:
			kind = "INLINE"
		case intro <= 0xef:
			kind = fmt.Sprintf("RUN%d", intro-0xd0)
		default:
			kind = fmt.Sprintf("LEN(%d)", len(data))
		}
		fmt.Printf("@%-4d %-9s %-28s %-10s |%s|%s\n",
			p, kind, hex.EncodeToString(data), le(data), asc(data), classLabel(data))
		p += n
	}
}
