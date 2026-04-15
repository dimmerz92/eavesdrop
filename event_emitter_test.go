package eavesdrop_test

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/dimmerz92/eavesdrop"
)

func TestEmitter(t *testing.T) {
	tmp := t.TempDir()

	excluder := eavesdrop.NewExcluder(
		eavesdrop.WithDirs("ignore_me", "ignore_me2"),
		eavesdrop.WithFiles(".env", ".DS_Store"),
		eavesdrop.WithRegex("_test.go"),
	)

	emitter := eavesdrop.NewEmitter(tmp, eavesdrop.WithGlobalExcluder(excluder))

	go func() {
		emitter.Start(t.Context())
	}()

	time.Sleep(50 * time.Millisecond)

	t.Run("create events", func(t *testing.T) {
		tests := []struct {
			name  string
			dir   bool
			path  string
			event bool
		}{
			{name: "dir created and event emitted", dir: true, path: "watch_me", event: true},
			{name: "dir created and event not emitted", dir: true, path: "ignore_me"},
			{name: "file created and event emitted", path: "main.go", event: true},
			{name: "file created and event not emitted", path: ".env"},
			{name: "file created and event not emitted regex", path: "test_test.go"},
		}

		for _, test := range tests {
			t.Run(test.name, func(t *testing.T) {
				event := emitter.Subscribe()

				if test.dir {
					err := os.MkdirAll(filepath.Join(tmp, test.path), 0755)
					if err != nil {
						t.Fatalf("failed to make directory: %v", err)
					}
				} else {
					file, err := os.Create(filepath.Join(tmp, test.path))
					if err != nil {
						t.Fatalf("failed to make file: %v", err)
					}
					defer file.Close()
				}

				ctx, cancel := context.WithTimeout(t.Context(), 50*time.Millisecond)
				defer cancel()

				select {
				case <-ctx.Done():
					if test.event {
						t.Errorf("expected event, got none: %s", test.path)
					}

				case <-event:
					if !test.event {
						t.Errorf("got event, expected nothing: %s", test.path)
					}
				}
			})
		}
	})

	t.Run("rename events", func(t *testing.T) {
		tests := []struct {
			name  string
			from  string
			to    string
			event bool
		}{
			{name: "dir renamed and event emitted", from: "watch_me", to: "watch_me2", event: true},
			{name: "dir renamed and event not emitted", from: "ignore_me", to: "ignore_me2"},
			{name: "file renamed and event emitted", from: "main.go", to: "test.go", event: true},
			{name: "file renamed and event not emitted", from: ".env", to: ".DS_Store"},
		}

		for _, test := range tests {
			t.Run(test.name, func(t *testing.T) {
				event := emitter.Subscribe()

				err := os.Rename(filepath.Join(tmp, test.from), filepath.Join(tmp, test.to))
				if err != nil {
					t.Fatalf("failed to rename path: %v", err)
				}

				ctx, cancel := context.WithTimeout(t.Context(), 50*time.Millisecond)
				defer cancel()

				select {
				case <-ctx.Done():
					if test.event {
						t.Errorf("expected event, got none: %s -> %s", test.from, test.to)
					}

				case <-event:
					if !test.event {
						t.Errorf("got event, expected nothing: %s -> %s", test.from, test.to)
					}
				}
			})
		}
	})

	t.Run("delete events", func(t *testing.T) {
		tests := []struct {
			name  string
			path  string
			event bool
		}{
			{name: "dir deleted and event emitted", path: "watch_me2", event: true},
			{name: "dir deleted and event not emitted", path: "ignore_me2"},
			{name: "file deleted and event emitted", path: "test.go", event: true},
			{name: "file deleted and event not emitted", path: ".DS_Store"},
			{name: "file deleted and event not emitted", path: "test_test.go"},
		}

		for _, test := range tests {
			t.Run(test.name, func(t *testing.T) {
				event := emitter.Subscribe()

				err := os.Remove(filepath.Join(tmp, test.path))
				if err != nil {
					t.Fatalf("failed to delete path: %v", err)
				}

				ctx, cancel := context.WithTimeout(t.Context(), 50*time.Millisecond)
				defer cancel()

				select {
				case <-ctx.Done():
					if test.event {
						t.Errorf("expected event, got none: %s", test.path)
					}

				case <-event:
					if !test.event {
						t.Errorf("got event, expected nothing: %s", test.path)
					}
				}
			})
		}
	})
}
