package mtproto

import (
	"fmt"
	"io"
)

func Client(conn io.ReadWriteCloser, state ConnectionState) (*Conn, error) {
	h, err := newHeader(state)
	if err != nil {
		return nil, err
	}

	a := h.Authentication(nil)
	decoder, err := a.Local()
	if err != nil {
		return nil, err
	}
	encoder, err := a.Peer()
	if err != nil {
		return nil, err
	}

	var header Header
	encoder.XORKeyStream(header[:], h[:])
	copy(header[:56], h[:])

	_, err = conn.Write(header[:])
	if err != nil {
		return nil, err
	}

	c := new(Conn)
	c.StreamReader.R = conn
	c.StreamReader.S = decoder
	c.StreamWriter.W = conn
	c.StreamWriter.S = encoder
	c.state = state
	return c, nil
}

func Server(conn io.ReadWriteCloser, secret Secret) (*Conn, error) {
	var h Header
	_, err := io.ReadFull(conn, h[:])
	if err != nil {
		return nil, err
	}

	a := h.Authentication(&secret)
	encoder, err := a.Local()
	if err != nil {
		return nil, err
	}
	decoder, err := a.Peer()
	if err != nil {
		return nil, err
	}

	decoder.XORKeyStream(h[:], h[:])
	typ := h.ConnectionType()
	if typ == ConnectionTypeUnknown {
		return nil, fmt.Errorf("unknown connection type: %#x", h.connectionType())
	}
	dataCenter := h.DataCenterID()

	c := new(Conn)
	c.StreamReader.R = conn
	c.StreamReader.S = decoder
	c.StreamWriter.W = conn
	c.StreamWriter.S = encoder
	c.state.Type = typ
	c.state.DataCenter = dataCenter
	return c, nil
}
