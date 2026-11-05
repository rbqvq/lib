package rand

import (
	"gitlab.com/CoiaPrant/lib/unsafe"
	"gitlab.com/go-extension/rand"
)

func Int(min, max int) int {
	if min == max {
		return min
	}

	if min > max {
		min, max = max, min
	}

	return min + rand.Standard.IntN(max-min+1)
}

func Uint16(min, max uint16) uint16 {
	if min == max {
		return min
	}

	if min > max {
		min, max = max, min
	}

	return min + uint16(rand.Standard.Uint32N(uint32(max-min+1)))
}

func String(n int) string {
	chars := "0123456789abcdefghigklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"

	str := make([]byte, n)
	for i := range n {
		str[i] = chars[rand.Standard.IntN(62)]
	}

	return unsafe.DataString(str)
}
