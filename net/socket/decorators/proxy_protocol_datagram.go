package decorators

import (
	"io"
	"net"
	"sync"
)

var _ net.Conn = (*ProxyProtocolDatagram)(nil)
var _ io.ReaderFrom = (*ProxyProtocolDatagram)(nil)

// ProxyProtocolDatagram prepends a fixed PROXY Protocol header to every Write.
// Suitable for datagram sinks where the header must be present on each
// datagram (e.g. AWS NLB, Cloudflare SPP).
type ProxyProtocolDatagram struct {
	net.Conn

	mutex  sync.Mutex
	offset int    // len(header), where payload starts
	buf    []byte // header (fixed) + max payload
}

// WithProxyProtocolDatagram wraps conn so that every Write emits header
// followed by the payload as a single underlying Write. length is the
// maximum payload size per Write; the internal buffer is sized
// len(header) + length. header is copied once at construction.
func WithProxyProtocolDatagram(conn net.Conn, header []byte, length int) *ProxyProtocolDatagram {
	buf := make([]byte, len(header)+length)
	copy(buf, header)

	return &ProxyProtocolDatagram{
		Conn:   conn,
		offset: len(header),
		buf:    buf,
	}
}

func (c *ProxyProtocolDatagram) Write(p []byte) (int, error) {
	if len(p) == 0 {
		return 0, nil
	}

	if c.offset+len(p) > len(c.buf) {
		return 0, io.ErrShortWrite
	}

	c.mutex.Lock()
	defer c.mutex.Unlock()

	buf := c.buf
	n := copy(buf[c.offset:], p)

	_, err := c.Conn.Write(buf[:c.offset+n])
	if err != nil {
		return 0, err
	}

	return len(p), nil
}

// ReadFrom implements io.ReaderFrom. It reads one datagram at a time from r
// and writes each datagram to the underlying conn with the PROXY Protocol
// header prepended. It returns when r returns an error (typically io.EOF)
// or when a write fails.
//
// written counts bytes sent to the underlying conn, including the
// prepended PROXY Protocol header on each datagram.
//
// The whole read-write loop is serialized by c.mutex, so the caller must
// ensure no other goroutine writes to this conn concurrently. In practice
// that means "use Write or ReadFrom, not both, from different goroutines".
func (c *ProxyProtocolDatagram) ReadFrom(r io.Reader) (written int64, err error) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	buf := c.buf
	p := buf[c.offset:]

	var n int
	for err == nil {
		n, err = r.Read(p)
		if n <= 0 {
			continue
		}

		n, e := c.Conn.Write(buf[:c.offset+n])
		written += int64(n)
		if e != nil {
			return written, e
		}
	}

	switch err {
	case io.EOF:
		return written, nil
	default:
		return
	}
}
