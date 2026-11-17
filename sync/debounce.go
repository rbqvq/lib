package sync

import (
	"sync"
	"time"
)

// Debounce coalesces bursts of Trigger calls. Each Trigger blocks until
// either the quiet period d elapses with no newer Trigger or Stop, or it is
// superseded by one.
//
// It is safe for concurrent use. Its correctness relies on the following
// invariant: every Trigger that starts the timer must reach its select
// statement, because cancel is sent while holding mutex and depends on the
// pending waiter arriving shortly. Re-verify this before adding early
// returns, context cancellation or panic paths to Trigger.
//
// It behaves the same on Go 1.23+ (synchronous timer channels) and on earlier
// versions or asynctimerchan=1 (buffered timer channels).
type Debounce struct {
	mutex  sync.Mutex
	timer  *time.Timer
	cancel chan struct{}
	d      time.Duration
}

// NewDebounce returns a Debounce with quiet period duration.
// It panics if duration <= 0.
func NewDebounce(duration time.Duration) *Debounce {
	if duration <= 0 {
		panic("lib/sync: debounce duration must be greater than 0")
	}

	d := &Debounce{
		timer:  time.NewTimer(0),
		cancel: make(chan struct{}),
		d:      duration,
	}
	// The timer starts already fired; stop it and drain the value so it
	// begins in a clean, inactive state. There are no waiters yet, so the
	// result is ignored.
	d.stopTimer()
	return d
}

// stopTimer stops the timer and drains any undelivered value from timer.C.
// It reports whether the current round still had a pending waiter, meaning
// the timer was active, or it had fired but nobody had received the value yet.
// In that case the caller must notify the waiter through d.cancel.
//
// Draining under the mutex prevents a stale value from being consumed after
// the next Reset on pre-1.23 runtimes. On Go 1.23+ Stop already discards
// undelivered values, so the drain just hits default.
//
// The caller must hold d.mutex (except in NewDebounce, where there is no
// concurrency yet).
func (d *Debounce) stopTimer() bool {
	if d.timer.Stop() {
		return true
	}

	select {
	case <-d.timer.C:
		return true
	default:
		return false
	}
}

// Stop cancels the current round, if any. The pending Trigger, if there is
// one, returns false.
//
// It reports whether a pending Trigger was cancelled. It returns false if
// nothing was pending, including when the timer has already fired and its
// Trigger is returning true.
func (d *Debounce) Stop() bool {
	d.mutex.Lock()
	defer d.mutex.Unlock()

	if d.stopTimer() {
		d.cancel <- struct{}{}
		return true
	}

	return false
}

// Trigger (re)starts the quiet period and blocks until it resolves.
//
// It returns true if the full period d elapsed without a newer Trigger or
// Stop, which makes this call the winner of its burst. It returns false if a
// later Trigger or Stop superseded it.
//
// A previous pending Trigger, if any, is notified and returns false before
// this call starts its own timer.
func (d *Debounce) Trigger() bool {
	d.mutex.Lock()
	// Supersede the previous round. Holding the mutex until Reset guarantees
	// that at most one waiter is un-notified at any time.
	if d.stopTimer() {
		d.cancel <- struct{}{}
	}

	d.timer.Reset(d.d)
	d.mutex.Unlock()

	select {
	case <-d.timer.C:
		return true
	case <-d.cancel:
		return false
	}
}
