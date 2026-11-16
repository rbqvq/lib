package minecraft

import (
	"gitlab.com/go-extension/bytes"
)

type PingRequest struct {
	Timestamp int64
}

func parsePingRequest(version uint32, r *bytes.Reader) (*PingRequest, error) {
	v, err := readFull(r, 8)
	if err != nil {
		return nil, err
	}
	timestamp := long(v)

	return &PingRequest{
		Timestamp: timestamp,
	}, nil
}

func (m *PingRequest) Format(version uint32) ([]byte, error) {
	id := formatVarInt(1)

	size := len(id)
	size += 8

	w := bytes.NewBuffer(make([]byte, 0, size))
	w.Write(id)
	writeLong(w, m.Timestamp)
	return w.Bytes(), nil
}
