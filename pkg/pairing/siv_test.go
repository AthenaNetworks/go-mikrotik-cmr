package pairing

import (
	"bytes"
	"encoding/hex"
	"testing"
)

// RFC 5297 Appendix A.1 deterministic authenticated-encryption test vector.
func TestAESSIV_RFC5297(t *testing.T) {
	key, _ := hex.DecodeString("fffefdfcfbfaf9f8f7f6f5f4f3f2f1f0f0f1f2f3f4f5f6f7f8f9fafbfcfdfeff")
	ad, _ := hex.DecodeString("101112131415161718191a1b1c1d1e1f2021222324252627")
	pt, _ := hex.DecodeString("112233445566778899aabbccddee")
	want, _ := hex.DecodeString("85632d07c6e8f37f950acd320a2ecc9340c02b9690c4dc04daef7f6afe5c")

	got, err := sivEncrypt(key, [][]byte{ad}, pt)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(got, want) {
		t.Fatalf("SIV encrypt:\n got %x\nwant %x", got, want)
	}
	back, err := sivDecrypt(key, [][]byte{ad}, got)
	if err != nil {
		t.Fatalf("decrypt: %v", err)
	}
	if !bytes.Equal(back, pt) {
		t.Fatalf("decrypt mismatch: got %x want %x", back, pt)
	}
}

// AES-CMAC vectors from RFC 4493.
func TestAESCMAC_RFC4493(t *testing.T) {
	key, _ := hex.DecodeString("2b7e151628aed2a6abf7158809cf4f3c")
	blk := mustCipher(t, key)
	cases := []struct{ msg, mac string }{
		{"", "bb1d6929e95937287fa37d129b756746"},
		{"6bc1bee22e409f96e93d7e117393172a", "070a16b46b4d4144f79bdd9dd04a287c"},
		{"6bc1bee22e409f96e93d7e117393172aae2d8a571e03ac9c9eb76fac45af8e5130c81c46a35ce411", "dfa66747de9ae63030ca32611497c827"},
	}
	for _, c := range cases {
		msg, _ := hex.DecodeString(c.msg)
		want, _ := hex.DecodeString(c.mac)
		if got := aesCMAC(blk, msg); !bytes.Equal(got, want) {
			t.Errorf("CMAC(%s):\n got %x\nwant %x", c.msg, got, want)
		}
	}
}
