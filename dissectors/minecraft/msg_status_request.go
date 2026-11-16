package minecraft

import (
	"gitlab.com/go-extension/bytes"
)

type StatusRequest struct{}

func parseStatusRequest(version uint32, r *bytes.Reader) (*StatusRequest, error) {
	return &StatusRequest{}, nil
}

func (m *StatusRequest) Format(version uint32) ([]byte, error) {
	id := formatVarInt(0)

	size := len(id)

	w := bytes.NewBuffer(make([]byte, 0, size))
	w.Write(id)
	return w.Bytes(), nil
}
