package minecraft

import (
	"github.com/google/uuid"

	"gitlab.com/CoiaPrant/lib/unsafe"
	"gitlab.com/go-extension/bytes"
)

type LoginStart struct {
	UserName string
	UUID     uuid.UUID
}

func parseLoginStart(version uint32, r *bytes.Reader) (*LoginStart, error) {
	n, err := readVarInt(r)
	if err != nil {
		return nil, err
	}
	v, err := readFull(r, int(n))
	if err != nil {
		return nil, err
	}
	userName := unsafe.DataString(v)

	if version <= 759 {
		return &LoginStart{
			UserName: userName,
		}, nil
	}

	v, err = readFull(r, 16)
	if err != nil {
		return nil, err
	}
	uuid := uuid.UUID(v)

	return &LoginStart{
		UserName: userName,
		UUID:     uuid,
	}, nil
}

func (m *LoginStart) Format(version uint32) ([]byte, error) {
	id := formatVarInt(0)
	userNameLength := formatVarInt(uint32(len(m.UserName)))

	size := len(id)
	size += len(userNameLength)
	size += len(m.UserName)
	if version > 759 {
		size += 16
	}

	w := bytes.NewBuffer(make([]byte, 0, size))
	w.Write(id)
	w.Write(userNameLength)
	w.WriteString(m.UserName)
	if version > 759 {
		w.Write(m.UUID[:])
	}
	return w.Bytes(), nil
}
