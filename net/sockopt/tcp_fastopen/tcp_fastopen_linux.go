package tcp_fastopen

import (
	"syscall"

	"golang.org/x/sys/unix"
)

func ListenControl(network string, address string, c syscall.RawConn) error {
	switch network {
	case "tcp", "tcp4", "tcp6":
	default:
		return nil
	}

	return c.Control(func(fd uintptr) {
		syscall.SetsockoptInt(int(fd), syscall.SOL_TCP, unix.TCP_FASTOPEN, 4096)
	})
}

func DialControl(network, address string, c syscall.RawConn) error {
	switch network {
	case "tcp", "tcp4", "tcp6":
	default:
		return nil
	}

	return c.Control(func(fd uintptr) {
		syscall.SetsockoptInt(int(fd), syscall.SOL_TCP, unix.TCP_FASTOPEN_CONNECT, 1)
	})
}
