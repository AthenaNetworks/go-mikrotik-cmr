package pairing

import (
	"crypto/aes"
	"crypto/cipher"
	"testing"
)

func mustCipher(t *testing.T, key []byte) cipher.Block {
	t.Helper()
	b, err := aes.NewCipher(key)
	if err != nil {
		t.Fatal(err)
	}
	return b
}
