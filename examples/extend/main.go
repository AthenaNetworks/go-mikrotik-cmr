// Command extend shows how another project embeds the CMR controller: it
// supplies custom Hooks (authorization, per-peer codes, enrollment reactions)
// and watches the store, then runs one in-process enrollment to demonstrate the
// flow end to end. Swap the UDP transport for an 802.11 action-frame transport
// to talk to real devices (see docs/PROTOCOL.md).
package main

import (
	"context"
	"fmt"
	"log/slog"
	"net"
	"os"
	"time"

	"github.com/athenanetworks/cmr-server/pkg/model"
	"github.com/athenanetworks/cmr-server/pkg/pkex"
	"github.com/athenanetworks/cmr-server/pkg/server"
	"github.com/athenanetworks/cmr-server/pkg/transport"
)

func main() {
	log := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))

	// Controller transport.
	srvTr, err := transport.NewUDP(transport.SynthMAC(1), "127.0.0.1:0", "")
	must(err)
	defer srvTr.Close()

	store := model.NewStore()

	// Extensibility seam #1: react to every store change (feed a DB, a UI, ...).
	store.Watch(func(e model.Event) {
		fmt.Printf("[store] %s %s id=%d\n", e.Type, e.Collection, e.ID)
	})

	// Signalled by the OnEnroll hook so we can wait for the controller (which
	// finalizes enrollment on its own goroutine) before reading the store.
	enrolled := make(chan struct{}, 1)

	// Extensibility seam #2: controller hooks.
	srv := server.New(server.Config{
		Code:     []byte("default-code"),
		Identity: "example-controller",
		Logger:   log,
		Hooks: server.Hooks{
			// Only let locally-administered MACs enroll (example policy).
			Authorize: func(peer net.HardwareAddr) bool {
				return len(peer) == 6 && peer[0]&0x02 != 0
			},
			// Give a specific device its own code.
			CodeFor: func(peer net.HardwareAddr) []byte {
				if peer.String() == transport.SynthMAC(2).String() {
					return []byte("device-2-code")
				}
				return nil // fall back to Config.Code
			},
			OnEnroll: func(d *model.Device, res pkex.Result) {
				fmt.Printf("[hook] enrolled device id=%d mac=%s ip=%v peer_id=%s addrs=%d\n",
					d.ID, d.MACString(), d.IP, d.PeerID, len(res.Addrs))
				select {
				case enrolled <- struct{}{}:
				default:
				}
			},
		},
	}, store, srvTr)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	go srv.Run(ctx)

	// A device enrolls, using the per-peer code the hook assigns to SynthMAC(2).
	cliTr, err := transport.NewUDP(transport.SynthMAC(2), "127.0.0.1:0", srvTr.LocalAddr())
	must(err)
	defer cliTr.Close()

	en := server.NewEnroller(server.EnrollConfig{
		Code:       []byte("device-2-code"),
		Identity:   "example-device",
		Addrs:      [][]byte{[]byte("192.168.88.10")},
		Controller: transport.Broadcast,
		Logger:     log,
	}, cliTr)

	if err := en.Run(ctx); err != nil {
		fmt.Println("enrollment failed:", err)
		os.Exit(1)
	}
	// The client finished sending; wait for the controller to finalize the
	// device on its own goroutine (signalled by the OnEnroll hook).
	select {
	case <-enrolled:
	case <-time.After(2 * time.Second):
		fmt.Println("warning: timed out waiting for controller to finalize enrollment")
	}
	fmt.Printf("done: controller now manages %d device(s)\n", len(store.Devices()))
}

func must(err error) {
	if err != nil {
		panic(err)
	}
}
