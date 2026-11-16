package minecraft

import (
	"encoding/json"

	"github.com/Tnze/go-mc/chat"
	"gitlab.com/go-extension/bytes"
)

type Disconnect struct {
	Reason chat.Message
}

func parseDisconnect(version uint32, r *bytes.Reader) (*Disconnect, error) {
	n, err := readVarInt(r)
	if err != nil {
		return nil, err
	}
	v, err := readFull(r, int(n))
	if err != nil {
		return nil, err
	}

	var m Disconnect
	err = json.Unmarshal(v, &m.Reason)
	if err != nil {
		return nil, err
	}
	return &m, nil
}

func (m *Disconnect) Format(version uint32) ([]byte, error) {
	reason, err := json.Marshal(m.Reason)
	if err != nil {
		return nil, err
	}

	id := formatVarInt(0)
	reasonLength := formatVarInt(uint32(len(reason)))

	size := len(id)
	size += len(reasonLength)
	size += len(reason)

	w := bytes.NewBuffer(make([]byte, 0, size))
	w.Write(id)
	w.Write(reasonLength)
	w.Write(reason)
	return w.Bytes(), nil
}
