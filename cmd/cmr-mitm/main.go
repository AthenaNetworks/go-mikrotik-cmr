// Command cmr-mitm is a transparent CMR control-channel proxy used to observe
// a real pairing (the password "key handshake") in plaintext. It terminates
// TLS from the MikroTik client with its own Ed25519 identity, opens a second
// TLS session to the real controller via pkg/mtls (presenting its own client
// identity), and relays m3 frames in both directions, logging and decoding
// each one. Frame boundaries are preserved (one TLS record per m3 frame), which
// the boundary-delimited m3 framing requires.
//
// Authorized interoperability research against your own lab devices only. Note
// that because the proxy presents its own identities to each side, an
// identity-bound handshake may fail at the confirmation step; the request and
// challenge frames are still captured.
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
	"sync"
	"time"

	"github.com/athenanetworks/cmr-server/pkg/m3"
	"github.com/athenanetworks/cmr-server/pkg/mtls"
)

func main() {
	listen := flag.String("listen", "10.99.0.1:54321", "address to listen on (point the client here)")
	upstream := flag.String("upstream", "10.99.0.113:54321", "real controller address")
	flag.Parse()

	srvCert := identity("proxy-server")
	cfg := &tls.Config{
		Certificates: []tls.Certificate{srvCert},
		MinVersion:   tls.VersionTLS12,
		MaxVersion:   tls.VersionTLS12,
		ClientAuth:   tls.NoClientCert, // client connects without presenting a cert
	}
	ln, err := tls.Listen("tcp", *listen, cfg)
	if err != nil {
		log.Fatalf("listen: %v", err)
	}
	log.Printf("cmr-mitm listening on %s, relaying to %s", *listen, *upstream)

	clientDER, clientKey, cn := identityRaw("proxy-client")
	log.Printf("proxy client identity CN=%s", cn)

	for {
		c, err := ln.Accept()
		if err != nil {
			log.Printf("accept: %v", err)
			continue
		}
		go relay(c.(*tls.Conn), *upstream, clientDER, clientKey)
	}
}

func relay(cli *tls.Conn, upstream string, clientDER []byte, clientKey ed25519.PrivateKey) {
	defer cli.Close()
	peer := cli.RemoteAddr()
	log.Printf("=== client %s connected ===", peer)
	if err := cli.Handshake(); err != nil {
		log.Printf("[%s] client TLS failed: %v", peer, err)
		return
	}
	up, err := mtls.DialConfig(upstream, 10*time.Second, &mtls.Config{ClientCertDER: clientDER, ClientKey: clientKey})
	if err != nil {
		log.Printf("[%s] upstream dial failed: %v", peer, err)
		return
	}
	defer up.Close()
	log.Printf("[%s] upstream connected", peer)

	var wg sync.WaitGroup
	wg.Add(2)
	// client -> controller
	go func() {
		defer wg.Done()
		buf := make([]byte, 16384)
		for {
			cli.SetReadDeadline(time.Now().Add(120 * time.Second))
			n, err := cli.Read(buf)
			if n > 0 {
				frame := append([]byte(nil), buf[:n]...)
				dump(peer.String(), "client->controller", frame)
				if werr := up.WriteFrame(frame); werr != nil {
					log.Printf("[%s] write upstream: %v", peer, werr)
					return
				}
			}
			if err != nil {
				log.Printf("[%s] client read ended: %v", peer, err)
				up.Close()
				return
			}
		}
	}()
	// controller -> client
	go func() {
		defer wg.Done()
		for {
			up.SetDeadline(time.Now().Add(120 * time.Second))
			frame, err := up.ReadFrame()
			if len(frame) > 0 {
				dump(peer.String(), "controller->client", frame)
				if _, werr := cli.Write(frame); werr != nil {
					log.Printf("[%s] write client: %v", peer, werr)
					return
				}
			}
			if err != nil {
				log.Printf("[%s] upstream read ended: %v", peer, err)
				cli.Close()
				return
			}
		}
	}()
	wg.Wait()
	log.Printf("=== client %s closed ===", peer)
}

func dump(peer, dir string, b []byte) {
	fmt.Printf("\n[%s] %s (%d bytes)\n%s", peer, dir, len(b), hexdump(b))
	if msg, err := m3.Decode(b); err == nil {
		for _, f := range msg.Fields {
			asc := ascii(f.Value)
			fmt.Printf("    id=%-4d type=%d len=%-3d %s |%s|\n", f.ID, f.Type, len(f.Value), hex.EncodeToString(f.Value), asc)
			if len(f.Value) >= 6 {
				if sub, e := m3.Decode(f.Value); e == nil && len(sub.Fields) > 1 && eq(sub.Encode(), f.Value) {
					for _, sf := range sub.Fields {
						fmt.Printf("        id=%-4d type=%d len=%-3d %s |%s|\n", sf.ID, sf.Type, len(sf.Value), hex.EncodeToString(sf.Value), ascii(sf.Value))
					}
				}
			}
		}
	}
}

func identity(tag string) tls.Certificate {
	der, key, _ := identityRaw(tag)
	leaf, _ := x509.ParseCertificate(der)
	return tls.Certificate{Certificate: [][]byte{der}, PrivateKey: key, Leaf: leaf}
}

func identityRaw(tag string) ([]byte, ed25519.PrivateKey, string) {
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
		log.Fatalf("make %s cert: %v", tag, err)
	}
	return der, priv, cn
}

func ascii(b []byte) string {
	s := ""
	for _, c := range b {
		if c >= 32 && c < 127 {
			s += string(c)
		} else {
			s += "."
		}
	}
	return s
}

func eq(a, b []byte) bool {
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
		out += "    " + hexs + "\n"
	}
	return out
}
