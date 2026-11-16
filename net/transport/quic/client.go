package quic

import (
	"context"
	"crypto/tls"
	"net"
	_ "unsafe"

	"github.com/quic-go/quic-go"
	"gitlab.com/CoiaPrant/lib/net/socket/adapters/connected"
)

// DialAddr establishes a new QUIC connection to a server.
// It resolves the address, and then creates a new UDP connection to dial the QUIC server.
// When the QUIC connection is closed, this UDP connection is closed.
// See [Dial] for more details.
func DialAddr(ctx context.Context, laddr *net.UDPAddr, addr string, tlsConf *tls.Config, conf *quic.Config) (*quic.Conn, error) {
	udpAddr, err := net.ResolveUDPAddr("udp", addr)
	if err != nil {
		return nil, err
	}
	udpConn, err := net.DialUDP("udp", laddr, udpAddr)
	if err != nil {
		return nil, err
	}
	tr, err := setupTransport(connected.NewUDPConn(udpConn), tlsConf, true)
	if err != nil {
		return nil, err
	}
	conn, err := dial(tr, ctx, udpAddr, addr, tlsConf, conf, false)
	if err != nil {
		tr.Close()
		return nil, err
	}
	return conn, nil
}

// DialAddrEarly establishes a new 0-RTT QUIC connection to a server.
// See [DialAddr] for more details.
func DialAddrEarly(ctx context.Context, laddr *net.UDPAddr, addr string, tlsConf *tls.Config, conf *quic.Config) (*quic.Conn, error) {
	udpAddr, err := net.ResolveUDPAddr("udp", addr)
	if err != nil {
		return nil, err
	}
	udpConn, err := net.DialUDP("udp", laddr, udpAddr)
	if err != nil {
		return nil, err
	}
	tr, err := setupTransport(connected.NewUDPConn(udpConn), tlsConf, true)
	if err != nil {
		return nil, err
	}
	conn, err := dial(tr, ctx, udpAddr, addr, tlsConf, conf, true)
	if err != nil {
		tr.Close()
		return nil, err
	}
	return conn, nil
}

//go:linkname setupTransport github.com/quic-go/quic-go.setupTransport
func setupTransport(c net.PacketConn, tlsConf *tls.Config, createdPacketConn bool) (*quic.Transport, error)

//go:linkname dial github.com/quic-go/quic-go.(*Transport).dial
func dial(t *quic.Transport, ctx context.Context, addr net.Addr, host string, tlsConf *tls.Config, conf *quic.Config, use0RTT bool) (*quic.Conn, error)
