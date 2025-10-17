package eavesdrop_test

import (
	"sync"
	"testing"
	"time"

	"github.com/dimmerz92/eavesdrop"
)

func TestDebouncer_Run_SingleCall(t *testing.T) {
	var called bool
	d := &eavesdrop.Debouncer{Delay: 10 * time.Millisecond}

	done := make(chan struct{})

	d.Run(func() {
		called = true
		close(done)
	})

	select {
	case <-done:
		if !called {
			t.Errorf("Function was not called after delay")
		}

	case <-time.After(50 * time.Millisecond):
		t.Errorf("Function was not called within expected time")
	}
}

func TestDebouncer_Run_ResetTimer(t *testing.T) {
	d := &eavesdrop.Debouncer{Delay: 30 * time.Millisecond}

	var mu sync.Mutex
	count := 0
	done := make(chan struct{})

	fn := func() {
		mu.Lock()
		count++
		mu.Unlock()
		close(done)
	}

	d.Run(fn)

	time.Sleep(10 * time.Millisecond)

	d.Run(fn)

	select {
	case <-done:
		mu.Lock()
		defer mu.Unlock()
		if count != 1 {
			t.Errorf("Function should have been called once, but was called %d times", count)
		}

	case <-time.After(100 * time.Millisecond):
		t.Errorf("Function was not called within expected time")
	}
}

func TestDebouncer_Run_MultipleRapidCalls(t *testing.T) {
	d := &eavesdrop.Debouncer{Delay: 30 * time.Millisecond}

	var mu sync.Mutex
	count := 0
	done := make(chan struct{})

	fn := func() {
		mu.Lock()
		count++
		mu.Unlock()
		close(done)
	}

	for range 5 {
		d.Run(fn)
		time.Sleep(5 * time.Millisecond)
	}

	select {
	case <-done:
		mu.Lock()
		defer mu.Unlock()
		if count != 1 {
			t.Errorf("Expected function to run only once, ran %d times", count)
		}

	case <-time.After(100 * time.Millisecond):
		t.Errorf("Function was not called within expected time")
	}
}
