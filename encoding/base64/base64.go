package base64

import (
	"encoding/base64"

	"gitlab.com/CoiaPrant/lib/unsafe"
)

type Encoding = *base64.Encoding

var (
	StdEncoding    Encoding = base64.StdEncoding
	URLEncoding    Encoding = base64.URLEncoding
	RawStdEncoding Encoding = base64.RawStdEncoding
	RawURLEncoding Encoding = base64.RawURLEncoding
)

func Encode[T ~[]byte](encoding Encoding, src T) []byte {
	dst := make([]byte, encoding.EncodedLen(len(src)))
	encoding.Encode(dst, src)
	return dst
}

func EncodeToString[T ~[]byte](encoding Encoding, src T) string {
	dst := make([]byte, encoding.EncodedLen(len(src)))
	encoding.Encode(dst, src)
	return unsafe.DataString(dst)
}

func EncodeString[T ~string](encoding Encoding, src T) []byte {
	return Encode(encoding, unsafe.StringData(src))
}

func StringEncode[T ~string](encoding Encoding, src T) string {
	return EncodeToString(encoding, unsafe.StringData(src))
}

func Decode[T ~[]byte](encoding Encoding, src T) ([]byte, error) {
	dst := make([]byte, encoding.DecodedLen(len(src)))

	n, err := encoding.Decode(dst, src)
	if err != nil {
		return nil, err
	}

	return dst[:n], nil
}

func DecodeToString[T ~[]byte](encoding Encoding, src T) (string, error) {
	dst := make([]byte, encoding.DecodedLen(len(src)))

	n, err := encoding.Decode(dst, src)
	if err != nil {
		return "", err
	}

	return unsafe.DataString(dst[:n]), nil
}

func DecodeString[T ~string](encoding Encoding, src T) ([]byte, error) {
	return Decode(encoding, unsafe.StringData(src))
}

func StringDecode[T ~string](encoding Encoding, src T) (string, error) {
	return DecodeToString(encoding, unsafe.StringData(src))
}
