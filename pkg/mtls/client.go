// Package mtls is a minimal, pure-Go TLS 1.2 ECDHE client used to speak to a
// MikroTik CMR controller on tcp/54321.
//
// Why not crypto/tls: MikroTik's identity certificates encode a zero-length
// signatureAlgorithm OID, which both Go's crypto/x509 and OpenSSL reject when
// they parse the *server* certificate during the handshake — and neither
// exposes a hook to tolerate it. For an ECDHE cipher suite the server cert is
// only used to authenticate the key exchange; it is not needed to derive the
// session keys. This client therefore performs the ECDHE handshake and simply
// does not parse (or verify) the server certificate. It is deliberately scoped
// to exactly one cipher suite, TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256
// (0xc02b), which is what RouterOS 7.26 negotiates.
//
// This is for authorized interoperability/security research against your own
// devices. It performs no certificate validation and must not be used as a
// general-purpose TLS client.
package mtls

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/ecdh"
	"crypto/ed25519"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"net"
	"time"
)

const (
	recHandshake      = 22
	recChangeCipher   = 20
	recAlert          = 21
	recAppData        = 23
	hsClientHello     = 1
	hsServerHello     = 2
	hsCertificate     = 11
	hsServerKeyExch   = 12
	hsCertRequest     = 13
	hsServerHelloDone = 14
	hsClientKeyExch   = 16
	hsCertVerify      = 15
	hsFinished        = 20
	cipherECDHE_ECDSA = 0xc02b
	sigSchemeEd25519  = 0x0807 // TLS SignatureScheme ed25519 (hash=Intrinsic, sig=ed25519)
)

// Config carries an optional client identity for mutual TLS. If the server
// requests a client certificate and a Cert/Key are set, they are presented
// with a CertificateVerify; otherwise an empty certificate is sent.
type Config struct {
	ClientCertDER []byte             // leaf certificate (DER)
	ClientKey     ed25519.PrivateKey // its Ed25519 private key
}

// Conn is an established CMR TLS session carrying m3 frames. It is not safe for
// concurrent use.
type Conn struct {
	raw        net.Conn
	gcmOut     cipher.AEAD
	gcmIn      cipher.AEAD
	ivOut      [4]byte
	ivIn       [4]byte
	seqOut     uint64
	seqIn      uint64
	readBuf       []byte // decrypted app-data not yet returned
	ServerCert    []byte // raw leaf certificate DER (unparsed), for out-of-band analysis
	CertRequested bool   // the server sent a CertificateRequest
}

// Dial connects to addr and completes the handshake with no client identity.
func Dial(addr string, timeout time.Duration) (*Conn, error) {
	return DialConfig(addr, timeout, nil)
}

// DialConfig connects to addr and completes the TLS 1.2 ECDHE handshake,
// presenting the client identity in cfg if the server asks for one. The server
// certificate is never validated (see package docs).
func DialConfig(addr string, timeout time.Duration, cfg *Config) (*Conn, error) {
	if cfg == nil {
		cfg = &Config{}
	}
	raw, err := net.DialTimeout("tcp", addr, timeout)
	if err != nil {
		return nil, err
	}
	if timeout > 0 {
		raw.SetDeadline(time.Now().Add(timeout))
	}
	c := &Conn{raw: raw}
	if err := c.handshake(cfg); err != nil {
		raw.Close()
		return nil, err
	}
	raw.SetDeadline(time.Time{})
	return c, nil
}

func (c *Conn) handshake(cfg *Config) error {
	var transcript []byte // all handshake-message bytes in order

	clientRandom := make([]byte, 32)
	rand.Read(clientRandom)
	ch := clientHello(clientRandom)
	transcript = append(transcript, ch...)
	if err := c.writeRecord(recHandshake, ch); err != nil {
		return fmt.Errorf("send ClientHello: %w", err)
	}

	// Read server flight up to ServerHelloDone.
	var serverRandom []byte
	var serverPoint []byte
	hs := newHSReader(c)
	for {
		typ, body, err := hs.next()
		if err != nil {
			return fmt.Errorf("read server flight: %w", err)
		}
		transcript = append(transcript, msgBytes(typ, body)...)
		switch typ {
		case hsServerHello:
			if len(body) < 35 {
				return errors.New("short ServerHello")
			}
			serverRandom = append([]byte(nil), body[2:34]...)
			// skip session id
			p := 34
			sidLen := int(body[p])
			p += 1 + sidLen
			if p+2 > len(body) {
				return errors.New("ServerHello truncated at cipher")
			}
			cs := binary.BigEndian.Uint16(body[p:])
			if cs != cipherECDHE_ECDSA {
				return fmt.Errorf("server chose cipher %#04x, want %#04x", cs, cipherECDHE_ECDSA)
			}
		case hsCertificate:
			c.ServerCert = firstCert(body)
		case hsServerKeyExch:
			pt, err := parseServerKeyExchange(body)
			if err != nil {
				return err
			}
			serverPoint = pt
		case hsCertRequest:
			// Server asks for a client cert; we have none it would accept.
			// TLS requires us to answer with an (empty) Certificate message.
			c.CertRequested = true
		case hsServerHelloDone:
			goto haveFlight
		}
	}
haveFlight:
	if serverRandom == nil || serverPoint == nil {
		return errors.New("missing ServerHello/ServerKeyExchange")
	}

	// Our ECDHE key pair (P-256) and the shared secret.
	curve := ecdh.P256()
	priv, err := curve.GenerateKey(rand.Reader)
	if err != nil {
		return err
	}
	srvPub, err := curve.NewPublicKey(serverPoint)
	if err != nil {
		return fmt.Errorf("server ECDHE point: %w", err)
	}
	preMaster, err := priv.ECDH(srvPub)
	if err != nil {
		return fmt.Errorf("ECDH: %w", err)
	}

	// If the server requested a client certificate, present ours (or empty).
	haveCert := c.CertRequested && len(cfg.ClientCertDER) > 0 && cfg.ClientKey != nil
	if c.CertRequested {
		var certMsg []byte
		if haveCert {
			certMsg = msgBytes(hsCertificate, certificateBody(cfg.ClientCertDER))
		} else {
			certMsg = msgBytes(hsCertificate, []byte{0, 0, 0})
		}
		transcript = append(transcript, certMsg...)
		if err := c.writeRecord(recHandshake, certMsg); err != nil {
			return fmt.Errorf("send Certificate: %w", err)
		}
	}

	// ClientKeyExchange with our uncompressed point.
	myPub := priv.PublicKey().Bytes() // 65-byte uncompressed
	cke := make([]byte, 0, 1+len(myPub))
	cke = append(cke, byte(len(myPub)))
	cke = append(cke, myPub...)
	ckeMsg := msgBytes(hsClientKeyExch, cke)
	transcript = append(transcript, ckeMsg...)
	if err := c.writeRecord(recHandshake, ckeMsg); err != nil {
		return fmt.Errorf("send ClientKeyExchange: %w", err)
	}

	// CertificateVerify proves possession of the client key (Ed25519 signs the
	// handshake transcript directly, per RFC 8422).
	if haveCert {
		sig := ed25519.Sign(cfg.ClientKey, transcript)
		cv := make([]byte, 0, 4+len(sig))
		cv = append(cv, byte(sigSchemeEd25519>>8), byte(sigSchemeEd25519&0xff), byte(len(sig)>>8), byte(len(sig)))
		cv = append(cv, sig...)
		cvMsg := msgBytes(hsCertVerify, cv)
		transcript = append(transcript, cvMsg...)
		if err := c.writeRecord(recHandshake, cvMsg); err != nil {
			return fmt.Errorf("send CertificateVerify: %w", err)
		}
	}

	// Key schedule (classic, no extended master secret).
	master := prf(preMaster, "master secret", concat(clientRandom, serverRandom), 48)
	kb := prf(master, "key expansion", concat(serverRandom, clientRandom), 40)
	cwk, swk := kb[0:16], kb[16:32]
	copy(c.ivOut[:], kb[32:36])
	copy(c.ivIn[:], kb[36:40])
	if c.gcmOut, err = newGCM(cwk); err != nil {
		return err
	}
	if c.gcmIn, err = newGCM(swk); err != nil {
		return err
	}

	// ChangeCipherSpec, then encrypted Finished.
	if err := c.writeRecord(recChangeCipher, []byte{1}); err != nil {
		return err
	}
	clientVerify := prf(master, "client finished", sha256sum(transcript), 12)
	finMsg := msgBytes(hsFinished, clientVerify)
	transcript = append(transcript, finMsg...)
	if err := c.writeEncrypted(recHandshake, finMsg); err != nil {
		return fmt.Errorf("send Finished: %w", err)
	}

	// Server ChangeCipherSpec + encrypted Finished.
	typ, payload, err := c.readRecord()
	if err != nil {
		return fmt.Errorf("read server CCS: %w", err)
	}
	if typ != recChangeCipher {
		return fmt.Errorf("want server CCS, got record type %d", typ)
	}
	_ = payload
	typ, serverFin, err := c.readEncrypted()
	if err != nil {
		return fmt.Errorf("read server Finished: %w", err)
	}
	if typ != recHandshake || len(serverFin) < 1 || serverFin[0] != hsFinished {
		return fmt.Errorf("want server Finished, got type %d", typ)
	}
	expect := prf(master, "server finished", sha256sum(transcript), 12)
	if len(serverFin) >= 4 && !hmac.Equal(serverFin[4:], expect) {
		return errors.New("server Finished verify_data mismatch")
	}
	return nil
}

// WriteFrame sends one application record (one m3 frame per TLS record).
func (c *Conn) WriteFrame(b []byte) error { return c.writeEncrypted(recAppData, b) }

// ReadFrame reads one application record.
func (c *Conn) ReadFrame() ([]byte, error) {
	typ, b, err := c.readEncrypted()
	if err != nil {
		return nil, err
	}
	if typ == recAlert {
		return nil, fmt.Errorf("tls alert: % x", b)
	}
	if typ != recAppData {
		return nil, fmt.Errorf("unexpected record type %d", typ)
	}
	return b, nil
}

// SetDeadline sets the underlying connection deadline.
func (c *Conn) SetDeadline(t time.Time) error { return c.raw.SetDeadline(t) }

// Close closes the connection.
func (c *Conn) Close() error { return c.raw.Close() }

// --- record layer ---

func (c *Conn) writeRecord(typ byte, payload []byte) error {
	hdr := []byte{typ, 3, 3, byte(len(payload) >> 8), byte(len(payload))}
	if _, err := c.raw.Write(append(hdr, payload...)); err != nil {
		return err
	}
	return nil
}

func (c *Conn) readRecord() (byte, []byte, error) {
	var hdr [5]byte
	if _, err := io.ReadFull(c.raw, hdr[:]); err != nil {
		return 0, nil, err
	}
	n := int(binary.BigEndian.Uint16(hdr[3:5]))
	if n > 1<<16 {
		return 0, nil, errors.New("record too large")
	}
	buf := make([]byte, n)
	if _, err := io.ReadFull(c.raw, buf); err != nil {
		return 0, nil, err
	}
	return hdr[0], buf, nil
}

func (c *Conn) writeEncrypted(typ byte, plaintext []byte) error {
	var nonce [12]byte
	copy(nonce[:4], c.ivOut[:])
	binary.BigEndian.PutUint64(nonce[4:], c.seqOut)
	aad := aead(c.seqOut, typ, len(plaintext))
	ct := c.gcmOut.Seal(nil, nonce[:], plaintext, aad)
	payload := make([]byte, 8+len(ct))
	binary.BigEndian.PutUint64(payload[:8], c.seqOut)
	copy(payload[8:], ct)
	c.seqOut++
	return c.writeRecord(typ, payload)
}

func (c *Conn) readEncrypted() (byte, []byte, error) {
	typ, payload, err := c.readRecord()
	if err != nil {
		return 0, nil, err
	}
	if len(payload) < 8+16 {
		return 0, nil, errors.New("short encrypted record")
	}
	var nonce [12]byte
	copy(nonce[:4], c.ivIn[:])
	copy(nonce[4:], payload[:8]) // explicit nonce from wire
	ct := payload[8:]
	aad := aead(c.seqIn, typ, len(ct)-16)
	pt, err := c.gcmIn.Open(nil, nonce[:], ct, aad)
	if err != nil {
		return 0, nil, fmt.Errorf("gcm open (seq %d): %w", c.seqIn, err)
	}
	c.seqIn++
	return typ, pt, nil
}

func aead(seq uint64, typ byte, n int) []byte {
	a := make([]byte, 13)
	binary.BigEndian.PutUint64(a[:8], seq)
	a[8] = typ
	a[9], a[10] = 3, 3
	binary.BigEndian.PutUint16(a[11:], uint16(n))
	return a
}

// --- handshake message reassembly (messages may span records) ---

type hsReader struct {
	c   *Conn
	buf []byte
}

func newHSReader(c *Conn) *hsReader { return &hsReader{c: c} }

func (h *hsReader) next() (byte, []byte, error) {
	for {
		if len(h.buf) >= 4 {
			l := int(h.buf[1])<<16 | int(h.buf[2])<<8 | int(h.buf[3])
			if len(h.buf) >= 4+l {
				typ := h.buf[0]
				body := append([]byte(nil), h.buf[4:4+l]...)
				h.buf = h.buf[4+l:]
				return typ, body, nil
			}
		}
		typ, payload, err := h.c.readRecord()
		if err != nil {
			return 0, nil, err
		}
		if typ == recAlert {
			return 0, nil, fmt.Errorf("tls alert during handshake: % x", payload)
		}
		if typ != recHandshake {
			return 0, nil, fmt.Errorf("unexpected record type %d during handshake", typ)
		}
		h.buf = append(h.buf, payload...)
	}
}

// --- helpers ---

func clientHello(random []byte) []byte {
	var ext []byte
	// supported_groups: secp256r1 (0x0017)
	ext = appendExt(ext, 0x000a, []byte{0x00, 0x02, 0x00, 0x17})
	// ec_point_formats: uncompressed
	ext = appendExt(ext, 0x000b, []byte{0x01, 0x00})
	// signature_algorithms
	sa := []byte{0x08, 0x07, 0x04, 0x03, 0x04, 0x01, 0x08, 0x04, 0x05, 0x01}
	ext = appendExt(ext, 0x000d, append([]byte{byte(len(sa) >> 8), byte(len(sa))}, sa...))
	// renegotiation_info (empty)
	ext = appendExt(ext, 0xff01, []byte{0x00})

	var b []byte
	b = append(b, 3, 3)
	b = append(b, random...)
	b = append(b, 0x00) // session id len
	b = append(b, 0x00, 0x02, 0xc0, 0x2b)
	b = append(b, 0x01, 0x00) // compression: null
	b = append(b, byte(len(ext)>>8), byte(len(ext)))
	b = append(b, ext...)
	return msgBytes(hsClientHello, b)
}

func appendExt(b []byte, typ uint16, data []byte) []byte {
	b = append(b, byte(typ>>8), byte(typ), byte(len(data)>>8), byte(len(data)))
	return append(b, data...)
}

func msgBytes(typ byte, body []byte) []byte {
	out := make([]byte, 4+len(body))
	out[0] = typ
	out[1] = byte(len(body) >> 16)
	out[2] = byte(len(body) >> 8)
	out[3] = byte(len(body))
	copy(out[4:], body)
	return out
}

func parseServerKeyExchange(body []byte) ([]byte, error) {
	// ServerECDHParams: curve_type(1), named_curve(2), point_len(1), point.
	if len(body) < 4 || body[0] != 3 {
		return nil, errors.New("ServerKeyExchange: not a named_curve")
	}
	nc := binary.BigEndian.Uint16(body[1:3])
	if nc != 0x0017 {
		return nil, fmt.Errorf("ServerKeyExchange: curve %#04x, want secp256r1", nc)
	}
	pl := int(body[3])
	if 4+pl > len(body) {
		return nil, errors.New("ServerKeyExchange: short point")
	}
	return append([]byte(nil), body[4:4+pl]...), nil
}

// certificateBody builds a TLS Certificate message body carrying one cert.
func certificateBody(der []byte) []byte {
	cl := len(der)
	listLen := 3 + cl
	out := make([]byte, 0, 6+cl)
	out = append(out, byte(listLen>>16), byte(listLen>>8), byte(listLen))
	out = append(out, byte(cl>>16), byte(cl>>8), byte(cl))
	return append(out, der...)
}

func firstCert(body []byte) []byte {
	// Certificate: certs_len(3), then cert_len(3)+cert...
	if len(body) < 6 {
		return nil
	}
	cl := int(body[3])<<16 | int(body[4])<<8 | int(body[5])
	if 6+cl > len(body) {
		return nil
	}
	return append([]byte(nil), body[6:6+cl]...)
}

func newGCM(key []byte) (cipher.AEAD, error) {
	blk, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	return cipher.NewGCM(blk)
}

func prf(secret []byte, label string, seed []byte, n int) []byte {
	full := append([]byte(label), seed...)
	out := make([]byte, 0, n)
	a := full
	for len(out) < n {
		h := hmac.New(sha256.New, secret)
		h.Write(a)
		a = h.Sum(nil)
		h2 := hmac.New(sha256.New, secret)
		h2.Write(a)
		h2.Write(full)
		out = append(out, h2.Sum(nil)...)
	}
	return out[:n]
}

func sha256sum(b []byte) []byte { s := sha256.Sum256(b); return s[:] }
func concat(a, b []byte) []byte { return append(append([]byte(nil), a...), b...) }
