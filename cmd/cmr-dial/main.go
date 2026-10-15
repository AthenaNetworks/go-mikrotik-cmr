// Command cmr-dial connects to a live MikroTik CMR controller as a client,
// using the minimal cert-tolerant TLS client in pkg/mtls, and logs/decodes the
// controller's m3 replies. It is the inverse of cmr-learn: cmr-learn is the
// controller side, cmr-dial is the enrollee side, used to observe what a real
// controller sends so our server can reply with byte parity.
//
// Authorized interoperability research against your own lab devices only.
package main

import (
	"crypto/ed25519"
	"crypto/rand"
	"crypto/sha256"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/hex"
	"flag"
	"fmt"
	"log"
	"math/big"
	"time"

	"github.com/athenanetworks/cmr-server/pkg/m3"
	"github.com/athenanetworks/cmr-server/pkg/mtls"
)

func main() {
	addr := flag.String("connect", "10.99.0.113:54321", "controller address")
	frames := flag.String("frames", "11dcd21101d8707eae265b81f4c7,09e5b1d41bce15ad35d84b60be6a01621a071dd3434852,31d8b1d4d0893efe1d07,09d8b1d41508f0fc1d07", "comma-separated client frames (hex) to send in order")
	drain := flag.Duration("drain", 3*time.Second, "how long to wait for replies after each send")
	wait := flag.Duration("wait", 10*time.Second, "final wait for replies")
	withCert := flag.Bool("cert", true, "present an Ed25519 client certificate (mutual TLS)")
	flag.Parse()

	cfg := &mtls.Config{}
	if *withCert {
		der, key, cn := identity()
		cfg.ClientCertDER, cfg.ClientKey = der, key
		log.Printf("client identity CN=%s", cn)
	}
	conn, err := mtls.DialConfig(*addr, 10*time.Second, cfg)
	if err != nil {
		log.Fatalf("dial %s: %v", *addr, err)
	}
	defer conn.Close()
	log.Printf("TLS up to %s (server requested client cert: %v)", *addr, conn.CertRequested)
	if len(conn.ServerCert) > 0 {
		log.Printf("server leaf cert: %d bytes (DER) sha-less dump below)", len(conn.ServerCert))
		log.Printf("  %s", hex.EncodeToString(conn.ServerCert))
	}

	// Send each client frame, draining server replies after each.
	for _, h := range splitNonEmpty(*frames) {
		b, err := hex.DecodeString(h)
		if err != nil {
			log.Fatalf("bad frame hex %q: %v", h, err)
		}
		dump("client->server", b)
		conn.SetDeadline(time.Now().Add(5 * time.Second))
		if err := conn.WriteFrame(b); err != nil {
			log.Printf("send failed: %v", err)
			return
		}
		drainReplies(conn, *drain)
	}
	// Final wait for any trailing server frames.
	drainReplies(conn, *wait)
	log.Printf("done")
}

func drainReplies(conn *mtls.Conn, d time.Duration) {
	deadline := time.Now().Add(d)
	for {
		conn.SetDeadline(deadline)
		frame, err := conn.ReadFrame()
		if err != nil {
			if time.Now().Before(deadline) {
				log.Printf("read ended: %v", err)
			}
			return
		}
		dump("server->client", frame)
	}
}

func splitNonEmpty(s string) []string {
	var out []string
	for _, p := range splitComma(s) {
		if p != "" {
			out = append(out, p)
		}
	}
	return out
}

func splitComma(s string) []string {
	var out []string
	cur := ""
	for _, r := range s {
		if r == ',' {
			out = append(out, cur)
			cur = ""
		} else {
			cur += string(r)
		}
	}
	return append(out, cur)
}

// identity builds a self-signed Ed25519 cert like a MikroTik figman identity.
func identity() ([]byte, ed25519.PrivateKey, string) {
	pub, priv, _ := ed25519.GenerateKey(rand.Reader)
	spki, _ := x509.MarshalPKIXPublicKey(pub)
	id := sha256.Sum256(spki)
	cn := "figman-" + hex.EncodeToString(id[:2])
	tmpl := &x509.Certificate{
		SerialNumber:          big.NewInt(time.Now().UnixNano()),
		Subject:               pkix.Name{CommonName: cn},
		Issuer:                pkix.Name{CommonName: cn},
		NotBefore:             time.Unix(0, 0),
		NotAfter:              time.Date(9999, 12, 31, 23, 59, 59, 0, time.UTC),
		IsCA:                  true,
		BasicConstraintsValid: true,
	}
	der, err := x509.CreateCertificate(rand.Reader, tmpl, tmpl, pub, priv)
	if err != nil {
		log.Fatalf("make cert: %v", err)
	}
	return der, priv, cn
}

func dump(dir string, b []byte) {
	fmt.Printf("\n=== %s (%d bytes) ===\n%s\n", dir, len(b), hexdump(b))
	if msg, err := m3.Decode(b); err == nil {
		printMsg(msg, 1)
	} else {
		fmt.Printf("  (m3 decode: %v)\n", err)
	}
}

func printMsg(m *m3.Message, depth int) {
	ind := ""
	for i := 0; i < depth; i++ {
		ind += "  "
	}
	for _, f := range m.Fields {
		asc := ""
		for _, c := range f.Value {
			if c >= 32 && c < 127 {
				asc += string(c)
			} else {
				asc += "."
			}
		}
		fmt.Printf("%sid=%-3d type=%d len=%-3d %s |%s|\n", ind, f.ID, f.Type, len(f.Value), hex.EncodeToString(f.Value), asc)
		// Only descend when the value cleanly round-trips as a sub-message with
		// more than one field (avoids spurious recursion into opaque blobs).
		if len(f.Value) >= 6 {
			if sub, err := m3.Decode(f.Value); err == nil && len(sub.Fields) > 1 && bytesEq(sub.Encode(), f.Value) {
				printMsg(sub, depth+1)
			}
		}
	}
}

func hexdump(b []byte) string {
	var out string
	for i := 0; i < len(b); i += 16 {
		end := i + 16
		if end > len(b) {
			end = len(b)
		}
		hexs := ""
		for _, ch := range b[i:end] {
			hexs += fmt.Sprintf("%02x ", ch)
		}
		out += "  " + hexs + "\n"
	}
	return out
}

func bytesEq(a, b []byte) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}
