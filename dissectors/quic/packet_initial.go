package quic

import (
	"fmt"

	"gitlab.com/go-extension/bytes"
)

type InitialPacket struct {
	Version         Version
	PacketProtector *PacketProtector
	PacketNumber    int64
	Frames          []FrameCrypto
}

type parseInitialPacketOption struct {
	Version         Version
	Recv            bool
	PacketProtector *PacketProtector
}

type ParseInitialPacketOption func(*parseInitialPacketOption)

func WithVersion(version Version) ParseInitialPacketOption {
	return func(option *parseInitialPacketOption) {
		option.Version = version
	}
}

func WithRecv() ParseInitialPacketOption {
	return func(option *parseInitialPacketOption) {
		option.Recv = true
	}
}

func WithPacketProtector(p *PacketProtector) ParseInitialPacketOption {
	return func(option *parseInitialPacketOption) {
		option.PacketProtector = p
	}
}

func WithInitialPacket(p *InitialPacket) ParseInitialPacketOption {
	return func(option *parseInitialPacketOption) {
		option.Version = p.Version
		option.PacketProtector = p.PacketProtector
	}
}

func ParseInitialPacket(packet []byte, options ...ParseInitialPacketOption) (*InitialPacket, error) {
	if !isQUIC(packet) {
		return nil, ErrNotQUIC
	}

	option := &parseInitialPacketOption{}
	for _, o := range options {
		o(option)
	}

	v := option.Version
	if v == 0 {
		var err error
		v, err = ParseVersion(packet)
		if err != nil {
			return nil, err
		}
	}
	if !v.IsSupportedVersion() {
		return nil, fmt.Errorf("unsupported version %d or not a QUIC packet", v)
	}

	typeByte := packet[0]
	isGoogleFrame := false
	if v.isGoogle() {
		if isGoogleHeader(typeByte) {
			return parseGoogleLegacyInitialPacket(packet, v)
		}

		if v < VersionQ050 {
			return parseGoogleInitialPacket(packet, option.Recv)
		}

		isGoogleFrame = true
	}

	b := bytes.NewReader(packet)
	h, err := parseLongHeader(b)
	if err != nil {
		return nil, err
	}
	if h.PacketType != packetTypeInitial {
		return nil, ErrNotInitialPacket
	}
	headerLen := int64(b.Offset())

	p := option.PacketProtector
	if p == nil {
		p, err = newClientInitialProtector(h.Version, h.DestConnectionID)
		if err != nil {
			return nil, err
		}
	}

	maxHeaderLen := headerLen + 4
	header := make([]byte, maxHeaderLen)
	copy(header, packet[:maxHeaderLen])

	sample := packet[maxHeaderLen : maxHeaderLen+16]

	err = p.DecryptLongHeader(h, sample, header[headerLen:maxHeaderLen])
	if err != nil {
		return nil, err
	}
	header[0] = h.Flags.raw
	dataEnd := headerLen + h.Length
	headerLen += h.Flags.PacketNumberLen

	header = header[:headerLen]
	encrypted := packet[headerLen:dataEnd]

	decrypted, err := p.Open(nil, h.PacketNumber, encrypted, header)
	if err != nil {
		return nil, err
	}
	b.Reset(decrypted)

	var frames []FrameCrypto
	if isGoogleFrame {
		frames, err = parseGoogleFrames(b, h.Version, h.Flags.PacketNumberLen)
	} else {
		frames, err = parseFrames(b)
	}
	if err != nil {
		return nil, err
	}

	return &InitialPacket{
		Version:         v,
		PacketProtector: p,
		PacketNumber:    h.PacketNumber,
		Frames:          frames,
	}, nil
}
