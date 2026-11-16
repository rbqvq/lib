package tls

import "crypto/tls"

// TLS record types.
type RecordType uint8

const (
	RecordTypeChangeCipherSpec RecordType = 20
	RecordTypeAlert            RecordType = 21
	RecordTypeHandshake        RecordType = 22
	RecordTypeApplicationData  RecordType = 23
)

type RecordHeader struct {
	Type    RecordType
	Version uint16
	Length  uint16
}

func ParseRecordHeader(hdr []byte) (RecordHeader, error) {
	typ := RecordType(hdr[0])
	switch typ {
	case RecordTypeChangeCipherSpec, RecordTypeAlert, RecordTypeHandshake, RecordTypeApplicationData:
	default:
		return RecordHeader{}, ErrNotTLS
	}

	ver := beUint16(hdr[1:3])
	switch ver {
	case tls.VersionSSL30, tls.VersionTLS10, tls.VersionTLS11, tls.VersionTLS12:
	default:
		return RecordHeader{}, ErrNotTLS
	}

	return RecordHeader{
		Type:    typ,
		Version: ver,
		Length:  beUint16(hdr[3:5]),
	}, nil
}
