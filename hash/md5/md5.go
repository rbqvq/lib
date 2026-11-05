package md5

import (
	"crypto/md5"
	"encoding/hex"

	"gitlab.com/CoiaPrant/lib/unsafe"
)

func Sum[T ~[]byte](data T) [16]byte {
	return md5.Sum(data)
}

func SumHex[T ~[]byte](data T) string {
	hash := md5.Sum(data)
	return hex.EncodeToString(hash[:])
}

func String[T ~string](data T) [16]byte {
	hash := md5.Sum(unsafe.StringData(data))
	return hash
}

func StringHex[T ~string](data T) string {
	hash := md5.Sum(unsafe.StringData(data))
	return hex.EncodeToString(hash[:])
}
