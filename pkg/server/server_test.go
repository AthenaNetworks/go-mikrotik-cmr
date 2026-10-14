package server

import (
	"context"
	"io"
	"log/slog"
	"testing"
	"time"

	"github.com/athenanetworks/cmr-server/pkg/model"
	"github.com/athenanetworks/cmr-server/pkg/pkex"
	"github.com/athenanetworks/cmr-server/pkg/transport"
)

func quiet() *slog.Logger { return slog.New(slog.NewTextHandler(io.Discard, nil)) }

func TestEnrollOverUDP(t *testing.T) {
	code := []byte("lab-code")

	srvTr, err := transport.NewUDP(transport.SynthMAC(1), "127.0.0.1:0", "")
	if err != nil {
		t.Fatalf("server transport: %v", err)
	}
	defer srvTr.Close()

	store := model.NewStore()
	enrolled := make(chan *model.Device, 1)
	srv := New(Config{
		Code:     code,
		Identity: "controller-1",
		Logger:   quiet(),
		Hooks:    Hooks{OnEnroll: func(d *model.Device, _ pkex.Result) { enrolled <- d }},
	}, store, srvTr)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	go srv.Run(ctx)

	cliTr, err := transport.NewUDP(transport.SynthMAC(2), "127.0.0.1:0", srvTr.LocalAddr())
	if err != nil {
		t.Fatalf("client transport: %v", err)
	}
	defer cliTr.Close()

	en := NewEnroller(EnrollConfig{
		Code:       code,
		Identity:   "device-1",
		Addrs:      [][]byte{[]byte("10.0.0.9")},
		Controller: transport.Broadcast,
		Logger:     quiet(),
	}, cliTr)

	if err := en.Run(ctx); err != nil {
		t.Fatalf("enroll: %v", err)
	}

	select {
	case d := <-enrolled:
		if !d.Enrolled {
			t.Fatal("device not marked enrolled")
		}
		if d.IP == nil || d.IP.String() != "10.0.0.9" {
			t.Fatalf("advertised IP not recorded: %v", d.IP)
		}
		if d.PeerID == "" {
			t.Fatal("peer id not recorded")
		}
	case <-time.After(3 * time.Second):
		t.Fatal("timed out waiting for enrollment")
	}

	if n := len(store.Devices()); n != 1 {
		t.Fatalf("store has %d devices, want 1", n)
	}
}

func TestEnrollWrongCodeDenied(t *testing.T) {
	srvTr, _ := transport.NewUDP(transport.SynthMAC(3), "127.0.0.1:0", "")
	defer srvTr.Close()
	store := model.NewStore()
	srv := New(Config{Code: []byte("right"), Identity: "c", Logger: quiet()}, store, srvTr)
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	go srv.Run(ctx)

	cliTr, _ := transport.NewUDP(transport.SynthMAC(4), "127.0.0.1:0", srvTr.LocalAddr())
	defer cliTr.Close()
	en := NewEnroller(EnrollConfig{Code: []byte("wrong"), Identity: "d", Controller: transport.Broadcast, Logger: quiet()}, cliTr)

	err := en.Run(ctx)
	if err == nil {
		t.Fatal("enrollment succeeded with wrong code")
	}
	if n := len(store.Devices()); n != 0 {
		t.Fatalf("store has %d devices after failed enroll, want 0", n)
	}
}
