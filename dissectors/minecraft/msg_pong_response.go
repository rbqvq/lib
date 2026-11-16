package minecraft

import (
	"gitlab.com/go-extension/bytes"
)

type PongRequest struct {
	Timestamp int64
}

func parsePongRequest(version uint32, r *bytes.Reader) (*PongRequest, error) {
	v, err := readFull(r, 8)
	if err != nil {
		return nil, err
	}
	timestamp := long(v)

	return &PongRequest{
		Timestamp: timestamp,
	}, nil
}

func (m *PongRequest) Format(version uint32) ([]byte, error) {
	id := formatVarInt(1)

	size := len(id)
	size += 8

	w := bytes.NewBuffer(make([]byte, 0, size))
	w.Write(id)
	writeLong(w, m.Timestamp)
	return w.Bytes(), nil
}
