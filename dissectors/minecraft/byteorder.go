package minecraft

import (
	"encoding/binary"
	"io"
)

func unsignedShort(b []byte) uint16 {
	return binary.BigEndian.Uint16(b)
}

func writeUnsignedShort(w io.Writer, v uint16) {
	binary.Write(w, binary.BigEndian, v)
}

func long(b []byte) int64 {
	_ = b[7] // bounds check hint to compiler; see golang.org/issue/14808
	return int64(b[7]) | int64(b[6])<<8 | int64(b[5])<<16 | int64(b[4])<<24 |
		int64(b[3])<<32 | int64(b[2])<<40 | int64(b[1])<<48 | int64(b[0])<<56
}

func writeLong(w io.Writer, v int64) {
	var b [8]byte
	b[0] = byte(v >> 56)
	b[1] = byte(v >> 48)
	b[2] = byte(v >> 40)
	b[3] = byte(v >> 32)
	b[4] = byte(v >> 24)
	b[5] = byte(v >> 16)
	b[6] = byte(v >> 8)
	b[7] = byte(v)
	w.Write(b[:])
}
