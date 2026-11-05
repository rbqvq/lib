package sockopt

import (
	"syscall"

	"golang.org/x/sys/unix"
)

func ListenControl(network, address string, c syscall.RawConn) error {
	return c.Control(func(fd uintptr) {
		unix.SetsockoptInt(int(fd), unix.SOL_SOCKET, unix.SO_REUSEADDR, 1)
	})
}

func DialControl(network, address string, c syscall.RawConn) error {
	return c.Control(func(fd uintptr) {
		unix.SetsockoptInt(int(fd), unix.SOL_SOCKET, unix.SO_REUSEADDR, 1)
		unix.SetsockoptInt(int(fd), unix.SOL_SOCKET, unix.SO_REUSEPORT, 1)
	})
}
