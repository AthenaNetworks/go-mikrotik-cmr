// cfgdump decodes a CMR config object graph (the id30 write payload). By default
// it prints the linear m3 chunk stream: for each chunk the offset, intro kind
// (inline / run / length-prefixed), raw bytes, little-endian integer and ASCII,
// labelling the known class ids. With -tree it decodes the graph structurally via
// pkg/cmr.DecodeGraph — per object, the Entry and Value fields — and, when the
// graph is a vlan config, prints the ParseVLANConfig result. It is the inverse
// view of pkg/cmr.BuildVLANConfig and the tool for diffing captured graphs.
//
//	cfgdump [-tree] <hexfile|hexstring>
package main

import (
	"encoding/binary"
	"encoding/hex"
	"flag"
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
	tree := flag.Bool("tree", false, "decode structurally (objects and fields) via pkg/cmr.DecodeGraph")
	flag.Parse()
	if flag.NArg() < 1 {
		fmt.Fprintln(os.Stderr, "usage: cfgdump [-tree] <hexfile|hexstring>")
		os.Exit(2)
	}
	src := flag.Arg(0)
	if raw, err := os.ReadFile(src); err == nil {
		src = string(raw)
	}
	b, err := hex.DecodeString(strings.TrimSpace(src))
	if err != nil {
		fmt.Fprintln(os.Stderr, "bad hex:", err)
		os.Exit(1)
	}
	if *tree {
		dumpTree(b)
		return
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

// dumpTree prints the structural decode: each object's Entry and Value fields,
// and the VLANConfig if the graph parses as one.
func dumpTree(b []byte) {
	g, err := cmr.DecodeGraph(b)
	if err != nil {
		fmt.Fprintln(os.Stderr, "decode:", err)
		os.Exit(1)
	}
	fmt.Printf("# graph: %d objects, trailer %x\n", len(g.Objects), g.Trailer)
	for i, o := range g.Objects {
		fmt.Printf("object %d:\n", i)
		for j, c := range o.Entry {
			fmt.Printf("  entry[%d] %-20s %s %s\n", j, hex.EncodeToString(c), le(c), field(c))
		}
		for j, c := range o.Value {
			fmt.Printf("  value[%d] %-20s %s %s\n", j, hex.EncodeToString(c), le(c), field(c))
		}
	}
	if v, ok := cmr.ParseVLANConfig(b); ok {
		fmt.Printf("# VLANConfig: vlan-id=%d bridge=%q\n", v.VLANID, v.BridgeName)
	}
	if rt := g.Encode(); !strings.EqualFold(hex.EncodeToString(rt), hex.EncodeToString(b)) {
		fmt.Fprintln(os.Stderr, "WARNING: re-encode does not match input")
	}
}

// field annotates a chunk with its ASCII rendering when it is a printable string
// (two or more printable bytes, e.g. a bridge name), not a short numeric value.
func field(c []byte) string {
	if len(c) < 2 {
		return ""
	}
	a := asc(c)
	if !strings.Contains(a, ".") {
		return "|" + a + "|"
	}
	return ""
}
