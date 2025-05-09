package notify_test

import (
	"os"
	"path/filepath"
	"reflect"
	"slices"
	"testing"
	"time"

	"github.com/dimmerz92/eavesdrop/internal/config"
	"github.com/dimmerz92/eavesdrop/internal/notify"
	"github.com/fsnotify/fsnotify"
)

func TestHandleNewDirectory(t *testing.T) {
	tmp := t.TempDir()
	tests := []string{filepath.Join(tmp, ".git"), filepath.Join(tmp, "internal"), filepath.Join(tmp, "cmd")}
	expected := []string{tmp, filepath.Join(tmp, "internal"), filepath.Join(tmp, "cmd")}
	watcher := notify.NewWatcher(config.DefaultConfig(""))
	watcher.Config.Root = tmp
	watcher.Add(tmp)

	for _, test := range tests {
		if err := os.Mkdir(test, 0777); err != nil {
			t.Fatalf("faile to create folder: %v", err)
		}
	}

	var done bool
	for !done {
		select {
		case event, ok := <-watcher.Events:
			if !ok {
				t.Fatalf("event channel closed")
			}
			if event.Has(fsnotify.Create) {
				notify.HandleNewDirectory(event.Name, watcher)
			}
		case err, ok := <-watcher.Errors:
			if !ok {
				t.Fatalf("error channel closed")
			} else if err != nil {
				t.Fatalf("message on error chan: %v", err)
			}
		case <-time.Tick(time.Second):
			done = true
		}
	}

	slices.Sort(expected)
	watched := watcher.WatchList()
	slices.Sort(watched)
	if !reflect.DeepEqual(expected, watched) {
		t.Errorf("expected %v\ngot %v", expected, watched)
	}
}

func TestHandleRemoveDirectory(t *testing.T) {
	tmp := t.TempDir()
	watcher := notify.NewWatcher(config.DefaultConfig(""))
	watcher.Config.Root = tmp
	watcher.Add(tmp)

	t.Run("test folder delete", func(t *testing.T) {
		tests := []string{filepath.Join(tmp, "keep"), filepath.Join(tmp, "delete"), filepath.Join(tmp, "delete/delete")}
		expected := []string{tmp, filepath.Join(tmp, "keep")}
		for _, test := range tests {
			if err := os.Mkdir(test, 0777); err != nil {
				t.Fatalf("failed to make directory: %v", err)
			}
			notify.HandleNewDirectory(test, watcher)
		}

		time.Sleep(time.Millisecond)

		if err := os.RemoveAll(tests[1]); err != nil {
			t.Fatalf("failed to remove directory: %v", err)
		}

		time.Sleep(time.Millisecond)

		var done bool
		for !done {
			select {
			case event, ok := <-watcher.Events:
				if !ok {
					t.Fatalf("event channel closed")
				}
				if event.Has(fsnotify.Remove) {
					notify.HandleNewDirectory(event.Name, watcher)
				}
			case err, ok := <-watcher.Errors:
				if !ok {
					t.Fatalf("event channel closed")
				} else if err != nil {
					t.Fatalf("message on error chan: %v", err)
				}
			case <-time.Tick(time.Second):
				done = true
			}
		}

		slices.Sort(expected)
		watched := watcher.WatchList()
		slices.Sort(watched)
		if !reflect.DeepEqual(expected, watched) {
			t.Errorf("expected %v\ngot %v", expected, watched)
		}
	})
}
