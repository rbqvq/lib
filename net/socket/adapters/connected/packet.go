package connected

import (
	"net"
)

var _ net.Conn = (*PacketConn)(nil)
var _ net.PacketConn = (*PacketConn)(nil)

type PacketConn struct {
	net.Conn
	remoteAddr net.Addr
}

func NewPacketConn(conn net.Conn) *PacketConn {
	return &PacketConn{
		Conn:       conn,
		remoteAddr: conn.RemoteAddr(),
	}
}

func (c *PacketConn) ReadFrom(p []byte) (n int, addr net.Addr, err error) {
	n, err = c.Conn.Read(p)
	return n, c.remoteAddr, err
}

func (c *PacketConn) WriteTo(p []byte, _ net.Addr) (n int, err error) {
	return c.Conn.Write(p)
}
