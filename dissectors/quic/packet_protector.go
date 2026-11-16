package quic

import (
	"crypto"
	"crypto/tls"
	"encoding/binary"
	"fmt"

	"golang.org/x/crypto/chacha20"
	"golang.org/x/crypto/chacha20poly1305"
)

type PacketProtector struct {
	headerProtector
	aead *xorNonceAEAD
}

func (p *PacketProtector) Overhead() int {
	return p.aead.Overhead()
}

func (p *PacketProtector) Seal(out []byte, packetNumber int64, plaintext, additionalData []byte) []byte {
	var nonce [8]byte
	binary.BigEndian.PutUint64(nonce[:], uint64(packetNumber))
	return p.aead.Seal(out, nonce[:], plaintext, additionalData)
}

func (p *PacketProtector) Open(out []byte, packetNumber int64, ciphertext, additionalData []byte) ([]byte, error) {
	var nonce [8]byte
	binary.BigEndian.PutUint64(nonce[:], uint64(packetNumber))
	return p.aead.Open(out, nonce[:], ciphertext, additionalData)
}

func newPacketProtector(suite uint16, trafficSecret []byte, keyLabel, ivLabel, hpLabel string) (*PacketProtector, error) {
	switch suite {
	case tls.TLS_AES_128_GCM_SHA256:
		key := hkdfExpandLabel(crypto.SHA256, trafficSecret, nil, keyLabel, 16)
		iv := hkdfExpandLabel(crypto.SHA256, trafficSecret, nil, ivLabel, aeadNonceLength)
		aead, err := aeadAESGCMTLS13(key, iv)
		if err != nil {
			return nil, err
		}

		hpKey := hkdfExpandLabel(crypto.SHA256, trafficSecret, nil, hpLabel, 16)
		hp, err := newAESHeaderProtector(hpKey)
		if err != nil {
			return nil, err
		}

		return &PacketProtector{aead: aead, headerProtector: hp}, nil
	case tls.TLS_AES_256_GCM_SHA384:
		key := hkdfExpandLabel(crypto.SHA384, trafficSecret, nil, keyLabel, 32)
		iv := hkdfExpandLabel(crypto.SHA384, trafficSecret, nil, ivLabel, aeadNonceLength)
		aead, err := aeadAESGCMTLS13(key, iv)
		if err != nil {
			return nil, err
		}

		hpKey := hkdfExpandLabel(crypto.SHA384, trafficSecret, nil, hpLabel, 32)
		hp, err := newAESHeaderProtector(hpKey)
		if err != nil {
			return nil, err
		}

		return &PacketProtector{aead: aead, headerProtector: hp}, nil
	case tls.TLS_CHACHA20_POLY1305_SHA256:
		key := hkdfExpandLabel(crypto.SHA256, trafficSecret, nil, keyLabel, chacha20poly1305.KeySize)
		iv := hkdfExpandLabel(crypto.SHA256, trafficSecret, nil, ivLabel, aeadNonceLength)
		aead, err := aeadChaCha20Poly1305(key, iv)
		if err != nil {
			return nil, err
		}

		hpKey := hkdfExpandLabel(crypto.SHA256, trafficSecret, nil, hpLabel, chacha20.KeySize)
		hp, err := newChaChaHeaderProtector(hpKey)
		if err != nil {
			return nil, err
		}

		return &PacketProtector{aead: aead, headerProtector: hp}, nil
	default:
		return nil, fmt.Errorf("quic.newPacketProtector: Invalid cipher suite id: %d", suite)
	}
}
