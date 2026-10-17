package pairing

import (
	"crypto/elliptic"
	"crypto/hkdf"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"errors"
	"math/big"
)

// CMR's password handshake is a Wi-Fi-DPP-style PKEX over NIST P-256,
// reverse-engineered from cmr + libucrypto.so (PKEXImpl::{init,encrypt,decrypt,
// reveal1,reveal2,authDeriveZ,auth1,auth2}). Mechanics confirmed from the
// binary and the plaintext capture (docs/samples/pairing-handshake.txt):
//
//   - password -> scalar s (HKPDF/HKDF-SHA256); Qpw = s*G is the "password point"
//   - each side picks ephemeral d, public X = d*G, and sends the MASKED point
//     M = X + Qpw, encoded as [32-byte big-endian x][1-byte y-parity]
//   - the peer recovers the ephemeral point as X = M - Qpw and computes the
//     ECDH shared point K = d_self * X_peer; Z = HKPDF over K.x
//   - mutual confirmation is HMAC-SHA256 over the exchanged points (auth1/auth2);
//     later attributes are wrapped with AES-SIV under a key from Z
//
// The P-256 mechanics, point encoding, masking, ECDH and the AES-SIV/HMAC/HKDF
// primitives are implemented and self-consistent here. The exact MikroTik label
// inputs to the three KDF/MAC steps — marked CALIBRATE below — are assembled at
// runtime in the binary (scope "com.mikrotik.cmr", username, role tags) and are
// the final piece to match byte-for-byte against a live controller; see
// docs/PROTOCOL.md for the crypto-site addresses.

var curve = elliptic.P256()

// scope is the PKEX context string seen throughout the capture.
const scope = "com.mikrotik.cmr"

// Point is an affine P-256 point.
type Point struct{ X, Y *big.Int }

// EncodePoint serializes a point as [32-byte big-endian X][1-byte (Y&1)],
// the encoding cmr's ECPoint::toBin(0x20, compressed) produces on the wire.
func EncodePoint(p Point) []byte {
	out := make([]byte, 33)
	p.X.FillBytes(out[:32])
	out[32] = byte(p.Y.Bit(0))
	return out
}

// DecodePoint parses the [32-byte X][1-byte parity] encoding, recovering Y.
func DecodePoint(b []byte) (Point, error) {
	if len(b) != 33 {
		return Point{}, errors.New("pkex: point must be 33 bytes")
	}
	x := new(big.Int).SetBytes(b[:32])
	y, err := recoverY(x, b[32]&1)
	if err != nil {
		return Point{}, err
	}
	if !curve.IsOnCurve(x, y) {
		return Point{}, errors.New("pkex: decoded point not on curve")
	}
	return Point{X: x, Y: y}, nil
}

// recoverY solves y^2 = x^3 - 3x + b (mod p) for the given parity. P-256's p ≡ 3
// (mod 4), so the square root is v^((p+1)/4) mod p.
func recoverY(x *big.Int, parity byte) (*big.Int, error) {
	p := curve.Params().P
	three := big.NewInt(3)
	rhs := new(big.Int).Exp(x, three, p)        // x^3
	rhs.Sub(rhs, new(big.Int).Mul(three, x))    // - 3x
	rhs.Add(rhs, curve.Params().B)              // + b
	rhs.Mod(rhs, p)
	exp := new(big.Int).Add(p, big.NewInt(1))
	exp.Rsh(exp, 2) // (p+1)/4
	y := new(big.Int).Exp(rhs, exp, p)
	if new(big.Int).Exp(y, big.NewInt(2), p).Cmp(rhs) != 0 {
		return nil, errors.New("pkex: no square root (bad x)")
	}
	if byte(y.Bit(0)) != parity {
		y.Sub(p, y)
	}
	return y, nil
}

func scalarMultBase(s []byte) Point {
	x, y := curve.ScalarBaseMult(s)
	return Point{X: x, Y: y}
}

func scalarMult(p Point, s []byte) Point {
	x, y := curve.ScalarMult(p.X, p.Y, s)
	return Point{X: x, Y: y}
}

func pointAdd(a, b Point) Point {
	x, y := curve.Add(a.X, a.Y, b.X, b.Y)
	return Point{X: x, Y: y}
}

func pointNeg(p Point) Point {
	return Point{X: new(big.Int).Set(p.X), Y: new(big.Int).Sub(curve.Params().P, p.Y)}
}

func pointSub(a, b Point) Point { return pointAdd(a, pointNeg(b)) }

// randScalar returns a uniform scalar in [1, n-1].
func randScalar() ([]byte, error) {
	n := curve.Params().N
	for {
		b := make([]byte, 32)
		if _, err := rand.Read(b); err != nil {
			return nil, err
		}
		k := new(big.Int).SetBytes(b)
		if k.Sign() != 0 && k.Cmp(n) < 0 {
			return k.FillBytes(make([]byte, 32)), nil
		}
	}
}

// PasswordScalar maps the pairing password to a P-256 scalar s such that
// Qpw = s*G. CALIBRATE: the exact HKPDF salt/info (scope, username, role tag)
// are assembled at runtime in cmr (HKPDF site 0x8055a9d); here we use
// HKDF-SHA256(secret=password, salt=scope, info="cmr pkex password") reduced
// mod n — the structure is right, the label bytes are the piece to match.
func PasswordScalar(password string) []byte {
	raw, err := hkdf.Key(sha256.New, []byte(password), []byte(scope), "cmr pkex password", 48)
	if err != nil {
		panic("pkex: hkdf: " + err.Error())
	}
	n := curve.Params().N
	s := new(big.Int).SetBytes(raw)
	s.Mod(s, new(big.Int).Sub(n, big.NewInt(1)))
	s.Add(s, big.NewInt(1))
	return s.FillBytes(make([]byte, 32))
}

// Qpw returns the password point s*G.
func Qpw(password string) Point { return scalarMultBase(PasswordScalar(password)) }

// Session is one side of the PKEX exchange.
type Session struct {
	password string
	priv     []byte // ephemeral private scalar
	Public   Point  // ephemeral public X = priv*G
	qpw      Point
}

// NewSession creates a session with a fresh ephemeral key.
func NewSession(password string) (*Session, error) {
	priv, err := randScalar()
	if err != nil {
		return nil, err
	}
	return &Session{password: password, priv: priv, Public: scalarMultBase(priv), qpw: Qpw(password)}, nil
}

// Masked returns the masked public point M = X + Qpw, wire-encoded.
func (s *Session) Masked() []byte { return EncodePoint(pointAdd(s.Public, s.qpw)) }

// Agree takes the peer's masked point, recovers the peer ephemeral public key,
// and returns the ECDH shared secret Z.
//
// Confirmed byte-level from figman authDeriveZ (0x806e57e): K = ownEphemeral *
// ecPeerProt (ctx+0xb0, the recovered peer point) via WCurve::mul->toEC; then
// z = HKPDF(hash=createHash(5)=SHA256, ikm=K.x [BigNum::bin], salt=S, info=I,
// len=0x20). ctx+0xa0 (kProt) and ctx+0xb8 (code, the password element) are the
// other live context fields. CALIBRATE: the salt S and info I vectors — their
// exact contents (which of code/kProt/the masked points, and order) are the
// remaining bytes; placeholder info "cmr pkex z" keeps the mechanics consistent.
func (s *Session) Agree(peerMasked []byte) (z []byte, peerPub Point, err error) {
	m, err := DecodePoint(peerMasked)
	if err != nil {
		return nil, Point{}, err
	}
	peerPub = pointSub(m, s.qpw)
	if !curve.IsOnCurve(peerPub.X, peerPub.Y) {
		return nil, Point{}, errors.New("pkex: recovered peer point not on curve")
	}
	k := scalarMult(peerPub, s.priv)
	kx := k.X.FillBytes(make([]byte, 32))
	z, err = hkdf.Key(sha256.New, kx, nil, "cmr pkex z", 32)
	if err != nil {
		return nil, Point{}, err
	}
	return z, peerPub, nil
}

// Confirm computes the HMAC-SHA256 confirmation tag.
//
// Confirmed byte-level from figman reveal1 (0x806e820): J = ownScalar(ctx+0x98) *
// peerPoint(ctx+0x58) via WCurve::mul->toEC; tag = HMAC(key=J.x [BigNum::bin],
// hash=SHA256@0x807bcb0) updated with THREE BigNum::bin values, in order:
//
//	1. ctx+0x60              (own protocol/masked value)
//	2. ctx+0xb0 ecPeerProt   (peer ephemeral point)
//	3. ctx+0xa0 kProt + 4    (own protocol key field)
//
// then HMAC::digest. CALIBRATE: the field->Session mapping (which concrete
// point/scalar each ctx offset holds) is the piece to pin by lab iteration; the
// current two-masked-point transcript is a placeholder.
func Confirm(z, initiatorMasked, responderMasked []byte) []byte {
	mac := hmac.New(sha256.New, z)
	mac.Write(initiatorMasked)
	mac.Write(responderMasked)
	return mac.Sum(nil)
}
