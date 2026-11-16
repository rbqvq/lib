package tcping

import (
	"net"
	"time"

	libnet "gitlab.com/CoiaPrant/lib/net"
)

const defaultTimeout = 3 * time.Second

var defaultDialer = &net.Dialer{
	Timeout: defaultTimeout,
}

func CheckWithDialer(dialer libnet.Dialer, raddr string) bool {
	conn, err := dialer.Dial("tcp", raddr)
	if err != nil {
		return false
	}
	defer conn.Close()

	return true
}

func CheckWithLAddr(laddr net.Addr, raddr string) bool {
	dialer := &net.Dialer{
		Timeout:   defaultTimeout,
		LocalAddr: laddr,
	}

	return CheckWithDialer(dialer, raddr)
}

func Check(raddr string) bool {
	return CheckWithDialer(defaultDialer, raddr)
}

func dialPing(dialer libnet.Dialer, raddr string) (net.Conn, time.Duration, error) {
	begin := time.Now()

	conn, err := dialer.Dial("tcp", raddr)
	elapsed := time.Since(begin)
	return conn, elapsed, err
}

func PingWithDialer(dialer libnet.Dialer, raddr string) (time.Duration, error) {
	conn, elapsed, err := dialPing(dialer, raddr)
	if err != nil {
		return 0, err
	}
	defer conn.Close()

	return elapsed, nil
}

func PingWithLAddr(laddr net.Addr, raddr string) (time.Duration, error) {
	dialer := &net.Dialer{
		Timeout:   defaultTimeout,
		LocalAddr: laddr,
	}

	return PingWithDialer(dialer, raddr)
}

func Ping(raddr string) (time.Duration, error) {
	return PingWithDialer(defaultDialer, raddr)
}
