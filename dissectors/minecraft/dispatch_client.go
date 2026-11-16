package minecraft

import (
	"fmt"

	"gitlab.com/go-extension/bytes"
)

func ClientDispatchStatus(version uint32, pkt *Packet) (Message, error) {
	r := bytes.NewReader(pkt.Data)

	id, err := readVarInt(r)
	if err != nil {
		return nil, err
	}

	switch id {
	case 0:
		return parseStatusResponse(version, r)
	case 1:
		return parsePongRequest(version, r)
	default:
		return nil, fmt.Errorf("dissectors/minecraft: bad packet id (Status, id=%d)", id)
	}
}
