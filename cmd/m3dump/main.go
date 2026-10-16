// m3dump: recursively decode hex m3 frames (one "<outerid> <hex>" per line) for RE.
package main

import (
	"bufio"
	"encoding/hex"
	"fmt"
	"os"
	"strings"

	"github.com/athenanetworks/cmr-server/pkg/m3"
)

func dump(b []byte, depth int) {
	msg, err := m3.Decode(b)
	if err != nil || msg == nil {
		return
	}
	ind := strings.Repeat("  ", depth)
	for _, f := range msg.Fields {
		v := f.Value
		asc := ""
		for _, c := range v {
			if c >= 0x20 && c < 0x7f {
				asc += string(c)
			} else {
				asc += "."
			}
		}
		if len(v) > 40 {
			asc = asc[:40] + "…"
		}
		fmt.Printf("%sid=%-3d len=%-3d %x |%s|\n", ind, f.ID, len(v), v, asc)
		if len(v) >= 2 && depth < 4 { // try nested
			if sub, e := m3.Decode(v); e == nil && sub != nil && len(sub.Fields) > 0 {
				// only recurse if it round-trips (real nesting)
				if bytesEq(sub.Encode(), v) {
					dump(v, depth+1)
				}
			}
		}
	}
}
func bytesEq(a, b []byte) bool {
	if len(a) != len(b) { return false }
	for i := range a { if a[i] != b[i] { return false } }
	return true
}

func main() {
	sc := bufio.NewScanner(os.Stdin)
	sc.Buffer(make([]byte, 1<<20), 1<<20)
	for sc.Scan() {
		line := strings.TrimSpace(sc.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		parts := strings.Fields(line)
		h := parts[len(parts)-1]
		b, err := hex.DecodeString(h)
		if err != nil {
			continue
		}
		fmt.Printf(">>> %s\n", line[:min(len(line), 30)])
		dump(b, 0)
	}
}
func min(a, b int) int { if a < b { return a }; return b }
