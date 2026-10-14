package pkex

import (
	"io"
	"log/slog"
)

// SendFunc transmits one frame body under the given frame type to the peer.
type SendFunc func(frameType byte, body []byte) error

// ResultFunc is called once when a session completes enrollment successfully.
type ResultFunc func(Result)

// Config parameterizes a Session.
type Config struct {
	Role    Role      // RoleClient or RoleServer
	Code    []byte    // shared PKEX code/password
	LocalID []byte    // local identity (hashed into the 32-byte peer id)
	Addrs   [][]byte  // addresses to advertise in Msg4 (client role)
	Suite   Suite     // crypto suite; defaults to ReferenceSuite
	Rng     io.Reader // randomness; defaults to crypto/rand
	Logger  *slog.Logger
}

// Session runs one peer handshake to completion. It is not safe for concurrent
// use; the owner (see pkg/server) serializes frames per peer.
type Session struct {
	cfg      Config
	ctx      Context
	state    State
	send     SendFunc
	onResult ResultFunc
	peer     string
	result   *Result
}

// NewSession builds a session for one peer. send transmits frames; onResult (if
// non-nil) fires on successful completion.
func NewSession(cfg Config, peer string, send SendFunc, onResult ResultFunc) *Session {
	if cfg.Suite == nil {
		cfg.Suite = ReferenceSuite{}
	}
	if cfg.Logger == nil {
		cfg.Logger = slog.Default()
	}
	return &Session{
		cfg:      cfg,
		ctx:      cfg.Suite.NewContext(cfg.Role, cfg.Code, cfg.LocalID, cfg.Rng),
		state:    StateInit,
		send:     send,
		onResult: onResult,
		peer:     peer,
	}
}

// State returns the current handshake state.
func (s *Session) State() State { return s.state }

// Result returns the completion result, if the session finished successfully.
func (s *Session) Result() (Result, bool) {
	if s.result != nil {
		return *s.result, true
	}
	return Result{}, false
}

// Start begins the client handshake by sending DiscoveryReq. It is a no-op for
// the server role (which waits for an inbound DiscoveryReq).
func (s *Session) Start() error {
	if s.cfg.Role != RoleClient {
		return nil
	}
	body, err := s.ctx.BuildDiscovery()
	if err != nil {
		return s.fail(err)
	}
	s.state = StateDiscovery
	return s.send(FrameDiscoveryReq, body)
}

// Handle processes one inbound frame and advances the state machine.
func (s *Session) Handle(frameType byte, body []byte) error {
	if s.state == StateDone || s.state == StateFailed {
		return ErrDone
	}
	if s.cfg.Role == RoleServer {
		return s.handleServer(frameType, body)
	}
	return s.handleClient(frameType, body)
}

func (s *Session) handleClient(frameType byte, body []byte) error {
	switch frameType {
	case FrameDiscoveryResp:
		if s.state != StateDiscovery {
			return s.unexpected(frameType)
		}
		if err := s.ctx.ConsumeDiscovery(body); err != nil {
			return s.fail(err)
		}
		s.cfg.Logger.Info("server found", "peer", s.peer)
		s.state = StateExchange
		return nil
	case FrameMsg1:
		if s.state != StateExchange {
			return s.unexpected(frameType)
		}
		if err := s.ctx.ConsumeMsg1(body); err != nil {
			return s.fail(err)
		}
		out, err := s.ctx.BuildMsg2()
		if err != nil {
			return s.fail(err)
		}
		s.state = StateAuth
		return s.send(FrameMsg2, out)
	case FrameMsg3:
		if s.state != StateAuth {
			return s.unexpected(frameType)
		}
		if err := s.ctx.ConsumeMsg3(body); err != nil {
			return s.fail(err)
		}
		out, err := s.ctx.BuildMsg4(s.cfg.Addrs)
		if err != nil {
			return s.fail(err)
		}
		if err := s.send(FrameMsg4, out); err != nil {
			return s.fail(err)
		}
		return s.finish(Result{PeerID: s.ctx.PeerID()})
	default:
		return s.unexpected(frameType)
	}
}

func (s *Session) handleServer(frameType byte, body []byte) error {
	switch frameType {
	case FrameDiscoveryReq:
		if s.state != StateInit {
			return s.unexpected(frameType)
		}
		if err := s.ctx.ConsumeDiscovery(body); err != nil {
			return s.fail(err)
		}
		s.cfg.Logger.Info("session started", "peer", s.peer)
		resp, err := s.ctx.BuildDiscovery()
		if err != nil {
			return s.fail(err)
		}
		if err := s.send(FrameDiscoveryResp, resp); err != nil {
			return s.fail(err)
		}
		out, err := s.ctx.BuildMsg1()
		if err != nil {
			return s.fail(err)
		}
		if err := s.send(FrameMsg1, out); err != nil {
			return s.fail(err)
		}
		s.state = StateExchange
		return nil
	case FrameMsg2:
		if s.state != StateExchange {
			return s.unexpected(frameType)
		}
		if err := s.ctx.ConsumeMsg2(body); err != nil {
			return s.fail(err)
		}
		out, err := s.ctx.BuildMsg3()
		if err != nil {
			return s.fail(err)
		}
		s.state = StateAuth
		return s.send(FrameMsg3, out)
	case FrameMsg4:
		if s.state != StateAuth {
			return s.unexpected(frameType)
		}
		id, addrs, err := s.ctx.ConsumeMsg4(body)
		if err != nil {
			return s.fail(err)
		}
		s.cfg.Logger.Info("discovery finished", "peer", s.peer)
		return s.finish(Result{PeerID: id, Addrs: addrs})
	default:
		return s.unexpected(frameType)
	}
}

func (s *Session) finish(r Result) error {
	s.state = StateDone
	s.result = &r
	if s.onResult != nil {
		s.onResult(r)
	}
	return nil
}

func (s *Session) fail(err error) error {
	s.state = StateFailed
	s.cfg.Logger.Warn("pkex session failed", "peer", s.peer, "err", err)
	return err
}

func (s *Session) unexpected(frameType byte) error {
	s.cfg.Logger.Warn("unhandled msg", "peer", s.peer, "type", FrameName(frameType), "state", s.state)
	return ErrUnexpectedFrame
}
