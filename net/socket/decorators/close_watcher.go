package decorators

import (
	"net"
	"sync/atomic"
)

var _ net.Conn = (*CloseWatcher)(nil)

type CloseWatcher struct {
	net.Conn

	isClosed atomic.Bool
	done     chan struct{}
}

func WithCloseWatcher(conn net.Conn) *CloseWatcher {
	return &CloseWatcher{
		Conn: conn,
		done: make(chan struct{}),
	}
}

func (conn *CloseWatcher) Close() error {
	if conn.isClosed.Swap(true) {
		return net.ErrClosed
	}

	close(conn.done)
	return conn.Conn.Close()
}

func (conn *CloseWatcher) IsClosed() bool {
	return conn.isClosed.Load()
}

func (conn *CloseWatcher) Done() <-chan struct{} {
	return conn.done
}
