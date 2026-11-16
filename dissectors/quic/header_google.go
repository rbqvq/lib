package quic

import (
	"encoding/binary"
	"io"

	"gitlab.com/go-extension/bytes"
)

func isGoogleHeader(typeByte byte) bool {
	return typeByte&0b11000000 == 0
}

func isGoogleHeaderWithVersion(typeByte byte) bool {
	return typeByte&0b11000001 == 0b00000001
}

type GoogleHeaderFlags struct {
	Version         bool
	Reset           bool
	Nonce           bool
	ConnectionIDLen int
	PacketNumberLen int64
	MultiPath       bool // Reserved and never be used
	Reserved        byte // Must be zero
}

func parseGoogleHeaderFlags(typeByte byte) (flags GoogleHeaderFlags) {
	flags.Version = (typeByte & 0b00000001) != 0
	flags.Reset = (typeByte & 0b00000010) != 0
	switch (typeByte & 0b00110000) >> 4 {
	case 0:
		flags.PacketNumberLen = 1
	case 1:
		flags.PacketNumberLen = 2
	case 2:
		flags.PacketNumberLen = 4
	case 3:
		flags.PacketNumberLen = 6
	}
	flags.MultiPath = (typeByte & 0b01000000) != 0
	flags.Reserved = typeByte >> 7
	return
}

func (flags *GoogleHeaderFlags) parseLegacy(typeByte byte) {
	switch (typeByte & 0b00001100) >> 2 {
	case 1:
		flags.ConnectionIDLen = 1
	case 2:
		flags.ConnectionIDLen = 4
	case 3:
		flags.ConnectionIDLen = 8
	}
}

func (flags *GoogleHeaderFlags) parseNew(typeByte byte) {
	flags.Nonce = (typeByte & 0b00000100) != 0
	if (typeByte & 0b00001000) != 0 {
		flags.ConnectionIDLen = 8
	}
}

type GoogleHeader struct {
	Flags                     GoogleHeaderFlags
	Version                   Version
	ConnectionID              []byte
	PacketNumber              uint64
	MessageAuthenticationHash []byte
}

func parseGoogleHeader(b *bytes.Reader, v Version) (*GoogleHeader, error) {
	typeByte, err := b.ReadByte()
	if err != nil {
		return nil, err
	}

	flags := parseGoogleHeaderFlags(typeByte)
	if v == 0 {
		return nil, ErrNotQUIC
	}
	if v < VersionQ033 {
		flags.parseLegacy(typeByte)
	} else {
		flags.parseNew(typeByte)
	}

	h := &GoogleHeader{Flags: flags}

	h.ConnectionID, err = readFull(b, flags.ConnectionIDLen)
	if err != nil {
		return nil, err
	}

	if flags.Version {
		ver, err := readUint32(binary.BigEndian, b)
		if err != nil {
			return nil, err
		}
		h.Version = Version(ver)
	} else {
		h.Version = v
	}

	var byteOrder binary.ByteOrder = binary.BigEndian
	if h.Version < VersionQ039 {
		byteOrder = binary.LittleEndian
	}

	h.PacketNumber, err = readUint(byteOrder, b, flags.PacketNumberLen)
	if err != nil {
		return nil, err
	}

	h.MessageAuthenticationHash, err = readFull(b, 12)
	if err != nil {
		return nil, err
	}

	if h.Version < VersionQ034 {
		b.Seek(1, io.SeekCurrent) // No Private Flags after Q034
	}

	return h, err
}

type GoogleLongHeader struct {
	Flags                     LongHeaderFlags
	PacketType                packetType
	Version                   Version
	ConnectionID              []byte
	PacketNumber              uint64
	MessageAuthenticationHash []byte
}

func parseGoogleLongHeader(b *bytes.Reader, recv bool) (*GoogleLongHeader, error) {
	typeByte, err := b.ReadByte()
	if err != nil {
		return nil, err
	}

	h := &GoogleLongHeader{Flags: parseLongHeaderFlags(typeByte, true)}

	ver, err := readUint32(binary.BigEndian, b)
	if err != nil {
		return nil, err
	}
	h.Version = Version(ver)

	lenByte, err := b.ReadByte()
	if err != nil {
		return nil, err
	}
	var connIDLen int
	if recv {
		connIDLen = int(lenByte & 0b00001111)
	} else {
		connIDLen = int(lenByte >> 4)
	}
	if connIDLen > 0 {
		connIDLen += 3

		h.ConnectionID, err = readFull(b, connIDLen)
		if err != nil {
			return nil, err
		}
	}
	h.PacketNumber, err = beReadUint(b, h.Flags.PacketNumberLen)
	if err != nil {
		return nil, err
	}

	h.MessageAuthenticationHash, err = readFull(b, 12)
	if err != nil {
		return nil, err
	}

	switch h.Flags.PacketType {
	case 0b00:
		h.PacketType = packetTypeInitial
	case 0b01:
		h.PacketType = packetType0RTT
	case 0b10:
		h.PacketType = packetTypeHandshake
	case 0b11:
		h.PacketType = packetTypeRetry
	}

	return h, err
}
