package unsafe

import "unsafe"

func DataString[T ~[]byte](b T) string {
	return unsafe.String(unsafe.SliceData(b), len(b))
}

func StringData[T ~string](v T) []byte {
	s := *(*string)(unsafe.Pointer(&v))
	return unsafe.Slice(unsafe.StringData(s), len(s))
}
