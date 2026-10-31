package controller

import (
	"crypto/ed25519"
	"crypto/rand"
	"crypto/sha256"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/hex"
	"math/big"
	"os"
	"time"
)

// Identity returns the controller's TLS identity as a self-signed Ed25519
// certificate. If keyfile is non-empty, the private key seed is loaded from it,
// or generated and written there (0600) if absent, so the device-pinned
// fingerprint survives restarts. An empty keyfile yields an ephemeral identity.
func Identity(keyfile string) (tls.Certificate, error) {
	var pub ed25519.PublicKey
	var priv ed25519.PrivateKey
	if keyfile != "" {
		if seed, err := os.ReadFile(keyfile); err == nil && len(seed) == ed25519.SeedSize {
			priv = ed25519.NewKeyFromSeed(seed)
			pub = priv.Public().(ed25519.PublicKey)
		}
	}
	if priv == nil {
		var err error
		pub, priv, err = ed25519.GenerateKey(rand.Reader)
		if err != nil {
			return tls.Certificate{}, err
		}
		if keyfile != "" {
			if err := os.WriteFile(keyfile, priv.Seed(), 0o600); err != nil {
				return tls.Certificate{}, err
			}
		}
	}
	spki, err := x509.MarshalPKIXPublicKey(pub)
	if err != nil {
		return tls.Certificate{}, err
	}
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
	if err != nil {
		return tls.Certificate{}, err
	}
	leaf, _ := x509.ParseCertificate(der)
	return tls.Certificate{Certificate: [][]byte{der}, PrivateKey: priv, Leaf: leaf}, nil
}

// Fingerprint returns the hex SHA-256 of the certificate's SPKI — the value a
// device pins for the controller. Useful for logging which identity is in use.
func Fingerprint(cert tls.Certificate) string {
	leaf := cert.Leaf
	if leaf == nil && len(cert.Certificate) > 0 {
		leaf, _ = x509.ParseCertificate(cert.Certificate[0])
	}
	if leaf == nil {
		return ""
	}
	spki, err := x509.MarshalPKIXPublicKey(leaf.PublicKey)
	if err != nil {
		return ""
	}
	sum := sha256.Sum256(spki)
	return hex.EncodeToString(sum[:])
}
