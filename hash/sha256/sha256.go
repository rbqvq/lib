package sha256

import (
	"crypto/sha256"
	"encoding/hex"

	"gitlab.com/CoiaPrant/lib/unsafe"
)

func Sum[T ~[]byte](data T) [32]byte {
	return sha256.Sum256(data)
}

func SumHex[T ~[]byte](data T) string {
	hash := sha256.Sum256(data)
	return hex.EncodeToString(hash[:])
}

func String[T ~string](data T) [32]byte {
	return sha256.Sum256(unsafe.StringData(data))
}

func StringHex[T ~string](data T) string {
	hash := sha256.Sum256(unsafe.StringData(data))
	return hex.EncodeToString(hash[:])
}
