package quic

import (
	"crypto"
	"crypto/tls"
	"fmt"
)

func newClientInitialProtector(version Version, dcid []byte) (*PacketProtector, error) {
	salt, keyLabel, ivLabel, hpLabel := initialHDKF(version)
	if salt == nil {
		return nil, fmt.Errorf("newClientHeanderProtecter: unsupported quic version %x", version)
	}

	initialSecret := hkdfExtract(crypto.SHA256.New, dcid, salt)
	clientSecret := hkdfExpandLabel(crypto.SHA256, initialSecret, nil, "client in", crypto.SHA256.Size())

	return newPacketProtector(tls.TLS_AES_128_GCM_SHA256, clientSecret, keyLabel, ivLabel, hpLabel)
}
