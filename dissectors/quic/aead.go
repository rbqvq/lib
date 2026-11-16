package quic

import (
	"crypto/aes"
	"crypto/cipher"
	"errors"
	"fmt"

	"golang.org/x/crypto/chacha20poly1305"
)

const (
	aeadNonceLength   = 12
	noncePrefixLength = 4
)

// xorNonceAEAD wraps an AEAD by XORing in a fixed pattern to the nonce
// before each call.
type xorNonceAEAD struct {
	nonceMask [aeadNonceLength]byte
	aead      cipher.AEAD
}

func (f *xorNonceAEAD) NonceSize() int { return 8 } // 64-bit sequence number
func (f *xorNonceAEAD) Overhead() int  { return f.aead.Overhead() }

func (f *xorNonceAEAD) Seal(out, nonce, plaintext, additionalData []byte) []byte {
	for i, b := range nonce {
		f.nonceMask[4+i] ^= b
	}
	result := f.aead.Seal(out, f.nonceMask[:], plaintext, additionalData)
	for i, b := range nonce {
		f.nonceMask[4+i] ^= b
	}

	return result
}

func (f *xorNonceAEAD) Open(out, nonce, ciphertext, additionalData []byte) ([]byte, error) {
	for i, b := range nonce {
		f.nonceMask[4+i] ^= b
	}
	result, err := f.aead.Open(out, f.nonceMask[:], ciphertext, additionalData)
	for i, b := range nonce {
		f.nonceMask[4+i] ^= b
	}

	return result, err
}

func aeadAESGCMTLS13(key, nonceMask []byte) (*xorNonceAEAD, error) {
	if len(nonceMask) != aeadNonceLength {
		return nil, errors.New("quic: internal error: wrong nonce length")
	}
	aes, err := aes.NewCipher(key)
	if err != nil {
		return nil, fmt.Errorf("quic: error creating new cipher: %s", err)
	}
	aead, err := cipher.NewGCM(aes)
	if err != nil {
		return nil, fmt.Errorf("quic: error creating new aead cipher: %s", err)
	}

	ret := &xorNonceAEAD{aead: aead}
	copy(ret.nonceMask[:], nonceMask)
	return ret, nil
}

func aeadChaCha20Poly1305(key, nonceMask []byte) (*xorNonceAEAD, error) {
	if len(nonceMask) != aeadNonceLength {
		return nil, errors.New("quic: internal error: wrong nonce length")
	}
	aead, err := chacha20poly1305.New(key)
	if err != nil {
		return nil, fmt.Errorf("quic: error creating new aead cipher: %s", err)
	}

	ret := &xorNonceAEAD{aead: aead}
	copy(ret.nonceMask[:], nonceMask)
	return ret, nil
}
