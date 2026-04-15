package eavesdrop

import (
	"sync"
	"time"
)

const (
	DefaultDelay = 300 * time.Millisecond
)

type Debouncer struct {
	delay time.Duration
	mu    sync.Mutex
	timer *time.Timer
}

func NewDebouncer(delay time.Duration) *Debouncer {
	debouncer := &Debouncer{delay: DefaultDelay}

	if delay > 0 {
		debouncer.delay = delay
	}

	return debouncer
}

func (d *Debouncer) Do(f func()) {
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
