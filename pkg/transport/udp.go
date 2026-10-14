package transport

import (
	"context"
	"encoding/binary"
	"fmt"
	"net"
	"sync"
	"time"
)

// udpHeader is this transport's own envelope so it can carry a (srcMAC,dstMAC)
// pair over connectionless UDP without ever inspecting the opaque Payload.
// Layout: srcMAC[6] dstMAC[6] payload...
const udpHeaderLen = 12

// UDPTransport is the built-in reference Transport. It is intended for tests,
// local development, and controller<->controller links — NOT for over-the-air
// interop with real devices (that needs an 802.11 action-frame transport).
//
// It learns peer UDP endpoints from inbound datagrams, so a device/agent only
// needs to know the controller's address to begin discovery; the controller
// replies to the learned endpoint. Discovery broadcasts go to BroadcastAddr.
type UDPTransport struct {
	local         net.HardwareAddr
	conn          *net.UDPConn
	broadcastAddr *net.UDPAddr

	mu    sync.RWMutex
	peers map[string]*net.UDPAddr // macString -> last-seen UDP endpoint
}

// NewUDP creates a UDP transport bound to listen, presenting localMAC as the
// controller address. broadcast is where discovery frames are sent before any
// peer endpoint is known (e.g. a subnet broadcast or multicast group:port).
func NewUDP(localMAC net.HardwareAddr, listen string, broadcast string) (*UDPTransport, error) {
	laddr, err := net.ResolveUDPAddr("udp", listen)
	if err != nil {
		return nil, fmt.Errorf("resolve listen %q: %w", listen, err)
	}
	conn, err := net.ListenUDP("udp", laddr)
	if err != nil {
		return nil, fmt.Errorf("listen %q: %w", listen, err)
	}
	var baddr *net.UDPAddr
	if broadcast != "" {
		if baddr, err = net.ResolveUDPAddr("udp", broadcast); err != nil {
			conn.Close()
			return nil, fmt.Errorf("resolve broadcast %q: %w", broadcast, err)
		}
	}
	return &UDPTransport{
		local:         localMAC,
		conn:          conn,
		broadcastAddr: baddr,
		peers:         map[string]*net.UDPAddr{},
	}, nil
}

// LocalMAC implements Transport.
func (t *UDPTransport) LocalMAC() net.HardwareAddr { return t.local }

// LocalAddr returns the UDP address the transport is listening on (useful when
// bound to :0 and another party needs the resolved port).
func (t *UDPTransport) LocalAddr() string { return t.conn.LocalAddr().String() }

// Start implements Transport: reads datagrams, decodes the envelope, and
// delivers frames addressed to us (or broadcast) to h.
func (t *UDPTransport) Start(ctx context.Context, h Handler) error {
	go func() {
		<-ctx.Done()
		t.conn.SetReadDeadline(time.Now())
	}()
	buf := make([]byte, 65535)
	for {
		n, src, err := t.conn.ReadFromUDP(buf)
		if err != nil {
			if ctx.Err() != nil {
				return nil
			}
			if ne, ok := err.(net.Error); ok && ne.Timeout() {
				if ctx.Err() != nil {
					return nil
				}
				continue
			}
			return err
		}
		if n < udpHeaderLen {
			continue
		}
		srcMAC := net.HardwareAddr(append([]byte(nil), buf[0:6]...))
		dstMAC := net.HardwareAddr(append([]byte(nil), buf[6:12]...))
		payload := append([]byte(nil), buf[udpHeaderLen:n]...)

		t.mu.Lock()
		t.peers[srcMAC.String()] = src
		t.mu.Unlock()

		// Accept frames for us or broadcast.
		if !IsBroadcast(dstMAC) && !macEqual(dstMAC, t.local) {
			continue
		}
		h(Frame{Peer: srcMAC, Payload: payload})
	}
}

// Send implements Transport.
func (t *UDPTransport) Send(f Frame) error {
	pkt := make([]byte, udpHeaderLen+len(f.Payload))
	copy(pkt[0:6], t.local)
	if IsBroadcast(f.Peer) || f.Peer == nil {
		copy(pkt[6:12], Broadcast)
	} else {
		copy(pkt[6:12], f.Peer)
	}
	copy(pkt[udpHeaderLen:], f.Payload)

	if IsBroadcast(f.Peer) || f.Peer == nil {
		if t.broadcastAddr == nil {
			return fmt.Errorf("udp transport: broadcast requested but no broadcast address configured")
		}
		_, err := t.conn.WriteToUDP(pkt, t.broadcastAddr)
		return err
	}
	t.mu.RLock()
	dst, ok := t.peers[f.Peer.String()]
	t.mu.RUnlock()
	if !ok {
		return fmt.Errorf("udp transport: unknown peer %s (no endpoint learned yet)", f.Peer)
	}
	_, err := t.conn.WriteToUDP(pkt, dst)
	return err
}

// Close implements Transport.
func (t *UDPTransport) Close() error { return t.conn.Close() }

func macEqual(a, b net.HardwareAddr) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

// synthMAC derives a stable locally-administered MAC from a seed string, handy
// for tests and single-host multi-instance setups.
func SynthMAC(seed uint16) net.HardwareAddr {
	m := make(net.HardwareAddr, 6)
	m[0] = 0x02 // locally administered, unicast
	binary.BigEndian.PutUint16(m[4:], seed)
	return m
}
