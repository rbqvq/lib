//go:build !(linux || darwin || windows || freebsd)

package sockopt

import "syscall"

func ListenControl(network, address string, rawConn syscall.RawConn) error {
	return nil
}

func DialControl(network, address string, rawConn syscall.RawConn) error {
	return nil
}