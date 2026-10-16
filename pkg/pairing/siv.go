// Package pairing implements MikroTik's CMR password "key handshake" — a
// Wi-Fi-DPP-style PKEX over NIST P-256, reverse-engineered from the cmr binary
// and libucrypto.so. See docs/PROTOCOL.md for the protocol and the crypto map.
//
// This file implements AES-SIV (RFC 5297), the authenticated encryption cmr
// uses to wrap PKEX attributes (AESSIV::encrypt/decrypt in libucrypto).
package pairing

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/subtle"
	"errors"
)

// aesCMAC computes AES-CMAC (RFC 4493) of msg under block cipher c.
func aesCMAC(c cipher.Block, msg []byte) []byte {
	const bs = 16
	// Subkey generation.
	l := make([]byte, bs)
	c.Encrypt(l, l)
	k1 := dbl(l)
	k2 := dbl(k1)

	var last []byte
	n := len(msg) / bs
	rem := len(msg) % bs
	complete := len(msg) > 0 && rem == 0
	if complete {
		last = xor(msg[(n-1)*bs:], k1)
	} else {
		pad := make([]byte, bs)
		copy(pad, msg[n*bs:])
		pad[rem] = 0x80
		last = xor(pad, k2)
	}

	x := make([]byte, bs)
	blocks := len(msg) / bs
	if !complete {
		// process all full blocks before the padded last
		blocks = n
	} else {
		blocks = n - 1
	}
	for i := 0; i < blocks; i++ {
		x = xor(x, msg[i*bs:(i+1)*bs])
		c.Encrypt(x, x)
	}
	y := xor(last, x)
	out := make([]byte, bs)
	c.Encrypt(out, y)
	return out
}

// dbl doubles a 128-bit value in GF(2^128) (RFC 5297 §2.3 / RFC 4493).
func dbl(b []byte) []byte {
	out := make([]byte, 16)
	var carry byte
	for i := 15; i >= 0; i-- {
		out[i] = b[i]<<1 | carry
		carry = b[i] >> 7
	}
	if b[0]&0x80 != 0 {
		out[15] ^= 0x87
	}
	return out
}

func xor(a, b []byte) []byte {
	n := len(a)
	if len(b) < n {
		n = len(b)
	}
	out := make([]byte, len(a))
	copy(out, a)
	for i := 0; i < n; i++ {
		out[i] = a[i] ^ b[i]
	}
	return out
}

// xorend XORs b into the end of a (len(a) >= len(b)).
func xorend(a, b []byte) []byte {
	out := make([]byte, len(a))
	copy(out, a)
	off := len(a) - len(b)
	for i := 0; i < len(b); i++ {
		out[off+i] ^= b[i]
	}
	return out
}

// s2v is the S2V construction (RFC 5297 §2.4).
func s2v(c cipher.Block, ads [][]byte, plaintext []byte) []byte {
	zero := make([]byte, 16)
	if len(ads) == 0 && len(plaintext) == 0 {
		one := make([]byte, 16)
		one[15] = 1
		return aesCMAC(c, one)
	}
	d := aesCMAC(c, zero)
	for _, ad := range ads {
		d = xor(dbl(d), aesCMAC(c, ad))
	}
	var t []byte
	if len(plaintext) >= 16 {
		t = xorend(plaintext, d)
	} else {
		pad := make([]byte, 16)
		copy(pad, plaintext)
		pad[len(plaintext)] = 0x80
		t = xor(dbl(d), pad)
	}
	return aesCMAC(c, t)
}

// sivEncrypt performs AES-SIV encryption (RFC 5297). key is 2*aesKeySize
// (32 for AES-128-SIV, 64 for AES-256-SIV); the first half is the S2V/CMAC key
// and the second half the CTR key. Output is V(16) || ciphertext.
func sivEncrypt(key []byte, ads [][]byte, plaintext []byte) ([]byte, error) {
	if len(key)%2 != 0 {
		return nil, errors.New("aes-siv: odd key length")
	}
	half := len(key) / 2
	macBlk, err := aes.NewCipher(key[:half])
	if err != nil {
		return nil, err
	}
	ctrBlk, err := aes.NewCipher(key[half:])
	if err != nil {
		return nil, err
	}
	v := s2v(macBlk, ads, plaintext)
	q := make([]byte, 16)
	copy(q, v)
	q[8] &= 0x7f
	q[12] &= 0x7f
	ct := make([]byte, len(plaintext))
	cipher.NewCTR(ctrBlk, q).XORKeyStream(ct, plaintext)
	return append(append([]byte(nil), v...), ct...), nil
}

// sivDecrypt reverses sivEncrypt, returning the plaintext or an error if the
// synthetic IV does not verify.
func sivDecrypt(key []byte, ads [][]byte, in []byte) ([]byte, error) {
	if len(in) < 16 {
		return nil, errors.New("aes-siv: input too short")
	}
	half := len(key) / 2
	macBlk, err := aes.NewCipher(key[:half])
	if err != nil {
		return nil, err
	}
	ctrBlk, err := aes.NewCipher(key[half:])
	if err != nil {
		return nil, err
	}
	v := in[:16]
	ct := in[16:]
	q := make([]byte, 16)
	copy(q, v)
	q[8] &= 0x7f
	q[12] &= 0x7f
	pt := make([]byte, len(ct))
	cipher.NewCTR(ctrBlk, q).XORKeyStream(pt, ct)
	expect := s2v(macBlk, ads, pt)
	if subtle.ConstantTimeCompare(expect, v) != 1 {
		return nil, errors.New("aes-siv: authentication failed")
	}
	return pt, nil
}
