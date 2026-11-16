package socks

type Version uint8

const (
	Version4 Version = 0x04
	Version5 Version = 0x05
)

func (v Version) Valid() bool {
	switch v {
	case Version4, Version5:
		return true
	default:
		return false
	}
}
