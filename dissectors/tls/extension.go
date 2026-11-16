package tls

import "gitlab.com/go-extension/bytes"

// TLS extension numbers
const (
	extensionServerName              uint16 = 0
	extensionStatusRequest           uint16 = 5
	extensionSupportedCurves         uint16 = 10 // supported_groups in TLS 1.3, see RFC 8446, Section 4.2.7
	extensionSupportedPoints         uint16 = 11
	extensionSignatureAlgorithms     uint16 = 13
	extensionALPN                    uint16 = 16
	extensionSCT                     uint16 = 18
	extensionPadding                 uint16 = 21
	extensionExtendedMasterSecret    uint16 = 23
	extensionCompressCert            uint16 = 27 // compress_certificate in TLS 1.3
	extensionSessionTicket           uint16 = 35
	extensionPreSharedKey            uint16 = 41
	extensionEarlyData               uint16 = 42
	extensionSupportedVersions       uint16 = 43
	extensionCookie                  uint16 = 44
	extensionPSKModes                uint16 = 45
	extensionCertificateAuthorities  uint16 = 47
	extensionSignatureAlgorithmsCert uint16 = 50
	extensionKeyShare                uint16 = 51
	extensionQUICTransportParameters uint16 = 57
	extensionALPS                    uint16 = 17513
	extensionRenegotiationInfo       uint16 = 0xff01
	extensionECH                     uint16 = 0xfe0d // draft-ietf-tls-esni-13
	extensionECHOuterExtensions      uint16 = 0xfd00 // draft-ietf-tls-esni-13
)

func parseExtensionServerName(r *bytes.Reader) (string, error) {
	length, err := readUint16(r)
	if err != nil {
		return "", err
	}

	if length == 0 {
		return "", nil
	}

	data, err := readFull(r, int(length))
	if err != nil {
		return "", err
	}
	r = bytes.NewReader(data)

	for r.Len() > 0 {
		nameType, err := r.ReadByte()
		if err != nil {
			return "", err
		}

		length, err := readUint16(r)
		if err != nil {
			return "", err
		}

		value, err := readFull(r, int(length))
		if err != nil {
			return "", err
		}

		if nameType == 0 {
			return string(value), nil
		}
	}
	return "", nil
}

func parseExtensionALPN(r *bytes.Reader) ([]string, error) {
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

	var alpn []string
	for r.Len() > 0 {
		length, err := r.ReadByte()
		if err != nil {
			return nil, err
		}

		value, err := readFull(r, int(length))
		if err != nil {
			return nil, err
		}

		alpn = append(alpn, string(value))
	}

	return alpn, nil
}

type KeyShare struct {
	Group    uint16
	Exchange []byte
}

func parseExtensionKeyShare(r *bytes.Reader) ([]KeyShare, error) {
	length, err := readUint16(r)
	if err != nil {
		return nil, err
	}

	if length == 0 {
		return nil, ErrDecode
	}

	data, err := readFull(r, int(length))
	if err != nil {
		return nil, err
	}
	r = bytes.NewReader(data)

	var keyShares []KeyShare
	for r.Len() > 0 {
		group, err := readUint16(r)
		if err != nil {
			return nil, err
		}

		length, err := readUint16(r)
		if err != nil {
			return nil, err
		}

		exchange, err := readFull(r, int(length))
		if err != nil {
			return nil, err
		}

		keyShares = append(keyShares, KeyShare{Group: group, Exchange: exchange})
	}

	return keyShares, nil
}

type ExtensionECH struct {
	KDF  uint16
	AEAD uint16
}

func parseExtensionECH(r *bytes.Reader) (*ExtensionECH, error) {
	typ, err := r.ReadByte()
	if err != nil {
		return nil, err
	}

	if typ != 0 {
		return nil, ErrDecode
	}

	kdf, err := readUint16(r)
	if err != nil {
		return nil, err
	}

	aead, err := readUint16(r)
	if err != nil {
		return nil, err
	}

	return &ExtensionECH{
		KDF:  kdf,
		AEAD: aead,
	}, nil
}
