package eavesdrop_test

import (
	"bytes"
	"os"
	"runtime"
	"testing"
	"time"

	"github.com/dimmerz92/eavesdrop"
)

func TestShell(t *testing.T) {
	newShell := func() (eavesdrop.Shell, *os.File, *os.File, func()) {
		stdout := os.Stdout
		r, w, _ := os.Pipe()
		os.Stdout = w

		shell := eavesdrop.NewShell(t.Context(),
			eavesdrop.WithTaskTimeout(50*time.Millisecond),
			eavesdrop.WithServiceTimeout(50*time.Millisecond),
		)

		return shell, r, w, func() { os.Stdout = stdout }
	}

	t.Run("test ExecAndReturn", func(t *testing.T) {
		tests := []struct {
			name     string
			task     string
			expected string
			err      bool
		}{
			{name: "empty task", task: " ", err: true},
			{name: "echo 'hello' to stdout", task: "echo -n 'hello'", expected: "hello"},
			{name: "echo 'hello' to stderr", task: "echo -n 'hello' >&2", expected: "hello"},
			{name: "task timeout", task: "sleep 100; echo 'hello'", err: true},
		}

		for _, test := range tests {
			t.Run(test.name, func(t *testing.T) {
				shell, r, w, restore := newShell()
				defer restore()

				err := shell.ExecAndWait(test.task)
				w.Close()
				if test.err && err == nil {
					t.Error("expected error")
				} else if !test.err && err != nil {
					t.Errorf("unexpected error: %v", err)
				}

				if !test.err {
					var buf bytes.Buffer
					buf.ReadFrom(r)

					if stdout := buf.String(); stdout != test.expected {
						t.Errorf("expected %s, got %s", test.expected, stdout)
					}
				}
			})
		}
	})

	t.Run("ExecAndReturn terminations", func(t *testing.T) {
		t.Run("TerminateProcessGroup", func(t *testing.T) {
			service := `trap "exit 0" TERM; while true; do sleep 100; done`
			if runtime.GOOS == "windows" {
				service = ""
			}

			shell := eavesdrop.NewShell(t.Context(),
				eavesdrop.WithTaskTimeout(50*time.Second),
				eavesdrop.WithServiceTimeout(50*time.Millisecond),
			)

			err := shell.ExecAndReturn(service)
			if err != nil {
				t.Fatalf("failed to run service: %v", err)
			}

			time.Sleep(10 * time.Millisecond)

			err = shell.TerminateProcessGroup()
			if err != nil {
				t.Fatalf("failed graceful shutdown: %v", err)
			}
		})

		t.Run("KillProcessGroup", func(t *testing.T) {
			service := `trap "" TERM; while true; do sleep 100; done`
			if runtime.GOOS == "windows" {
				service = ""
			}

			shell := eavesdrop.NewShell(t.Context(),
				eavesdrop.WithTaskTimeout(50*time.Second),
				eavesdrop.WithServiceTimeout(50*time.Millisecond),
			)

			err := shell.ExecAndReturn(service)
			if err != nil {
				t.Fatalf("failed to run service: %v", err)
			}

			time.Sleep(10 * time.Millisecond)

			err = shell.KillProcessGroup()
			if err != nil {
				t.Fatalf("failed hard shutdown: %v", err)
			}
		})

		t.Run("StopService", func(t *testing.T) {
			service := `trap "" TERM; while true; do sleep 1; done`
			if runtime.GOOS == "windows" {
				t.Skip("I have no idea how to test this on windows, lmao")
			}

			shell := eavesdrop.NewShell(t.Context(),
				eavesdrop.WithTaskTimeout(50*time.Millisecond),
				eavesdrop.WithServiceTimeout(50*time.Millisecond),
			)

			err := shell.ExecAndReturn(service)
			if err != nil {
				t.Fatalf("failed to run service: %v", err)
			}

			time.Sleep(10 * time.Millisecond)

			start := time.Now()

			err = shell.Stop()
			if err != nil {
				t.Fatalf("failed to stop service: %v", err)
			}

			duration := time.Since(start)

			if duration <= 50*time.Millisecond {
				t.Errorf("expected termination to wait ~50ms, happened too fast: %v", duration)
			}
		})
	})
}
