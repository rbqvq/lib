package mtproto

type ConnectionType uint8

const (
	ConnectionTypeUnknown ConnectionType = iota
	ConnectionTypeAbridged
	ConnectionTypeIntermediate
	ConnectionTypePaddedIntermediate
)

// https://core.telegram.org/mtproto/mtproto-transports
var (
	ConnectionTagAbridged           = []byte{0xef, 0xef, 0xef, 0xef}
	ConnectionTagIntermediate       = []byte{0xee, 0xee, 0xee, 0xee}
	ConnectionTagPaddedIntermediate = []byte{0xdd, 0xdd, 0xdd, 0xdd}
)

func (t ConnectionType) Tag() []byte {
	switch t {
	case ConnectionTypeAbridged:
		return ConnectionTagAbridged
	case ConnectionTypeIntermediate:
		return ConnectionTagIntermediate
	case ConnectionTypePaddedIntermediate:
		return ConnectionTagPaddedIntermediate
	default:
		return nil
	}
}

type ConnectionState struct {
	Type       ConnectionType
	DataCenter uint16
}
