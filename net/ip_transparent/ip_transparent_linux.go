//go:build linux

package ip_transparent

import (
	"net"
	"syscall"
	"time"

	"gitlab.com/CoiaPrant/lib/net/sockopt"
	"golang.org/x/sys/unix"
)

var control = func(network, address string, c syscall.RawConn) error {
	err := sockopt.DialControl(network, address, c)
	if err != nil {
		return err
	}

	return c.Control(func(fd uintptr) {
		unix.SetsockoptInt(int(fd), unix.SOL_IP, unix.IP_TRANSPARENT, 1)
	})
}

func DialTimeout(network string, laddr net.Addr, raddr string, timeout time.Duration) (net.Conn, error) {
	dialer := &net.Dialer{
		Timeout:   timeout,
		LocalAddr: laddr,
		Control:   control,
	}

	return dialer.Dial(network, raddr)
}
