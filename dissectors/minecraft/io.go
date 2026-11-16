package minecraft

import (
	"io"

	"gitlab.com/go-extension/bytes"
)

func readByte(r io.Reader) (byte, error) {
	if r, ok := r.(io.ByteReader); ok {
		return r.ReadByte()
	}

	if r, ok := r.(*bytes.Reader); ok {
		v, err := readFull(r, 1)
		if err != nil {
			return 0, err
		}
		return v[0], nil
	}

	v := make([]byte, 1)
	_, err := io.ReadFull(r, v)
	if err != nil {
		return 0, err
	}

	return v[0], err
}

func readFull(r io.Reader, n int) ([]byte, error) {
	if n < 0 {
		return nil, io.ErrShortBuffer
	}
	if n == 0 {
		return nil, nil
	}
	switch r := r.(type) {
	case *bytes.Reader:
		b := r.Next(n)
		if len(b) < n {
			return nil, io.ErrUnexpectedEOF
		}
		return b[:n], nil
	default:
		b := make([]byte, n)
		_, err := io.ReadFull(r, b)
		if err != nil {
			return nil, err
		}
		return b, nil
	}
}
