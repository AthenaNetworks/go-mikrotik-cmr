// Command embed shows how another project runs a CMR controller in-process with
// pkg/controller: a stock MikroTik device pairs over open pairing, the program
// prints what the device reports about itself, and pushes a vlan config once the
// device reaches steady state. This whole file is the integration surface — an
// application swaps the callbacks for its own device registry and config source.
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

	srv, err := controller.New(controller.Config{
		Certificate: cert,
		Logger:      logger,

		// Push vlan 100 into every device that pairs, once it is steady. For
		// dynamic, per-device config call Session.WriteConfig from a callback or
		// from your own control loop over srv.Sessions() instead.
		AutoWrite: cmr.BuildVLANConfig(cmr.VLANConfig{VLANID: 100, BridgeName: "cmr-bridge"}),

		// React to the device as it reports its inventory (interfaces, identity,
		// board, config objects). This is where an app builds its device registry.
		OnObject: func(s *controller.Session, o controller.Object) {
			if len(o.Key.Strings) > 0 {
				logger.Info("device object", "peer", s.Peer(), "class", cmr.ClassName(o.Key.Class), "strings", o.Key.Strings)
			}
		},
		OnDisconnect: func(s *controller.Session, err error) {
			logger.Info("device gone", "peer", s.Peer(), "objects_seen", len(s.Objects()))
		},
	})
	if err != nil {
		logger.Error("controller", "err", err)
		os.Exit(1)
	}

	logger.Info("cmr controller embedded; point a device at this host:54321",
		"fingerprint", controller.Fingerprint(cert))
	if err := srv.ListenAndServe("0.0.0.0:54321"); err != nil {
		logger.Error("serve", "err", err)
		os.Exit(1)
	}
}
