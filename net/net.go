package net

import (
	"errors"
	"fmt"
	"net"
	"strings"
)

var IPv4Loopback = net.IPv4(127, 0, 0, 1)

func JoinHostPort(host string, port uint16) string {
	if strings.HasPrefix(host, "[") && strings.HasSuffix(host, "]") {
		return fmt.Sprintf("%s:%d", host, port)
	}

	if strings.Contains(host, ":") {
		return fmt.Sprintf("[%s]:%d", host, port)
	}

	return fmt.Sprintf("%s:%d", host, port)
}

func ResolveAddr(network string, forward string) (net.Addr, error) {
	switch network {
	case "tcp", "tcp4", "tcp6":
		return net.ResolveTCPAddr(network, forward)
	case "udp", "udp4", "udp6":
		return net.ResolveUDPAddr(network, forward)
	default:
		return nil, errors.ErrUnsupported
	}
}
