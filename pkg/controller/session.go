package controller

import (
	"context"
	"crypto/tls"
	"encoding/hex"
	"errors"
	"log/slog"
	"net"
	"sync"
	"time"

	"github.com/athenanetworks/cmr-server/pkg/cmr"
	"github.com/athenanetworks/cmr-server/pkg/m3"
)

// Session is one device's control-channel connection. It is created by the
// Server; interact with the device through its methods, which are safe for
// concurrent use.
type Session struct {
	srv  *Server
	tc   *tls.Conn
	peer string
	in   *m3.StreamDecoder

	mu              sync.Mutex // serialises writes to tc and the mutable state below
	discSeq         []byte     // last discovery/resync outer id21, echoed in verAck
	devFlag         byte       // device's query flag; pushed objects must echo it
	paired          bool
	wroteCfg        bool           // the config push has happened this connection
	configScheduled bool           // the config-push timer has been started
	devID           DeviceID       // registry key, once identified (registry mode)
	descr           cmr.DeviceInfo // last descriptor the device reported

	objMu   sync.Mutex
	objects []Object
}

// Object is a device-reported management object (the read path): its handle and
// flags plus the decoded key/value content the device reports about itself.
type Object struct {
	FieldID  uint32     // the frame opcode that carried it (id0/id1/id9)
	Handle   []byte     // object handle (id21)
	Flag     []byte     // object flag (id29)
	Index    []byte     // object index (id64)
	Key      cmr.Object // decoded key object (class, strings, MACs)
	Value    cmr.Object // decoded value object (may be empty)
	RawKey   []byte
	RawValue []byte
}

// Peer returns the device's remote address (host:port).
func (s *Session) Peer() string { return s.peer }

// remoteIP returns the device's source IP (the host part of Peer, without the
// port). For a device behind NAT this is the NAT address, not the device's own.
func (s *Session) remoteIP() string {
	host, _, err := net.SplitHostPort(s.peer)
	if err != nil {
		return s.peer
	}
	return host
}

// DevFlag returns the device's query flag (0 until its discovery arrives).
func (s *Session) DevFlag() byte {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.devFlag
}

// Paired reports whether the device has been promoted to paired (SUCCESS sent).
func (s *Session) Paired() bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.paired
}

// DeviceID returns the session's registry key, or "" until the device is
// identified (only populated in registry mode).
func (s *Session) DeviceID() DeviceID {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.devID
}

// Device returns the identified device's registry record, if any.
func (s *Session) Device() (Device, bool) {
	id := s.DeviceID()
	if id == "" {
		return Device{}, false
	}
	return s.srv.Device(id)
}

// Objects returns a snapshot of the objects the device has reported so far.
func (s *Session) Objects() []Object {
	s.objMu.Lock()
	defer s.objMu.Unlock()
	return append([]Object(nil), s.objects...)
}

// Send writes raw control frames to the device under the write lock. The device
// decodes the controller->device direction as one continuous m3 stream, so the
// lock keeps frames from interleaving mid-stream. Frame builders such as
// GetAllQuery and SubscribeQuery produce suitable input.
func (s *Session) Send(frames ...[]byte) error {
	if len(frames) == 0 {
		return nil
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.sendLocked(frames)
}

// sendLocked writes frames; the caller holds s.mu.
func (s *Session) sendLocked(frames [][]byte) error {
	for _, out := range frames {
		s.srv.log.Debug("tx frame", "peer", s.peer, "len", len(out))
		if _, err := s.tc.Write(out); err != nil {
			s.srv.log.Warn("write failed", "peer", s.peer, "err", err)
			return err
		}
	}
	return nil
}

// WriteConfig pushes an id30 config object graph (e.g. from cmr.BuildVLANConfig)
// into the device's own config scope. The device reconciles its managed config
// to the graph. It uses the server's configured write scope and flag.
func (s *Session) WriteConfig(graph []byte) error {
	if len(graph) == 0 {
		return errors.New("controller: empty config graph")
	}
	frame := buildWriteConfig(s.srv.cfg.writeScope(), s.srv.cfg.writeFlag(), graph)
	s.srv.log.Info("config WRITE", "peer", s.peer, "scope", s.srv.cfg.writeScope(), "bytes", len(graph))
	return s.Send(frame)
}

// handle runs the full control-channel lifecycle for one device connection.
func (s *Server) handle(tc *tls.Conn) {
	defer tc.Close()
	sess := &Session{srv: s, tc: tc, peer: tc.RemoteAddr().String(), in: m3.NewStreamDecoder()}
	s.track(sess)
	defer s.untrack(sess)

	s.log.Info("device connected", "peer", sess.peer)
	if err := tc.Handshake(); err != nil {
		s.log.Warn("TLS handshake failed", "peer", sess.peer, "err", err)
		if s.cfg.OnDisconnect != nil {
			s.cfg.OnDisconnect(sess, err)
		}
		return
	}
	st := tc.ConnectionState()
	s.log.Info("TLS up", "peer", sess.peer, "version", st.Version, "cipher", st.CipherSuite)
	if s.cfg.OnConnect != nil {
		s.cfg.OnConnect(sess)
	}

	// Keep the control channel alive; the device holds "connected" as long as this
	// socket is up and stays silent when idle.
	if tcp, ok := tc.NetConn().(*net.TCPConn); ok {
		tcp.SetKeepAlive(true)
		tcp.SetKeepAlivePeriod(25 * time.Second)
	}

	buf := make([]byte, 16384)
	for {
		if s.cfg.IdleTimeout > 0 {
			tc.SetReadDeadline(time.Now().Add(s.cfg.IdleTimeout))
		} else {
			tc.SetReadDeadline(time.Time{})
		}
		n, err := tc.Read(buf)
		if n > 0 {
			sess.in.Feed(buf[:n])
			for {
				f, ok := sess.in.Next()
				if !ok {
					break
				}
				s.logFrame(sess, f)
				wasPaired := sess.Paired()
				frames := sess.step(f)
				if len(frames) > 0 {
					sess.mu.Lock()
					werr := sess.sendLocked(frames)
					sess.mu.Unlock()
					if werr != nil {
						break
					}
				}
				// Fire OnPair once, after the SUCCESS frame is on the wire, so a
				// callback that sends further frames (e.g. a getall) orders after it.
				if !wasPaired && sess.Paired() && s.cfg.OnPair != nil {
					s.cfg.OnPair(sess)
				}
			}
		}
		if err != nil {
			s.log.Info("connection ended", "peer", sess.peer, "err", err, "buffered", sess.in.Buffered())
			if id := sess.DeviceID(); id != "" {
				s.markDisconnected(id)
			}
			if s.cfg.OnDisconnect != nil {
				s.cfg.OnDisconnect(sess, err)
			}
			return
		}
	}
}

func (s *Server) logFrame(sess *Session, f m3.Field) {
	if !s.log.Enabled(context.Background(), slog.LevelDebug) {
		return
	}
	s.log.Debug("rx frame", "peer", sess.peer, "id", f.ID, "len", len(f.Value),
		"hex", hex.EncodeToString(f.Value), "dump", m3.Dump(f.Value, 6))
}

// step maps one top-level device field to the controller's reply frames.
func (s *Session) step(f m3.Field) [][]byte {
	srv := s.srv
	// Object-bearing fields (id0/id1/id9 whose body carries an id74 key) are the
	// device REPORTING state, not control frames — route them to the read path.
	if f.ID == idDesc || f.ID == idVer || f.ID == idDev {
		if body, _ := m3.Decode(f.Value); body != nil {
			if _, ok := body.GetBytes(idKey); ok {
				s.readObject(f, body)
				return nil
			}
		}
	}
	switch f.ID {
	case idHello: // C hello -> S hello
		return [][]byte{serverHello()}

	case idVer: // id1: version (has id25), the post-discovery nonce/seq, or an ack
		if inner, _ := m3.Decode(f.Value); inner != nil {
			if _, ok := inner.GetBytes(idBuildV); ok {
				return [][]byte{serverVersion()} // C version -> S version
			}
		}
		s.mu.Lock()
		if s.paired {
			s.mu.Unlock()
			// Already paired: a bare id1{nonce,seq} is a control ack / keepalive.
			// Absorb it — re-running the SUCCESS path would re-pair and storm.
			return nil
		}
		// First post-discovery id1 -> verAck, then id10 SUCCESS (promotion to
		// paired). SUCCESS alone pairs; roots ride the SUCCESS frame when known.
		s.paired = true
		discSeq := append([]byte(nil), s.discSeq...)
		devFlag := s.devFlag
		s.mu.Unlock()

		out := [][]byte{verAck(discSeq, 0x01)}
		if devFlag != 0 {
			synth := len(srv.cfg.Roots) == 0
			out = append(out, successWithRootsFrame(srv.cfg.rootsBlob(), devFlag))
			srv.log.Info("paired (SUCCESS+roots)", "peer", s.peer, "flag", devFlag, "synthesized", synth)
		} else {
			// No query flag yet: can't frame roots the device will accept.
			out = append(out, successFrame())
			srv.log.Info("paired (SUCCESS)", "peer", s.peer)
		}
		// OnPair fires from the read loop once these frames are on the wire. The
		// config push is scheduled here (simple/AutoWrite mode) or once the device
		// is identified (registry mode); see maybeScheduleConfig.
		s.maybeScheduleConfig()
		return out

	case idDisc, idResync: // C discovery/resync -> S ack; remember seq + flag
		if inner, _ := m3.Decode(f.Value); inner != nil {
			if b, ok := inner.GetBytes(idSeq); ok {
				s.mu.Lock()
				s.discSeq = append([]byte(nil), b...)
				s.mu.Unlock()
			}
			// The device's query carries an id29 flag inside its nested id4 that the
			// controller MUST echo in every object it pushes, or the device rejects
			// them "not found". Capture it here.
			if q, ok := inner.GetNested(idQuery); ok {
				if b, ok := q.GetBytes(idDevFl); ok && len(b) == 1 {
					s.mu.Lock()
					s.devFlag = b[0]
					s.mu.Unlock()
				}
			}
		}
		if f.ID == idResync {
			// A resync (id7) is the device's getall for the controller's scope; a
			// real controller answers it by streaming its config id9 burst, which is
			// what makes the device reciprocate with its own state dump. The burst is
			// synthesised from structure unless a capture is configured.
			burst := srv.cfg.configBurst()
			if len(burst) > 0 {
				flag := s.DevFlag()
				srv.log.Info("getall -> config burst", "peer", s.peer, "bytes", len(burst),
					"flag", flag, "synthesized", len(srv.cfg.ConfigBurst) == 0)
				return [][]byte{patchObjectFlag(burst, flag)}
			}
			return nil
		}
		return [][]byte{discAck(0x15)}

	case idDesc: // id0: device descriptor, config object, or an error report
		if body, _ := m3.Decode(f.Value); body != nil {
			if b, ok := body.GetBytes(idStatus); ok && len(b) == 1 && b[0] != 0 {
				srv.log.Info("device status report", "peer", s.peer, "status", b[0], "msg", strField(body, 28))
			}
			if cmr.IsDescriptor(body) {
				if d, err := cmr.ParseDescriptor(f.Value); err == nil {
					srv.log.Info("device descriptor", "peer", s.peer, "descriptor", d.String())
					s.mu.Lock()
					s.descr = d
					id := s.devID
					s.mu.Unlock()
					if id != "" { // enrich an already-identified record
						srv.upsertDevice(id, "", s.remoteIP(), d)
					}
				}
			}
		}
		return nil

	default:
		return nil
	}
}

// maybeScheduleConfig starts the one-shot config push once the device is ready:
// paired, its query flag known, and — in registry mode — identified. It is
// idempotent, so it can be called both at pairing and at identification.
func (s *Session) maybeScheduleConfig() {
	s.mu.Lock()
	ready := s.paired && !s.configScheduled && s.devFlag != 0
	if ready && s.srv.registryEnabled() {
		ready = s.devID != ""
	}
	if !ready {
		s.mu.Unlock()
		return
	}
	s.configScheduled = true
	s.mu.Unlock()
	go s.configTimer()
}

// configTimer waits for the device to reach steady state, then pushes its config
// exactly once per connection.
func (s *Session) configTimer() {
	time.Sleep(s.srv.cfg.writeDelay())
	graph := s.configToApply()
	if len(graph) == 0 {
		return
	}
	s.mu.Lock()
	already := s.wroteCfg
	s.mu.Unlock()
	if already {
		return
	}
	if err := s.applyConfig(graph); err != nil {
		s.srv.log.Warn("config push failed", "peer", s.peer, "err", err)
	}
}

// configToApply returns the config graph to push: in registry mode the
// identified device's DesiredConfig, falling back to AutoWrite as the default for
// a device with none; in simple mode, AutoWrite for every device.
func (s *Session) configToApply() []byte {
	if s.srv.registryEnabled() {
		if id := s.DeviceID(); id != "" {
			if g := s.srv.desiredConfig(id); len(g) > 0 {
				return g
			}
		}
	}
	return s.srv.cfg.AutoWrite
}

// applyConfig pushes graph and records that this connection has been configured.
func (s *Session) applyConfig(graph []byte) error {
	s.mu.Lock()
	s.wroteCfg = true
	s.mu.Unlock()
	return s.WriteConfig(graph)
}

// identify derives the device's registry key from a reported object and records
// it (registry mode only). On first identification this connection it enriches
// the device record, fires OnIdentify, and schedules the config push.
func (s *Session) identify(obj Object) {
	if !s.srv.registryEnabled() {
		return
	}
	s.mu.Lock()
	if s.devID != "" {
		s.mu.Unlock()
		return
	}
	id, ok := s.srv.deviceKey(obj)
	if !ok {
		s.mu.Unlock()
		return
	}
	s.devID = id
	descr := s.descr
	s.mu.Unlock()

	mac := ""
	if len(obj.Key.MACs) > 0 {
		mac = obj.Key.MACs[0]
	}
	dev, _ := s.srv.upsertDevice(id, mac, s.remoteIP(), descr)
	s.srv.log.Info("device identified", "peer", s.peer, "id", id, "board", dev.Board)
	if s.srv.cfg.OnIdentify != nil {
		s.srv.cfg.OnIdentify(s, dev)
	}
	s.maybeScheduleConfig()
}

// readObject records a device-reported object and fires the OnObject callback.
func (s *Session) readObject(f m3.Field, body *m3.Message) {
	handle, _ := body.GetBytes(idSeq)
	flagb, _ := body.GetBytes(idDevFl)
	idxb, _ := body.GetBytes(idDevIx)
	key, _ := body.GetBytes(idKey)
	val, _ := body.GetBytes(idVal)
	obj := Object{
		FieldID:  f.ID,
		Handle:   handle,
		Flag:     flagb,
		Index:    idxb,
		RawKey:   key,
		RawValue: val,
	}
	if len(key) > 0 {
		obj.Key = cmr.ParseObject(key)
	}
	if len(val) > 0 {
		obj.Value = cmr.ParseObject(val)
	}
	s.objMu.Lock()
	s.objects = append(s.objects, obj)
	s.objMu.Unlock()

	s.srv.log.Info("device object", "peer", s.peer, "id", f.ID,
		"handle", hex.EncodeToString(handle), "key", obj.Key.String())
	if s.srv.cfg.OnObject != nil {
		s.srv.cfg.OnObject(s, obj)
	}
	s.identify(obj)
}

func strField(m *m3.Message, id uint32) string {
	if b, ok := m.GetBytes(id); ok {
		return string(b)
	}
	return ""
}
