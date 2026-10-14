// Command cmr-server runs a standalone CMR controller: it accepts PKEX
// enrollment from devices over the built-in UDP transport and keeps them in an
// in-memory store, optionally exposing a read-only HTTP status endpoint.
//
// This is a clean-room reimplementation derived from static analysis of
// MikroTik's cmr binary; see docs/PROTOCOL.md for what is verified vs. pending.
// The bundled crypto suite and UDP transport interoperate with this project's
// own client (pkg/server.Enroller), not (yet) with MikroTik devices.
package main

import (
	"context"
	"encoding/json"
	"flag"
	"hash/fnv"
	"log/slog"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/athenanetworks/cmr-server/pkg/model"
	"github.com/athenanetworks/cmr-server/pkg/pkex"
	"github.com/athenanetworks/cmr-server/pkg/server"
	"github.com/athenanetworks/cmr-server/pkg/transport"
)

func main() {
	var (
		listen   = flag.String("listen", "0.0.0.0:14700", "UDP address to listen on")
		code     = flag.String("code", "", "shared PKEX enrollment code (required)")
		identity = flag.String("identity", hostnameOr("cmr-controller"), "controller identity")
		httpAddr = flag.String("http", "", "optional HTTP status address, e.g. 127.0.0.1:8080")
		verbose  = flag.Bool("v", false, "verbose logging")
	)
	flag.Parse()

	level := slog.LevelInfo
	if *verbose {
		level = slog.LevelDebug
	}
	log := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: level}))

	if *code == "" {
		log.Error("missing -code (shared PKEX enrollment code)")
		os.Exit(2)
	}

	localMAC := transport.SynthMAC(uint16(fnvHash(*identity)))
	tr, err := transport.NewUDP(localMAC, *listen, "")
	if err != nil {
		log.Error("failed to create transport", "err", err)
		os.Exit(1)
	}
	defer tr.Close()

	store := model.NewStore()
	store.Watch(func(e model.Event) {
		log.Debug("store event", "type", e.Type, "collection", e.Collection, "id", e.ID)
	})

	srv := server.New(server.Config{
		Code:     []byte(*code),
		Identity: *identity,
		Logger:   log,
		Hooks: server.Hooks{
			OnEnroll: func(d *model.Device, _ pkex.Result) {
				log.Info("enrolled device", "id", d.ID, "mac", d.MACString(), "ip", d.IP, "peer_id", d.PeerID)
			},
			OnSessionError: func(peer net.HardwareAddr, err error) {
				log.Warn("enrollment error", "peer", peer, "err", err)
			},
		},
	}, store, tr)

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	if *httpAddr != "" {
		go serveHTTP(ctx, *httpAddr, store, log)
	}

	log.Info("cmr-server listening", "listen", *listen, "identity", *identity, "mac", localMAC)
	if err := srv.Run(ctx); err != nil && ctx.Err() == nil {
		log.Error("server stopped", "err", err)
		os.Exit(1)
	}
	log.Info("shutting down", "devices", len(store.Devices()))
}

func serveHTTP(ctx context.Context, addr string, store *model.Store, log *slog.Logger) {
	mux := http.NewServeMux()
	mux.HandleFunc("/healthz", func(w http.ResponseWriter, _ *http.Request) { w.Write([]byte("ok\n")) })
	mux.HandleFunc("/devices", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(store.Devices())
	})
	srv := &http.Server{Addr: addr, Handler: mux, ReadHeaderTimeout: 5 * time.Second}
	go func() {
		<-ctx.Done()
		sctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()
		_ = srv.Shutdown(sctx)
	}()
	log.Info("http status endpoint", "addr", addr)
	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Warn("http server stopped", "err", err)
	}
}

func hostnameOr(def string) string {
	if h, err := os.Hostname(); err == nil && h != "" {
		return h
	}
	return def
}

func fnvHash(s string) uint32 {
	h := fnv.New32a()
	_, _ = h.Write([]byte(s))
	return h.Sum32()
}
