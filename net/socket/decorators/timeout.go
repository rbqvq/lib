package decorators

import (
	"net"
	"time"
)

var _ net.Conn = (*timeout)(nil)

type timeout struct {
	net.Conn
	d time.Duration
}

func WithTimeout(conn net.Conn, d time.Duration) net.Conn {
	return &timeout{
		Conn: conn,
		d:    d,
	}
}

func (c *timeout) setSessionTimeout() {
	c.Conn.SetReadDeadline(time.Now().Add(c.d))
}

func (c *timeout) Read(p []byte) (n int, err error) {
	c.setSessionTimeout()
	return c.Conn.Read(p)
}

func (c *timeout) Write(p []byte) (n int, err error) {
	c.setSessionTimeout()
	return c.Conn.Write(p)
}
