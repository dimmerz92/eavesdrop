package components_test

import (
	"sync/atomic"
	"testing"
	"time"

	"github.com/dimmerz92/eavesdrop/v2/internal/components"
)

func TestNewDebouncer(t *testing.T) {
	tests := []struct {
		name        string
		delay       uint
		expectedErr bool
	}{
		{"positive delay", 1, false},
		{"zero delay", 0, true},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			d := components.NewDebouncer(test.delay)
			if !test.expectedErr && d == nil {
				t.Error("NewDebouncer() returned nil Debouncer without error")
			}
		})
	}
}

func TestDebouncer_UpdateDelay(t *testing.T) {
	tests := []struct {
		name          string
		initialDelay  uint
		updatedDelay  uint
		waitAfter     time.Duration
		expectedFires int32
	}{
		{
			name:          "longer delay defers callback",
			initialDelay:  50,
			updatedDelay:  200,
			waitAfter:     100 * time.Millisecond,
			expectedFires: 0,
		},
		{
			name:          "longer delay fires after new delay elapses",
			initialDelay:  50,
			updatedDelay:  200,
			waitAfter:     400 * time.Millisecond,
			expectedFires: 1,
		},
		{
			name:          "shorter delay fires before original delay elapses",
			initialDelay:  200,
			updatedDelay:  50,
			waitAfter:     100 * time.Millisecond,
			expectedFires: 1,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			d := components.NewDebouncer(test.initialDelay)
			d.UpdateDelay(test.updatedDelay)

			var count atomic.Int32
			d.Do(func() { count.Add(1) })

			time.Sleep(test.waitAfter)

			if got := count.Load(); got != test.expectedFires {
				t.Errorf("callback fired %d time(s), expected %d", got, test.expectedFires)
			}
		})
	}
}

func TestDebouncer_Do(t *testing.T) {
	const delay = 50

	tests := []struct {
		name          string
		calls         int
		callInterval  time.Duration
		waitAfter     time.Duration
		expectedFires int32
	}{
		{
			name:          "single call fires after delay",
			calls:         1,
			callInterval:  0,
			waitAfter:     delay * 3 * time.Millisecond,
			expectedFires: 1,
		},
		{
			name:          "rapid calls fire only once",
			calls:         5,
			callInterval:  5 * time.Millisecond,
			waitAfter:     delay * 3 * time.Millisecond,
			expectedFires: 1,
		},
		{
			name:          "callback not fired before delay elapses",
			calls:         1,
			callInterval:  0,
			waitAfter:     delay / 5 * time.Millisecond,
			expectedFires: 0,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			d := components.NewDebouncer(delay)

			var count atomic.Int32
			for range test.calls {
				d.Do(func() { count.Add(1) })
				if test.callInterval > 0 {
					time.Sleep(test.callInterval)
				}
			}

			time.Sleep(test.waitAfter)

			if got := count.Load(); got != test.expectedFires {
				t.Errorf("callback fired %d time(s), expected %d", got, test.expectedFires)
			}
		})
	}
}
