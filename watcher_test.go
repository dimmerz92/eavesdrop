package ev_test

import (
	"sync/atomic"
	"testing"
	"time"

	ev "github.com/dimmerz92/eavesdrop"
)

const (
	testRoot      = "/tmp/ev-watcher-test"
	debounceDelay = 10
	debounceWait  = 50 * time.Millisecond
)

func fileEvent(path string, op ev.Op) ev.Event {
	return ev.NewEvent(op, testRoot+"/"+path, mockFileInfo{name: path})
}

func TestNewWatcher(t *testing.T) {
	t.Run("PanicsOnEmptyName", func(t *testing.T) {
		defer func() {
			if r := recover(); r == nil {
				t.Error("expected panic on empty name")
			}
		}()
		ev.NewWatcher("", ".")
	})

	t.Run("PanicsOnWhitespaceName", func(t *testing.T) {
		defer func() {
			if r := recover(); r == nil {
				t.Error("expected panic on whitespace-only name")
			}
		}()
		ev.NewWatcher("   ", ".")
	})

	t.Run("PanicsOnDuplicateName", func(t *testing.T) {
		ev.NewWatcher(t.Name(), ".")
		defer func() {
			if r := recover(); r == nil {
				t.Error("expected panic on duplicate name")
			}
		}()
		ev.NewWatcher(t.Name(), ".")
	})

	t.Run("ReturnsNonNil", func(t *testing.T) {
		w := ev.NewWatcher(t.Name(), ".")
		if w == nil {
			t.Error("NewWatcher() returned nil")
		}
	})
}

func TestWatcher_Watched(t *testing.T) {
	tests := []struct {
		name     string
		setup    func(*ev.Watcher) *ev.Watcher
		event    ev.Event
		expected bool
	}{
		{
			name:     "nil info always watched",
			setup:    func(w *ev.Watcher) *ev.Watcher { return w },
			event:    ev.NewEvent(ev.WRITE, testRoot+"/main.go", nil),
			expected: true,
		},
		{
			name:     "path outside root",
			setup:    func(w *ev.Watcher) *ev.Watcher { return w.WithFiletypes(".go") },
			event:    ev.NewEvent(ev.WRITE, "/other/path/main.go", mockFileInfo{name: "main.go"}),
			expected: false,
		},
		{
			name:     "filetype match",
			setup:    func(w *ev.Watcher) *ev.Watcher { return w.WithFiletypes(".go") },
			event:    fileEvent("main.go", ev.WRITE),
			expected: true,
		},
		{
			name:     "filetype mismatch",
			setup:    func(w *ev.Watcher) *ev.Watcher { return w.WithFiletypes(".go") },
			event:    fileEvent("styles.css", ev.WRITE),
			expected: false,
		},
		{
			name:     "file match",
			setup:    func(w *ev.Watcher) *ev.Watcher { return w.WithFiles("config.toml") },
			event:    fileEvent("config.toml", ev.WRITE),
			expected: true,
		},
		{
			name:     "file not in watch list",
			setup:    func(w *ev.Watcher) *ev.Watcher { return w.WithFiles("config.toml") },
			event:    fileEvent("other.toml", ev.WRITE),
			expected: false,
		},
		{
			name:     "dir match",
			setup:    func(w *ev.Watcher) *ev.Watcher { return w.WithDirs("src") },
			event:    fileEvent("src/main.go", ev.WRITE),
			expected: true,
		},
		{
			name:     "file not inside watched dir",
			setup:    func(w *ev.Watcher) *ev.Watcher { return w.WithDirs("src") },
			event:    fileEvent("other/main.go", ev.WRITE),
			expected: false,
		},
		{
			name:     "no filetype file or dir configured",
			setup:    func(w *ev.Watcher) *ev.Watcher { return w },
			event:    fileEvent("main.go", ev.WRITE),
			expected: false,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			w := ev.NewWatcher(t.Name(), testRoot)
			w = test.setup(w)
			if got := w.Watched(test.event); got != test.expected {
				t.Errorf("Watched() = %v, expected %v", got, test.expected)
			}
		})
	}
}

func TestWatcher_Handle(t *testing.T) {
	tests := []struct {
		name          string
		setup         func(*ev.Watcher) *ev.Watcher
		event         ev.Event
		calls         int
		expectedFires int32
	}{
		{
			name: "watched event fires onChange",
			setup: func(w *ev.Watcher) *ev.Watcher {
				return w.
					WithFiletypes(".go").
					WithDebounceDelay(debounceDelay).
					WithExcluder(ev.NewExcluder(testRoot))
			},
			event:         fileEvent("main.go", ev.WRITE),
			calls:         1,
			expectedFires: 1,
		},
		{
			name: "unwatched event does not fire onChange",
			setup: func(w *ev.Watcher) *ev.Watcher {
				return w.
					WithFiletypes(".go").
					WithDebounceDelay(debounceDelay).
					WithExcluder(ev.NewExcluder(testRoot))
			},
			event:         fileEvent("styles.css", ev.WRITE),
			calls:         1,
			expectedFires: 0,
		},
		{
			name: "excluded event does not fire onChange",
			setup: func(w *ev.Watcher) *ev.Watcher {
				return w.
					WithFiletypes(".go").
					WithDebounceDelay(debounceDelay).
					WithExcluder(ev.NewExcluder(testRoot).WithFiles(testRoot + "/main.go"))
			},
			event:         fileEvent("main.go", ev.WRITE),
			calls:         1,
			expectedFires: 0,
		},
		{
			name: "rapid calls debounce to single fire",
			setup: func(w *ev.Watcher) *ev.Watcher {
				return w.
					WithFiletypes(".go").
					WithDebounceDelay(debounceDelay).
					WithExcluder(ev.NewExcluder(testRoot))
			},
			event:         fileEvent("main.go", ev.WRITE),
			calls:         5,
			expectedFires: 1,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			var count atomic.Int32
			w := ev.NewWatcher(t.Name(), testRoot)
			w = test.setup(w)
			w.WithOnChange(func(_ ev.Event) { count.Add(1) })

			for range test.calls {
				w.Handle(test.event)
			}

			time.Sleep(debounceWait)

			if got := count.Load(); got != test.expectedFires {
				t.Errorf("onChange fired %d time(s), expected %d", got, test.expectedFires)
			}
		})
	}
}

func TestWatcher_Trigger(t *testing.T) {
	var called atomic.Bool
	w := ev.NewWatcher(t.Name(), ".")
	w.WithOnChange(func(_ ev.Event) { called.Store(true) })

	w.Trigger()

	if !called.Load() {
		t.Error("Trigger() did not call onChange")
	}
}

func TestWatcher_Builders_Chainable(t *testing.T) {
	w := ev.NewWatcher(t.Name(), ".")
	tests := []struct {
		name string
		got  *ev.Watcher
	}{
		{"WithFiletypes", w.WithFiletypes(".go")},
		{"WithDirs", w.WithDirs("src")},
		{"WithFiles", w.WithFiles("main.go")},
		{"WithOnChange", w.WithOnChange(func(_ ev.Event) {})},
		{"WithDebounceDelay", w.WithDebounceDelay(50)},
		{"WithExcluder", w.WithExcluder(ev.NewExcluder("."))},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if test.got != w {
				t.Errorf("%s() did not return same *Watcher", test.name)
			}
		})
	}
}
