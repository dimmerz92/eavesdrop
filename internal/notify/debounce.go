package notify

import "time"

type Debouncer struct {
	timer *time.Timer
	used  bool
}

// Run runs a given func after the given delay. If Run is called within the
// delay period, the delay will reset.
// Ensures the given func is run only once after the last call.
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
