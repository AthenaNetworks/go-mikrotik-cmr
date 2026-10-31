// Package controller implements the controller side of MikroTik's CMR protocol
// as an embeddable library: a stock MikroTik CMR device pairs and connects to a
// Server with no MikroTik/Nova controller in the path.
//
// It implements the OPEN pairing flow (controller pairing-requirement=none):
// TLS 1.2 on tcp/54321 with a self-signed Ed25519 identity, the nv/m3 control
// plane bring-up, and an id10 SUCCESS (status=0x00) that promotes the device to
// "paired" — no PKEX / password handshake. The frame shapes and sequencing were
// reconstructed byte-for-byte from live open-pairing captures and verified
// against pkg/m3 (see docs/re and the byte-exact tests in frames_test.go).
//
// Minimal server:
//
//	cert, _ := controller.Identity("ctrl.ed25519.seed")
//	srv, _ := controller.New(controller.Config{
//		Certificate: cert,
//		OnPair: func(s *controller.Session) {
//			s.WriteConfig(cmr.BuildVLANConfig(cmr.VLANConfig{VLANID: 100}))
//		},
//		OnObject: func(s *controller.Session, o controller.Object) {
//			log.Printf("%s reported %s", s.Peer(), o.Key)
//		},
//	})
//	srv.ListenAndServe("0.0.0.0:54321")
package controller

import (
	"crypto/tls"
	"errors"
	"io"
	"log/slog"
	"net"
	"sync"
	"time"
)

// m3 field ids on the CMR control channel.
const (
	idHello  = 2  // hello frame opcode
	idVer    = 1  // version / nonce-seq / verAck opcode (also nested inside verAck)
	idDisc   = 6  // discovery + discovery-ack opcode
	idResync = 7  // discovery resync / getall / config-write opcode
	idDev    = 9  // controller->device object row
	idSucc   = 10 // pairing SUCCESS/status opcode
	idDesc   = 0  // device descriptor / error / config object opcode

	idQuery  = 4  // query/selector body nested inside a discovery/resync/getall
	idSub    = 11 // subscribe opcode (nv::subscribe)
	idScope  = 23 // scope string, e.g. "com.mikrotik.cmr"
	idStatus = 27 // status byte (0x00=success, 0x02=password required, 0x83=error)
	idBoard  = 24 // board identity string (in version)
	idBuildV = 25 // build-time + version
	idNonce  = 22 // per-message nonce
	idSeq    = 21 // session sequence / object handle
	idDevFl  = 29 // object/query flag
	idDevIx  = 64 // object row index
	idBoard2 = 76 // board string (in a device row)

	idKey = 74 // object key (the typed attribute blob)
	idVal = 75 // object value

	nonceSize = 4
)

// Defaults for the config-write direction.
const (
	DefaultWriteScope = "com.mikrotik.cmr.config" // the device's own config scope
	DefaultWriteFlag  = 0x1c                      // id29 flag a real controller uses for a config write
	DefaultWriteDelay = 6 * time.Second           // let the device reach steady state before writing
)

// Config configures a Server. Only Certificate is required.
type Config struct {
	// Certificate is the controller's TLS identity. Use Identity to load or
	// create a persistent one so the device-pinned fingerprint survives restarts.
	Certificate tls.Certificate

	// IdleTimeout closes a connection after this long with no data from the
	// device. Zero means never (hold the control channel open).
	IdleTimeout time.Duration

	// Logger receives structured logs; nil discards them. Lifecycle, pairing and
	// read/write events log at Info; per-frame detail at Debug.
	Logger *slog.Logger

	// Roots is the controller's root object blob (id9 rows) streamed as part of
	// the SUCCESS frame — this is what makes a paired device begin its state
	// dump. Empty means SUCCESS is sent alone. See docs/samples/controller-roots.hex.
	Roots []byte

	// ConfigBurst is the controller's config object blob (id9 rows) streamed in
	// answer to the device's getall for the controller scope. Empty means the
	// getall is not answered. See docs/samples/controller-config.hex.
	ConfigBurst []byte

	// AutoWrite, if set, is an id30 config object graph (e.g. from
	// cmr.BuildVLANConfig) pushed into each device's config scope once, WriteDelay
	// after it pairs. For dynamic control use Session.WriteConfig instead.
	AutoWrite  []byte
	WriteScope string        // config-write scope; empty = DefaultWriteScope
	WriteFlag  byte          // config-write id29 flag; zero = DefaultWriteFlag
	WriteDelay time.Duration // delay after pairing before AutoWrite; zero = DefaultWriteDelay

	// Store persists the device registry (enrolments + each device's desired
	// config) across restarts. When set — or when DeviceKey is set — the server is
	// registry-driven: it identifies devices by a stable key, tracks them
	// (Devices/Device) and pushes each device's DesiredConfig on connect, instead
	// of the apply-to-all AutoWrite. Nil keeps the simple apply-to-all mode.
	Store Store

	// DeviceKey derives a stable DeviceID from a reported object; nil uses the
	// first MAC the device reports. Setting it also enables the registry.
	DeviceKey func(Object) (DeviceID, bool)

	// Callbacks (all optional) run on the device's connection goroutine, so they
	// must not block for long. Interact with the device through Session.
	OnConnect    func(*Session)         // device connected, TLS up (not yet paired)
	OnPair       func(*Session)         // device paired (SUCCESS sent)
	OnObject     func(*Session, Object) // device reported an object (read path)
	OnIdentify   func(*Session, Device) // device identified in the registry (first time this connection)
	OnDisconnect func(*Session, error)  // connection ended
}

func (c *Config) writeScope() string {
	if c.WriteScope == "" {
		return DefaultWriteScope
	}
	return c.WriteScope
}

func (c *Config) writeFlag() byte {
	if c.WriteFlag == 0 {
		return DefaultWriteFlag
	}
	return c.WriteFlag
}

func (c *Config) writeDelay() time.Duration {
	if c.WriteDelay == 0 {
		return DefaultWriteDelay
	}
	return c.WriteDelay
}

// Server is a CMR controller. Create it with New and run it with ListenAndServe
// or Serve. A Server handles many devices concurrently, one Session each.
type Server struct {
	cfg Config
	log *slog.Logger
	tls *tls.Config

	mu   sync.Mutex
	ln   net.Listener
	sess map[*Session]struct{}
	done bool

	devMu   sync.Mutex
	devices map[DeviceID]*Device
}

// New validates cfg, loads the device registry (if a Store is configured) and
// returns a Server ready to serve.
func New(cfg Config) (*Server, error) {
	if len(cfg.Certificate.Certificate) == 0 {
		return nil, errors.New("controller: Config.Certificate is required (see Identity)")
	}
	log := cfg.Logger
	if log == nil {
		log = slog.New(slog.NewTextHandler(io.Discard, nil))
	}
	s := &Server{
		cfg: cfg,
		log: log,
		tls: &tls.Config{
			Certificates: []tls.Certificate{cfg.Certificate},
			MinVersion:   tls.VersionTLS12,
			MaxVersion:   tls.VersionTLS12,
			// MikroTik's client cert carries an empty signatureAlgorithm OID that
			// crypto/x509 rejects and that cannot be rewritten on the wire without
			// breaking the Finished MAC, so trust is the app-layer pinned
			// fingerprint, not a TLS-layer client cert.
			ClientAuth: tls.NoClientCert,
		},
		sess:    make(map[*Session]struct{}),
		devices: make(map[DeviceID]*Device),
	}
	if err := s.loadRegistry(); err != nil {
		return nil, err
	}
	return s, nil
}

// ListenAndServe listens on addr (TLS 1.2) and serves until Close.
func (s *Server) ListenAndServe(addr string) error {
	ln, err := net.Listen("tcp", addr)
	if err != nil {
		return err
	}
	return s.Serve(ln)
}

// Serve accepts connections on ln until Close. A plain (non-TLS) connection is
// wrapped with the server's TLS config, so ln may be an ordinary net.Listener.
func (s *Server) Serve(ln net.Listener) error {
	s.mu.Lock()
	if s.done {
		s.mu.Unlock()
		return errors.New("controller: server closed")
	}
	s.ln = ln
	s.mu.Unlock()

	s.log.Info("cmr controller listening", "addr", ln.Addr().String(), "tls", "1.2", "pairing", "open")
	for {
		c, err := ln.Accept()
		if err != nil {
			s.mu.Lock()
			done := s.done
			s.mu.Unlock()
			if done {
				return nil
			}
			s.log.Warn("accept failed", "err", err)
			continue
		}
		tc, ok := c.(*tls.Conn)
		if !ok {
			tc = tls.Server(c, s.tls)
		}
		go s.handle(tc)
	}
}

// Sessions returns a snapshot of the currently connected device sessions.
func (s *Server) Sessions() []*Session {
	s.mu.Lock()
	defer s.mu.Unlock()
	out := make([]*Session, 0, len(s.sess))
	for sess := range s.sess {
		out = append(out, sess)
	}
	return out
}

// Close stops accepting new connections and closes the listener. Existing
// device connections are left to drain (close them via their own lifecycle).
func (s *Server) Close() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.done {
		return nil
	}
	s.done = true
	if s.ln != nil {
		return s.ln.Close()
	}
	return nil
}

func (s *Server) track(sess *Session) {
	s.mu.Lock()
	s.sess[sess] = struct{}{}
	s.mu.Unlock()
}

func (s *Server) untrack(sess *Session) {
	s.mu.Lock()
	delete(s.sess, sess)
	s.mu.Unlock()
}
