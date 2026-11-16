package net

import (
	"context"
	"net"
	"net/netip"
	"os"
	"syscall"
)

type Dialer interface {
	Dial(network, address string) (net.Conn, error)
}

type DialerContext interface {
	Dialer
	DialContext(ctx context.Context, network, address string) (net.Conn, error)
}

type UDPConn interface {
	net.Conn
	net.PacketConn
	syscall.Conn
	File() (f *os.File, err error)

	SetReadBuffer(bytes int) error
	ReadFromUDP(b []byte) (n int, addr *net.UDPAddr, err error)
	ReadFromUDPAddrPort(b []byte) (n int, addr netip.AddrPort, err error)
	ReadMsgUDP(b []byte, oob []byte) (n int, oobn int, flags int, addr *net.UDPAddr, err error)
	ReadMsgUDPAddrPort(b []byte, oob []byte) (n int, oobn int, flags int, addr netip.AddrPort, err error)

	SetWriteBuffer(bytes int) error
	WriteMsgUDP(b []byte, oob []byte, addr *net.UDPAddr) (n int, oobn int, err error)
	WriteMsgUDPAddrPort(b []byte, oob []byte, addr netip.AddrPort) (n int, oobn int, err error)
	WriteToUDP(b []byte, addr *net.UDPAddr) (int, error)
	WriteToUDPAddrPort(b []byte, addr netip.AddrPort) (int, error)
}
