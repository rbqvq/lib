package tls

import (
	"fmt"

	"gitlab.com/go-extension/bytes"
)

type HandshakeType uint8

// TLS handshake message types.
const (
	TypeHelloRequest          HandshakeType = 0
	TypeClientHello           HandshakeType = 1
	TypeServerHello           HandshakeType = 2
	TypeNewSessionTicket      HandshakeType = 4
	TypeEndOfEarlyData        HandshakeType = 5
	TypeEncryptedExtensions   HandshakeType = 8
	TypeCertificate           HandshakeType = 11
	TypeServerKeyExchange     HandshakeType = 12
	TypeCertificateRequest    HandshakeType = 13
	TypeServerHelloDone       HandshakeType = 14
	TypeCertificateVerify     HandshakeType = 15
	TypeClientKeyExchange     HandshakeType = 16
	TypeFinished              HandshakeType = 20
	TypeCertificateStatus     HandshakeType = 22
	TypeKeyUpdate             HandshakeType = 24
	TypeCompressedCertificate HandshakeType = 25
	TypeNextProtocol          HandshakeType = 67  // Not IANA assigned
	TypeMessageHash           HandshakeType = 254 // synthetic message
)

type HandshakeHeader struct {
	Type   HandshakeType
	Length uint32
}

func ParseHandshakeHeader(r *bytes.Reader) (HandshakeHeader, error) {
	typ, err := r.ReadByte()
	if err != nil {
		return HandshakeHeader{}, err
	}

	length, err := readUint24(r)
	if err != nil {
		return HandshakeHeader{}, err
	}

	return HandshakeHeader{
		Type:   HandshakeType(typ),
		Length: length,
	}, nil
}

func ParseHandshakeClientHello(r *bytes.Reader) (*ClientHello, error) {
	h, err := ParseHandshakeHeader(r)
	if err != nil {
		return nil, err
	}

	if h.Type != TypeClientHello {
		return nil, fmt.Errorf("bad handshake type: %d", h.Type)
	}

	msgData, err := readFull(r, int(h.Length))
	if err != nil {
		return nil, err
	}

	r.Reset(msgData)

	return ParseClientHello(r)
}
