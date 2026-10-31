package controller

import (
	"encoding/hex"
	"testing"
)

// hx decodes a hex string or fails the test.
func hx(t *testing.T, s string) []byte {
	t.Helper()
	b, err := hex.DecodeString(s)
	if err != nil {
		t.Fatalf("bad hex %q: %v", s, err)
	}
	return b
}

// withNonce runs fn with randBytes stubbed to return the given bytes (a capture's
// nonce), so a builder's output is byte-deterministic.
func withNonce(nonce []byte, fn func() []byte) []byte {
	old := randBytes
	randBytes = func(n int) []byte {
		if n == len(nonce) {
			return append([]byte(nil), nonce...)
		}
		return make([]byte, n)
	}
	defer func() { randBytes = old }()
	return fn()
}

// The frames below are the exact controller->device payloads from live
// open-pairing captures, with the controller's random nonce held fixed so every
// non-random byte must match.

func TestServerVersionBytes(t *testing.T) {
	want := hx(t, "09e5b1d4c3fc4cf235d84b60be6a01621a071dd3434852")
	got := withNonce(hx(t, "c3fc4cf2"), serverVersion)
	if !eq(got, want) {
		t.Fatalf("serverVersion\n got %x\nwant %x", got, want)
	}
}

func TestDiscAckBytes(t *testing.T) {
	want := hx(t, "31d8b1d45255d5f81d15")
	got := withNonce(hx(t, "5255d5f8"), func() []byte { return discAck(0x15) })
	if !eq(got, want) {
		t.Fatalf("discAck\n got %x\nwant %x", got, want)
	}
}

func TestVerAckBytes(t *testing.T) {
	want := hx(t, "09e309d7d1d9016dd23301b1d41af662c71dd23401")
	got := withNonce(hx(t, "1af662c7"), func() []byte { return verAck([]byte{0x34, 0x01}, 0x01) })
	if !eq(got, want) {
		t.Fatalf("verAck\n got %x\nwant %x", got, want)
	}
}

func TestControllerRowsBytes(t *testing.T) {
	// The exact 73-byte frame the live controller sent after verAck:
	// id10 SUCCESS (nonce=e500a2d3) + id9 row1 + id9 row2.
	want := hx(t, "51eab9e0636f6d2e6d696b726f74696b2e636d7245005dd4e500a2d3"+
		"1de1d1e949d2010201a9d81400000000000004"+
		"05e8d26102d3434852d1e949d2010205a9d81500000000000004")
	got := withNonce(hx(t, "e500a2d3"), controllerRows)
	if !eq(got, want) {
		t.Fatalf("controllerRows\n got %x\nwant %x", got, want)
	}
}

func TestServerHelloShape(t *testing.T) {
	got := serverHello()
	want := hx(t, "11dcd21101d8acec0a1b7c1ad815")
	if !eq(got, want) {
		t.Fatalf("serverHello\n got %x\nwant %x", got, want)
	}
}

func eq(a, b []byte) bool {
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
