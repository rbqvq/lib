package quic

import (
	"crypto"
	"crypto/aes"
	"crypto/cipher"
	"crypto/tls"
	"fmt"

	"golang.org/x/crypto/chacha20"
)

type headerProtector interface {
	DecryptLongHeader(h *LongHeader, sample []byte, hdrBytes []byte) error
	DecryptShortHeader(h *ShortHeader, sample []byte, hdrBytes []byte) error
}

func newHeaderProtector(suite uint16, trafficSecret []byte, hkdfLabel string) (headerProtector, error) {
	switch suite {
	case tls.TLS_AES_128_GCM_SHA256:
		hpKey := hkdfExpandLabel(crypto.SHA256, trafficSecret, nil, hkdfLabel, 16)
		return newAESHeaderProtector(hpKey)
	case tls.TLS_AES_256_GCM_SHA384:
		hpKey := hkdfExpandLabel(crypto.SHA384, trafficSecret, nil, hkdfLabel, 32)
		return newAESHeaderProtector(hpKey)
	case tls.TLS_CHACHA20_POLY1305_SHA256:
		hpKey := hkdfExpandLabel(crypto.SHA256, trafficSecret, nil, hkdfLabel, 32)
		return newChaChaHeaderProtector(hpKey)
	default:
		return nil, fmt.Errorf("quic.newHeaderProtector: Invalid cipher suite id: %d", suite)
	}
}

type aesHeaderProtector struct {
	block         cipher.Block
	highestRcvdPN int64
}

var _ headerProtector = (*aesHeaderProtector)(nil)

func newAESHeaderProtector(hpKey []byte) (headerProtector, error) {
	block, err := aes.NewCipher(hpKey)
	if err != nil {
		return nil, fmt.Errorf("quic.newAESHeaderProtector: error creating new AES cipher: %s", err)
	}
	return &aesHeaderProtector{block: block}, nil
}

func (p *aesHeaderProtector) DecryptLongHeader(h *LongHeader, sample []byte, hdrBytes []byte) error {
	if size := len(sample); size != 16 {
		return fmt.Errorf("invalid sample size: %d", size)
	}

	var mask [aes.BlockSize]byte
	p.block.Encrypt(mask[:], sample)

	typeByte := h.Flags.raw
	typeByte ^= mask[0] & 0b00001111
	flags := parseLongHeaderFlags(typeByte, true)
	h.Flags = flags

	unmaskPacketNumber(h.Flags.PacketNumberLen, hdrBytes, mask[:])
	packetNumber := int64(readPacketNumber(h.Flags.PacketNumberLen, hdrBytes))
	h.PacketNumber = decodePacketNumber(h.Flags.PacketNumberLen, p.highestRcvdPN, packetNumber)
	p.highestRcvdPN = max(p.highestRcvdPN, h.PacketNumber)
	return nil
}

func (p *aesHeaderProtector) DecryptShortHeader(h *ShortHeader, sample []byte, hdrBytes []byte) error {
	if size := len(sample); size != 16 {
		return fmt.Errorf("invalid sample size: %d", size)
	}

	var mask [aes.BlockSize]byte
	p.block.Encrypt(mask[:], sample)

	typeByte := h.Flags.raw
	typeByte ^= mask[0] & 0b00011111
	flags := parseShortHeaderFlags(typeByte, true)
	h.Flags = flags

	unmaskPacketNumber(h.Flags.PacketNumberLen, hdrBytes, mask[:])
	packetNumber := int64(readPacketNumber(h.Flags.PacketNumberLen, hdrBytes))
	h.PacketNumber = decodePacketNumber(h.Flags.PacketNumberLen, p.highestRcvdPN, packetNumber)
	p.highestRcvdPN = max(p.highestRcvdPN, h.PacketNumber)
	return nil
}

type chachaHeaderProtector struct {
	key           []byte
	highestRcvdPN int64
}

var _ headerProtector = (*chachaHeaderProtector)(nil)

func newChaChaHeaderProtector(hpKey []byte) (headerProtector, error) {
	return &chachaHeaderProtector{key: hpKey}, nil
}

func (p *chachaHeaderProtector) DecryptLongHeader(h *LongHeader, sample []byte, hdrBytes []byte) error {
	if size := len(sample); size != 16 {
		return fmt.Errorf("invalid sample size: %d", size)
	}

	cipher, err := chacha20.NewUnauthenticatedCipher(p.key[:], sample[4:])
	if err != nil {
		return err
	}
	cipher.SetCounter(leUint32(sample[:4]))

	var mask [5]byte
	cipher.XORKeyStream(mask[:], mask[:])

	typeByte := h.Flags.raw
	typeByte ^= mask[0] & 0b00001111
	flags := parseLongHeaderFlags(typeByte, true)
	h.Flags = flags

	unmaskPacketNumber(h.Flags.PacketNumberLen, hdrBytes, mask[:])
	packetNumber := int64(readPacketNumber(h.Flags.PacketNumberLen, hdrBytes))
	h.PacketNumber = decodePacketNumber(h.Flags.PacketNumberLen, p.highestRcvdPN, packetNumber)
	p.highestRcvdPN = max(p.highestRcvdPN, h.PacketNumber)
	return nil
}

func (p *chachaHeaderProtector) DecryptShortHeader(h *ShortHeader, sample []byte, hdrBytes []byte) error {
	if size := len(sample); size != 16 {
		return fmt.Errorf("invalid sample size: %d", size)
	}

	cipher, err := chacha20.NewUnauthenticatedCipher(p.key[:], sample[4:])
	if err != nil {
		return err
	}
	cipher.SetCounter(leUint32(sample[:4]))

	var mask [5]byte
	cipher.XORKeyStream(mask[:], mask[:])

	typeByte := h.Flags.raw
	typeByte ^= mask[0] & 0b00011111
	flags := parseShortHeaderFlags(typeByte, true)
	h.Flags = flags

	unmaskPacketNumber(h.Flags.PacketNumberLen, hdrBytes, mask[:])
	packetNumber := int64(readPacketNumber(h.Flags.PacketNumberLen, hdrBytes))
	h.PacketNumber = decodePacketNumber(h.Flags.PacketNumberLen, p.highestRcvdPN, packetNumber)
	p.highestRcvdPN = max(p.highestRcvdPN, h.PacketNumber)
	return nil
}

// decodePacketNumber calculates the packet number based its length and the last seen packet number
// This function is taken from https://www.rfc-editor.org/rfc/rfc9000.html#section-a.3.
func decodePacketNumber(length, largest, truncated int64) int64 {
	expected := int64(largest + 1)
	win := int64(1 << (length * 8))
	hwin := win / 2
	mask := win - 1
	candidate := (expected & ^mask) | truncated
	if candidate <= expected-hwin && candidate < 1<<62-win {
		return candidate + win
	}
	if candidate > expected+hwin && candidate >= win {
		return candidate - win
	}
	return candidate
}

func unmaskPacketNumber(packetNumberLen int64, hdrBytes, mask []byte) {
	for i := range packetNumberLen {
		hdrBytes[i] ^= mask[i+1]
	}
}

func readPacketNumber(packetNumberLen int64, hdrBytes []byte) uint64 {
	switch packetNumberLen {
	case 1:
		return uint64(hdrBytes[0])
	case 2:
		return uint64(hdrBytes[0])<<8 | uint64(hdrBytes[1])
	case 3:
		return uint64(hdrBytes[0])<<16 | uint64(hdrBytes[1])<<8 | uint64(hdrBytes[2])
	case 4:
		return uint64(hdrBytes[0])<<24 | uint64(hdrBytes[1])<<16 | uint64(hdrBytes[2])<<8 | uint64(hdrBytes[3])
	default:
		panic("unreachable")
	}
}
