package cmr

import (
	"encoding/hex"
	"strings"
	"testing"
)

// Two real device objects captured from a live fresh sync (the device reporting
// its own state over the nv control plane).
const (
	// a config/system object: identity, version, board model
	objSystem = "d44a00c03fd086d21823d24980d39261058500d2857f0185018501d48901c03f" +
		"d35b5a08d28d0102d2e140d343485285d9372e32366265746131d28d01d3434852" +
		"85fe2a4348522051454d55205374616e646172642050432028693434304658202b" +
		"20504949582c203139393629d2e580d400c0550f8dd400000030d28501d400c041" +
		"018dd40010403dd2059ed2500985d2f00a85d2500985d2f00a85d2500985d2f00a"
	// a state object: MAC, vendor, version timestamp
	objState = "d4aa00803ffe5efe5cb1d6525400cc0101b50249d40a630071d2a101d3434852" +
		"d20501d84d696b726f54696bd20502d34348528ddb44695331776f4451786142" +
		"d2e101ed372e3236626574613120323032362d31302d30312031333a32393a3437d2890505"
)

func TestParseObjectState(t *testing.T) {
	body, err := hex.DecodeString(objState)
	if err != nil {
		t.Fatal(err)
	}
	o := ParseObject(body)
	if o.Class != ClassState {
		t.Errorf("Class = %#x, want ClassState %#x", o.Class, ClassState)
	}
	joined := strings.Join(o.Strings, "|")
	for _, want := range []string{"CHR", "MikroTik", "7.26beta1 2026-10-01 13:29:47"} {
		if !strings.Contains(joined, want) {
			t.Errorf("Strings missing %q; got %q", want, joined)
		}
	}
	if !contains(o.MACs, "52:54:00:cc:01:01") {
		t.Errorf("MACs = %v, want 52:54:00:cc:01:01", o.MACs)
	}
	t.Logf("%s", o)
}

func TestParseObjectSystem(t *testing.T) {
	body, err := hex.DecodeString(objSystem)
	if err != nil {
		t.Fatal(err)
	}
	o := ParseObject(body)
	if o.Class != ClassConfig {
		t.Errorf("Class = %#x, want ClassConfig %#x", o.Class, ClassConfig)
	}
	joined := strings.Join(o.Strings, "|")
	for _, want := range []string{"CHR", "7.26beta1", "CHR QEMU Standard PC (i440FX + PIIX, 1996)"} {
		if !strings.Contains(joined, want) {
			t.Errorf("Strings missing %q; got %q", want, joined)
		}
	}
	t.Logf("%s", o)
}

func contains(xs []string, s string) bool {
	for _, x := range xs {
		if x == s {
			return true
		}
	}
	return false
}
