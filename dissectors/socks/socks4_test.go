package socks

import (
	"fmt"
	"os"
	"testing"
)

func TestParseRequestV4(t *testing.T) {
	payload, err := os.ReadFile("testdata/socks4.bin")
	if err != nil {
		panic(err)
	}

	req, err := ParseRequestV4(payload)
	if err != nil {
		panic(err)
	}

	fmt.Printf("socks4 request parsed, cmd: %d, port: %d, ip: %s, user: %s\n", req.Command, req.Port, req.IP, req.User)
}

func TestParseRequestV4A(t *testing.T) {
	payload, err := os.ReadFile("testdata/socks4a.bin")
	if err != nil {
		panic(err)
	}

	req, err := ParseRequestV4(payload)
	if err != nil {
		panic(err)
	}

	fmt.Printf("socks4a request parsed, cmd: %d, port: %d, ip: %s, user: %s, addr: %s\n", req.Command, req.Port, req.IP, req.User, req.Addr)
}
