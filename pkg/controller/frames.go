package controller

import (
	"crypto/rand"

	"github.com/athenanetworks/cmr-server/pkg/m3"
)

// CMR control-channel constants recovered from the live protocol.
var (
	helloMagic = []byte{0xd2, 0x11, 0x01, 0xd8} // hello token prefix
	// helloToken is the controller's 8-byte hello suffix: a FIXED protocol
	// constant (a version/capability identifier), not a nonce. The real RouterOS
	// 7.26 controller always sends acec0a1b7c1ad815 (the device always sends
	// 707eae265b81f4c7). Randomising it still pairs but makes the device treat us
	// as an incompatible peer (it adds an id26 to its version and then answers
	// object requests "not found").
	helloToken = []byte{0xac, 0xec, 0x0a, 0x1b, 0x7c, 0x1a, 0xd8, 0x15}
	buildVer   = []byte{0x4b, 0x60, 0xbe, 0x6a, 0x01, 0x62, 0x1a, 0x07} // build-time + version
	boardName  = []byte("CHR")
	scopeName  = []byte("com.mikrotik.cmr")
)

// randBytes is a package var so tests can substitute deterministic nonces.
var randBytes = func(n int) []byte {
	b := make([]byte, n)
	_, _ = rand.Read(b)
	return b
}

func serverHello() []byte {
	token := append(append([]byte(nil), helloMagic...), helloToken...)
	return m3.New().PutBytes(idHello, token).Encode()
}

func serverVersion() []byte {
	inner := m3.New().PutBytes(idNonce, randBytes(nonceSize)).PutBytes(idBuildV, buildVer).PutBytes(idBoard, boardName)
	return m3.New().PutNested(idVer, inner).Encode()
}

func discAck(seq byte) []byte {
	inner := m3.New().PutBytes(idNonce, randBytes(nonceSize)).PutBytes(idSeq, []byte{seq})
	return m3.New().PutNested(idDisc, inner).Encode()
}

// verAck is the controller's id1 reply after the device's post-discovery id1:
// id1 { id1{id27:status, id21:innerSeq}, id22:nonce, id21:outerSeq } where
// outerSeq echoes the device's discovery seq and innerSeq = [outerSeq[0]-1, ...].
func verAck(discSeq []byte, status byte) []byte {
	outerSeq := discSeq
	if len(outerSeq) == 0 {
		outerSeq = []byte{0x01, 0x01}
	}
	innerSeq := append([]byte(nil), outerSeq...)
	innerSeq[0]-- // the controller's own message seq is one below the device's
	sub := m3.New().PutU32(idStatus, uint32(status)).PutBytes(idSeq, innerSeq)
	body := m3.New().PutNested(idVer, sub).PutBytes(idNonce, randBytes(nonceSize)).PutBytes(idSeq, outerSeq)
	return m3.New().PutNested(idVer, body).Encode()
}

func successInner() *m3.Message {
	return m3.New().PutBytes(idScope, scopeName).PutU32(idStatus, 0).PutBytes(idNonce, randBytes(nonceSize))
}

// successFrame is the id10 pairing SUCCESS alone: {id23:scope, id27:0x00, id22:nonce}.
func successFrame() []byte {
	return m3.New().PutNested(idSucc, successInner()).Encode()
}

// successWithRootsFrame builds ONE continuous m3 stream: id10 SUCCESS followed by
// the controller's root id9 objects (parsed out of rootBlob), with each object's
// id29 flag rewritten to the device's query flag. The id9 fields delta-chain off
// the id10 inside a single Encode(), which is how a real controller frames it and
// what the device needs to treat the objects as ADDs rather than lookups.
func successWithRootsFrame(rootBlob []byte, flag byte) []byte {
	m := m3.New().PutNested(idSucc, successInner())
	for _, f := range m3.StreamAll(patchObjectFlag(rootBlob, flag)) {
		if f.ID == idDev {
			m.PutBytes(idDev, f.Value)
		}
	}
	return m.Encode()
}

// controllerRows is the SUCCESS frame plus two id9 device-table rows (the
// controller advertising itself), delta-chained 10->9->9 as one message — the
// shape a live controller sends after verAck. Retained for reference and pinned
// by a byte-exact test; the production path uses successWithRootsFrame.
func controllerRows() []byte {
	r1 := m3.New().PutBytes(idDevFl, []byte{0x49}).PutBytes(idDevIx, []byte{0x01}).
		PutBytes(idSeq, []byte{0x14, 0, 0, 0, 0, 0, 0, 0x04})
	r2 := m3.New().PutBytes(idBoard2, boardName).PutBytes(idDevFl, []byte{0x49}).PutBytes(idDevIx, []byte{0x05}).
		PutBytes(idSeq, []byte{0x15, 0, 0, 0, 0, 0, 0, 0x04})
	return m3.New().PutNested(idSucc, successInner()).PutNested(idDev, r1).PutNested(idDev, r2).Encode()
}

// GetAllQuery builds an id7 getall mirroring the device's own resync: id7{
// id4{id21 seq, id23 scope, id29 flag, [id30 selector, id31], id27=0}, id22
// nonce, id21 seq }. With no selector it is byte-shaped exactly like the device's
// resync (which makes a peer push its objects). Exposed for enumeration
// experiments via Session.Send.
func GetAllQuery(scope string, flag byte, selector []byte) []byte {
	q := m3.New().
		Put(idSeq, []byte{0x01, 0x01}).
		PutBytes(idScope, []byte(scope)).
		Put(idDevFl, []byte{flag})
	if len(selector) > 0 {
		q.PutBytes(30, selector).PutBytes(31, nil)
	}
	q.PutU32(idStatus, 0)
	return m3.New().PutNested(idResync, q).PutBytes(idNonce, randBytes(nonceSize)).Put(idSeq, []byte{0x02, 0x01}).Encode()
}

// SubscribeQuery builds an id11 subscribe for a scope: id11{ id23 scope, [id32
// user\0], id21 seq }. Exposed for pull-stream experiments via Session.Send.
func SubscribeQuery(scope, user string) []byte {
	q := m3.New().PutBytes(idScope, []byte(scope))
	if user != "" {
		q.PutBytes(32, append([]byte(user), 0))
	}
	q.Put(idSeq, []byte{0x30})
	return m3.New().PutNested(idSub, q).Encode()
}

// buildWriteConfig builds the controller->device config WRITE: an id7 whose id4
// names the device's own config scope and carries the desired object graph in
// id30. Byte-shaped exactly like the write a real RouterOS controller emits when
// its managed config changes (captured live: `/cmr/vlan/add vlan-ids=100` -> the
// device builds a MANAGED cmr-bridge with ether1 at pvid=100). The nonce and
// reqid live INSIDE the id7 envelope (siblings of id4), not outside.
//
//	id7{ id4{ id21 seq, id23 scope, id29 flag, id30 graph }, id22 nonce, id21 reqid }
func buildWriteConfig(scope string, flag byte, payload []byte) []byte {
	q := m3.New().
		Put(idSeq, []byte{0x34}).
		PutBytes(idScope, []byte(scope)).
		Put(idDevFl, []byte{flag}).
		PutBytes(30, payload)
	inner := m3.New().
		PutNested(idQuery, q).
		PutBytes(idNonce, randBytes(nonceSize)).
		Put(idSeq, []byte{0x35})
	return m3.New().PutNested(idResync, inner).Encode()
}

// patchObjectFlag rewrites every object's id29 flag (encoded `d1 e9 <flag>` in a
// captured blob, originally 0x10) to flag. The device's discovery/getall query
// dictates this flag and answers objects carrying a different one "not found".
// flag==0 means "not captured" -> leave the blob unchanged.
func patchObjectFlag(blob []byte, flag byte) []byte {
	if flag == 0 {
		return blob
	}
	needle := []byte{0xd1, 0xe9, 0x10}
	out := append([]byte(nil), blob...)
	for i := 0; i+len(needle) <= len(out); i++ {
		if out[i] == needle[0] && out[i+1] == needle[1] && out[i+2] == needle[2] {
			out[i+2] = flag
		}
	}
	return out
}
