package pkex

import "errors"

// Role is the handshake role.
type Role int

const (
	// RoleClient is the enrollee/initiator (PKEXClntImpl in the binary).
	RoleClient Role = iota
	// RoleServer is the controller/responder (PKEXSrvImpl in the binary).
	RoleServer
)

func (r Role) String() string {
	if r == RoleServer {
		return "server"
	}
	return "client"
}

// Frame type bytes, observed in PKEXSrvImpl::handleFrame / PKEXClntImpl::handleFrame.
const (
	FrameDiscoveryReq  byte = 0x40 // client -> server
	FrameDiscoveryResp byte = 0x41 // server -> client
	FrameMsg1          byte = 0x42 // server -> client
	FrameMsg2          byte = 0x43 // client -> server
	FrameMsg3          byte = 0x44 // server -> client
	FrameMsg4          byte = 0x45 // client -> server
)

// FrameName returns a human label for a frame type.
func FrameName(t byte) string {
	switch t {
	case FrameDiscoveryReq:
		return "DiscoveryReq"
	case FrameDiscoveryResp:
		return "DiscoveryResp"
	case FrameMsg1:
		return "Msg1"
	case FrameMsg2:
		return "Msg2"
	case FrameMsg3:
		return "Msg3"
	case FrameMsg4:
		return "Msg4"
	default:
		return "Unknown"
	}
}

// State mirrors the server state machine variable (this+0xc8, values 0..5).
type State int

const (
	StateInit      State = iota // 0
	StateDiscovery              // client: awaiting DiscoveryResp
	StateExchange               // awaiting a masked-point message
	StateAuth                   // awaiting an auth-tag message
	StateDone                   // 5 (success)
	StateFailed
)

func (s State) String() string {
	switch s {
	case StateInit:
		return "init"
	case StateDiscovery:
		return "discovery"
	case StateExchange:
		return "exchange"
	case StateAuth:
		return "auth"
	case StateDone:
		return "done"
	default:
		return "failed"
	}
}

// Protocol errors. ErrVersionMismatch and ErrAuthFailed mirror the binary's
// "protocol version missmatch" and "auth failed (N)" gates.
var (
	ErrVersionMismatch = errors.New("pkex: protocol version mismatch")
	ErrAuthFailed      = errors.New("pkex: authentication failed")
	ErrUnexpectedFrame = errors.New("pkex: unexpected frame for current state")
	ErrDone            = errors.New("pkex: session already finished")
)

// Result is produced by the server when a peer completes enrollment.
type Result struct {
	PeerID []byte   // 32-byte authenticated peer identity
	Addrs  [][]byte // addresses the peer advertised in Msg4
}
