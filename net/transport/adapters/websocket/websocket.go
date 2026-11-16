package websocket

import (
	"io"
	"net"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

var _ net.Conn = (*WebsocketConn)(nil)
var _ io.ReaderFrom = (*WebsocketConn)(nil)

type WebsocketConn struct {
	in, out sync.Mutex
	conn    *websocket.Conn

	r io.Reader
}

func New(c *websocket.Conn) *WebsocketConn {
	return &WebsocketConn{conn: c}
}

func (c *WebsocketConn) next() error {
	_, r, err := c.conn.NextReader()
	switch err {
	case nil:
		c.r = r
		return nil
	default:
		c.r = nil
		return err
	}
}

func (c *WebsocketConn) Read(b []byte) (n int, err error) {
	c.in.Lock()
	defer c.in.Unlock()

	if c.r == nil {
		goto next
	}

read:
	n, err = c.r.Read(b)
	if err != io.EOF {
		return n, err
	}
	if n != 0 {
		return n, nil
	}

next:
	err = c.next()
	if err != nil {
		return 0, err
	}
	goto read
}

func (c *WebsocketConn) Write(b []byte) (int, error) {
	c.out.Lock()
	defer c.out.Unlock()

	err := c.conn.WriteMessage(websocket.BinaryMessage, b)
	if err != nil {
		return 0, err
	}

	return len(b), nil
}

func (c *WebsocketConn) ReadFrom(r io.Reader) (int64, error) {
	c.out.Lock()
	defer c.out.Unlock()

	w, err := c.conn.NextWriter(websocket.BinaryMessage)
	if err != nil {
		return 0, err
	}
	defer w.Close()

	return io.Copy(w, r)
}

func (c *WebsocketConn) Ping() error {
	c.out.Lock()
	defer c.out.Unlock()

	return c.conn.WriteMessage(websocket.PingMessage, nil)
}

func (c *WebsocketConn) Close() error {
	if c.out.TryLock() {
		defer c.out.Unlock()

		c.conn.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
	}

	return c.conn.Close()
}

func (c *WebsocketConn) LocalAddr() net.Addr {
	return c.conn.LocalAddr()
}

func (c *WebsocketConn) RemoteAddr() net.Addr {
	return c.conn.RemoteAddr()
}

func (c *WebsocketConn) SetDeadline(t time.Time) error {
	if err := c.SetReadDeadline(t); err != nil {
		return err
	}

	return c.SetWriteDeadline(t)
}

func (c *WebsocketConn) SetReadDeadline(t time.Time) error {
	return c.conn.SetReadDeadline(t)
}

func (c *WebsocketConn) SetWriteDeadline(t time.Time) error {
	return c.conn.SetWriteDeadline(t)
}
