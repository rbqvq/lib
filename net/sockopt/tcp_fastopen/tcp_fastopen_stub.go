//go:build !(linux || darwin || windows || freebsd)

package tcp_fastopen

import "syscall"

func ListenControl(network string, address string, c syscall.RawConn) error {
	return nil
}

func DialControl(network, address string, c syscall.RawConn) error {
	return nil
}
