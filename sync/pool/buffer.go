package pool

import (
	"bytes"
	"sync"
)

var _ Pool[*bytes.Buffer] = (*bufferPool)(nil)

type bufferPool struct {
	pool *sync.Pool
}

func NewBufferPool() Pool[*bytes.Buffer] {
	return &bufferPool{
		pool: &sync.Pool{
			New: func() any {
				return bytes.NewBuffer(nil)
			},
		},
	}
}

func (p *bufferPool) Get() *bytes.Buffer {
	buffer := p.pool.Get().(*bytes.Buffer)
	buffer.Reset()
	return buffer
}

func (p *bufferPool) Put(buffer *bytes.Buffer) {
	if buffer == nil {
		return
	}

	buffer.Reset()
	p.pool.Put(buffer)
}
