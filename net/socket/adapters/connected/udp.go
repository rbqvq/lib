package connected

import (
	"net"
	"net/netip"

	libnet "gitlab.com/CoiaPrant/lib/net"
)

var _ libnet.UDPConn = (*udpConn)(nil)

type udpConn struct {
	libnet.UDPConn
}

func NewUDPConn(conn libnet.UDPConn) libnet.UDPConn {
	return &udpConn{
		UDPConn: conn,
	}
}

func (c *udpConn) WriteTo(p []byte, _ net.Addr) (n int, err error) {
	return c.UDPConn.Write(p)
}

func (c *udpConn) WriteToUDP(p []byte, _ *net.UDPAddr) (n int, err error) {
	return c.UDPConn.Write(p)
}

func (c *udpConn) WriteToUDPAddrPort(p []byte, _ netip.AddrPort) (n int, err error) {
	return c.UDPConn.Write(p)
}

func (c *udpConn) WriteMsgUDP(b []byte, oob []byte, _ *net.UDPAddr) (n int, oobn int, err error) {
	return c.UDPConn.WriteMsgUDP(b, oob, nil)
}

func (c *udpConn) WriteMsgUDPAddrPort(b []byte, oob []byte, _ netip.AddrPort) (n int, oobn int, err error) {
	return c.UDPConn.WriteMsgUDP(b, oob, nil)
}
