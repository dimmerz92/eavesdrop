package notify_test

import (
	"testing"
	"time"

	"github.com/dimmerz92/eavesdrop/internal/notify"
)

func TestDebounce_Run(t *testing.T) {
	debouncer := notify.Debouncer{}

	var expected int
	for i := 0; i < 10; i++ {
		debouncer.Run(10*time.Millisecond, func() { expected = i + 1 })
	}

	time.Sleep(11 * time.Millisecond)

	if expected != 10 {
		t.Errorf("expected 10, got %d", expected)
	}

	debouncer.Run(10*time.Millisecond, func() { expected = 100 })

	time.Sleep(11 * time.Millisecond)

	if expected != 100 {
		t.Errorf("expected 100, got %d", expected)
	}
}
