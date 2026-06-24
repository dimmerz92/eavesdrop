package components

import (
	"fmt"
	"sync"
	"time"
)

type Debouncer struct {
	delay time.Duration
	timer *time.Timer
	mu    sync.Mutex
}

func NewDebouncer(delay time.Duration) (*Debouncer, error) {
	if delay <= 0 {
		return nil, fmt.Errorf("debounce delay must be a positive non-zero duration")
	}
	return &Debouncer{delay: delay}, nil
}

func (d *Debouncer) Do(callback func()) {
	d.mu.Lock()
	defer d.mu.Unlock()

	if d.timer != nil {
		d.timer.Stop()
	}

	d.timer = time.AfterFunc(d.delay, callback)
}
