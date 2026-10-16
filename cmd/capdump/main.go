// Command capdump reconstructs and recursively decodes the nv/m3 frames from a
// cmr-mitm log (read on stdin), annotating each with direction and frame index.
// It is a reverse-engineering aid for the CMR control/management plane.
//
//	cmr-mitm ... > cap.log
//	capdump < cap.log            # full recursive decode
//	capdump -dir C < cap.log     # only client->controller frames
//	capdump -id 0 < cap.log      # only frames whose top-level field is id0
package main

import (
	"bufio"
	"encoding/hex"
	"flag"
	"fmt"
	"os"
	"regexp"
	"strconv"
	"strings"

	"github.com/athenanetworks/cmr-server/pkg/m3"
)

// knownIDs annotates field ids whose meaning we've reversed, for readability.
var knownIDs = map[uint32]string{
	0:  "obj",      // management object / descriptor / error
	1:  "ver/seq",  // version or nonce-seq frame
	2:  "hello",    // hello token
	4:  "query",    // getall query body (scope+topic+selector)
	6:  "disc",     // discovery
	7:  "resync",   // discovery resync
	9:  "devrow",   // device-table row
	10: "status",   // pairing/status (SUCCESS)
	21: "seq",      // session sequence
	22: "nonce",    // per-message nonce
	23: "scope",    // scope string
	24: "board",    // board (version)
	25: "build",    // build-time+version
	27: "st",       // status byte
	28: "msg",      // status message string
	29: "flag",     // device/row flag
	64: "idx",      // row index
	65: "build",    // descriptor: build
	66: "arch",     // descriptor: arch (x86)
	67: "ident",    // descriptor: identity/board
	69: "os",       // descriptor: os (routeros)
	74: "key",      // config row key blob
	75: "val",      // config row value blob
	76: "board",    // device row: board
	79: "caps?",    // descriptor: capability/version word
	80: "blob",     // descriptor: 112B capability blob
}

var frameHdr = regexp.MustCompile(`(client->controller|controller->client) \((\d+) bytes\)`)
var hexRow = regexp.MustCompile(`^\s+([0-9a-fA-F]{2}(?: [0-9a-fA-F]{2})*) *$`)

func ann(id uint32) string {
	if n, ok := knownIDs[id]; ok {
		return "(" + n + ")"
	}
	return ""
}

func ascii(v []byte) string {
	s := ""
	for _, c := range v {
		if c >= 0x20 && c < 0x7f {
			s += string(c)
		} else {
			s += "."
		}
	}
	if len(s) > 48 {
		s = s[:48] + "…"
	}
	return s
}

func dump(b []byte, depth int, maxDepth int) {
	msg, err := m3.Decode(b)
	if err != nil || msg == nil {
		return
	}
	ind := strings.Repeat("  ", depth+1)
	for _, f := range msg.Fields {
		v := f.Value
		fmt.Printf("%sid=%-4d%-9s len=%-4d %x |%s|\n", ind, f.ID, ann(f.ID), len(v), clip(v), ascii(v))
		if len(v) >= 2 && depth < maxDepth {
			if sub, e := m3.Decode(v); e == nil && sub != nil && len(sub.Fields) > 0 && bytesEq(sub.Encode(), v) {
				dump(v, depth+1, maxDepth)
			}
		}
	}
}

func clip(v []byte) []byte {
	if len(v) > 48 {
		return v[:48]
	}
	return v
}

func bytesEq(a, b []byte) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

// unfold reverses the m3 tag encoding given the previous field id (mirror of the
// unexported m3.unfold) so capdump can decode a continuous stream with a running
// base that carries across TLS-write boundaries.
func unfold(folded, base uint32) (uint8, uint32) {
	t := uint8(folded & 3)
	if folded&4 == 0 {
		return t, folded >> 3
	}
	delta := folded >> 4
	if folded&8 != 0 {
		return t, base - delta
	}
	return t, base + delta
}

func leUint(b []byte) uint32 {
	var v uint32
	for i := 0; i < len(b) && i < 4; i++ {
		v |= uint32(b[i]) << (8 * i)
	}
	return v
}

// streamDecode treats raw as a single continuous field stream (base carried
// across the whole stream, not reset per message) and prints each field.
func streamDecode(raw []byte) {
	base := uint32(0)
	p := 0
	n := 0
	for p < len(raw) {
		tag, tn, err := m3.FirstChunk(raw[p:])
		if err != nil {
			fmt.Printf("  [tag error at %d: %v]\n", p, err)
			return
		}
		p += tn
		typ, id := unfold(leUint(tag), base)
		val, vn, err := m3.FirstChunk(raw[p:])
		if err != nil {
			fmt.Printf("  [val error at %d: %v]\n", p, err)
			return
		}
		p += vn
		base = id
		n++
		asc := make([]byte, 0, len(val))
		for _, c := range val {
			if c >= 0x20 && c < 0x7f {
				asc = append(asc, c)
			} else {
				asc = append(asc, '.')
			}
		}
		show := val
		trail := ""
		if len(show) > 40 {
			show = show[:40]
			trail = "…"
			asc = asc[:min(len(asc), 40)]
		}
		fmt.Printf("%5d: id=%-5d t=%d len=%-4d %x%s |%s|\n", n, id, typ, len(val), show, trail, asc)
	}
}

func main() {
	dirFilter := flag.String("dir", "", "only this direction: C (client->controller) or S (controller->client)")
	idFilter := flag.Int("id", -1, "only frames whose first top-level field id equals this")
	maxDepth := flag.Int("depth", 5, "max recursion depth")
	stream := flag.String("stream", "", "concatenate all frames of a direction (C or S) and decode as ONE continuous field stream (running base)")
	flag.Parse()

	if *stream != "" {
		runStream(*stream)
		return
	}

	sc := bufio.NewScanner(os.Stdin)
	sc.Buffer(make([]byte, 1<<20), 1<<20)

	var dir string
	var want int
	var hexParts []string
	frame := 0

	finalize := func() {
		if dir == "" {
			return
		}
		raw, err := hex.DecodeString(strings.ReplaceAll(strings.Join(hexParts, ""), " ", ""))
		dir, hexParts = "", nil
		if err != nil || len(raw) == 0 {
			return
		}
		top, _ := m3.Decode(raw)
		firstID := -1
		if top != nil && len(top.Fields) > 0 {
			firstID = int(top.Fields[0].ID)
		}
		d := "C->S"
		if strings.HasPrefix(dirSaved, "controller") {
			d = "S->C"
		}
		if *dirFilter == "C" && d != "C->S" {
			return
		}
		if *dirFilter == "S" && d != "S->C" {
			return
		}
		if *idFilter >= 0 && firstID != *idFilter {
			return
		}
		frame++
		fmt.Printf("\n#%d %s %dB (want %d)\n", frame, d, len(raw), want)
		dump(raw, 0, *maxDepth)
	}

	for sc.Scan() {
		line := sc.Text()
		if m := frameHdr.FindStringSubmatch(line); m != nil {
			finalize()
			dirSaved = m[1]
			dir = m[1]
			want, _ = strconv.Atoi(m[2])
			hexParts = nil
			continue
		}
		if dir != "" {
			if hexRow.MatchString(line) {
				hexParts = append(hexParts, strings.TrimSpace(line))
				continue
			}
			if strings.Contains(line, "id=") || strings.TrimSpace(line) == "" {
				// decoded section or blank ends the hex block; keep until next header
				continue
			}
		}
	}
	finalize()
}

// dirSaved carries the raw direction label across finalize (set at header time).
var dirSaved string

// runStream reads the whole log, concatenates the raw bytes of every frame in
// the requested direction (in order), and decodes the result as one continuous
// m3 field stream to test whether the delta base carries across TLS writes.
func runStream(which string) {
	sc := bufio.NewScanner(os.Stdin)
	sc.Buffer(make([]byte, 1<<20), 1<<20)
	var dir string
	var raw []byte
	want := which // "C" or "S"
	for sc.Scan() {
		line := sc.Text()
		if m := frameHdr.FindStringSubmatch(line); m != nil {
			dir = "C"
			if strings.HasPrefix(m[1], "controller") {
				dir = "S"
			}
			continue
		}
		if dir == want && hexRow.MatchString(line) {
			b, err := hex.DecodeString(strings.ReplaceAll(strings.TrimSpace(line), " ", ""))
			if err == nil {
				raw = append(raw, b...)
			}
		}
	}
	fmt.Printf("# %s stream: %d bytes\n", want, len(raw))
	streamDecode(raw)
}
