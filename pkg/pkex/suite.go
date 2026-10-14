package pkex

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/elliptic"
	"crypto/hkdf"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"errors"
	"fmt"
	"io"
	"math/big"

	"github.com/athenanetworks/cmr-server/pkg/m3"
)

// Suite abstracts the PKEX cryptographic operations so the exact MikroTik DPP
// key schedule can be substituted once it is validated against a capture (see
// docs/PROTOCOL.md). The bundled ReferenceSuite is a correct, self-consistent
// P-256 PKEX-shaped handshake: it interoperates between this library's client
// and server roles, and is NOT claimed to interoperate with MikroTik devices.
type Suite interface {
	Name() string
	NewContext(role Role, code, localID []byte, rng io.Reader) Context
}

// Context carries one peer handshake. Build* produce an m3 body to send;
// Consume* ingest a received m3 body. Session calls these in role order.
type Context interface {
	BuildDiscovery() ([]byte, error)
	ConsumeDiscovery(body []byte) error
	BuildMsg1() ([]byte, error)
	ConsumeMsg1(body []byte) error
	BuildMsg2() ([]byte, error)
	ConsumeMsg2(body []byte) error
	BuildMsg3() ([]byte, error)
	ConsumeMsg3(body []byte) error
	BuildMsg4(addrs [][]byte) ([]byte, error)
	ConsumeMsg4(body []byte) (peerID []byte, addrs [][]byte, err error)
	PeerID() []byte
}

// m3 field ids used inside PKEX frame bodies by the reference suite.
const (
	fVersion uint32 = 1
	fNonce   uint32 = 0x10
	fPoint   uint32 = 0x11
	fTag     uint32 = 0x12
	fEnc     uint32 = 0x13
	fAddrs   uint32 = 0x15
)

// ProtocolVersion is carried in fVersion (observed: checked == 1 in handleMsg1).
const ProtocolVersion uint32 = 1

const (
	nonceLen  = 16
	peerIDLen = 32 // Srv::onPKEXResult copies a 32-byte peer id
)

// ReferenceSuite implements Suite with P-256 + HKDF-SHA256 + AES-256-GCM.
// It uses crypto/elliptic point ops (Marshal/Add/ScalarMult); those APIs are
// deprecated but present, and adequate for a reference implementation.
type ReferenceSuite struct{}

// Name implements Suite.
func (ReferenceSuite) Name() string { return "reference-p256-v0" }

// NewContext implements Suite.
func (ReferenceSuite) NewContext(role Role, code, localID []byte, rng io.Reader) Context {
	if rng == nil {
		rng = rand.Reader
	}
	return &refCtx{role: role, code: code, localID: localID, rng: rng, curve: elliptic.P256()}
}

type refCtx struct {
	role    Role
	code    []byte
	localID []byte
	rng     io.Reader
	curve   elliptic.Curve

	priv       []byte   // our ephemeral scalar
	myX, myY   *big.Int // our ephemeral public point
	myMasked   []byte   // our masked point as put on the wire
	peerMasked []byte   // peer masked point as received

	myNonce   []byte
	peerNonce []byte

	peerX, peerY *big.Int // recovered peer ephemeral public point
	kConfirm     []byte
	kEnc         []byte

	peerID []byte
}

// --- password element ------------------------------------------------------

func (c *refCtx) pwPoint() (x, y *big.Int) {
	h := sha256.Sum256(append([]byte("CMR-PKEX-pw\x00"), c.code...))
	n := c.curve.Params().N
	s := new(big.Int).Mod(new(big.Int).SetBytes(h[:]), n)
	return c.curve.ScalarBaseMult(s.Bytes())
}

func (c *refCtx) ensureEphemeral() error {
	if c.priv != nil {
		return nil
	}
	s, err := randScalar(c.curve, c.rng)
	if err != nil {
		return err
	}
	c.priv = s
	c.myX, c.myY = c.curve.ScalarBaseMult(s)
	return nil
}

// --- key schedule ----------------------------------------------------------

func (c *refCtx) ensureKeys() error {
	if c.kConfirm != nil {
		return nil
	}
	if c.priv == nil || c.peerX == nil {
		return errors.New("pkex: keys requested before point exchange")
	}
	zx, _ := c.curve.ScalarMult(c.peerX, c.peerY, c.priv)
	ni, nr := c.orderedNonces()
	okm, err := hkdf.Key(sha256.New, zx.Bytes(), append(append([]byte{}, ni...), nr...), "CMR-PKEX-v0", 64)
	if err != nil {
		return err
	}
	c.kConfirm, c.kEnc = okm[:32], okm[32:]
	return nil
}

func (c *refCtx) orderedNonces() (clientNonce, serverNonce []byte) {
	if c.role == RoleClient {
		return c.myNonce, c.peerNonce
	}
	return c.peerNonce, c.myNonce
}

// transcript binds both nonces and both on-wire masked points, client-first,
// so both roles compute an identical value.
func (c *refCtx) transcript() []byte {
	ni, nr := c.orderedNonces()
	cM, sM := c.myMasked, c.peerMasked
	if c.role == RoleServer {
		cM, sM = c.peerMasked, c.myMasked
	}
	out := make([]byte, 0, len(ni)+len(nr)+len(cM)+len(sM))
	out = append(out, ni...)
	out = append(out, nr...)
	out = append(out, cM...)
	out = append(out, sM...)
	return out
}

func (c *refCtx) tag(label byte) []byte {
	mac := hmac.New(sha256.New, c.kConfirm)
	mac.Write([]byte{label})
	mac.Write(c.transcript())
	return mac.Sum(nil)
}

// --- discovery -------------------------------------------------------------

func (c *refCtx) BuildDiscovery() ([]byte, error) {
	if c.myNonce == nil {
		c.myNonce = make([]byte, nonceLen)
		if _, err := io.ReadFull(c.rng, c.myNonce); err != nil {
			return nil, err
		}
	}
	return m3.New().PutU32(fVersion, ProtocolVersion).PutBytes(fNonce, c.myNonce).Encode(), nil
}

func (c *refCtx) ConsumeDiscovery(body []byte) error {
	msg, err := m3.Decode(body)
	if err != nil {
		return err
	}
	if v, ok := msg.U32(fVersion); !ok || v != ProtocolVersion {
		return ErrVersionMismatch
	}
	n, ok := msg.GetBytes(fNonce)
	if !ok || len(n) != nonceLen {
		return errors.New("pkex: bad discovery nonce")
	}
	c.peerNonce = n
	return nil
}

// --- point messages (msg1 server->client, msg2 client->server) -------------

func (c *refCtx) BuildMsg1() ([]byte, error)    { return c.buildPointMsg() }
func (c *refCtx) ConsumeMsg1(body []byte) error { return c.consumePointMsg(body) }
func (c *refCtx) BuildMsg2() ([]byte, error)    { return c.buildPointMsg() }
func (c *refCtx) ConsumeMsg2(body []byte) error { return c.consumePointMsg(body) }

func (c *refCtx) buildPointMsg() ([]byte, error) {
	if err := c.ensureEphemeral(); err != nil {
		return nil, err
	}
	pwx, pwy := c.pwPoint()
	mx, my := c.curve.Add(c.myX, c.myY, pwx, pwy)
	pt := elliptic.Marshal(c.curve, mx, my) //nolint:staticcheck // reference suite
	c.myMasked = pt
	return m3.New().PutU32(fVersion, ProtocolVersion).PutBytes(fPoint, pt).Encode(), nil
}

func (c *refCtx) consumePointMsg(body []byte) error {
	msg, err := m3.Decode(body)
	if err != nil {
		return err
	}
	if v, ok := msg.U32(fVersion); !ok || v != ProtocolVersion {
		return ErrVersionMismatch
	}
	pt, ok := msg.GetBytes(fPoint)
	if !ok {
		return errors.New("pkex: missing masked point")
	}
	mx, my := elliptic.Unmarshal(c.curve, pt) //nolint:staticcheck // reference suite
	if mx == nil {
		return errors.New("pkex: bad masked point encoding")
	}
	c.peerMasked = append([]byte(nil), pt...)
	// P = M - Pw
	pwx, pwy := c.pwPoint()
	negY := new(big.Int).Sub(c.curve.Params().P, pwy)
	px, py := c.curve.Add(mx, my, pwx, negY)
	if !c.curve.IsOnCurve(px, py) {
		return errors.New("pkex: recovered peer point not on curve")
	}
	c.peerX, c.peerY = px, py
	return nil
}

// --- auth messages (msg3 server->client tag, msg4 client->server tag+id) ---

func (c *refCtx) BuildMsg3() ([]byte, error) {
	if err := c.ensureKeys(); err != nil {
		return nil, err
	}
	return m3.New().PutBytes(fTag, c.tag('S')).Encode(), nil
}

func (c *refCtx) ConsumeMsg3(body []byte) error {
	if err := c.ensureKeys(); err != nil {
		return err
	}
	msg, err := m3.Decode(body)
	if err != nil {
		return err
	}
	tag, ok := msg.GetBytes(fTag)
	if !ok || !hmac.Equal(tag, c.tag('S')) {
		return ErrAuthFailed
	}
	return nil
}

func (c *refCtx) BuildMsg4(addrs [][]byte) ([]byte, error) {
	if err := c.ensureKeys(); err != nil {
		return nil, err
	}
	h := sha256.Sum256(c.localID)
	enc, err := c.seal(h[:])
	if err != nil {
		return nil, err
	}
	msg := m3.New().PutBytes(fTag, c.tag('C')).PutBytes(fEnc, enc)
	if len(addrs) > 0 {
		msg.PutVector(fAddrs, addrs)
	}
	return msg.Encode(), nil
}

func (c *refCtx) ConsumeMsg4(body []byte) ([]byte, [][]byte, error) {
	if err := c.ensureKeys(); err != nil {
		return nil, nil, err
	}
	msg, err := m3.Decode(body)
	if err != nil {
		return nil, nil, err
	}
	tag, ok := msg.GetBytes(fTag)
	if !ok || !hmac.Equal(tag, c.tag('C')) {
		return nil, nil, ErrAuthFailed
	}
	enc, ok := msg.GetBytes(fEnc)
	if !ok {
		return nil, nil, errors.New("pkex: missing encrypted fields in msg4")
	}
	id, err := c.open(enc)
	if err != nil {
		return nil, nil, err
	}
	if len(id) != peerIDLen {
		return nil, nil, fmt.Errorf("pkex: peer id length %d != %d", len(id), peerIDLen)
	}
	c.peerID = id
	addrs, _ := msg.GetVector(fAddrs)
	return id, addrs, nil
}

func (c *refCtx) PeerID() []byte { return c.peerID }

// --- AEAD ------------------------------------------------------------------

func (c *refCtx) aead() (cipher.AEAD, error) {
	block, err := aes.NewCipher(c.kEnc)
	if err != nil {
		return nil, err
	}
	return cipher.NewGCM(block)
}

func (c *refCtx) seal(pt []byte) ([]byte, error) {
	a, err := c.aead()
	if err != nil {
		return nil, err
	}
	nonce := make([]byte, a.NonceSize())
	if _, err := io.ReadFull(c.rng, nonce); err != nil {
		return nil, err
	}
	return append(nonce, a.Seal(nil, nonce, pt, nil)...), nil
}

func (c *refCtx) open(ct []byte) ([]byte, error) {
	a, err := c.aead()
	if err != nil {
		return nil, err
	}
	if len(ct) < a.NonceSize() {
		return nil, errors.New("pkex: short ciphertext")
	}
	return a.Open(nil, ct[:a.NonceSize()], ct[a.NonceSize():], nil)
}

// randScalar returns a uniformly random scalar in [1, n-1].
func randScalar(curve elliptic.Curve, rng io.Reader) ([]byte, error) {
	n := curve.Params().N
	for i := 0; i < 64; i++ {
		b := make([]byte, (n.BitLen()+7)/8)
		if _, err := io.ReadFull(rng, b); err != nil {
			return nil, err
		}
		k := new(big.Int).SetBytes(b)
		if k.Sign() > 0 && k.Cmp(n) < 0 {
			return k.Bytes(), nil
		}
	}
	return nil, errors.New("pkex: failed to sample scalar")
}
