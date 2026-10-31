package controller

import (
	"time"

	"github.com/athenanetworks/cmr-server/pkg/cmr"
)

// DeviceID is a stable per-device key. By default it is the first hardware MAC
// the device reports (stable across reconnects); override via Config.DeviceKey.
type DeviceID string

// Device is the server's record of a known CMR device. The exported fields
// persist (see Store); zero values mean "not yet learned".
type Device struct {
	ID             DeviceID  `json:"id"`
	MAC            string    `json:"mac,omitempty"`
	Identity       string    `json:"identity,omitempty"` // device system name (descriptor id67)
	Board          string    `json:"board,omitempty"`
	Arch           string    `json:"arch,omitempty"`
	OS             string    `json:"os,omitempty"`
	LastRemoteAddr string    `json:"last_remote_addr,omitempty"` // source IP of the device's last connection (may be a NAT address)
	DesiredConfig  []byte    `json:"desired_config,omitempty"`   // id30 graph pushed on connect
	FirstSeen      time.Time `json:"first_seen"`
	LastSeen       time.Time `json:"last_seen"`

	connected bool // runtime only
}

// Store persists the device registry so enrolments and desired config survive
// restarts. Save receives the full set; implementations may snapshot it.
type Store interface {
	Load() ([]Device, error)
	Save(devices []Device) error
}

// defaultDeviceKey derives a DeviceID from a reported object: the first MAC it
// carries. Returns false when the object has no MAC.
func defaultDeviceKey(o Object) (DeviceID, bool) {
	if len(o.Key.MACs) > 0 {
		return DeviceID(o.Key.MACs[0]), true
	}
	return "", false
}

// registryEnabled reports whether per-device state is tracked (a Store or a
// custom key function makes the server registry-driven rather than apply-to-all).
func (s *Server) registryEnabled() bool {
	return s.cfg.Store != nil || s.cfg.DeviceKey != nil
}

func (s *Server) deviceKey(o Object) (DeviceID, bool) {
	if s.cfg.DeviceKey != nil {
		return s.cfg.DeviceKey(o)
	}
	return defaultDeviceKey(o)
}

// Devices returns a snapshot of every known device, connected or not.
func (s *Server) Devices() []Device {
	s.devMu.Lock()
	defer s.devMu.Unlock()
	out := make([]Device, 0, len(s.devices))
	for _, d := range s.devices {
		out = append(out, *d)
	}
	return out
}

// Device returns the record for id, if known.
func (s *Server) Device(id DeviceID) (Device, bool) {
	s.devMu.Lock()
	defer s.devMu.Unlock()
	d, ok := s.devices[id]
	if !ok {
		return Device{}, false
	}
	return *d, true
}

// SetDesiredConfig sets the config graph the server pushes to a device whenever
// it connects, persists it, and — if that device is connected and paired — pushes
// it now. The graph is typically from cmr.BuildVLANConfig. A nil graph clears it.
func (s *Server) SetDesiredConfig(id DeviceID, graph []byte) error {
	s.devMu.Lock()
	d, ok := s.devices[id]
	if !ok {
		d = &Device{ID: id, FirstSeen: time.Now()}
		s.devices[id] = d
	}
	d.DesiredConfig = append([]byte(nil), graph...)
	sess := s.sessionFor(id)
	s.devMu.Unlock()

	s.persist()
	if len(graph) > 0 && sess != nil && sess.Paired() {
		return sess.applyConfig(graph)
	}
	return nil
}

// Forget removes a device's registry record (identity + desired config) and
// persists the change, returning whether the device was known. It does not drop
// a live connection: a device that is currently connected, or that reconnects
// under open pairing, re-enrols on its next reported object — so Forget is for
// removing a known device, not for denying one access (do that at the network).
func (s *Server) Forget(id DeviceID) bool {
	s.devMu.Lock()
	_, ok := s.devices[id]
	if ok {
		delete(s.devices, id)
	}
	s.devMu.Unlock()
	if ok {
		s.persist()
	}
	return ok
}

// sessionFor returns the live session for a device id, or nil. Caller holds devMu.
func (s *Server) sessionFor(id DeviceID) *Session {
	s.mu.Lock()
	defer s.mu.Unlock()
	for sess := range s.sess {
		sess.mu.Lock()
		match := sess.devID == id
		sess.mu.Unlock()
		if match {
			return sess
		}
	}
	return nil
}

// upsertDevice records what a session has learned about its device and marks it
// connected. It returns a snapshot of the device and whether this call is the one
// that first identified it this connection.
func (s *Server) upsertDevice(id DeviceID, mac, remoteAddr string, info cmr.DeviceInfo) (Device, bool) {
	s.devMu.Lock()
	d, known := s.devices[id]
	if !known {
		d = &Device{ID: id, FirstSeen: time.Now()}
		s.devices[id] = d
	}
	if mac != "" {
		d.MAC = mac
	}
	if remoteAddr != "" {
		d.LastRemoteAddr = remoteAddr
	}
	if info.Board != "" {
		d.Board = info.Board
	}
	if info.Arch != "" {
		d.Arch = info.Arch
	}
	if info.OS != "" {
		d.OS = info.OS
	}
	if info.Identity != "" {
		d.Identity = info.Identity
	}
	d.LastSeen = time.Now()
	d.connected = true
	snap := *d
	s.devMu.Unlock()
	s.persist()
	return snap, !known
}

// markDisconnected clears the connected flag and stamps LastSeen.
func (s *Server) markDisconnected(id DeviceID) {
	s.devMu.Lock()
	if d, ok := s.devices[id]; ok {
		d.connected = false
		d.LastSeen = time.Now()
	}
	s.devMu.Unlock()
	s.persist()
}

// desiredConfig returns a device's desired config graph, if any.
func (s *Server) desiredConfig(id DeviceID) []byte {
	s.devMu.Lock()
	defer s.devMu.Unlock()
	if d, ok := s.devices[id]; ok {
		return d.DesiredConfig
	}
	return nil
}

// persist writes the current registry through the Store, if one is configured.
func (s *Server) persist() {
	if s.cfg.Store == nil {
		return
	}
	if err := s.cfg.Store.Save(s.Devices()); err != nil {
		s.log.Warn("registry persist failed", "err", err)
	}
}

// loadRegistry populates the in-memory registry from the Store at startup.
func (s *Server) loadRegistry() error {
	if s.cfg.Store == nil {
		return nil
	}
	devs, err := s.cfg.Store.Load()
	if err != nil {
		return err
	}
	s.devMu.Lock()
	for i := range devs {
		d := devs[i]
		d.connected = false
		s.devices[d.ID] = &d
	}
	n := len(s.devices)
	s.devMu.Unlock()
	s.log.Info("registry loaded", "devices", n)
	return nil
}
