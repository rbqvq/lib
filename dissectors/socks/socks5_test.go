package socks

import (
	"fmt"
	"os"
	"testing"
)

func TestCheckAuthRequestV5(t *testing.T) {
	payload, err := os.ReadFile("testdata/socks5.bin")
	if err != nil {
		panic(err)
	}

	err = CheckAuthRequestV5(payload)
	if err != nil {
		panic(err)
	}

	fmt.Println("socks5 auth request detected")
}

func TestParseAuthRequestV5(t *testing.T) {
	payload, err := os.ReadFile("testdata/socks5.bin")
	if err != nil {
		panic(err)
	}

	methods, err := ParseAuthRequestV5(payload)
	if err != nil {
		panic(err)
	}

	fmt.Printf("socks5 auth request parsed, methods: %d\n", methods)
}
