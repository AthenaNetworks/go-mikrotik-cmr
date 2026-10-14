// Package server is the standalone CMR controller: it accepts PKEX enrollment
// from devices over a pluggable transport and records them in a model.Store,
// exposing hooks so other projects can authorize peers, supply per-peer codes,
// and react to enrollment without forking this code.
package server

import (
	"context"
	"fmt"
	"log/slog"
	"net"
	"sync"
	"time"

	"github.com/athenanetworks/cmr-server/pkg/m3"
	"github.com/athenanetworks/cmr-server/pkg/model"
	"github.com/athenanetworks/cmr-server/pkg/pkex"
	"github.com/athenanetworks/cmr-server/pkg/transport"
)

// Hooks are optional callbacks an embedding project provides. All are nil-safe.
type Hooks struct {
	// Authorize decides whether a peer may begin enrollment. Nil allows all.
	Authorize func(peer net.HardwareAddr) bool
	// CodeFor returns the PKEX code to use for a specific peer. Nil (or a nil
	// return) falls back to Config.Code.
	CodeFor func(peer net.HardwareAddr) []byte
	// OnEnroll fires after a device is added to the store on success.
	OnEnroll func(dev *model.Device, res pkex.Result)
	// OnSessionError fires when a peer handshake fails.
	OnSessionError func(peer net.HardwareAddr, err error)
}

// Config parameterizes the controller.
type Config struct {
	Code     []byte     // default shared PKEX code
	Identity string     // controller identity (hashed into our peer id)
	Suite    pkex.Suite // crypto suite; defaults to pkex.ReferenceSuite
	Logger   *slog.Logger
	Hooks    Hooks
	Now      func() time.Time // injectable clock; defaults to time.Now
}

// Server is a CMR controller instance.
type Server struct {
	cfg   Config
	store *model.Store
	tr    transport.Transport
	log   *slog.Logger
	now   func() time.Time

	mu       sync.Mutex
	sessions map[string]*pkex.Session // keyed by peer MAC string
}

// New builds a controller over store and tr.
func New(cfg Config, store *model.Store, tr transport.Transport) *Server {
	if cfg.Logger == nil {
		cfg.Logger = slog.Default()
	}
	if cfg.Suite == nil {
		cfg.Suite = pkex.ReferenceSuite{}
	}
	if cfg.Now == nil {
		cfg.Now = time.Now
	}
	return &Server{
		cfg:      cfg,
		store:    store,
		tr:       tr,
		log:      cfg.Logger,
		now:      cfg.Now,
		sessions: map[string]*pkex.Session{},
	}
}

// Store returns the controller's managed-object store (for queries and hooks).
func (s *Server) Store() *model.Store { return s.store }

// Run starts serving until ctx is cancelled.
func (s *Server) Run(ctx context.Context) error {
	s.log.Info("cmr controller starting", "identity", s.cfg.Identity, "suite", s.cfg.Suite.Name(), "local_mac", s.tr.LocalMAC())
	return s.tr.Start(ctx, s.onFrame)
}

// onFrame routes an inbound frame to the per-peer PKEX session, creating a new
// server-role session when a peer opens with a DiscoveryReq.
func (s *Server) onFrame(f transport.Frame) {
	if len(f.Payload) < 1 {
		return
	}
	ft, body := f.Payload[0], f.Payload[1:]
	key := f.Peer.String()

	s.mu.Lock()
	sess := s.sessions[key]
	if sess == nil {
		if ft != pkex.FrameDiscoveryReq {
			s.mu.Unlock()
			s.log.Debug("ignoring frame from unknown peer", "peer", key, "type", pkex.FrameName(ft))
			return
		}
		if s.cfg.Hooks.Authorize != nil && !s.cfg.Hooks.Authorize(f.Peer) {
			s.mu.Unlock()
			s.log.Warn("enrollment denied by hook", "peer", key)
			return
		}
		sess = s.newSession(f.Peer)
		s.sessions[key] = sess
	}
	s.mu.Unlock()

	if err := sess.Handle(ft, body); err != nil {
		s.log.Warn("frame handling failed", "peer", key, "type", pkex.FrameName(ft), "err", err)
		if s.cfg.Hooks.OnSessionError != nil {
			s.cfg.Hooks.OnSessionError(f.Peer, err)
		}
	}
	if st := sess.State(); st == pkex.StateDone || st == pkex.StateFailed {
		s.mu.Lock()
		delete(s.sessions, key)
		s.mu.Unlock()
	}
}

func (s *Server) newSession(peer net.HardwareAddr) *pkex.Session {
	p := append(net.HardwareAddr(nil), peer...)
	code := s.cfg.Code
	if s.cfg.Hooks.CodeFor != nil {
		if c := s.cfg.Hooks.CodeFor(p); c != nil {
			code = c
		}
	}
	send := func(t byte, b []byte) error {
		return s.tr.Send(transport.Frame{Peer: p, Payload: append([]byte{t}, b...)})
	}
	return pkex.NewSession(pkex.Config{
		Role:    pkex.RoleServer,
		Code:    code,
		LocalID: []byte(s.cfg.Identity),
		Suite:   s.cfg.Suite,
		Logger:  s.log,
	}, p.String(), send, func(r pkex.Result) { s.onEnrolled(p, r) })
}

func (s *Server) onEnrolled(peer net.HardwareAddr, r pkex.Result) {
	now := s.now()
	dev, existed := s.store.DeviceByMAC(peer.String())
	if !existed {
		dev = model.Device{ID: s.store.NewID(), FirstSeen: now, IsMikrotik: true}
	}
	dev.MAC = peer
	dev.PeerID = fmt.Sprintf("%x", r.PeerID)
	dev.Enrolled = true
	dev.LastSeen = now
	// Addresses advertised in Msg4: use the first parseable one as the IP.
	for _, a := range r.Addrs {
		if ip := net.ParseIP(string(a)); ip != nil {
			dev.IP = ip
			break
		}
	}
	d := dev
	s.store.PutDevice(&d)
	s.store.AddAlert(&model.Alert{ID: s.store.NewID(), Kind: "enroll", Device: d.ID, Message: "device enrolled", Time: now})
	s.log.Info("device enrolled", "mac", peer, "id", d.ID, "peer_id", dev.PeerID)
	if s.cfg.Hooks.OnEnroll != nil {
		s.cfg.Hooks.OnEnroll(&d, r)
	}
}

// topicQuery mirrors the startup "fetching topic names" getall the binary
// issues; exposed so an embedding project can reproduce bus-style discovery if
// it bridges to a real RouterOS. It is not used by the standalone server.
func topicQuery() []byte {
	return m3.New().PutU32(m3.IDSel, 3).PutU32(m3.IDTopic, m3.IDPkgs).Encode()
}

var _ = topicQuery
