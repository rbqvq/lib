package mtproto

import (
	"crypto/cipher"
)

type Conn struct {
	cipher.StreamReader
	cipher.StreamWriter
	state ConnectionState
}

func (c *Conn) ConnectionState() ConnectionState {
	return c.state
}
