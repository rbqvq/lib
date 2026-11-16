package socks

import (
	"gitlab.com/go-extension/bytes"
)

type AuthMethodV5 uint8

const (
	AuthMethodV5_NoAuthenticationRequired AuthMethodV5 = 0x00
	AuthMethodV5_GSSAPI                   AuthMethodV5 = 0x01
	AuthMethodV5_UsernamePassword         AuthMethodV5 = 0x02
	AuthMethodV5_NoAcceptableMethods      AuthMethodV5 = 0xFF
)

func (method AuthMethodV5) Valid() bool {
	switch method {
	case AuthMethodV5_NoAuthenticationRequired,
		AuthMethodV5_GSSAPI,
		AuthMethodV5_UsernamePassword,
		AuthMethodV5_NoAcceptableMethods:
		return true
	default:
		return false
	}
}

func parseAuthRequestV5(r *bytes.Reader) ([]AuthMethodV5, error) {
	length, err := r.ReadByte()
	if err != nil {
		return nil, err
	}
	if length == 0 {
		return nil, ErrNotSOCKS
	}

	data, err := readFull(r, int(length))
	if err != nil {
		return nil, err
	}
	b := bytes.NewReader(data)

	var methods []AuthMethodV5
	for b.Len() > 0 {
		method, err := b.ReadByte()
		if err != nil {
			return nil, err
		}

		methods = append(methods, AuthMethodV5(method))
	}

	return methods, nil
}
