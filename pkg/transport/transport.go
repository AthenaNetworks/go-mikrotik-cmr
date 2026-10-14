// Package transport abstracts how CMR provisioning/control frames reach peers.
//
// The real MikroTik cmr binary does not own a socket: it hands frames to the
// RouterOS wifi/bus stack, which emits them as link-layer frames to a peer
// MAC across the relevant interfaces (the binary's PKEXPushButton::getIfs /
// send(..., MacAddr) path). A standalone server cannot assume that stack, so
// transport is an interface the integrating project implements:
//
//   - 802.11 public action frames (true over-the-air interop with devices),
//   - raw L2 frames over an Ethernet interface (wired backhaul / lab),
//   - UDP (the built-in reference transport, for tests and server-to-server).
//
// The server treats a transport purely as a bidirectional pipe of Frames keyed
// by peer MAC; all protocol meaning lives in pkg/pkex and pkg/m3.
package transport

import (
	"context"
	"net"
)

// Frame is one CMR protocol frame as seen by a transport: an opaque payload
// (the m3-encoded PKEX frame, including its one-byte type) addressed to or
// from a peer link-layer address. Transports MUST NOT interpret Payload.
type Frame struct {
	Peer    net.HardwareAddr // destination (outbound) or source (inbound)
	Iface   string           // interface name, if the transport is link-scoped
	Payload []byte
}

// Handler consumes inbound frames. It must be non-blocking or fan out to its
// own goroutine; transports may call it from their read loop.
type Handler func(Frame)

// Transport is the pluggable I/O seam of the server.
type Transport interface {
	// Start begins delivering inbound frames to h until ctx is cancelled.
	Start(ctx context.Context, h Handler) error
	// Send transmits one frame to f.Peer (and f.Iface if the transport is
	// link-scoped). A broadcast peer address means "all peers on all ifaces".
	Send(f Frame) error
	// LocalMAC returns the address this transport presents as the controller.
	LocalMAC() net.HardwareAddr
	// Close stops the transport and releases resources.
	Close() error
}

// Broadcast is the link-layer broadcast address used for discovery frames.
var Broadcast = net.HardwareAddr{0xff, 0xff, 0xff, 0xff, 0xff, 0xff}

// IsBroadcast reports whether addr is the all-ones broadcast address.
func IsBroadcast(addr net.HardwareAddr) bool {
	if len(addr) != 6 {
		return false
	}
	for _, b := range addr {
		if b != 0xff {
			return false
		}
	}
	return true
}
