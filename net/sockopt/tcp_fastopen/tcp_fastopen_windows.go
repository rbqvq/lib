package tcp_fastopen

import (
	"syscall"

	"golang.org/x/sys/windows"
)

func ListenControl(network string, address string, c syscall.RawConn) error {
	switch network {
	case "tcp", "tcp4", "tcp6":
	default:
		return nil
	}

	return c.Control(func(fd uintptr) {
		syscall.SetsockoptInt(syscall.Handle(fd), syscall.IPPROTO_TCP, windows.TCP_FASTOPEN, 1)
	})
}

func DialControl(network, address string, c syscall.RawConn) error {
	switch network {
	case "tcp", "tcp4", "tcp6":
	default:
		return nil
	}

	return c.Control(func(fd uintptr) {
		syscall.SetsockoptInt(syscall.Handle(fd), syscall.IPPROTO_TCP, windows.TCP_FASTOPEN, 1)
	})
}
