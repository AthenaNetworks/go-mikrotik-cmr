// Command cmr-learn is a pure-Go CMR controller TLS endpoint used to observe
// and probe the inner (post-TLS) m3 control protocol spoken by real MikroTik
// CMR devices. No external tools (no socat/openssl) — it ships as one binary.
//
// Observed from lab CHRs: the CMR control plane is TLS 1.2 on tcp/54321 with
// self-signed Ed25519 identities; the device id is the SHA-256 of the identity.
//
// Client-cert note: MikroTik's certs encode an empty signatureAlgorithm OID
// that Go's crypto/x509 rejects, and the cert can't be rewritten on the wire
// (it would break the TLS Finished MAC). So by default we DON'T request a
// client cert at the TLS layer (-mutual to try anyway); CMR's trust is the
// app-layer pinned-fingerprint pairing, and the device identity travels in the
// inner protocol. Use -capture-cert to grab the client cert out-of-band.
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
	"time"

	"github.com/athenanetworks/cmr-server/pkg/m3"
)

func main() {
	addr := flag.String("listen", "10.99.0.1:54321", "TLS listen address")
	mutual := flag.Bool("mutual", false, "request a client cert at TLS layer (fails on MikroTik's empty-OID certs)")
	reply := flag.String("reply", "mirror", "reply strategy for the first inner frame: mirror|serverhello|none")
	flag.Parse()

	cert := identity()
	cfg := &tls.Config{
		Certificates: []tls.Certificate{cert},
		MinVersion:   tls.VersionTLS12,
		MaxVersion:   tls.VersionTLS12,
		ClientAuth:   tls.NoClientCert,
	}
	if *mutual {
		cfg.ClientAuth = tls.RequireAnyClientCert
		cfg.VerifyPeerCertificate = func(rawCerts [][]byte, _ [][]*x509.Certificate) error {
			for _, rc := range rawCerts {
				fingerprints(rc)
			}
			return nil
		}
	}

	ln, err := tls.Listen("tcp", *addr, cfg)
	must(err)
	log.Printf("cmr-learn listening on %s (TLS 1.2, mutual=%v, reply=%s)", *addr, *mutual, *reply)
	for {
		c, err := ln.Accept()
		if err != nil {
			log.Printf("accept: %v", err)
			continue
		}
		go handle(c.(*tls.Conn), *reply)
	}
}

func handle(tc *tls.Conn, reply string) {
	defer tc.Close()
	peer := tc.RemoteAddr()
	log.Printf("--- connection from %s ---", peer)
	if err := tc.Handshake(); err != nil {
		log.Printf("[%s] handshake failed: %v", peer, err)
		return
	}
	st := tc.ConnectionState()
	log.Printf("[%s] TLS up version=%#x cipher=%#x", peer, st.Version, st.CipherSuite)

	buf := make([]byte, 8192)
	first := true
	for {
		tc.SetReadDeadline(time.Now().Add(20 * time.Second))
		n, err := tc.Read(buf)
		if n > 0 {
			frame := append([]byte(nil), buf[:n]...)
			log.Printf("[%s] client->server %d bytes:\n%s", peer, n, hexdump(frame))
			if msg, derr := m3.Decode(frame); derr == nil {
				log.Printf("[%s]   m3.Decode OK: %d field(s): %+v", peer, len(msg.Fields), msg.Fields)
			}
			if out := craftReply(reply, frame, first); out != nil {
				log.Printf("[%s] server->client %d bytes (reply=%s):\n%s", peer, len(out), reply, hexdump(out))
				if _, werr := tc.Write(out); werr != nil {
					log.Printf("[%s] write failed: %v", peer, werr)
					return
				}
			}
			first = false
		}
		if err != nil {
			log.Printf("[%s] read ended: %v", peer, err)
			return
		}
	}
}

// craftReply builds a probe reply to a client frame. first marks the opening
// frame of the connection.
func craftReply(strategy string, in []byte, first bool) []byte {
	switch strategy {
	case "mirror":
		if !first {
			return nil
		}
		return append([]byte(nil), in...)
	case "mirror-all":
		return append([]byte(nil), in...)
	case "serverhello":
		if !first {
			return nil
		}
		// Same framing as the observed hello (magic[4] ver[1] body[9]) with our
		// own random 9-byte body, to see if the client treats us as a peer.
		if len(in) >= 5 {
			out := append([]byte(nil), in[:5]...)
			body := make([]byte, len(in)-5)
			_, _ = rand.Read(body)
			return append(out, body...)
		}
		return nil
	default: // "none"
		return nil
	}
}

func identity() tls.Certificate {
	pub, priv, err := ed25519.GenerateKey(rand.Reader)
	must(err)
	spki, err := x509.MarshalPKIXPublicKey(pub)
	must(err)
	id := sha256.Sum256(spki)
	cn := "figman-" + hex.EncodeToString(id[:2])
	tmpl := &x509.Certificate{
		SerialNumber: big.NewInt(time.Now().UnixNano()),
		Subject:      pkix.Name{CommonName: cn},
		Issuer:       pkix.Name{CommonName: cn},
		NotBefore:    time.Unix(0, 0),
		NotAfter:     time.Date(2098, 1, 1, 0, 0, 0, 0, time.UTC),
	}
	der, err := x509.CreateCertificate(rand.Reader, tmpl, tmpl, pub, priv)
	must(err)
	log.Printf("our identity CN=%s id(SPKI)=%s", cn, hex.EncodeToString(id[:]))
	leaf, _ := x509.ParseCertificate(der)
	return tls.Certificate{Certificate: [][]byte{der}, PrivateKey: priv, Leaf: leaf}
}

func fingerprints(raw []byte) {
	cert, err := x509.ParseCertificate(raw)
	if err != nil {
		log.Printf("  peer cert parse error (expected for MikroTik empty-OID): %v", err)
		return
	}
	der := sha256.Sum256(raw)
	spki := sha256.Sum256(cert.RawSubjectPublicKeyInfo)
	log.Printf("  peer CN=%q id(DER)=%s id(SPKI)=%s", cert.Subject.CommonName,
		hex.EncodeToString(der[:]), hex.EncodeToString(spki[:]))
}

func hexdump(b []byte) string {
	var out string
	for i := 0; i < len(b); i += 16 {
		end := i + 16
		if end > len(b) {
			end = len(b)
		}
		row := b[i:end]
		hexs, asc := "", ""
		for _, ch := range row {
			hexs += fmt.Sprintf("%02x ", ch)
			if ch >= 32 && ch < 127 {
				asc += string(ch)
			} else {
				asc += "."
			}
		}
		out += fmt.Sprintf("  %-48s %s\n", hexs, asc)
	}
	return out
}

func must(err error) {
	if err != nil {
		log.Fatal(err)
	}
}
