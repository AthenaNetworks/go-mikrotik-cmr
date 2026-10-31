package controller

import (
	"crypto/tls"
	"net"
	"path/filepath"
	"testing"
	"time"

	"github.com/athenanetworks/cmr-server/pkg/cmr"
	"github.com/athenanetworks/cmr-server/pkg/m3"
)

func TestFileStoreRoundTrip(t *testing.T) {
	fs := NewFileStore(filepath.Join(t.TempDir(), "reg.json"))

	// A missing file loads as an empty registry, not an error.
	if got, err := fs.Load(); err != nil || len(got) != 0 {
		t.Fatalf("Load of missing file = %v, %v; want nil, nil", got, err)
	}

	now := time.Now().Round(time.Second)
	in := []Device{{
		ID: "52:54:00:cc:01:01", MAC: "52:54:00:cc:01:01", Board: "CHR", Arch: "x86",
		DesiredConfig: cmr.BuildVLANConfig(cmr.VLANConfig{VLANID: 42}),
		FirstSeen:     now, LastSeen: now,
	}}
	if err := fs.Save(in); err != nil {
		t.Fatal(err)
	}
	got, err := fs.Load()
	if err != nil || len(got) != 1 {
		t.Fatalf("Load = %v, %v; want 1 device", got, err)
	}
	if got[0].ID != in[0].ID || got[0].Board != "CHR" || !eq(got[0].DesiredConfig, in[0].DesiredConfig) {
		t.Fatalf("round-trip mismatch: %+v", got[0])
	}
}

// TestRegistryReconcile preloads a device's desired config, connects a mock
// device that identifies by MAC, and asserts the server pushes that device's
// persisted config graph once it is identified.
func TestRegistryReconcile(t *testing.T) {
	storePath := filepath.Join(t.TempDir(), "reg.json")
	want := cmr.BuildVLANConfig(cmr.VLANConfig{VLANID: 177, BridgeName: "regress-br"})
	const mac = "52:54:00:cc:01:01"
	if err := NewFileStore(storePath).Save([]Device{{ID: DeviceID(mac), DesiredConfig: want, FirstSeen: time.Now()}}); err != nil {
		t.Fatal(err)
	}

	cert, err := Identity("")
	if err != nil {
		t.Fatal(err)
	}
	srv, err := New(Config{
		Certificate: cert,
		Store:       NewFileStore(storePath),
		WriteDelay:  100 * time.Millisecond,
	})
	if err != nil {
		t.Fatal(err)
	}
	if d, ok := srv.Device(DeviceID(mac)); !ok || !eq(d.DesiredConfig, want) {
		t.Fatalf("preloaded device not in registry: %+v ok=%v", d, ok)
	}

	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer ln.Close()
	go srv.Serve(ln)

	conn, readFrame, write := mockDevice(t, ln.Addr().String())
	defer conn.Close()
	bringUp(t, readFrame, write)

	// Report a state object carrying the device MAC -> the server identifies it.
	stateBody := append(
		m3.AppendChunk(nil, []byte{0xaa, 0x00, 0x80, 0x3f}), // ClassState
		m3.AppendChunk(nil, []byte{0x52, 0x54, 0x00, 0xcc, 0x01, 0x01})...)
	obj := m3.New().PutBytes(21, []byte{0x90, 0x24, 0, 0, 0, 0, 0, 0x04}).PutBytes(74, stateBody)
	write(m3.New().PutNested(9, obj).Encode())

	// The next controller->device frame must be the config write (id7).
	var got m3.Field
	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		f := readFrame()
		if f.ID == 7 {
			got = f
			break
		}
	}
	if got.ID != 7 {
		t.Fatal("no config write (id7) after identification")
	}
	// Decode id7 -> id4 -> id30 and confirm it is the device's desired graph.
	inner, _ := m3.Decode(got.Value)
	q, _ := inner.GetNested(4)
	graph, _ := q.GetBytes(30)
	if !eq(graph, want) {
		t.Fatalf("pushed graph mismatch\n got %x\nwant %x", graph, want)
	}

	if d, ok := srv.Device(DeviceID(mac)); !ok || !d.connected || d.MAC != mac {
		t.Fatalf("device not registered/connected: %+v ok=%v", d, ok)
	}
}

// mockDevice dials the server as a device and returns frame read/write helpers.
func mockDevice(t *testing.T, addr string) (*tls.Conn, func() m3.Field, func([]byte)) {
	t.Helper()
	conn, err := tls.Dial("tcp", addr, &tls.Config{
		InsecureSkipVerify: true, MinVersion: tls.VersionTLS12, MaxVersion: tls.VersionTLS12,
	})
	if err != nil {
		t.Fatal(err)
	}
	dec := m3.NewStreamDecoder()
	readFrame := func() m3.Field {
		conn.SetReadDeadline(time.Now().Add(3 * time.Second))
		for {
			if f, ok := dec.Next(); ok {
				return f
			}
			buf := make([]byte, 4096)
			n, err := conn.Read(buf)
			if n > 0 {
				dec.Feed(buf[:n])
			}
			if err != nil {
				t.Fatalf("read: %v", err)
			}
		}
	}
	write := func(b []byte) {
		if _, err := conn.Write(b); err != nil {
			t.Fatalf("write: %v", err)
		}
	}
	return conn, readFrame, write
}

// bringUp walks the pairing handshake so the device is paired with devFlag 0x14.
func bringUp(t *testing.T, readFrame func() m3.Field, write func([]byte)) {
	t.Helper()
	write(m3.New().PutBytes(2, []byte{0xd2, 0x11, 0x01, 0xd8, 0x70, 0x7e, 0xae, 0x26, 0x5b, 0x81, 0xf4, 0xc7}).Encode())
	readFrame() // hello
	write(m3.New().PutNested(1, m3.New().PutBytes(22, []byte{1, 2, 3, 4}).PutBytes(25, make([]byte, 8)).PutBytes(24, []byte("CHR"))).Encode())
	readFrame() // version
	disc := m3.New().PutBytes(21, []byte{0x34, 0x01}).PutNested(4, m3.New().Put(29, []byte{0x14})).PutBytes(22, []byte{9, 9, 9, 9})
	write(m3.New().PutNested(6, disc).Encode())
	readFrame() // disc ack
	write(m3.New().PutNested(1, m3.New().PutBytes(22, []byte{5, 6, 7, 8}).PutBytes(21, []byte{0x34, 0x01})).Encode())
	readFrame() // verAck
	readFrame() // SUCCESS
}
