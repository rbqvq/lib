package minecraft

import (
	"errors"
	"io"
)

const maxVarIntLen = 5

func readVarInt(r io.Reader) (val uint32, err error) {
	n := 0
	for sec := byte(0x80); sec&0x80 != 0; n++ {
		if n > maxVarIntLen {
			return 0, errors.New("VarInt is too big")
		}

		sec, err = readByte(r)
		if err != nil {
			return 0, err
		}

		val |= uint32(sec&0x7F) << uint32(7*n)
	}
	return
}

func formatVarInt(val uint32) []byte {
	vi := make([]byte, 0, maxVarIntLen)
	for {
		b := val & 0x7F
		val >>= 7
		if val != 0 {
			b |= 0x80
		}

		vi = append(vi, byte(b))
		if val == 0 {
			break
		}
	}

	return vi
}
