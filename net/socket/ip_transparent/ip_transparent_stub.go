//go:build !linux

package ip_transparent

import (
	"errors"
	"net"
	"time"
)

func DialTimeout(network string, laddr net.Addr, raddr string, timeout time.Duration) (net.Conn, error) {
	return nil, errors.ErrUnsupported
}
