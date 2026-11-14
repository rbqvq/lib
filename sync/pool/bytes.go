// MIT License
//
// Copyright (c) 2016-2017 xtaci
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in all
// copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.

package pool

import (
	"sync"
)

type Bytes = []byte

var _ Pools[*Bytes] = (*bytesPool)(nil)

var (
	debruijnPos = [...]byte{0, 9, 1, 10, 13, 21, 2, 29, 11, 14, 16, 18, 22, 25, 3, 30, 8, 12, 20, 28, 15, 17, 24, 7, 19, 27, 23, 6, 26, 5, 4, 31}
)

type bytesPool struct {
	pools []sync.Pool
}

// NewBytesPool initiates a bytes pool for frames less than 65536 bytes,
// the waste(memory fragmentation) of space allocation is guaranteed to be
// no more than 50%.
func NewBytesPool() Pools[*Bytes] {
	p := new(bytesPool)

	p.pools = make([]sync.Pool, 17) // 1B -> 64K
	for k := range p.pools {
		i := k
		p.pools[k].New = func() any {
			buf := make(Bytes, 1<<uint32(i))
			return &buf
		}
	}
	return p
}

// Get a bytes from pool with most appropriate cap
func (p *bytesPool) Get(size int) *Bytes {
	if size <= 0 || size > 65536 {
		b := make(Bytes, size)
		return &b
	}

	bits := msb(size)
	if size == 1<<bits {
		buf := p.pools[bits].Get().(*Bytes)
		*buf = (*buf)[:size]
		return buf
	}

	buf := p.pools[bits+1].Get().(*Bytes)
	*buf = (*buf)[:size]
	return buf
}

// Put returns a bytes to pool for future use,
// which the cap must be exactly 2^n
func (p *bytesPool) Put(pBuf *Bytes) {
	if pBuf == nil {
		return
	}

	capacity := cap(*pBuf)
	bits := msb(capacity)
	if capacity == 0 || capacity > 65536 || capacity != 1<<bits {
		return
	}

	p.pools[bits].Put(pBuf)
}

// msb returns the pos of most significant bit
// http://supertech.csail.mit.edu/papers/debruijn.pdf
func msb(size int) byte {
	v := uint32(size)
	v |= v >> 1
	v |= v >> 2
	v |= v >> 4
	v |= v >> 8
	v |= v >> 16
	return debruijnPos[(v*0x07C4ACDD)>>27]
}
