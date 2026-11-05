package netio

import (
	"io"
	"net"
)

func Copy(dst, src net.Conn) (int64, error) {
	defer src.Close()
	defer dst.Close()

	return io.Copy(dst, src)
}
