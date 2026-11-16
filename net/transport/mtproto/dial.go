package mtproto

import (
	"fmt"
	"net"

	libnet "gitlab.com/CoiaPrant/lib/net"
)

func Dial(dialer libnet.Dialer, dataCenter uint16) (net.Conn, error) {
	var addrs []string
	switch dataCenter {
	case 1:
		addrs = []string{"149.154.175.50:443", "[2001:b28:f23d:f001::a]:443"}
	case 2, 0:
		addrs = []string{"149.154.167.51:443", "95.161.76.100:443", "[2001:67c:4e8:f002::a]:443"}
	case 3:
		addrs = []string{"149.154.175.100:443", "[2001:b28:f23d:f003::a]:443"}
	case 4:
		addrs = []string{"149.154.167.91:443", "[2001:67c:4e8:f004::a]:443"}
	case 5:
		addrs = []string{"149.154.171.5:443", "[2001:b28:f23f:f005::a]:443"}
	case 203:
		addrs = []string{"91.105.192.100:443"}
	case 10001:
		addrs = []string{"149.154.175.10:443", "[2001:b28:f23d:f001::e]:443"}
	case 10002:
		addrs = []string{"149.154.167.40:443", "[2001:67c:4e8:f002::e]:443"}
	case 10003:
		addrs = []string{"149.154.175.117:443", "[2001:b28:f23d:f003::e]:443"}
	default:
		return nil, fmt.Errorf("bad DC: %d", dataCenter)
	}

	for _, addr := range addrs {
		proxy, err := net.Dial("tcp", addr)
		if err != nil {
			continue
		}

		return proxy, nil
	}

	return nil, fmt.Errorf("cannot dial to the DC %d", dataCenter)
}
