package eavesdrop_test

import (
	"testing"
	"time"

	"github.com/dimmerz92/eavesdrop"
)

func TestDebouncer(t *testing.T) {
	delay := 10 * time.Millisecond
	debouncer := eavesdrop.NewDebouncer(delay)

	t.Run("single call", func(t *testing.T) {
		var count int = 0
		f := func() { count++ }

		debouncer.Do(f)

		if count != 0 {
			t.Errorf("got %d, expected 0", count)
		}

		time.Sleep(delay + 2*time.Millisecond)

		if count != 1 {
			t.Errorf("got %d, expected 1", count)
		}
	})

	t.Run("multiple calls", func(t *testing.T) {
		var count int = 0
		f := func() { count++ }

		go func() {
			for range 10 {
				debouncer.Do(f)
				time.Sleep(5 * time.Millisecond)
			}
		}()

		if count != 0 {
			t.Errorf("got %d, expected 0", count)
		}

		time.Sleep(delay + 50*time.Millisecond)

		if count != 1 {
			t.Errorf("got %d, expected 1", count)
		}
	})
}
