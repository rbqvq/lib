package quic

import "errors"

var (
	ErrNotQUIC          = errors.New("not a QUIC packet")
	ErrNotInitialPacket = errors.New("not initial packet")
)
