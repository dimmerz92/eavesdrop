package eavesdrop

import (
	"sync"
	"time"
)

const (
	DefaultDelay = 300 * time.Millisecond
)

type Debouncer interface {
	Do(f func())
}

type debouncer struct {
	delay time.Duration
	mu    sync.Mutex
	timer *time.Timer
}

// NewDebouncer returns an instance of the default Debouncer implementation.
func NewDebouncer(delay time.Duration) *debouncer {
	debouncer := &debouncer{delay: DefaultDelay}

	if delay > 0 {
		debouncer.delay = delay
	}

	return debouncer
}

// Do runs the given function after the configured delay.
// Calls to Do while the debounce delay is active resets the timer and the replaces the f with the newest function.
func (d *debouncer) Do(f func()) {
	d.mu.Lock()
	defer d.mu.Unlock()

	if d.timer == nil {
		d.timer = time.AfterFunc(d.delay, func() {
			d.mu.Lock()
			defer d.mu.Unlock()

			f()
			d.timer = nil
		})
		return
	}

	d.timer.Reset(d.delay)
}
