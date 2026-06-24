package components_test

import (
	"sync/atomic"
	"testing"
	"time"

	"github.com/dimmerz92/eavesdrop/internal/components"
)

func TestNewDebouncer(t *testing.T) {
	tests := []struct {
		name        string
		delay       time.Duration
		expectedErr bool
	}{
		{"positive delay", time.Millisecond, false},
		{"zero delay", 0, true},
		{"negative delay", -time.Millisecond, true},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			d, err := components.NewDebouncer(test.delay)
			if (err != nil) != test.expectedErr {
				t.Errorf("NewDebouncer(%v) error = %v, expectedErr %v", test.delay, err, test.expectedErr)
			}
			if !test.expectedErr && d == nil {
				t.Error("NewDebouncer() returned nil Debouncer without error")
			}
		})
	}
}

func TestDebouncer_Do(t *testing.T) {
	const delay = 50 * time.Millisecond

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
			waitAfter:     delay * 3,
			expectedFires: 1,
		},
		{
			name:          "rapid calls fire only once",
			calls:         5,
			callInterval:  5 * time.Millisecond,
			waitAfter:     delay * 3,
			expectedFires: 1,
		},
		{
			name:          "callback not fired before delay elapses",
			calls:         1,
			callInterval:  0,
			waitAfter:     delay / 5,
			expectedFires: 0,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			d, err := components.NewDebouncer(delay)
			if err != nil {
				t.Fatalf("NewDebouncer() unexpected error: %v", err)
			}

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
