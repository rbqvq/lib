package mtproto

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
)

type Secret [16]byte

func (s Secret) String() string {
	return hex.EncodeToString(s[:])
}

func (s Secret) MarshalJSON() ([]byte, error) {
	return json.Marshal(s.String())
}

func (s *Secret) UnmarshalJSON(data []byte) error {
	var content string
	if err := json.Unmarshal(data, &content); err != nil {
		return err
	}
	raw, err := hex.DecodeString(content)
	if err != nil {
		return err
	}
	if len(raw) != len(s) {
		return fmt.Errorf("secret: expected %d bytes, got %d", len(s), len(raw))
	}
	copy(s[:], raw)
	return nil
}

type Authentication struct {
	LocalKey, PeerKey     [32]byte
	LocalNonce, PeerNonce [16]byte
}

func (a *Authentication) Local() (cipher.Stream, error) {
	block, err := aes.NewCipher(a.LocalKey[:])
	if err != nil {
		return nil, err
	}

	return cipher.NewCTR(block, a.LocalNonce[:]), nil
}

func (a *Authentication) Peer() (cipher.Stream, error) {
	block, err := aes.NewCipher(a.PeerKey[:])
	if err != nil {
		return nil, err
	}

	return cipher.NewCTR(block, a.PeerNonce[:]), nil
}

func computeKey(dst, key []byte, secret *Secret) {
	if secret == nil {
		copy(dst, key)
		return
	}

	hash := sha256.New()
	hash.Write(key)
	hash.Write(secret[:])
	hash.Sum(dst[:0])
}
