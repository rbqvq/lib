package sync

import (
	"time"
)

// Debouncer runs f once, in its own goroutine, after a quiet period d with
// no further Trigger calls. It is safe for concurrent use.
//
// Like time.AfterFunc, it doesn't wait for a running f. If Trigger is called
// after f has started, f can run again concurrently with the previous run.
// Guard f with your own mutex if it must not overlap itself.
type Debouncer struct {
	timer *time.Timer
	d     time.Duration
	f     func()
}

// NewDebouncer returns a Debouncer that calls f after duration of quiet.
// It panics if duration <= 0 or f is nil.
// f is not called until the first Trigger.
func NewDebouncer(duration time.Duration, f func()) *Debouncer {
	if duration <= 0 {
		panic("lib/sync: debouncer duration must be greater than 0")
	}
	if f == nil {
		panic("lib/sync: debouncer func must not be nil")
	}

	d := &Debouncer{d: duration}

	// Put the timer into the "already fired, f finished" state without a
	// magic duration. Ordering matters here:
	//  1. d.f = d.init is written before AfterFunc, so invoke sees init.
	//  2. init reads d.timer.C while d.timer is still the placeholder.
	//  3. The unbuffered send below completes only after init has received,
	//     so the writes to d.timer and d.f after it are ordered after init's reads.
	// Do not reorder these assignments.
	d.f = d.init
	c := make(chan time.Time)
	d.timer = &time.Timer{C: c}
	timer := time.AfterFunc(0, d.invoke)

	c <- time.Time{} // blocks until init has really run

	d.timer = timer
	d.f = f
	return d
}

// init is the placeholder callback used only during construction.
func (d *Debouncer) init() {
	<-d.timer.C
}

func (d *Debouncer) invoke() {
	d.f()
}

// Stop cancels the pending call, if any. It reports whether it stopped one,
// and returns false if f has already started or nothing was pending.
// It doesn't wait for a running f.
func (d *Debouncer) Stop() bool {
	return d.timer.Stop()
}

// Trigger (re)starts the quiet period. It reports whether a pending call was
// pushed back (true), or a new round was started (false), including when the
// previous f has already started.
func (d *Debouncer) Trigger() bool {
	return d.timer.Reset(d.d)
}
