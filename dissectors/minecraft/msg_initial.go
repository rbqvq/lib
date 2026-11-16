package minecraft

import (
	"fmt"
	"strings"

	"gitlab.com/CoiaPrant/lib/unsafe"
	"gitlab.com/go-extension/bytes"
)

type Initial struct {
	Version    uint32
	ServerHost string
	ClientType ClientType
	ServerPort uint16
	Intent     Intent
}

func parseInitial(r *bytes.Reader) (*Initial, error) {
	version, err := readVarInt(r)
	if err != nil {
		return nil, err
	}

	n, err := readVarInt(r)
	if err != nil {
		return nil, err
	}
	v, err := readFull(r, int(n))
	if err != nil {
		return nil, err
	}
	serverHost, clientType, _ := strings.Cut(unsafe.DataString(v), "\x00")

	v, err = readFull(r, 2)
	if err != nil {
		return nil, err
	}
	serverPort := unsignedShort(v)

	intent, err := readVarInt(r)
	if err != nil {
		return nil, err
	}

	return &Initial{
		Version:    version,
		ServerHost: serverHost,
		ClientType: ClientType(clientType),
		ServerPort: serverPort,
		Intent:     Intent(intent),
	}, nil
}

func (m *Initial) Format(vers uint32) ([]byte, error) {
	if vers != m.Version {
		return nil, fmt.Errorf("unexpected version, want=%d, got=%d", m.Version, vers)
	}

	hostSize := len(m.ServerHost)
	if m.ClientType != "" {
		hostSize += 1
		hostSize += len(m.ClientType)
	}

	id := formatVarInt(0)
	version := formatVarInt(m.Version)
	serverHostLength := formatVarInt(uint32(hostSize))
	intent := formatVarInt(uint32(m.Intent))

	size := len(id)
	size += len(version)
	size += len(serverHostLength)
	size += hostSize
	size += 2
	size += len(intent)

	w := bytes.NewBuffer(make([]byte, 0, size))
	w.Write(id)
	w.Write(version)
	w.Write(serverHostLength)
	w.WriteString(m.ServerHost)
	if m.ClientType != "" {
		w.WriteByte(0)
		w.WriteString(string(m.ClientType))
	}
	writeUnsignedShort(w, m.ServerPort)
	w.Write(intent)
	return w.Bytes(), nil
}
