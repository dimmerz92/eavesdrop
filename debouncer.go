package eavesdrop

import "time"

type Debouncer struct {
	timer *time.Timer
	used  bool
}

// Run executes the function after the delay has passed.
// Repeat calls to Run will reset the timer.
// Run does not check if the given function is always the same.
// args:
// - delay is used to define the time before the function is run.
// - f is the function to be run after the delay time elapses.
func (d *Debouncer) Run(delay time.Duration, f func()) {
	if d.used {
		d.timer.Stop()
	} else {
		d.used = true
	}

	d.timer = time.AfterFunc(delay, func() {
		d.used = false
		f()
	})
}
