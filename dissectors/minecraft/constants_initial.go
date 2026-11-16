package minecraft

import "strings"

type ClientType string

func (typ ClientType) String() string {
	if typ == "" {
		return "Official"
	}

	return strings.TrimSuffix(string(typ), "\x00")
}

type Intent uint32

const (
	_ Intent = iota
	IntentStatus
	IntentLogin
	IntentTransfer
)
