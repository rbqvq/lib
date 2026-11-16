package quic

import (
	"encoding/binary"
	"io"

	"gitlab.com/go-extension/bytes"
)

type CryptoGoogle struct {
	Tag       string
	TagValues map[string][]byte
}

func ParseCryptoGoogle(crypto []byte) (*CryptoGoogle, error) {
	b := bytes.NewReader(crypto)

	// Tag
	tag, err := readFull(b, 4)
	if err != nil {
		return nil, err
	}

	// Tag Number
	tagNumber, err := readUint16(binary.LittleEndian, b)
	if err != nil {
		return nil, err
	}

	// Padding
	b.Seek(2, io.SeekCurrent)

	tagValues := make(map[string][]byte)
	{
		tagsLength := int((4 + 4) * tagNumber)

		if b.Len() < tagsLength {
			return nil, io.ErrUnexpectedEOF
		}

		tags := bytes.NewReader(b.Next(tagsLength))

		var valueOffsetBegin uint32
		for tags.Len() > 0 {
			// Tag Type
			tagType, err := readFull(tags, 4)
			if err != nil {
				return nil, err
			}

			// Tag Offset End
			valueOffsetEnd, err := readUint32(binary.LittleEndian, tags)
			if err != nil {
				return nil, err
			}

			valueLength := valueOffsetEnd - valueOffsetBegin
			valueOffsetBegin = valueOffsetEnd

			tagValue, err := readFull(b, int(valueLength))
			if err != nil {
				return nil, err
			}

			tagValues[string(tagType)] = tagValue
		}
	}

	return &CryptoGoogle{Tag: string(tag), TagValues: tagValues}, nil
}
