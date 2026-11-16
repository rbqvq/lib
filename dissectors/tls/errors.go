package tls

import "errors"

var (
	ErrNotTLS = errors.New("not a TLS packet")
	ErrDecode = errors.New("decode error")
)
