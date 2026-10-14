package server

import (
	"context"
	"log/slog"
	"net"
	"sync"

	"github.com/athenanetworks/cmr-server/pkg/pkex"
	"github.com/athenanetworks/cmr-server/pkg/transport"
)

// EnrollConfig parameterizes the device/initiator side of enrollment.
type EnrollConfig struct {
	Code       []byte           // shared PKEX code (must match the controller)
	Identity   string           // this device's identity
	Addrs      [][]byte         // addresses to advertise to the controller (Msg4)
	Controller net.HardwareAddr // controller address, or transport.Broadcast to discover
	Suite      pkex.Suite       // defaults to pkex.ReferenceSuite
	Logger     *slog.Logger
}

// Enroller drives the client side (PKEXClntImpl) of a CMR enrollment. It is
// provided so an embedding project (or a device agent) can join a controller
// using the same protocol code path the controller validates against.
type Enroller struct {
	cfg  EnrollConfig
	tr   transport.Transport
	log  *slog.Logger
	sess *pkex.Session
}

// NewEnroller builds an enroller over tr.
func NewEnroller(cfg EnrollConfig, tr transport.Transport) *Enroller {
	if cfg.Logger == nil {
		cfg.Logger = slog.Default()
	}
	if cfg.Suite == nil {
		cfg.Suite = pkex.ReferenceSuite{}
	}
	if cfg.Controller == nil {
		cfg.Controller = transport.Broadcast
	}
	return &Enroller{cfg: cfg, tr: tr, log: cfg.Logger}
}

// Run performs one enrollment and returns when it completes, fails, ctx is
// cancelled, or the transport stops.
func (e *Enroller) Run(ctx context.Context) error {
	peer := append(net.HardwareAddr(nil), e.cfg.Controller...)
	send := func(t byte, b []byte) error {
		return e.tr.Send(transport.Frame{Peer: peer, Payload: append([]byte{t}, b...)})
	}

	done := make(chan error, 1)
	var once sync.Once
	finish := func(err error) { once.Do(func() { done <- err }) }

	e.sess = pkex.NewSession(pkex.Config{
		Role:    pkex.RoleClient,
		Code:    e.cfg.Code,
		LocalID: []byte(e.cfg.Identity),
		Addrs:   e.cfg.Addrs,
		Suite:   e.cfg.Suite,
		Logger:  e.log,
	}, peer.String(), send, func(pkex.Result) { finish(nil) })

	handler := func(f transport.Frame) {
		if len(f.Payload) < 1 {
			return
		}
		// Learn the controller's real address from its first reply if we began
		// with a broadcast discovery.
		if transport.IsBroadcast(peer) && f.Peer != nil {
			peer = append(net.HardwareAddr(nil), f.Peer...)
		}
		if err := e.sess.Handle(f.Payload[0], f.Payload[1:]); err != nil {
			finish(err)
			return
		}
		if e.sess.State() == pkex.StateDone {
			finish(nil)
		}
	}

	errc := make(chan error, 1)
	go func() { errc <- e.tr.Start(ctx, handler) }()

	if err := e.sess.Start(); err != nil {
		return err
	}
	select {
	case <-ctx.Done():
		return ctx.Err()
	case err := <-done:
		return err
	case err := <-errc:
		return err
	}
}
