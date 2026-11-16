package minecraft

import (
	"encoding/json"

	"github.com/Tnze/go-mc/chat"
	"github.com/Tnze/go-mc/server"
	"gitlab.com/go-extension/bytes"
)

type StatusResponse struct {
	Version struct {
		Name     string `json:"name"`
		Protocol int    `json:"protocol"`
	} `json:"version"`
	Players struct {
		Max    int                   `json:"max"`
		Online int                   `json:"online"`
		Sample []server.PlayerSample `json:"sample"`
	} `json:"players"`
	Description *chat.Message `json:"description"`
	Favicon     string        `json:"favicon,omitempty"`
}

func parseStatusResponse(version uint32, r *bytes.Reader) (*StatusResponse, error) {
	n, err := readVarInt(r)
	if err != nil {
		return nil, err
	}
	v, err := readFull(r, int(n))
	if err != nil {
		return nil, err
	}

	var m StatusResponse
	err = json.Unmarshal(v, &m)
	return &m, err
}

func (m *StatusResponse) Format(version uint32) ([]byte, error) {
	data, err := json.Marshal(m)
	if err != nil {
		return nil, err
	}

	id := formatVarInt(0)
	length := formatVarInt(uint32(len(data)))

	size := len(id)
	size += len(length)
	size += len(data)

	w := bytes.NewBuffer(make([]byte, 0, size))
	w.Write(id)
	w.Write(length)
	w.Write(data)
	return w.Bytes(), nil
}
