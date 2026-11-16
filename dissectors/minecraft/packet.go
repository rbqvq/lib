package minecraft

import (
	"io"

	"gitlab.com/go-extension/bytes"
)

type Packet struct {
	Data []byte
}

func Parse(r io.Reader) (*Packet, error) {
	length, err := readVarInt(r)
	if err != nil {
		return nil, err
	}

	data, err := readFull(r, int(length))
	if err != nil {
		return nil, err
	}

	return &Packet{
		Data: data,
	}, nil
}

func (pkt *Packet) WriteTo(w io.Writer) (int64, error) {
	size := len(pkt.Data)

	length := formatVarInt(uint32(size))
	size += len(length)

	b := bytes.NewBuffer(make([]byte, 0, size))
	b.Write(length)
	b.Write(pkt.Data)
	return b.WriteTo(w)
}
