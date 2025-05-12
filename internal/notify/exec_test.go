package notify_test

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/dimmerz92/eavesdrop/internal/notify"
)

func TestExec_Build(t *testing.T) {
	e := notify.Exec{}
	out, err := e.Build("echo 'done'")
	if err != nil {
		t.Errorf("failed with error: %s", out)
	} else if out != "done" {
		t.Errorf("expected hello, got %s", out)
	}
}

func TestExec_Run(t *testing.T) {
	tmp := t.TempDir()

	t.Run("successful", func(t *testing.T) {
		path := filepath.Join(tmp, "test1")

		e := notify.Exec{}
		if err := e.Run("sleep 2 && echo 'done' >> " + path); err != nil {
			t.Fatalf("run failed with error: %v", err)
		}

		if err := e.Cmd.Wait(); err != nil {
			t.Fatalf("wait failed with error: %v", err)
		}

		out, err := os.ReadFile(path)
		if err != nil {
			t.Fatalf("read file failed with error: %v", err)
		}

		if string(out) != "done\n" {
			t.Errorf("expected 'done' got '%s'", string(out))
		}
	})

	t.Run("error", func(t *testing.T) {
		e := notify.Exec{}
		if err := e.Run("sleep 2 && exit 1"); err != nil {
			t.Fatalf("failed with error: %v", err)
		}

		if err := e.Cmd.Wait(); err == nil {
			t.Error("expected error, got none")
		}
	})
}

func TestExec_Kill(t *testing.T) {
	tmp := t.TempDir()

	t.Run("successful", func(t *testing.T) {
		path := filepath.Join(tmp, "test1")

		e := notify.Exec{}
		if err := e.Run("sleep 2 && echo 'done' >> " + path); err != nil {
			t.Fatalf("failed with error: %v", err)
		}

		if err := e.Kill(); err != nil {
			t.Fatalf("kill failed with error: %v", err)
		}

		if _, err := os.Stat(path); err == nil {
			t.Error("expected error, got none")
		}
	})

	t.Run("successful with timeout", func(t *testing.T) {
		path := filepath.Join(tmp, "test2")

		e := notify.Exec{}
		if err := e.Run("sleep 20 && echo 'done' >> " + path); err != nil {
			t.Fatalf("failed with error: %v", err)
		}

		if err := e.Kill(); err != nil {
			t.Errorf("kill failed with error: %v", err)
		}

		if _, err := os.Stat(path); err == nil {
			t.Error("expected error, got none")
		}
	})

	t.Run("successful with timeout and output", func(t *testing.T) {
		path := filepath.Join(tmp, "test3")

		e := notify.Exec{}
		if err := e.Run("sleep 2 && echo 'done' >> " + path); err != nil {
			t.Fatalf("failed with error: %v", err)
		}

		time.Sleep(3 * time.Second)

		if err := e.Kill(); err != nil {
			t.Fatalf("kill failed with error: %v", err)
		}

		out, err := os.ReadFile(path)
		if err != nil {
			t.Fatalf("read file failed with error: %v", err)
		}

		if string(out) != "done\n" {
			t.Errorf("expected 'done' got '%s'", string(out))
		}
	})
}
