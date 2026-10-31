package controller

import (
	"crypto/subtle"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"sort"
	"strings"
	"time"

	"github.com/athenanetworks/cmr-server/pkg/cmr"
)

// AdminHandler returns an http.Handler exposing the device registry as a small
// JSON API: list and inspect devices, set or clear each device's desired config,
// and list live sessions. It is the operator surface over the registry — the
// same thing SetDesiredConfig/Devices do in-process, reachable over HTTP.
//
// It is deliberately unauthenticated so the host application can mount it where
// it likes and apply its own policy. For a shared-token guard wrap it with
// BearerAuth; otherwise bind it to a loopback or internal address.
//
// The API reflects the registry, so it is only useful when the server is
// registry-driven (a Store or DeviceKey is configured). Setting a device's
// config pre-enrols it: the record is created now and applied when that device
// next connects and identifies.
//
// Routes:
//
//	GET    /healthz                liveness + device count
//	GET    /devices                known devices (connected or not)
//	GET    /devices/{id}           one device
//	PUT    /devices/{id}/config    set desired config — {"vlan":N,"bridge":"..."} or {"graph":"<hex>"}
//	DELETE /devices/{id}/config    clear desired config
//	GET    /sessions               live device sessions
func (s *Server) AdminHandler() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /healthz", func(w http.ResponseWriter, _ *http.Request) {
		writeJSON(w, http.StatusOK, map[string]any{"status": "ok", "devices": len(s.Devices())})
	})
	mux.HandleFunc("GET /devices", s.adminListDevices)
	mux.HandleFunc("GET /devices/{id}", s.adminGetDevice)
	mux.HandleFunc("PUT /devices/{id}/config", s.adminSetConfig)
	mux.HandleFunc("DELETE /devices/{id}/config", s.adminClearConfig)
	mux.HandleFunc("GET /sessions", s.adminListSessions)
	return mux
}

// BearerAuth wraps h so every request must carry "Authorization: Bearer <token>".
// It is a convenience for guarding AdminHandler with a shared token; for anything
// richer, put the handler behind your own middleware. An empty token disables the
// check and returns h unwrapped.
func BearerAuth(token string, h http.Handler) http.Handler {
	if token == "" {
		return h
	}
	want := []byte("Bearer " + token)
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		got := []byte(r.Header.Get("Authorization"))
		if subtle.ConstantTimeCompare(got, want) != 1 {
			w.Header().Set("WWW-Authenticate", "Bearer")
			httpError(w, http.StatusUnauthorized, "unauthorized")
			return
		}
		h.ServeHTTP(w, r)
	})
}

// deviceView is the JSON shape of a Device — the persisted fields plus runtime
// connected state and the decoded desired config (never the raw []byte, which
// would marshal as opaque base64).
type deviceView struct {
	ID        DeviceID    `json:"id"`
	MAC       string      `json:"mac,omitempty"`
	Identity  string      `json:"identity,omitempty"`
	Board     string      `json:"board,omitempty"`
	Arch      string      `json:"arch,omitempty"`
	OS        string      `json:"os,omitempty"`
	Connected bool        `json:"connected"`
	FirstSeen time.Time   `json:"first_seen"`
	LastSeen  time.Time   `json:"last_seen"`
	Config    *configView `json:"config,omitempty"`
}

// configView is the JSON shape of a desired config graph: the raw id30 graph as
// hex plus, when it is a recognised vlan graph, its decoded form.
type configView struct {
	Bytes int       `json:"bytes"`
	Graph string    `json:"graph"` // hex-encoded id30 object graph
	VLAN  *vlanView `json:"vlan,omitempty"`
}

type vlanView struct {
	ID     int    `json:"id"`
	Bridge string `json:"bridge"`
}

// sessionView is the JSON shape of a live connection.
type sessionView struct {
	Peer     string   `json:"peer"`
	DeviceID DeviceID `json:"device_id,omitempty"`
	Paired   bool     `json:"paired"`
	DevFlag  byte     `json:"dev_flag"`
	Objects  int      `json:"objects"`
}

// configRequest is the body of PUT /devices/{id}/config. Graph (hex id30 graph)
// takes precedence; otherwise a vlan graph is authored from VLAN/Bridge.
type configRequest struct {
	VLAN   int    `json:"vlan"`
	Bridge string `json:"bridge"`
	Graph  string `json:"graph"`
}

// graph resolves the request to an id30 config graph.
func (c configRequest) graph() ([]byte, error) {
	if c.Graph != "" {
		b, err := hex.DecodeString(strings.TrimSpace(c.Graph))
		if err != nil {
			return nil, fmt.Errorf("bad graph hex: %w", err)
		}
		if len(b) == 0 {
			return nil, errors.New("empty graph")
		}
		return b, nil
	}
	if c.VLAN > 0 {
		if c.VLAN > 4094 {
			return nil, errors.New("vlan out of range (1..4094)")
		}
		return cmr.BuildVLANConfig(cmr.VLANConfig{VLANID: c.VLAN, BridgeName: c.Bridge}), nil
	}
	return nil, errors.New("provide vlan (1..4094) or graph (hex)")
}

func (s *Server) adminListDevices(w http.ResponseWriter, _ *http.Request) {
	devs := s.Devices()
	out := make([]deviceView, 0, len(devs))
	for _, d := range devs {
		out = append(out, deviceToView(d))
	}
	sort.Slice(out, func(i, j int) bool { return out[i].ID < out[j].ID })
	writeJSON(w, http.StatusOK, out)
}

func (s *Server) adminGetDevice(w http.ResponseWriter, r *http.Request) {
	d, ok := s.Device(DeviceID(r.PathValue("id")))
	if !ok {
		httpError(w, http.StatusNotFound, "device not found")
		return
	}
	writeJSON(w, http.StatusOK, deviceToView(d))
}

func (s *Server) adminSetConfig(w http.ResponseWriter, r *http.Request) {
	id := DeviceID(r.PathValue("id"))
	var req configRequest
	if err := decodeJSON(w, r, &req); err != nil {
		httpError(w, http.StatusBadRequest, err.Error())
		return
	}
	graph, err := req.graph()
	if err != nil {
		httpError(w, http.StatusBadRequest, err.Error())
		return
	}
	if err := s.SetDesiredConfig(id, graph); err != nil {
		httpError(w, http.StatusInternalServerError, err.Error())
		return
	}
	d, _ := s.Device(id)
	writeJSON(w, http.StatusOK, deviceToView(d))
}

func (s *Server) adminClearConfig(w http.ResponseWriter, r *http.Request) {
	id := DeviceID(r.PathValue("id"))
	if _, ok := s.Device(id); !ok {
		httpError(w, http.StatusNotFound, "device not found")
		return
	}
	if err := s.SetDesiredConfig(id, nil); err != nil {
		httpError(w, http.StatusInternalServerError, err.Error())
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (s *Server) adminListSessions(w http.ResponseWriter, _ *http.Request) {
	sess := s.Sessions()
	out := make([]sessionView, 0, len(sess))
	for _, se := range sess {
		out = append(out, sessionView{
			Peer:     se.Peer(),
			DeviceID: se.DeviceID(),
			Paired:   se.Paired(),
			DevFlag:  se.DevFlag(),
			Objects:  len(se.Objects()),
		})
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Peer < out[j].Peer })
	writeJSON(w, http.StatusOK, out)
}

func deviceToView(d Device) deviceView {
	return deviceView{
		ID:        d.ID,
		MAC:       d.MAC,
		Identity:  d.Identity,
		Board:     d.Board,
		Arch:      d.Arch,
		OS:        d.OS,
		Connected: d.connected,
		FirstSeen: d.FirstSeen,
		LastSeen:  d.LastSeen,
		Config:    configToView(d.DesiredConfig),
	}
}

func configToView(graph []byte) *configView {
	if len(graph) == 0 {
		return nil
	}
	v := &configView{Bytes: len(graph), Graph: hex.EncodeToString(graph)}
	if vc, ok := cmr.ParseVLANConfig(graph); ok {
		v.VLAN = &vlanView{ID: vc.VLANID, Bridge: vc.BridgeName}
	}
	return v
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	_ = enc.Encode(v)
}

func httpError(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, map[string]string{"error": msg})
}

// decodeJSON reads a size-limited JSON body, rejecting unknown fields so typos
// in a config request surface as errors rather than silent no-ops.
func decodeJSON(w http.ResponseWriter, r *http.Request, v any) error {
	dec := json.NewDecoder(http.MaxBytesReader(w, r.Body, 1<<20))
	dec.DisallowUnknownFields()
	if err := dec.Decode(v); err != nil {
		if errors.Is(err, io.EOF) {
			return errors.New("empty request body")
		}
		return err
	}
	return nil
}
