package controller

import (
	"encoding/hex"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/athenanetworks/cmr-server/pkg/cmr"
)

// newAdminServer builds a registry-driven server with an ephemeral identity and
// returns its admin handler.
func newAdminServer(t *testing.T) (*Server, http.Handler) {
	t.Helper()
	cert, err := Identity("")
	if err != nil {
		t.Fatalf("identity: %v", err)
	}
	srv, err := New(Config{
		Certificate: cert,
		Store:       NewFileStore(t.TempDir() + "/devices.json"),
	})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	return srv, srv.AdminHandler()
}

func do(t *testing.T, h http.Handler, method, target, body string) *httptest.ResponseRecorder {
	t.Helper()
	var r io.Reader
	if body != "" {
		r = strings.NewReader(body)
	}
	req := httptest.NewRequest(method, target, r)
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)
	return rr
}

func decodeBody[T any](t *testing.T, rr *httptest.ResponseRecorder) T {
	t.Helper()
	var v T
	if err := json.Unmarshal(rr.Body.Bytes(), &v); err != nil {
		t.Fatalf("decode %q: %v", rr.Body.String(), err)
	}
	return v
}

func TestAdminDeviceLifecycle(t *testing.T) {
	_, h := newAdminServer(t)

	// Empty registry lists as an empty array, not null.
	rr := do(t, h, "GET", "/devices", "")
	if rr.Code != http.StatusOK {
		t.Fatalf("list: %d %s", rr.Code, rr.Body)
	}
	if got := strings.TrimSpace(rr.Body.String()); got != "[]" {
		t.Fatalf("empty list = %q, want []", got)
	}

	// Unknown device is a 404.
	if rr := do(t, h, "GET", "/devices/nope", ""); rr.Code != http.StatusNotFound {
		t.Fatalf("unknown device = %d, want 404", rr.Code)
	}

	// Set a vlan config: creates (pre-enrols) the device and decodes the graph.
	const id = "52:54:00:cc:01:01"
	rr = do(t, h, "PUT", "/devices/"+id+"/config", `{"vlan":100,"bridge":"cmr-bridge"}`)
	if rr.Code != http.StatusOK {
		t.Fatalf("set config: %d %s", rr.Code, rr.Body)
	}
	dv := decodeBody[deviceView](t, rr)
	if dv.ID != id {
		t.Fatalf("id = %q, want %q", dv.ID, id)
	}
	if dv.Config == nil || dv.Config.VLAN == nil || dv.Config.VLAN.ID != 100 {
		t.Fatalf("config not decoded: %+v", dv.Config)
	}
	// The hex graph must be exactly what BuildVLANConfig produces.
	want := hex.EncodeToString(cmr.BuildVLANConfig(cmr.VLANConfig{VLANID: 100, BridgeName: "cmr-bridge"}))
	if dv.Config.Graph != want {
		t.Fatalf("graph = %s, want %s", dv.Config.Graph, want)
	}

	// It now appears in the list and is fetchable by id.
	list := decodeBody[[]deviceView](t, do(t, h, "GET", "/devices", ""))
	if len(list) != 1 || list[0].ID != id {
		t.Fatalf("list = %+v, want one %q", list, id)
	}
	if rr := do(t, h, "GET", "/devices/"+id, ""); rr.Code != http.StatusOK {
		t.Fatalf("get device = %d", rr.Code)
	}

	// Clear it: 204, and the config is gone.
	if rr := do(t, h, "DELETE", "/devices/"+id+"/config", ""); rr.Code != http.StatusNoContent {
		t.Fatalf("clear = %d %s", rr.Code, rr.Body)
	}
	dv = decodeBody[deviceView](t, do(t, h, "GET", "/devices/"+id, ""))
	if dv.Config != nil {
		t.Fatalf("config still set after clear: %+v", dv.Config)
	}
}

func TestAdminSetConfigRawGraph(t *testing.T) {
	_, h := newAdminServer(t)
	graph := cmr.BuildVLANConfig(cmr.VLANConfig{VLANID: 250, BridgeName: "br0"})
	body := `{"graph":"` + hex.EncodeToString(graph) + `"}`
	rr := do(t, h, "PUT", "/devices/dev1/config", body)
	if rr.Code != http.StatusOK {
		t.Fatalf("set graph: %d %s", rr.Code, rr.Body)
	}
	dv := decodeBody[deviceView](t, rr)
	if dv.Config == nil || dv.Config.VLAN == nil || dv.Config.VLAN.ID != 250 || dv.Config.VLAN.Bridge != "br0" {
		t.Fatalf("raw graph not round-tripped: %+v", dv.Config)
	}
}

func TestAdminSetConfigBadRequest(t *testing.T) {
	_, h := newAdminServer(t)
	for _, body := range []string{
		``,               // empty body
		`{}`,             // neither vlan nor graph
		`{"vlan":9000}`,  // out of range
		`{"graph":"zz"}`, // bad hex
		`{"bogus":1}`,    // unknown field
	} {
		if rr := do(t, h, "PUT", "/devices/dev1/config", body); rr.Code != http.StatusBadRequest {
			t.Fatalf("body %q = %d, want 400", body, rr.Code)
		}
	}
}

func TestAdminBearerAuth(t *testing.T) {
	_, inner := newAdminServer(t)
	h := BearerAuth("s3cret", inner)

	if rr := do(t, h, "GET", "/devices", ""); rr.Code != http.StatusUnauthorized {
		t.Fatalf("no token = %d, want 401", rr.Code)
	}

	req := httptest.NewRequest("GET", "/devices", nil)
	req.Header.Set("Authorization", "Bearer wrong")
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)
	if rr.Code != http.StatusUnauthorized {
		t.Fatalf("wrong token = %d, want 401", rr.Code)
	}

	req = httptest.NewRequest("GET", "/devices", nil)
	req.Header.Set("Authorization", "Bearer s3cret")
	rr = httptest.NewRecorder()
	h.ServeHTTP(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("right token = %d, want 200", rr.Code)
	}

	// An empty token disables the guard (handler returned unwrapped).
	if open := BearerAuth("", inner); open != inner {
		t.Fatal("empty token should return the handler unwrapped")
	}
}

func TestAdminHealthAndSessions(t *testing.T) {
	_, h := newAdminServer(t)
	if rr := do(t, h, "GET", "/healthz", ""); rr.Code != http.StatusOK {
		t.Fatalf("healthz = %d", rr.Code)
	}
	// No live connections: sessions is an empty array.
	if got := strings.TrimSpace(do(t, h, "GET", "/sessions", "").Body.String()); got != "[]" {
		t.Fatalf("sessions = %q, want []", got)
	}
}
