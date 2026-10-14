// Command cmr-learn is a CMR controller TLS endpoint used to observe the inner
// (post-TLS) m3 control protocol that real MikroTik CMR devices speak.
//
// Observed from lab CHRs: the CMR control plane is mutual TLS 1.2 on tcp/54321
// with self-signed Ed25519 identity certs; a device's CMR id is the SHA-256
// fingerprint of its identity. This endpoint presents such an identity, accepts
// any client cert (an open pairing window — the operator points a device at us
// on purpose and still supplies the pairing password at the app layer), logs
// the peer's identity, and dumps + attempts to m3-decode the first inner bytes.
//
// It never forces a device to connect and never bypasses the pairing password.
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
	"time"

	"github.com/athenanetworks/cmr-server/pkg/m3"
)

func main() {
	addr := flag.String("listen", "10.99.0.1:54321", "TLS listen address")
	flag.Parse()

	cert := identity()
	cfg := &tls.Config{
		Certificates: []tls.Certificate{cert},
		MinVersion:   tls.VersionTLS12,
		MaxVersion:   tls.VersionTLS12,
		ClientAuth:   tls.RequireAnyClientCert, // CMR sends a CertificateRequest
		VerifyPeerCertificate: func(rawCerts [][]byte, _ [][]*x509.Certificate) error {
			for _, rc := range rawCerts {
				fingerprints(rc)
			}
			return nil
		},
	}
	ln, err := tls.Listen("tcp", *addr, cfg)
	must(err)
	log.Printf("cmr-learn listening on %s (mutual TLS 1.2)", *addr)
	for {
		c, err := ln.Accept()
		if err != nil {
			log.Printf("accept: %v", err)
			continue
		}
		go handle(c)
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
		log.Printf("  peer cert parse error: %v", err)
		return
	}
	der := sha256.Sum256(raw)
	spki := sha256.Sum256(cert.RawSubjectPublicKeyInfo)
	log.Printf("  peer CN=%q sigalg=%v", cert.Subject.CommonName, cert.SignatureAlgorithm)
	log.Printf("    id(DER) =%s", hex.EncodeToString(der[:]))
	log.Printf("    id(SPKI)=%s", hex.EncodeToString(spki[:]))
}

func handle(c net.Conn) {
	defer c.Close()
	tc := c.(*tls.Conn)
	log.Printf("--- connection from %s ---", c.RemoteAddr())
	if err := tc.Handshake(); err != nil {
		log.Printf("handshake failed: %v", err)
		return
	}
	st := tc.ConnectionState()
	log.Printf("TLS up version=%#x cipher=%#x", st.Version, st.CipherSuite)
	buf := make([]byte, 8192)
	tc.SetReadDeadline(time.Now().Add(60 * time.Second))
	for {
		n, err := tc.Read(buf)
		if n > 0 {
			log.Printf("client->server %d bytes:\n%s", n, hexdump(buf[:n]))
			if msg, derr := m3.Decode(buf[:n]); derr == nil {
				log.Printf("  m3.Decode OK: %d field(s): %+v", len(msg.Fields), msg.Fields)
			} else {
				log.Printf("  m3.Decode failed (expected until wire format aligned): %v", derr)
			}
		}
		if err != nil {
			log.Printf("read ended: %v", err)
			return
		}
		tc.SetReadDeadline(time.Now().Add(60 * time.Second))
	}
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
