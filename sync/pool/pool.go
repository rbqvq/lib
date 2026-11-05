package pool

import (
	"reflect"
	"sync"
)

type pool[T any] struct {
	pool  *sync.Pool
	reset func(item T)
}

func NewPool[T any](factory func() T, reset func(item T)) Pool[T] {
	return &pool[T]{
		pool: &sync.Pool{
			New: func() any {
				return factory()
			},
		},
		reset: reset,
	}
}

func (p *pool[T]) Get() T {
	item := p.pool.Get().(T)
	p.reset(item)
	return item
}

func (p *pool[T]) Put(item T) {
	if p.invalid(item) {
		return
	}

	p.reset(item)
	p.pool.Put(item)
}

func (p *pool[T]) invalid(item T) bool {
	v := reflect.ValueOf(item)

	switch v.Kind() {
	case reflect.Chan, reflect.Func, reflect.Map, reflect.Pointer, reflect.UnsafePointer,
		reflect.Interface, reflect.Slice:
		return v.IsNil()
	default:
		return false
	}
}
