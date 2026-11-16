package mtproto

import (
	"bytes"
	"crypto/rand"
	"encoding/binary"
	"io"
	"slices"
)

type Header [64]byte

func (h *Header) Authentication(secret *Secret) *Authentication {
	authentication := new(Authentication)

	auth := slices.Clone(h[8:56])
	computeKey(authentication.PeerKey[:], auth[:32], secret)
	copy(authentication.PeerNonce[:], auth[32:])

	slices.Reverse(auth)
	computeKey(authentication.LocalKey[:], auth[:32], secret)
	copy(authentication.LocalNonce[:], auth[32:])

	return authentication
}

func (h *Header) connectionType() []byte {
	return h[56:60]
}

func (h *Header) ConnectionType() ConnectionType {
	connectionType := h.connectionType()
	switch {
	case bytes.Equal(connectionType, ConnectionTagAbridged):
		return ConnectionTypeAbridged
	case bytes.Equal(connectionType, ConnectionTagIntermediate):
		return ConnectionTypeIntermediate
	case bytes.Equal(connectionType, ConnectionTagPaddedIntermediate):
		return ConnectionTypePaddedIntermediate
	default:
		return ConnectionTypeUnknown
	}
}

func (h *Header) DataCenterID() uint16 {
	x := ((int16(h[61]) << 8) | int16(h[60]))
	if x < 0 {
		x = -x
	}
	return uint16(x)
}

func newHeader(state ConnectionState) (*Header, error) {
	h := new(Header)
	for {
		_, err := io.ReadFull(rand.Reader, h[:])
		if err != nil {
			return nil, err
		}

		// https://github.com/tdlib/td/blob/master/td/mtproto/TcpTransport.cpp#L95
		switch h[0] {
		case
			0xef, // Abridged
			0x16: // TLS
			continue
		}

		switch binary.LittleEndian.Uint32(h[:4]) {
		case
			0x44414548, // HEAD
			0x54534f50, // POST
			0x20544547, // GET
			0x4954504f, // OPTI
			0xdddddddd, // PaddedIntermediate
			0xeeeeeeee: // Intermediate
			continue
		}

		if binary.LittleEndian.Uint32(h[4:8]) == 0 {
			continue
		}

		copy(h[56:60], state.Type.Tag())
		return h, nil
	}
}
