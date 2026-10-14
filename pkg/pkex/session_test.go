package pkex

import (
	"errors"
	"testing"
)

// drive runs the two sessions against an in-memory frame queue until quiescent.
func drive(t *testing.T, client, server *Session, q *[]qmsg) error {
	t.Helper()
	for i := 0; len(*q) > 0 && i < 100; i++ {
		m := (*q)[0]
		*q = (*q)[1:]
		var err error
		if m.to == RoleServer {
			err = server.Handle(m.ft, m.body)
		} else {
			err = client.Handle(m.ft, m.body)
		}
		if err != nil {
			return err
		}
	}
	return nil
}

type qmsg struct {
	to   Role
	ft   byte
	body []byte
}

func newPair(code []byte, q *[]qmsg, onResult ResultFunc) (client, server *Session) {
	clientSend := func(ft byte, body []byte) error { *q = append(*q, qmsg{RoleServer, ft, body}); return nil }
	serverSend := func(ft byte, body []byte) error { *q = append(*q, qmsg{RoleClient, ft, body}); return nil }
	server = NewSession(Config{Role: RoleServer, Code: code, LocalID: []byte("controller")}, "client", serverSend, onResult)
	client = NewSession(Config{Role: RoleClient, Code: code, LocalID: []byte("device-aa:bb:cc"), Addrs: [][]byte{[]byte("10.0.0.5")}}, "server", clientSend, nil)
	return client, server
}

func TestHandshakeSucceeds(t *testing.T) {
	var q []qmsg
	var got *Result
	client, server := newPair([]byte("shared-code"), &q, func(r Result) { got = &r })

	if err := client.Start(); err != nil {
		t.Fatalf("start: %v", err)
	}
	if err := drive(t, client, server, &q); err != nil {
		t.Fatalf("handshake: %v", err)
	}
	if client.State() != StateDone {
		t.Fatalf("client state = %v, want done", client.State())
	}
	if server.State() != StateDone {
		t.Fatalf("server state = %v, want done", server.State())
	}
	if got == nil {
		t.Fatal("server produced no result")
	}
	if len(got.PeerID) != peerIDLen {
		t.Fatalf("peer id len = %d, want %d", len(got.PeerID), peerIDLen)
	}
	if len(got.Addrs) != 1 || string(got.Addrs[0]) != "10.0.0.5" {
		t.Fatalf("addrs = %v, want [10.0.0.5]", got.Addrs)
	}
}

func TestHandshakeWrongCodeFailsClosed(t *testing.T) {
	// Different codes on each side: the exchange must not complete.
	var q []qmsg
	clientSend := func(ft byte, body []byte) error { q = append(q, qmsg{RoleServer, ft, body}); return nil }
	serverSend := func(ft byte, body []byte) error { q = append(q, qmsg{RoleClient, ft, body}); return nil }
	server := NewSession(Config{Role: RoleServer, Code: []byte("code-A"), LocalID: []byte("controller")}, "client", serverSend, nil)
	client := NewSession(Config{Role: RoleClient, Code: []byte("code-B"), LocalID: []byte("device")}, "server", clientSend, nil)

	if err := client.Start(); err != nil {
		t.Fatalf("start: %v", err)
	}
	var sawErr error
	for i := 0; len(q) > 0 && i < 100; i++ {
		m := q[0]
		q = q[1:]
		var err error
		if m.to == RoleServer {
			err = server.Handle(m.ft, m.body)
		} else {
			err = client.Handle(m.ft, m.body)
		}
		if err != nil {
			sawErr = err
			break
		}
	}
	if server.State() == StateDone {
		t.Fatal("server completed enrollment with mismatched code")
	}
	if sawErr == nil {
		t.Fatal("expected an error with mismatched code, got none")
	}
	// Most commonly surfaces as auth failure or an off-curve recovered point.
	if !errors.Is(sawErr, ErrAuthFailed) && sawErr.Error() != "pkex: recovered peer point not on curve" {
		t.Logf("note: failure surfaced as: %v", sawErr)
	}
}
