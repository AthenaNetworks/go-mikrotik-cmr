package m3

// StreamDecoder decodes one direction of the CMR control channel as a single
// continuous field stream. This is the real framing: the delta base is NOT
// reset per message — it carries across the entire direction's byte stream, and
// TLS writes chunk that stream arbitrarily (verified against a live capture: the
// whole controller->device and device->controller directions each decode as one
// 1800+ field stream with zero errors, whereas decoding per TLS-read mis-frames
// the heavy management burst).
//
// Feed bytes as they arrive from the socket; Next returns complete top-level
// fields one at a time, buffering any partial trailing field until more bytes
// arrive.
type StreamDecoder struct {
	base uint32
	buf  []byte
}

// NewStreamDecoder returns a decoder with a fresh (zero) base.
func NewStreamDecoder() *StreamDecoder { return &StreamDecoder{} }

// Feed appends freshly received bytes to the decoder's buffer.
func (d *StreamDecoder) Feed(p []byte) { d.buf = append(d.buf, p...) }

// Next returns the next complete field in the stream and advances the running
// base. ok is false when the buffer does not yet hold a complete field (feed
// more bytes and call again). On a malformed tag/value it also returns false;
// callers that need to distinguish can check Buffered.
func (d *StreamDecoder) Next() (f Field, ok bool) {
	tag, tn, done := chunk(d.buf, 0)
	if !done {
		return Field{}, false
	}
	val, vn, done := chunk(d.buf, tn)
	if !done {
		return Field{}, false
	}
	typ, id := unfold(uint32(leUint(tag)), d.base)
	d.base = id
	d.buf = d.buf[vn:]
	return Field{ID: id, Type: typ, Value: val}, true
}

// Buffered reports how many bytes are held pending a complete field.
func (d *StreamDecoder) Buffered() int { return len(d.buf) }

// chunk reads one m3 chunk starting at p, returning its data, the absolute new
// offset, and whether a complete chunk was available. It never allocates beyond
// the returned slice and reports done=false (rather than erroring) when the
// buffer is truncated mid-chunk, so a stream reader can wait for more bytes.
func chunk(b []byte, p int) (data []byte, next int, done bool) {
	for p < len(b) && b[p] == 0xff { // skip padding
		p++
	}
	if p >= len(b) {
		return nil, p, false
	}
	intro := b[p]
	p++
	switch {
	case intro <= 0xcf:
		return b[p-1 : p], p, true // the intro byte is the inline value
	case intro <= 0xef:
		n := int(intro - 0xd0)
		if p+n > len(b) {
			return nil, p, false
		}
		return b[p : p+n], p + n, true
	}
	var w int
	switch intro {
	case 0xfe:
		w = 1
	case 0xfd:
		w = 2
	case 0xfc:
		w = 4
	case 0xfb:
		w = 8
	case 0xfa:
		w = 16
	default:
		return nil, p, false
	}
	if p+w > len(b) {
		return nil, p, false
	}
	n := int(leUint(b[p : p+w]))
	p += w
	if n < 0 || p+n > len(b) {
		return nil, p, false
	}
	return b[p : p+n], p + n, true
}

// StreamAll decodes a complete, in-order direction buffer into its fields using
// a single running base. A trailing partial field (if any) is ignored.
func StreamAll(b []byte) []Field {
	d := NewStreamDecoder()
	d.Feed(b)
	var out []Field
	for {
		f, ok := d.Next()
		if !ok {
			return out
		}
		out = append(out, f)
	}
}
