// Command embed shows how another project runs a multi-device CMR controller
// in-process with pkg/controller: stock MikroTik devices pair over open pairing,
// the server identifies and persists each one in a registry, and pushes a
// per-device vlan config. This whole file is the integration surface — an
// application swaps the callbacks and config policy for its own.
package main

import (
	"log/slog"
	"os"

	"github.com/athenanetworks/cmr-server/pkg/cmr"
	"github.com/athenanetworks/cmr-server/pkg/controller"
)

func main() {
	logger := slog.New(slog.NewTextHandler(os.Stderr, nil))

	// A persistent identity keeps the device-pinned fingerprint stable across
	// restarts; pass "" for an ephemeral one.
	cert, err := controller.Identity("ctrl.ed25519.seed")
	if err != nil {
		logger.Error("identity", "err", err)
		os.Exit(1)
	}

	// A registry (any Store) makes the server multi-device: it identifies each
	// device by a stable key, tracks it, and persists enrolments + desired config.
	store := controller.NewFileStore("devices.json")

	srv, err := controller.New(controller.Config{
		Certificate: cert,
		Logger:      logger,
		Store:       store,

		// When a device is identified (first time per connection), decide its
		// config. Here every device gets vlan 100; a real app would look up the
		// device's role. SetDesiredConfig persists it and, since the device is
		// connected, pushes it now; on later reconnects it is re-applied from the
		// registry automatically.
		OnIdentify: func(s *controller.Session, d controller.Device) {
			logger.Info("device identified", "id", d.ID, "board", d.Board, "mac", d.MAC)
			if len(d.DesiredConfig) == 0 {
				_ = s.WriteConfig(cmr.BuildVLANConfig(cmr.VLANConfig{VLANID: 100}))
			}
		},
		OnDisconnect: func(s *controller.Session, err error) {
			logger.Info("device gone", "peer", s.Peer(), "id", s.DeviceID())
		},
	})
	if err != nil {
		logger.Error("controller", "err", err)
		os.Exit(1)
	}

	// srv.Devices() lists every known device (connected or not); srv.Device(id)
	// and srv.SetDesiredConfig(id, graph) drive one by its stable id — e.g. from
	// an HTTP admin handler or a control loop.
	logger.Info("cmr controller embedded; point devices at this host:54321",
		"fingerprint", controller.Fingerprint(cert), "known_devices", len(srv.Devices()))
	if err := srv.ListenAndServe("0.0.0.0:54321"); err != nil {
		logger.Error("serve", "err", err)
		os.Exit(1)
	}
}
