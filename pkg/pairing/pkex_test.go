package pairing

import (
	"bytes"
	"testing"
)

// TestPointRoundTrip checks the [32B x][1B parity] encoding decodes back to the
// same point, for a range of scalar multiples of G.
func TestPointRoundTrip(t *testing.T) {
	for i := 1; i <= 8; i++ {
		s := make([]byte, 32)
		s[31] = byte(i)
		p := scalarMultBase(s)
		enc := EncodePoint(p)
		if len(enc) != 33 {
			t.Fatalf("encoded length %d", len(enc))
		}
		got, err := DecodePoint(enc)
		if err != nil {
			t.Fatalf("decode %d*G: %v", i, err)
		}
		if got.X.Cmp(p.X) != 0 || got.Y.Cmp(p.Y) != 0 {
			t.Fatalf("round-trip mismatch for %d*G", i)
		}
	}
}

// TestMaskingRecovers verifies M = X + Qpw and X = M - Qpw round-trip.
func TestMaskingRecovers(t *testing.T) {
	sess, err := NewSession("LabPass123!")
	if err != nil {
		t.Fatal(err)
	}
	m, err := DecodePoint(sess.Masked())
	if err != nil {
		t.Fatal(err)
	}
	recovered := pointSub(m, Qpw("LabPass123!"))
	if recovered.X.Cmp(sess.Public.X) != 0 || recovered.Y.Cmp(sess.Public.Y) != 0 {
		t.Fatal("unmasking did not recover the ephemeral public key")
	}
}

// TestPKEXSelfConsistent runs both roles and checks they derive the same Z and
// matching confirmation tags — i.e. the mechanics (mask, exchange, ECDH, KDF,
// HMAC) are internally correct. (Byte-parity with MikroTik depends on the
// CALIBRATE label inputs documented in pkex.go.)
func TestPKEXSelfConsistent(t *testing.T) {
	const pw = "LabPass123!"
	initiator, err := NewSession(pw)
	if err != nil {
		t.Fatal(err)
	}
	responder, err := NewSession(pw)
	if err != nil {
		t.Fatal(err)
	}
	mi := initiator.Masked()
	mr := responder.Masked()

	zi, peerR, err := initiator.Agree(mr)
	if err != nil {
		t.Fatal(err)
	}
	zr, peerI, err := responder.Agree(mi)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(zi, zr) {
		t.Fatal("initiator and responder derived different Z")
	}
	// Each side recovered the other's true ephemeral public key.
	if peerR.X.Cmp(responder.Public.X) != 0 || peerI.X.Cmp(initiator.Public.X) != 0 {
		t.Fatal("recovered peer public keys do not match")
	}
	if !bytes.Equal(Confirm(zi, mi, mr), Confirm(zr, mi, mr)) {
		t.Fatal("confirmation tags differ")
	}
}
