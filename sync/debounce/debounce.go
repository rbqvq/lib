package debounce

import (
	"runtime"
	"time"
	"weak"

	"github.com/puzpuzpuz/xsync/v4"
	"gitlab.com/CoiaPrant/lib/sync"
)

type Group[K comparable] struct {
	d time.Duration
	m *xsync.Map[K, weak.Pointer[sync.Debounce]]
}

func New[K comparable](interval time.Duration) *Group[K] {
	return &Group[K]{
		d: interval,
		m: xsync.NewMap[K, weak.Pointer[sync.Debounce]](),
	}
}

func (g *Group[K]) invokeCleanup(key K) {
	g.m.Compute(key, func(v weak.Pointer[sync.Debounce], _ bool) (weak.Pointer[sync.Debounce], xsync.ComputeOp) {
		if v.Value() != nil {
			return v, xsync.CancelOp
		}

		return v, xsync.DeleteOp
	})
}

func (g *Group[K]) Stop(key K) bool {
	v, ok := g.m.Load(key)
	if !ok {
		return false
	}

	d := v.Value()
	if d == nil {
		return false
	}

	return d.Stop()
}

func (g *Group[K]) Trigger(key K) bool {
	var d *sync.Debounce
	g.m.Compute(key, func(v weak.Pointer[sync.Debounce], _ bool) (weak.Pointer[sync.Debounce], xsync.ComputeOp) {
		if d = v.Value(); d != nil {
			return v, xsync.CancelOp
		}

		d = sync.NewDebounce(g.d)
		runtime.AddCleanup(d, g.invokeCleanup, key)
		return weak.Make(d), xsync.UpdateOp
	})

	return d.Trigger()
}
