package socks

import (
	"net"

	"gitlab.com/go-extension/bytes"
)

type RequestV4 struct {
	Command Command
	Port    uint16
	IP      net.IP
	User    string
	Addr    string
}

func parseRequestBodyV4(r *bytes.Reader) (*RequestV4, error) {
	port, err := readUint16(r)
	if err != nil {
		return nil, err
	}

	ip, err := readFull(r, 4)
	if err != nil {
		return nil, err
	}

	user, err := r.ReadString(0)
	if err != nil {
		return nil, err
	}
	user = user[:len(user)-1]

	var addr string
	if r.Len() > 0 {
		addr, err = r.ReadString(0)
		if err != nil {
			return nil, err
		}
		addr = addr[:len(addr)-1]
	}

	return &RequestV4{
		Port: port,
		IP:   net.IP(ip),
		User: user,
		Addr: addr,
	}, nil
}
