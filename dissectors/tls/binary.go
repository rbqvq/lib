package tls

import (
	"encoding/binary"

	"gitlab.com/go-extension/bytes"
)

func beUint16(b []byte) uint16 {
	return binary.BigEndian.Uint16(b)
}

func readUint16(r *bytes.Reader) (uint16, error) {
	b, err := readFull(r, 2)
	if err != nil {
		return 0, err
	}
	return beUint16(b), nil
}

func beUint24(b []byte) uint32 {
	_ = b[2] // bounds check hint to compiler; see golang.org/issue/14808
	return uint32(b[2]) | uint32(b[1])<<8 | uint32(b[0])<<16
}

func readUint24(r *bytes.Reader) (uint32, error) {
	b, err := readFull(r, 3)
	if err != nil {
		return 0, err
	}
	return beUint24(b), nil
}

// func leUint32(b []byte) uint32 {
// 	return binary.LittleEndian.Uint32(b)
// }

// func beUint32(b []byte) uint32 {
// 	return binary.BigEndian.Uint32(b)
// }

// func readUint32(byteOrder binary.ByteOrder, r io.Reader) (uint32, error) {
// 	b := make([]byte, 4)
// 	if _, err := io.ReadFull(r, b); err != nil {
// 		return 0, err
// 	}
// 	return byteOrder.Uint32(b), nil
// }

// func readUint64(byteOrder binary.ByteOrder, r io.Reader) (uint64, error) {
// 	b := make([]byte, 8)
// 	if _, err := io.ReadFull(r, b); err != nil {
// 		return 0, err
// 	}
// 	return byteOrder.Uint64(b), nil
// }
