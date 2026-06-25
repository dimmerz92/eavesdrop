package ev_test

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	ev "github.com/dimmerz92/eavesdrop"
)

const eventTimeout = 2 * time.Second

func testWatcher(t *testing.T, dir string) (*ev.Watcher, <-chan ev.Event) {
	t.Helper()
	ch := make(chan ev.Event, 16)
	w := ev.NewWatcher(t.Context(), t.Name(), dir).
		WithFiletypes(".go").
		WithDebounceDelay(debounceDelay).
		WithExcluder(ev.NewExcluder(dir)).
		WithOnChange(func(e ev.Event) {
			select {
			case ch <- e:
			default:
			}
		})
	return w, ch
}

func awaitEvent(t *testing.T, ch <-chan ev.Event) ev.Event {
	t.Helper()
	select {
	case e := <-ch:
		return e
	case <-time.After(eventTimeout):
		t.Error("timed out waiting for onChange to fire")
		return ev.Event{}
	}
}

func TestNewEmitter(t *testing.T) {
	t.Run("PanicsOnEmptyRoot", func(t *testing.T) {
		defer func() {
			if r := recover(); r == nil {
				t.Error("expected panic on empty root")
			}
		}()
		ev.NewEmitter("")
	})

	t.Run("ReturnsNonNil", func(t *testing.T) {
		if e := ev.NewEmitter(t.TempDir()); e == nil {
			t.Error("NewEmitter() returned nil")
		}
	})
}

func TestEventEmitter_WithExcluder_Chainable(t *testing.T) {
	e := ev.NewEmitter(t.TempDir())
	if got := e.WithExcluder(ev.NewExcluder(".")); got != e {
		t.Error("WithExcluder() did not return same *EventEmitter")
	}
}

func TestEventEmitter_RecursiveWatch(t *testing.T) {
	dir := t.TempDir()
	if err := os.Mkdir(filepath.Join(dir, "sub"), 0o755); err != nil {
		t.Fatal(err)
	}

	e := ev.NewEmitter(dir)
	if err := e.RecursiveWatch(dir); err != nil {
		t.Errorf("RecursiveWatch() = %v, expected nil", err)
	}
}

func TestEventEmitter_RecursiveUnwatch(t *testing.T) {
	dir := t.TempDir()
	e := ev.NewEmitter(dir)
	if err := e.RecursiveWatch(dir); err != nil {
		t.Fatal(err)
	}
	if err := e.RecursiveUnwatch(dir); err != nil {
		t.Errorf("RecursiveUnwatch() = %v, expected nil", err)
	}
}

func TestEventEmitter_Start(t *testing.T) {
	tests := []struct {
		name    string
		prepare func(dir string) error
		trigger func(dir string) error
	}{
		{
			name: "file creation fires onChange",
			trigger: func(dir string) error {
				return os.WriteFile(filepath.Join(dir, "new.go"), []byte("hello"), 0o644)
			},
		},
		{
			name: "file write fires onChange",
			prepare: func(dir string) error {
				return os.WriteFile(filepath.Join(dir, "existing.go"), []byte("initial"), 0o644)
			},
			trigger: func(dir string) error {
				return os.WriteFile(filepath.Join(dir, "existing.go"), []byte("updated"), 0o644)
			},
		},
		{
			name: "file removal fires onChange",
			prepare: func(dir string) error {
				return os.WriteFile(filepath.Join(dir, "to_remove.go"), []byte("bye"), 0o644)
			},
			trigger: func(dir string) error {
				return os.Remove(filepath.Join(dir, "to_remove.go"))
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			dir := t.TempDir()

			if test.prepare != nil {
				if err := test.prepare(dir); err != nil {
					t.Fatal(err)
				}
			}

			w, ch := testWatcher(t, dir)
			e := ev.NewEmitter(dir)
			e.Subscribe(w)
			e.Start(t.Context())

			if err := test.trigger(dir); err != nil {
				t.Fatal(err)
			}

			awaitEvent(t, ch)
		})
	}
}

func TestEventEmitter_Start_ExcluderFiltersEvents(t *testing.T) {
	dir := t.TempDir()
	excludedFile := filepath.Join(dir, "ignored.go")
	watchedFile := filepath.Join(dir, "watched.go")

	w, ch := testWatcher(t, dir)
	e := ev.NewEmitter(dir).WithExcluder(ev.NewExcluder(dir).WithFiles(excludedFile))
	e.Subscribe(w)
	e.Start(t.Context())

	if err := os.WriteFile(excludedFile, []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}

	// Write a non-excluded file as a positive control, then verify
	// no event from the excluded file slipped through.
	if err := os.WriteFile(watchedFile, []byte("y"), 0o644); err != nil {
		t.Fatal(err)
	}
	awaitEvent(t, ch)

	time.Sleep(100 * time.Millisecond)
	for {
		select {
		case e := <-ch:
			if e.Path() == excludedFile {
				t.Errorf("received event for excluded file: op=%s", e.Op())
			}
		default:
			return
		}
	}
}
