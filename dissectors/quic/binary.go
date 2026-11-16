package quic

import (
	"encoding/binary"
	"io"

	"gitlab.com/go-extension/bytes"
)

type integer interface {
	~int | ~int8 | ~int16 | ~int32 | ~int64 |
		~uint | ~uint8 | ~uint16 | ~uint32 | ~uint64
}

func readUint[T integer](byteOrder binary.ByteOrder, br io.ByteReader, length T) (uint64, error) {
	if length <= 0 {
		return 0, nil
	}
	if byteOrder == binary.BigEndian {
		return beReadUint(br, length)
	}
	return leReadUint(br, length)
}

func beReadUint[T integer](br io.ByteReader, length T) (uint64, error) {
	if length <= 0 {
		return 0, nil
	}

	var res uint64
	for i := length - 1; i >= 0; i-- {
		b, err := br.ReadByte()
		if err != nil {
			return 0, err
		}
		res |= uint64(b) << (i * 8)
	}
	return res, nil
}

func leReadUint[T integer](br io.ByteReader, length T) (uint64, error) {
	var res uint64
	for i := T(0); i < length; i++ {
		b, err := br.ReadByte()
		if err != nil {
			return 0, err
		}
		res |= uint64(b) << (i * 8)
	}
	return res, nil
}

func readUint16(byteOrder binary.ByteOrder, r *bytes.Reader) (uint16, error) {
	b, err := readFull(r, 2)
	if err != nil {
		return 0, err
	}
	return byteOrder.Uint16(b), nil
}

func leUint32(b []byte) uint32 {
	return binary.LittleEndian.Uint32(b)
}

func beUint32(b []byte) uint32 {
	return binary.BigEndian.Uint32(b)
}

func readUint32(byteOrder binary.ByteOrder, r *bytes.Reader) (uint32, error) {
	b, err := readFull(r, 4)
	if err != nil {
		return 0, err
	}
	return byteOrder.Uint32(b), nil
}

func readUint64(byteOrder binary.ByteOrder, r *bytes.Reader) (uint64, error) {
	b, err := readFull(r, 8)
	if err != nil {
		return 0, err
	}
	return byteOrder.Uint64(b), nil
}
