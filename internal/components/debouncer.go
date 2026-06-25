package components

import (
	"sync"
	"time"
)

type Debouncer struct {
	delay time.Duration
	timer *time.Timer
	mu    sync.Mutex
}

func NewDebouncer(delayMs uint) *Debouncer {
	return &Debouncer{delay: time.Duration(delayMs) * time.Millisecond}
}

func (d *Debouncer) Do(callback func()) {
	d.mu.Lock()
	defer d.mu.Unlock()

	if d.timer != nil {
		d.timer.Stop()
	}

	d.timer = time.AfterFunc(d.delay, callback)
}

func (d *Debouncer) UpdateDelay(delayMs uint) {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.delay = time.Duration(delayMs) * time.Millisecond
}
