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

func main() {
	dirFilter := flag.String("dir", "", "only this direction: C (client->controller) or S (controller->client)")
	idFilter := flag.Int("id", -1, "only frames whose first top-level field id equals this")
	maxDepth := flag.Int("depth", 5, "max recursion depth")
	flag.Parse()

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
