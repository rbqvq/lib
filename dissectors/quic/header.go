package quic

import (
	"encoding/binary"
	"io"

	"gitlab.com/go-extension/bytes"

	"github.com/quic-go/quic-go/quicvarint"
)

func isHeader(typeByte byte) bool {
	return typeByte&0b01000000 != 0
}

func isLongHeader(typeByte byte) bool {
	return typeByte&0b10000000 != 0
}

type LongHeaderFlags struct {
	raw             byte
	PacketType      byte
	Reserved        byte
	PacketNumberLen int64
}

func parseLongHeaderFlags(typeByte byte, decrypted bool) (flags LongHeaderFlags) {
	flags.raw = typeByte
	flags.PacketType = (typeByte & 0b00110000) >> 4
	if decrypted {
		flags.Reserved = (typeByte & 0b00001100) >> 2
		flags.PacketNumberLen = int64(typeByte&0b00000011 + 1)
	}
	return
}

type LongHeader struct {
	Flags            LongHeaderFlags
	PacketType       packetType
	Version          Version
	SrcConnectionID  []byte
	DestConnectionID []byte
	Token            []byte
	Length           int64
	PacketNumber     int64
}

func parseLongHeader(b *bytes.Reader) (*LongHeader, error) {
	typeByte, err := b.ReadByte()
	if err != nil {
		return nil, err
	}

	h := &LongHeader{Flags: parseLongHeaderFlags(typeByte, false)}

	ver, err := readUint32(binary.BigEndian, b)
	if err != nil {
		return nil, err
	}
	h.Version = Version(ver)

	destConnIDLen, err := b.ReadByte()
	if err != nil {
		return nil, err
	}
	h.DestConnectionID, err = readFull(b, int(destConnIDLen))
	if err != nil {
		return nil, err
	}

	srcConnIDLen, err := b.ReadByte()
	if err != nil {
		return nil, err
	}
	h.SrcConnectionID, err = readFull(b, int(srcConnIDLen))
	if err != nil {
		return nil, err
	}

	if h.Version.v2() {
		switch h.Flags.PacketType {
		case 0b00:
			h.PacketType = packetTypeRetry
		case 0b01:
			h.PacketType = packetTypeInitial
		case 0b10:
			h.PacketType = packetType0RTT
		case 0b11:
			h.PacketType = packetTypeHandshake
		}
	} else {
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
	}

	switch h.PacketType {
	case packetTypeRetry:
		integrityTagLen := 16
		if h.Version.isRetryLegacy() {
			integrityTagLen = 0

			originalDestConnIDLen, err := b.ReadByte()
			if err != nil {
				return nil, err
			}
			b.Seek(int64(originalDestConnIDLen), io.SeekCurrent)
		}

		tokenLen := b.Len() - integrityTagLen
		if tokenLen <= 0 {
			return nil, io.EOF
		}
		h.Token, err = readFull(b, int(tokenLen))
		if err != nil {
			return nil, err
		}
		b.Seek(int64(integrityTagLen), io.SeekCurrent)
		return h, nil
	case packetTypeInitial:
		tokenLen, err := quicvarint.Read(b)
		if err != nil {
			return nil, err
		}
		h.Token, err = readFull(b, int(tokenLen))
		if err != nil {
			return nil, err
		}
	}

	pl, err := quicvarint.Read(b)
	if err != nil {
		return nil, err
	}
	h.Length = int64(pl)
	return h, err
}

type ShortHeaderFlags struct {
	raw             byte
	Spin            bool
	Reserved        byte
	KeyPhare        bool
	PacketNumberLen int64
}

func parseShortHeaderFlags(typeByte byte, decrypted bool) (flags ShortHeaderFlags) {
	flags.raw = typeByte
	flags.Spin = typeByte&0b00100000 != 0
	if decrypted {
		flags.Reserved = (typeByte & 0b00011000) >> 3
		flags.KeyPhare = typeByte&0b00000100 != 0
		flags.PacketNumberLen = int64(typeByte&0b00000011 + 1)
	}
	return
}

type ShortHeader struct {
	Flags            ShortHeaderFlags
	DestConnectionID []byte
	PacketNumber     int64
}

func parseShortHeader(b *bytes.Reader, destConnIDLen int) (*ShortHeader, error) {
	typeByte, err := b.ReadByte()
	if err != nil {
		return nil, err
	}
	h := &ShortHeader{Flags: parseShortHeaderFlags(typeByte, false)}

	h.DestConnectionID, err = readFull(b, int(destConnIDLen))
	if err != nil {
		return nil, err
	}

	return h, nil
}
