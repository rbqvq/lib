package socks

import "errors"

var (
	ErrNotSOCKS = errors.New("not a SOCKS packet")
)
