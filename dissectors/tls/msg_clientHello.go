package tls

import (
	"io"

	"gitlab.com/go-extension/bytes"
)

type ClientHello struct {
	Length       uint16
	Version      uint16
	CipherSuites []uint16
	ServerName   string
	ALPN         []string
	KeyShare     []KeyShare
	ECH          *ExtensionECH
	Extensions   []uint16
}

func ParseClientHello(r *bytes.Reader) (*ClientHello, error) {
	m := &ClientHello{Length: uint16(r.Len())}

	var err error
	m.Version, err = readUint16(r) // Version
	if err != nil {
		return nil, err
	}

	r.Seek(4+28, io.SeekCurrent) // Random (timestamp [4 bytes] + random [28 bytes])

	sessionIDLen, err := r.ReadByte() // Sessuin ID Length
	if err != nil {
		return nil, err
	}
	if sessionIDLen > 0 {
		r.Seek(int64(sessionIDLen), io.SeekCurrent) // Session ID
	}

	m.CipherSuites, err = parseCipherSuites(r)
	if err != nil {
		return nil, err
	}

	compressionMethodsLen, err := r.ReadByte() // Compression Methods Length
	if err != nil {
		return nil, err
	}
	if compressionMethodsLen > 0 {
		r.Seek(int64(compressionMethodsLen), io.SeekCurrent) // Compression Methods
	}

	if r.Len() == 0 {
		return m, nil
	}

	extensionLen, err := readUint16(r) // Extensions Length
	if err != nil {
		return nil, err
	}
	if extensionLen == 0 {
		return m, nil
	}

	extensionData, err := readFull(r, int(extensionLen))
	if err != nil {
		return nil, err
	}

	extension := bytes.NewReader(extensionData)
	for extension.Len() > 0 {
		extensionType, err := readUint16(extension) // Extension Type
		if err != nil {
			return nil, err
		}
		m.Extensions = append(m.Extensions, extensionType)

		extensionLen, err := readUint16(extension) // Extension Length
		if err != nil {
			return nil, err
		}

		if extensionLen == 0 {
			continue
		}

		extensionData, err := readFull(extension, int(extensionLen))
		if err != nil {
			return nil, err
		}

		extension := bytes.NewReader(extensionData)
		switch extensionType {
		case extensionServerName:
			m.ServerName, err = parseExtensionServerName(extension)
			if err != nil {
				return nil, err
			}
		case extensionALPN:
			m.ALPN, err = parseExtensionALPN(extension)
			if err != nil {
				return nil, err
			}
		case extensionKeyShare:
			m.KeyShare, err = parseExtensionKeyShare(extension)
			if err != nil {
				return nil, err
			}
		case extensionECH:
			m.ECH, err = parseExtensionECH(extension)
			if err != nil {
				return nil, err
			}
		}
	}

	return m, nil
}

func parseCipherSuites(r *bytes.Reader) ([]uint16, error) {
	length, err := readUint16(r)
	if err != nil {
		return nil, err
	}

	if length == 0 {
		return nil, nil
	}

	data, err := readFull(r, int(length))
	if err != nil {
		return nil, err
	}
	r = bytes.NewReader(data)

	var cipherSuites []uint16
	for r.Len() > 0 {
		cipherSuite, err := readUint16(r)
		if err != nil {
			return nil, err
		}

		cipherSuites = append(cipherSuites, cipherSuite)
	}
	return cipherSuites, nil
}
