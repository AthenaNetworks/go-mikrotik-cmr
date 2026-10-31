package controller

import (
	"crypto/tls"
	"net"
	"testing"
	"time"

	"github.com/athenanetworks/cmr-server/pkg/m3"
)

// TestServerBringUp drives the full pairing handshake end-to-end: a mock device
// connects over TLS and walks hello -> version -> discovery -> post-discovery
// id1, and the server must reply with hello, version, disc-ack, then verAck +
// SUCCESS, capture the device's query flag, fire OnPair and track the session.
func TestServerBringUp(t *testing.T) {
	cert, err := Identity("")
	if err != nil {
		t.Fatal(err)
	}
	paired := make(chan *Session, 1)
	srv, err := New(Config{
		Certificate: cert,
		OnPair:      func(s *Session) { paired <- s },
	})
	if err != nil {
		t.Fatal(err)
	}
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer ln.Close()
	go srv.Serve(ln)

	conn, err := tls.Dial("tcp", ln.Addr().String(), &tls.Config{
		InsecureSkipVerify: true,
		MinVersion:         tls.VersionTLS12,
		MaxVersion:         tls.VersionTLS12,
	})
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()

	dec := m3.NewStreamDecoder()
	readFrame := func() m3.Field {
		t.Helper()
		conn.SetReadDeadline(time.Now().Add(3 * time.Second))
		for {
			if f, ok := dec.Next(); ok {
				return f
			}
			buf := make([]byte, 4096)
			n, err := conn.Read(buf)
			if n > 0 {
				dec.Feed(buf[:n])
			}
			if err != nil {
				t.Fatalf("read: %v", err)
			}
		}
	}
	write := func(b []byte) {
		t.Helper()
		if _, err := conn.Write(b); err != nil {
			t.Fatalf("write: %v", err)
		}
	}

	// 1. hello -> server hello
	write(m3.New().PutBytes(2, []byte{0xd2, 0x11, 0x01, 0xd8, 0x70, 0x7e, 0xae, 0x26, 0x5b, 0x81, 0xf4, 0xc7}).Encode())
	if f := readFrame(); f.ID != 2 {
		t.Fatalf("after hello got id%d, want id2", f.ID)
	}

	// 2. version (inner carries id25 build) -> server version
	ver := m3.New().PutBytes(22, []byte{1, 2, 3, 4}).PutBytes(25, make([]byte, 8)).PutBytes(24, []byte("CHR"))
	write(m3.New().PutNested(1, ver).Encode())
	if f := readFrame(); f.ID != 1 {
		t.Fatalf("after version got id%d, want id1", f.ID)
	}

	// 3. discovery (id6{ id21 seq, id4{ id29 flag }, id22 nonce }) -> disc ack
	disc := m3.New().PutBytes(21, []byte{0x34, 0x01}).
		PutNested(4, m3.New().Put(29, []byte{0x14})).
		PutBytes(22, []byte{9, 9, 9, 9})
	write(m3.New().PutNested(6, disc).Encode())
	if f := readFrame(); f.ID != 6 {
		t.Fatalf("after discovery got id%d, want id6", f.ID)
	}

	// 4. post-discovery id1 (no id25) -> verAck (id1) + SUCCESS (id10)
	post := m3.New().PutBytes(22, []byte{5, 6, 7, 8}).PutBytes(21, []byte{0x34, 0x01})
	write(m3.New().PutNested(1, post).Encode())
	got := map[uint32]bool{}
	for i := 0; i < 2; i++ {
		got[readFrame().ID] = true
	}
	if !got[1] || !got[10] {
		t.Fatalf("after pairing got frames %v, want id1 (verAck) + id10 (SUCCESS)", got)
	}

	select {
	case s := <-paired:
		if s.DevFlag() != 0x14 {
			t.Errorf("captured devFlag=%#x, want 0x14", s.DevFlag())
		}
		if !s.Paired() {
			t.Error("session not marked paired")
		}
	case <-time.After(3 * time.Second):
		t.Fatal("OnPair never fired")
	}
	if n := len(srv.Sessions()); n != 1 {
		t.Errorf("Sessions()=%d, want 1", n)
	}
}

// TestNewRequiresCertificate checks New rejects a config with no identity.
func TestNewRequiresCertificate(t *testing.T) {
	if _, err := New(Config{}); err == nil {
		t.Fatal("New accepted a Config with no Certificate")
	}
}

// TestConfigWriteDefaults checks the write-direction defaults are applied.
func TestConfigWriteDefaults(t *testing.T) {
	var c Config
	if c.writeScope() != DefaultWriteScope {
		t.Errorf("writeScope=%q", c.writeScope())
	}
	if c.writeFlag() != DefaultWriteFlag {
		t.Errorf("writeFlag=%#x", c.writeFlag())
	}
	if c.writeDelay() != DefaultWriteDelay {
		t.Errorf("writeDelay=%v", c.writeDelay())
	}
}
