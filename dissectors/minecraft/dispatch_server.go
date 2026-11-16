package minecraft

import (
	"fmt"

	"gitlab.com/go-extension/bytes"
)

func ServerDispatch(pkt *Packet) (*Initial, error) {
	r := bytes.NewReader(pkt.Data)

	id, err := readVarInt(r)
	if err != nil {
		return nil, err
	}

	switch id {
	case 0:
		return parseInitial(r)
	default:
		return nil, fmt.Errorf("dissectors/minecraft: bad packet id (id=%d)", id)
	}
}

func ServerDispatchStatus(version uint32, pkt *Packet) (Message, error) {
	r := bytes.NewReader(pkt.Data)

	id, err := readVarInt(r)
	if err != nil {
		return nil, err
	}

	switch id {
	case 0:
		return parseStatusRequest(version, r)
	case 1:
		return parsePingRequest(version, r)
	default:
		return nil, fmt.Errorf("dissectors/minecraft: bad packet id (Status, id=%d)", id)
	}
}

func ServerDispatchLogin(version uint32, pkt *Packet) (Message, error) {
	r := bytes.NewReader(pkt.Data)

	id, err := readVarInt(r)
	if err != nil {
		return nil, err
	}

	switch id {
	case 0:
		return parseLoginStart(version, r)
	default:
		return nil, fmt.Errorf("dissectors/minecraft: bad packet id (Login, id=%d)", id)
	}
}

func ServerDispatchTransfer(version uint32, pkt *Packet) (Message, error) {
	if version < 766 {
		return nil, fmt.Errorf("dissectors/minecraft: transfer packet requires protocol >= 766, got %d", version)
	}

	r := bytes.NewReader(pkt.Data)

	id, err := readVarInt(r)
	if err != nil {
		return nil, err
	}

	switch id {
	case 0:
		return parseLoginStart(version, r)
	default:
		return nil, fmt.Errorf("dissectors/minecraft: bad packet id (Transfer, id=%d)", id)
	}
}
