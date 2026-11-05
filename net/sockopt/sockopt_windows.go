package sockopt

import (
	"syscall"
)

const (
	_SO_RANDOMIZE_PORT = 0x3005
)

func ListenControl(network, address string, c syscall.RawConn) error {
	return c.Control(func(fd uintptr) {
		syscall.SetsockoptInt(syscall.Handle(fd), syscall.SOL_SOCKET, syscall.SO_REUSEADDR, 1)
	})
}

func DialControl(network, address string, c syscall.RawConn) error {
	return c.Control(func(fd uintptr) {
		syscall.SetsockoptInt(syscall.Handle(fd), syscall.SOL_SOCKET, syscall.SO_REUSEADDR, 1)
		syscall.SetsockoptInt(syscall.Handle(fd), syscall.SOL_SOCKET, _SO_RANDOMIZE_PORT, 1)
	})
}
