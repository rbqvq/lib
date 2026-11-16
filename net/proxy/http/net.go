package http

import (
	"bufio"
	"io"
	"net"
)

type netConn struct {
	net.Conn
	*bufio.Reader
}

func (c *netConn) Read(b []byte) (int, error) {
	if c.Reader.Buffered() > 0 {
		return c.Reader.Read(b)
	}

	return c.Conn.Read(b)
}

func (c *netConn) WriteTo(w io.Writer) (int64, error) {
	return c.Reader.WriteTo(w)
}
