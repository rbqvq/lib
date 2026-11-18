package connected

import (
	"net"
	"net/netip"

	libnet "gitlab.com/CoiaPrant/lib/net"
)

var _ libnet.UDPConn = (*UDPConn)(nil)

type UDPConn struct {
	libnet.UDPConn
}

func NewUDPConn(conn libnet.UDPConn) *UDPConn {
	return &UDPConn{
		UDPConn: conn,
	}
}

func (c *UDPConn) WriteTo(p []byte, _ net.Addr) (n int, err error) {
	return c.UDPConn.Write(p)
}

func (c *UDPConn) WriteToUDP(p []byte, _ *net.UDPAddr) (n int, err error) {
	return c.UDPConn.Write(p)
}

func (c *UDPConn) WriteToUDPAddrPort(p []byte, _ netip.AddrPort) (n int, err error) {
	return c.UDPConn.Write(p)
}

func (c *UDPConn) WriteMsgUDP(b []byte, oob []byte, _ *net.UDPAddr) (n int, oobn int, err error) {
	return c.UDPConn.WriteMsgUDP(b, oob, nil)
}

func (c *UDPConn) WriteMsgUDPAddrPort(b []byte, oob []byte, _ netip.AddrPort) (n int, oobn int, err error) {
	return c.UDPConn.WriteMsgUDP(b, oob, nil)
}
