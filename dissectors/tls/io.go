package tls

import (
	"io"

	"gitlab.com/go-extension/bytes"
)

func readFull(r *bytes.Reader, n int) ([]byte, error) {
	if n < 0 {
		return nil, io.ErrShortBuffer
	}
	if n == 0 {
		return nil, nil
	}
	b := r.Next(n)
	if len(b) < n {
		return nil, io.ErrUnexpectedEOF
	}
	return b[:n], nil
}
