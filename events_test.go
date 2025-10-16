package eavesdrop_test

import (
	"os"
	"path/filepath"
	"reflect"
	"slices"
	"testing"
	"time"

	"github.com/dimmerz92/eavesdrop"
	"github.com/fsnotify/fsnotify"
)

func TestNotifier_HandleNewDir(t *testing.T) {
	tmp := t.TempDir()

	tests := []string{
		filepath.Join(tmp, "node_modules"),
		filepath.Join(tmp, ".git"),
		filepath.Join(tmp, "internal"),
		filepath.Join(tmp, "cmd"),
	}
	expected := []string{tmp, filepath.Join(tmp, "internal"), filepath.Join(tmp, "cmd")}

	cfg := eavesdrop.DefaultConfig()
	cfg.RootDir = tmp

	manager, err := eavesdrop.NewEventManager(cfg)
	if err != nil {
		t.Fatalf("failed to make event manager: %v", err)
	}

	for _, test := range tests {
		if err := os.Mkdir(test, 0777); err != nil {
			t.Fatalf("faile to create folder: %v", err)
		}
	}

	var done bool
	for !done {
		select {
		case event, ok := <-manager.Events:
			if !ok {
				t.Fatalf("event channel closed")
			}

			if event.Has(fsnotify.Create) {
				manager.HandleNewDir(event.Name)
			}

		case err, ok := <-manager.Errors:
			if !ok {
				t.Fatalf("error channel closed")
			}

			if err != nil {
				t.Fatalf("message on error chan: %v", err)
			}

		case <-time.Tick(time.Second):
			done = true
		}
	}

	watched := manager.WatchList()
	slices.Sort(watched)
	slices.Sort(expected)

	if !reflect.DeepEqual(expected, watched) {
		t.Errorf("expected\n%v\ngot\n%v", expected, watched)
	}
}

func TestNotifier_HandleRemovedDir(t *testing.T) {
	tmp := t.TempDir()

	cfg := eavesdrop.DefaultConfig()
	cfg.RootDir = tmp

	manager, err := eavesdrop.NewEventManager(cfg)
	if err != nil {
		t.Fatalf("failed to make event manager: %v", err)
	}

	t.Run("test folder delete", func(t *testing.T) {
		tests := []string{filepath.Join(tmp, "keep"), filepath.Join(tmp, "delete"), filepath.Join(tmp, "delete/delete")}
		expected := []string{tmp, filepath.Join(tmp, "keep")}

		for _, test := range tests {
			if err := os.Mkdir(test, 0777); err != nil {
				t.Fatalf("failed to make directory: %v", err)
			}
			manager.HandleNewDir(test)
		}

		time.Sleep(time.Millisecond)

		if err := os.RemoveAll(tests[1]); err != nil {
			t.Fatalf("failed to remove directory: %v", err)
		}

		time.Sleep(time.Millisecond)

		var done bool
		for !done {
			select {
			case event, ok := <-manager.Events:
				if !ok {
					t.Fatalf("event channel closed")
				}
				if event.Has(fsnotify.Remove) {
					manager.HandleNewDir(event.Name)
				}
			case err, ok := <-manager.Errors:
				if !ok {
					t.Fatalf("event channel closed")
				} else if err != nil {
					t.Fatalf("message on error chan: %v", err)
				}
			case <-time.Tick(time.Second):
				done = true
			}
		}

		watched := manager.WatchList()
		slices.Sort(watched)
		slices.Sort(expected)

		if !reflect.DeepEqual(expected, watched) {
			t.Errorf("expected\n%v\ngot\n%v", expected, watched)
		}
	})
}
