package socks

import (
	"fmt"

	"gitlab.com/go-extension/bytes"
)

func checkV5(r *bytes.Reader) error {
	v, err := readVersion(r)
	if err != nil {
		return err
	}
	if v != Version5 {
		return ErrNotSOCKS
	}
	return nil
}

func CheckAuthRequestV5(packet []byte) error {
	b := bytes.NewReader(packet)

	err := checkV5(b)
	if err != nil {
		return err
	}

	methods, err := parseAuthRequestV5(b)
	if err != nil {
		return ErrNotSOCKS
	}
	if b.Len() > 0 {
		return ErrNotSOCKS
	}

	for _, method := range methods {
		if !method.Valid() {
			return fmt.Errorf("invalid socks5 auth method: %d", method)
		}
	}

	return nil
}

func ParseAuthRequestV5(packet []byte) ([]AuthMethodV5, error) {
	b := bytes.NewReader(packet)

	err := checkV5(b)
	if err != nil {
		return nil, err
	}

	return parseAuthRequestV5(b)
}
