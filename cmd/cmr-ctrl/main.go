// Command cmr-ctrl is a pure-Go CMR *controller* that a real MikroTik CMR
// device will pair and connect to, with no MikroTik controller involved.
//
// It implements the OPEN pairing flow (controller pairing-requirement=none):
// mutual-TLS-free TLS 1.2 on tcp/54321 with a self-signed Ed25519 identity,
// then the nv/m3 control plane bring-up and an id10 SUCCESS (status=0x00) that
// promotes the device to "paired". No PKEX / password handshake is performed —
// that is the whole point of the open path (see docs/re/*/FINDINGS.md).
//
// The frame shapes and sequencing were reconstructed byte-for-byte from a live
// open-pairing capture (scratchpad mitm_open.log) and verified against pkg/m3.
//
// Flow (strictly alternating; the device waits for each reply):
//
//	C id2(hello)            -> S id2(hello)
//	C id1(version, has id25)-> S id1(version: nonce,build,board)
//	C id6(discovery, id4)   -> S id6(ack: nonce,seq)       [remember disc seq]
//	C id1(nonce/seq)        -> S id1(verAck) + S id10(SUCCESS) + S id9(rows)
//	... management plane ...
package main

import (
	"crypto/ed25519"
	"crypto/rand"
	"crypto/sha256"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/hex"
	"flag"
	"fmt"
	"log"
	"math/big"
	"net"
	"os"
	"time"

	"github.com/athenanetworks/cmr-server/pkg/m3"
)

// CMR control-channel constants recovered from the live protocol.
var (
	helloMagic = []byte{0xd2, 0x11, 0x01, 0xd8}                         // hello token prefix
	buildVer   = []byte{0x4b, 0x60, 0xbe, 0x6a, 0x01, 0x62, 0x1a, 0x07} // build-time + version
	boardName  = []byte("CHR")
	scopeName  = []byte("com.mikrotik.cmr")
)

// m3 field ids on the control channel.
const (
	idHello  = 2  // hello frame opcode
	idVer    = 1  // version / nonce-seq / verAck frame opcode (also nested inside verAck)
	idDisc   = 6  // discovery + discovery-ack frame opcode
	idResync = 7  // discovery resync opcode
	idDev    = 9  // device-table row
	idSucc   = 10 // pairing SUCCESS/status frame opcode
	idDesc   = 0  // device descriptor / error frame opcode

	idScope  = 23 // scope string "com.mikrotik.cmr"
	idStatus = 27 // status byte (0x00=success, 0x02=password required, 0x83=error)
	idBoard  = 24 // board identity string (in version)
	idBuildV = 25 // build-time + version
	idNonce  = 22 // per-message nonce
	idSeq    = 21 // session sequence
	idDevFl  = 29 // device row flag
	idDevIx  = 64 // device row index
	idBoard2 = 76 // board string (in device row)

	nonceSize = 4
)

func main() {
	addr := flag.String("listen", "10.99.0.1:54321", "TLS listen address")
	keyfile := flag.String("keyfile", "", "path to persist the Ed25519 identity (stable fingerprint across restarts); empty = ephemeral")
	idle := flag.Duration("idle", 0, "per-read idle timeout before closing a connection (0 = never, hold open)")
	flag.Parse()

	cert := identity(*keyfile)
	cfg := &tls.Config{
		Certificates: []tls.Certificate{cert},
		MinVersion:   tls.VersionTLS12,
		MaxVersion:   tls.VersionTLS12,
		ClientAuth:   tls.NoClientCert, // MikroTik's empty-OID client certs break crypto/x509; trust is app-layer
	}
	ln, err := tls.Listen("tcp", *addr, cfg)
	must(err)
	log.Printf("cmr-ctrl listening on %s (TLS 1.2, open pairing, pairing-requirement=none)", *addr)
	for {
		c, err := ln.Accept()
		if err != nil {
			log.Printf("accept: %v", err)
			continue
		}
		go handle(c.(*tls.Conn), *idle)
	}
}

// conn is the per-connection controller state machine.
type conn struct {
	tc       *tls.Conn
	peer     string
	discSeq []byte // last discovery/resync outer id21 (echoed back in verAck)
	paired  bool
}

func handle(tc *tls.Conn, idle time.Duration) {
	defer tc.Close()
	c := &conn{tc: tc, peer: tc.RemoteAddr().String()}
	log.Printf("--- connection from %s ---", c.peer)
	if err := tc.Handshake(); err != nil {
		log.Printf("[%s] handshake failed: %v", c.peer, err)
		return
	}
	st := tc.ConnectionState()
	log.Printf("[%s] TLS up version=%#x cipher=%#x", c.peer, st.Version, st.CipherSuite)

	// Keep the control channel alive with TCP keepalive; the device holds
	// "connected" as long as this socket is up and stays silent when idle.
	if tcp, ok := tc.NetConn().(*net.TCPConn); ok {
		tcp.SetKeepAlive(true)
		tcp.SetKeepAlivePeriod(25 * time.Second)
	}

	buf := make([]byte, 16384)
	for {
		if idle > 0 {
			tc.SetReadDeadline(time.Now().Add(idle))
		} else {
			tc.SetReadDeadline(time.Time{}) // no idle timeout: hold the connection open
		}
		n, err := tc.Read(buf)
		if n > 0 {
			frame := append([]byte(nil), buf[:n]...)
			msg, derr := m3.Decode(frame)
			if derr != nil || len(msg.Fields) == 0 {
				log.Printf("[%s] C->S %d bytes (undecodable: %v)\n%s", c.peer, n, derr, hexdump(frame))
			} else {
				log.Printf("[%s] C->S %d bytes: %s", c.peer, n, summary(msg))
				for _, out := range c.step(msg) {
					om, _ := m3.Decode(out)
					log.Printf("[%s] S->C %d bytes: %s", c.peer, len(out), summary(om))
					if _, werr := tc.Write(out); werr != nil {
						log.Printf("[%s] write failed: %v", c.peer, werr)
						return
					}
				}
			}
		}
		if err != nil {
			log.Printf("[%s] read ended: %v", c.peer, err)
			return
		}
	}
}

// step maps one decoded client frame to the controller's reply frames.
func (c *conn) step(msg *m3.Message) [][]byte {
	f := msg.Fields[0]
	switch f.ID {
	case idHello: // C hello -> S hello
		return [][]byte{serverHello()}

	case idVer: // id1: either the version (has id25) or the post-discovery nonce/seq
		inner, _ := m3.Decode(f.Value)
		if inner != nil {
			if _, ok := inner.GetBytes(idBuildV); ok {
				return [][]byte{serverVersion()} // C version -> S version
			}
		}
		// post-discovery id1 -> verAck, then id10 SUCCESS (promotion to "paired").
		// NB: do NOT send fabricated id9 device-table rows here; the device
		// replies id0{0x83 "not found"} to rows referencing devices it can't
		// resolve, then drops the session. SUCCESS alone is what pairs it.
		c.paired = true
		return [][]byte{verAck(c.discSeq, 0x01), successFrame()}

	case idDisc, idResync: // C discovery/resync -> S ack (remember the seq for verAck)
		// id22/id21 live inside the id6/id7 nested value, not at the top level.
		if inner, _ := m3.Decode(f.Value); inner != nil {
			if b, ok := inner.GetBytes(idSeq); ok {
				c.discSeq = append([]byte(nil), b...)
			}
		}
		if f.ID == idResync {
			// a resync (id7) wants a fresh verAck (status 0x00); keep the channel alive
			return [][]byte{verAck(c.discSeq, 0x00)}
		}
		return [][]byte{discAck(0x15)}

	case idDesc: // id0: device descriptor or an error report from the client
		if b, ok := msg.GetBytes(idStatus); ok && len(b) == 1 && b[0] != 0 {
			log.Printf("[%s]   client reports status=%#x (%q)", c.peer, b[0], strField(msg, 28))
		}
		return nil // ack nothing; keep channel open

	default:
		return nil
	}
}

func serverHello() []byte {
	token := append(append([]byte(nil), helloMagic...), randBytes(8)...)
	return m3.New().PutBytes(idHello, token).Encode()
}

func serverVersion() []byte {
	inner := m3.New().PutBytes(idNonce, randBytes(nonceSize)).PutBytes(idBuildV, buildVer).PutBytes(idBoard, boardName)
	return m3.New().PutNested(idVer, inner).Encode()
}

func discAck(seq byte) []byte {
	inner := m3.New().PutBytes(idNonce, randBytes(nonceSize)).PutBytes(idSeq, []byte{seq})
	return m3.New().PutNested(idDisc, inner).Encode()
}

// verAck is the controller's id1 reply after the client's post-discovery id1:
// id1 { id1{id27:status, id21:innerSeq}, id22:nonce, id21:outerSeq } where
// outerSeq echoes the client's discovery seq and innerSeq = [outerSeq[0]-1, ...].
func verAck(discSeq []byte, status byte) []byte {
	outerSeq := discSeq
	if len(outerSeq) == 0 {
		outerSeq = []byte{0x01, 0x01}
	}
	innerSeq := append([]byte(nil), outerSeq...)
	innerSeq[0]-- // controller's own message seq is one below the client's
	sub := m3.New().PutU32(idStatus, uint32(status)).PutBytes(idSeq, innerSeq)
	body := m3.New().PutNested(idVer, sub).PutBytes(idNonce, randBytes(nonceSize)).PutBytes(idSeq, outerSeq)
	return m3.New().PutNested(idVer, body).Encode()
}

// successFrame is the id10 pairing SUCCESS alone: {id23:scope, id27:0x00, id22:nonce}.
func successFrame() []byte {
	return m3.New().PutNested(idSucc, successInner()).Encode()
}

func successInner() *m3.Message {
	return m3.New().PutBytes(idScope, scopeName).PutU32(idStatus, 0).PutBytes(idNonce, randBytes(nonceSize))
}

// successAndRows is the single frame the live controller sends after verAck:
// id10 SUCCESS followed by two id9 device-table rows (the controller
// advertising itself). As one message the ids delta-chain 10->9->9.
func successAndRows() []byte {
	r1 := m3.New().PutBytes(idDevFl, []byte{0x49}).PutBytes(idDevIx, []byte{0x01}).
		PutBytes(idSeq, []byte{0x14, 0, 0, 0, 0, 0, 0, 0x04})
	r2 := m3.New().PutBytes(idBoard2, boardName).PutBytes(idDevFl, []byte{0x49}).PutBytes(idDevIx, []byte{0x05}).
		PutBytes(idSeq, []byte{0x15, 0, 0, 0, 0, 0, 0, 0x04})
	return m3.New().PutNested(idSucc, successInner()).PutNested(idDev, r1).PutNested(idDev, r2).Encode()
}

// randBytes is a package var so tests can substitute deterministic nonces.
var randBytes = func(n int) []byte {
	b := make([]byte, n)
	_, _ = rand.Read(b)
	return b
}

// identity returns the controller's TLS identity. With keyfile set, the Ed25519
// private key is persisted so the device-pinned fingerprint survives restarts.
func identity(keyfile string) tls.Certificate {
	var pub ed25519.PublicKey
	var priv ed25519.PrivateKey
	if keyfile != "" {
		if seed, err := os.ReadFile(keyfile); err == nil && len(seed) == ed25519.SeedSize {
			priv = ed25519.NewKeyFromSeed(seed)
			pub = priv.Public().(ed25519.PublicKey)
			log.Printf("loaded identity from %s", keyfile)
		}
	}
	if priv == nil {
		var err error
		pub, priv, err = ed25519.GenerateKey(rand.Reader)
		must(err)
		if keyfile != "" {
			must(os.WriteFile(keyfile, priv.Seed(), 0o600))
			log.Printf("generated and saved identity to %s", keyfile)
		}
	}
	spki, err := x509.MarshalPKIXPublicKey(pub)
	must(err)
	id := sha256.Sum256(spki)
	cn := "cmr-ctrl-" + hex.EncodeToString(id[:2])
	tmpl := &x509.Certificate{
		SerialNumber: big.NewInt(1),
		Subject:      pkix.Name{CommonName: cn},
		Issuer:       pkix.Name{CommonName: cn},
		NotBefore:    time.Unix(0, 0),
		NotAfter:     time.Date(2098, 1, 1, 0, 0, 0, 0, time.UTC),
	}
	der, err := x509.CreateCertificate(rand.Reader, tmpl, tmpl, pub, priv)
	must(err)
	log.Printf("controller identity CN=%s id(SPKI)=%s", cn, hex.EncodeToString(id[:]))
	leaf, _ := x509.ParseCertificate(der)
	return tls.Certificate{Certificate: [][]byte{der}, PrivateKey: priv, Leaf: leaf}
}

// summary renders a decoded m3 message compactly for logs.
func summary(m *m3.Message) string {
	if m == nil {
		return "(nil)"
	}
	s := ""
	for _, f := range m.Fields {
		v := f.Value
		if len(v) > 24 {
			v = v[:24]
		}
		s += fmt.Sprintf("id%d=%s ", f.ID, hex.EncodeToString(v))
	}
	return s
}

func strField(m *m3.Message, id uint32) string {
	if b, ok := m.GetBytes(id); ok {
		return string(b)
	}
	return ""
}

func hexdump(b []byte) string {
	var out string
	for i := 0; i < len(b); i += 16 {
		end := i + 16
		if end > len(b) {
			end = len(b)
		}
		out += "  " + hex.EncodeToString(b[i:end]) + "\n"
	}
	return out
}

func must(err error) {
	if err != nil {
		log.Fatal(err)
	}
}
