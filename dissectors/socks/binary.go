package socks

import (
	"encoding/binary"

	"gitlab.com/go-extension/bytes"
)

func readVersion(b *bytes.Reader) (Version, error) {
	v, err := b.ReadByte()
	if err != nil {
		return 0, err
	}

	return Version(v), nil
}

func readCommand(b *bytes.Reader) (Command, error) {
	v, err := b.ReadByte()
	if err != nil {
		return 0, err
	}

	return Command(v), nil
}

func beUint16(b []byte) uint16 {
	return binary.BigEndian.Uint16(b)
}

func readUint16(r *bytes.Reader) (uint16, error) {
	b, err := readFull(r, 2)
	if err != nil {
		return 0, err
	}
	return beUint16(b), nil
}
