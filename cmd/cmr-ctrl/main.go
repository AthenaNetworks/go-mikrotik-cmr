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
	"strings"
	"time"

	"github.com/athenanetworks/cmr-server/pkg/cmr"
	"github.com/athenanetworks/cmr-server/pkg/m3"
)

// runtime flags (set in main)
var (
	verbose   bool
	probe     bool
	enumScope string
	enumFlag  uint
	enumSel   string
	pushFile  string
	pushBlob  []byte
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
	flag.BoolVar(&verbose, "v", false, "log a full recursive decode of every frame")
	flag.BoolVar(&probe, "probe", false, "after SUCCESS, send id9 get-requests to elicit the device's objects (reading experiment)")
	flag.StringVar(&enumScope, "enum", "", "after SUCCESS, send an id7 getall for this scope to enumerate the device's objects (e.g. com.mikrotik.cmr)")
	flag.UintVar(&enumFlag, "enumflag", 0x18, "id29 flag byte for the getall query")
	flag.StringVar(&enumSel, "enumsel", "", "hex id30 selector for the getall query (empty = none)")
	flag.StringVar(&pushFile, "pushfile", "", "path to a hex blob of controller object-push bytes to send after a device resync (reading experiment: triggers the device to dump its objects)")
	flag.Parse()
	if pushFile != "" {
		b, err := os.ReadFile(pushFile)
		must(err)
		pushBlob = mustHex(strings.TrimSpace(string(b)))
		log.Printf("loaded push blob: %d bytes", len(pushBlob))
	}

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
	in      *m3.StreamDecoder // continuous decoder for the device->controller stream
	discSeq []byte            // last discovery/resync outer id21 (echoed back in verAck)
	paired  bool
}

func handle(tc *tls.Conn, idle time.Duration) {
	defer tc.Close()
	c := &conn{tc: tc, peer: tc.RemoteAddr().String(), in: m3.NewStreamDecoder()}
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

	// The device->controller direction is a single continuous m3 field stream
	// (the delta base carries across TLS reads), so decode it with a running
	// StreamDecoder and process one top-level field at a time.
	buf := make([]byte, 16384)
	for {
		if idle > 0 {
			tc.SetReadDeadline(time.Now().Add(idle))
		} else {
			tc.SetReadDeadline(time.Time{}) // no idle timeout: hold the connection open
		}
		n, err := tc.Read(buf)
		if n > 0 {
			c.in.Feed(buf[:n])
			for {
				f, ok := c.in.Next()
				if !ok {
					break
				}
				log.Printf("[%s] C->S id%d (%d bytes)", c.peer, f.ID, len(f.Value))
				if verbose {
					fmt.Printf("  id=%d len=%d %x\n%s", f.ID, len(f.Value), f.Value, m3.Dump(f.Value, 6))
				}
				for _, out := range c.step(f) {
					log.Printf("[%s] S->C %d bytes: %s", c.peer, len(out), summary(firstDecode(out)))
					if _, werr := tc.Write(out); werr != nil {
						log.Printf("[%s] write failed: %v", c.peer, werr)
						return
					}
				}
			}
		}
		if err != nil {
			log.Printf("[%s] read ended: %v (%d bytes buffered)", c.peer, err, c.in.Buffered())
			return
		}
	}
}

func firstDecode(b []byte) *m3.Message {
	m, _ := m3.Decode(b)
	return m
}

// step maps one top-level client field to the controller's reply frames.
func (c *conn) step(f m3.Field) [][]byte {
	// Object-bearing fields (id0/id1/id9 whose body carries an id74 key) are the
	// device REPORTING state, not control frames — route them to the reader.
	if f.ID == idDesc || f.ID == idVer || f.ID == idDev {
		if body, _ := m3.Decode(f.Value); body != nil {
			if _, ok := body.GetBytes(74); ok {
				c.readObject(f, body)
				return nil
			}
		}
	}
	switch f.ID {
	case idHello: // C hello -> S hello
		return [][]byte{serverHello()}

	case idVer: // id1: version (has id25), the post-discovery nonce/seq, or an ack
		inner, _ := m3.Decode(f.Value)
		if inner != nil {
			if _, ok := inner.GetBytes(idBuildV); ok {
				return [][]byte{serverVersion()} // C version -> S version
			}
		}
		if c.paired {
			// Already paired: a bare id1{nonce,seq} is a control ack / keepalive
			// (e.g. the device acking our getall). Absorb it — do NOT re-run the
			// SUCCESS path, or every ack would re-pair and storm the channel.
			return nil
		}
		// First post-discovery id1 -> verAck, then id10 SUCCESS (promotion to
		// "paired"). NB: in normal operation do NOT send fabricated id9
		// device-table rows here; the device replies id0{0x83 "not found"} and
		// drops the session. SUCCESS alone pairs. -probe/-enum are experiments.
		c.paired = true
		out := [][]byte{verAck(c.discSeq, 0x01)}
		switch {
		case probe:
			out = append(out, successAndRowsFresh()) // SUCCESS + id9 rows (fresh-capture flag/handles)
		default:
			out = append(out, successFrame())
		}
		if enumScope != "" {
			out = append(out, getAll(enumScope, byte(enumFlag), mustHex(enumSel)))
			log.Printf("[%s]   -> getall scope=%q flag=%#x sel=%dB", c.peer, enumScope, enumFlag, len(mustHex(enumSel)))
		}
		return out

	case idDisc, idResync: // C discovery/resync -> S ack (remember the seq for verAck)
		// id22/id21 live inside the id6/id7 nested value, not at the top level.
		if inner, _ := m3.Decode(f.Value); inner != nil {
			if b, ok := inner.GetBytes(idSeq); ok {
				c.discSeq = append([]byte(nil), b...)
			}
		}
		if f.ID == idResync {
			// A resync (id7) is the device asking to reconcile a scope. Ack it; then,
			// if we have a controller object-push blob, send it — the device dumps its
			// own objects (the read) as reciprocation to the controller's push.
			out := [][]byte{verAck(c.discSeq, 0x00)}
			if len(pushBlob) > 0 {
				out = append(out, pushBlob)
				log.Printf("[%s]   -> replaying %d-byte object push after resync", c.peer, len(pushBlob))
			}
			return out
		}
		return [][]byte{discAck(0x15)}

	case idDesc: // id0: device descriptor, config object, or an error report
		body, _ := m3.Decode(f.Value)
		if body != nil {
			if b, ok := body.GetBytes(idStatus); ok && len(b) == 1 && b[0] != 0 {
				log.Printf("[%s]   client reports status=%#x (%q)", c.peer, b[0], strField(body, 28))
			}
			if cmr.IsDescriptor(body) {
				if d, err := cmr.ParseDescriptor(f.Value); err == nil {
					log.Printf("[%s]   READ %s", c.peer, d)
				}
			}
		}
		return nil // ack nothing; keep channel open

	default:
		return nil
	}
}

const idQuery = 4 // query body inside a discovery/resync/getall opcode

// getAll builds an id7 getall query: id7{ id4{id21,id23 scope,id29 flag,
// [id30 selector],id31,id27=0}, id22 nonce, id21 seq }. It asks the device to
// enumerate its objects in scope.
func getAll(scope string, flag byte, selector []byte) []byte {
	q := m3.New().
		Put(idSeq, []byte{0x01}).
		PutBytes(idScope, []byte(scope)).
		Put(idDevFl, []byte{flag})
	if len(selector) > 0 {
		q.PutBytes(30, selector)
	}
	q.PutBytes(31, nil).PutU32(idStatus, 0)
	return m3.New().PutNested(idResync, q).PutBytes(idNonce, randBytes(nonceSize)).Put(idSeq, []byte{0x02}).Encode()
}

// readObject logs a device-reported object (the reading path): its handle, flag,
// index, and the key/value as their raw chunk lists (semantic decoding of the
// typed attributes is future work).
func (c *conn) readObject(f m3.Field, body *m3.Message) {
	handle, _ := body.GetBytes(idSeq)
	flagb, _ := body.GetBytes(idDevFl)
	idxb, _ := body.GetBytes(idDevIx)
	key, _ := body.GetBytes(74)
	val, _ := body.GetBytes(75)
	log.Printf("[%s]   READ obj id%d handle=%x flag=%x idx=%x", c.peer, f.ID, handle, flagb, idxb)
	if len(key) > 0 {
		log.Printf("[%s]       key: %s", c.peer, fmtChunks(key))
	}
	if len(val) > 0 {
		log.Printf("[%s]       val: %s", c.peer, fmtChunks(val))
	}
}

// fmtChunks splits a typed key/value blob into its m3 chunks and renders them as
// space-separated hex, the structural view of the nv attribute encoding.
func fmtChunks(b []byte) string {
	var parts []string
	for len(b) > 0 {
		d, n, err := m3.FirstChunk(b)
		if err != nil || n == 0 {
			parts = append(parts, "+"+hex.EncodeToString(b))
			break
		}
		parts = append(parts, hex.EncodeToString(d))
		b = b[n:]
	}
	return strings.Join(parts, " ")
}

func mustHex(s string) []byte {
	if s == "" {
		return nil
	}
	b, err := hex.DecodeString(s)
	if err != nil {
		log.Fatalf("bad hex %q: %v", s, err)
	}
	return b
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

// successAndRowsFresh mirrors the clean fresh-sync capture's SUCCESS+rows: flag
// 0x4b and handles 88 33../89 33.. (vs the old capture's 0x49 / 14../15..). It
// tests whether a freshly paired device accepts controller-chosen handles
// (returns its descriptor) or requires generation-matched ones.
func successAndRowsFresh() []byte {
	r1 := m3.New().PutBytes(idDevFl, []byte{0x4b}).PutBytes(idDevIx, []byte{0x01}).
		PutBytes(idSeq, []byte{0x88, 0x33, 0, 0, 0, 0, 0, 0x04})
	r2 := m3.New().PutBytes(idBoard2, boardName).PutBytes(idDevFl, []byte{0x4b}).PutBytes(idDevIx, []byte{0x05}).
		PutBytes(idSeq, []byte{0x89, 0x33, 0, 0, 0, 0, 0, 0x04})
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


func must(err error) {
	if err != nil {
		log.Fatal(err)
	}
}
