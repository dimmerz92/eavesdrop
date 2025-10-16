package eavesdrop_test

import (
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/dimmerz92/eavesdrop"
)

var OS = runtime.GOOS

func TestShell_Exec_Success(t *testing.T) {
	shell := eavesdrop.NewShell(1*time.Second, 1*time.Second)

	out, err := shell.Exec(`echo "hello world"`)
	if err != nil {
		t.Fatalf("Exec failed: %v", err)
	}

	if OS == "windows" {
		if strings.TrimSpace(out) != "\\\"hello world\\\"" {
			t.Errorf("Expected \\\"hello world\\\", got: %s", out)
		}
	} else if strings.TrimSpace(out) != "hello world" {
		t.Errorf("Expected 'hello world', got: %s", out)
	}
}

func TestShell_Exec_Timeout(t *testing.T) {
	shell := eavesdrop.NewShell(100*time.Millisecond, 1*time.Second)

	command := "sleep 1"
	if OS == "windows" {
		command = "timeout /T 1"
	}

	_, err := shell.Exec(command)
	if err == nil {
		t.Fatal("Expected timeout error, got nil")
	}
}

func TestShell_Run_And_Kill_Graceful(t *testing.T) {
	if OS == "windows" {
		t.Skip("I have no idea how to test this on windows, lmao")
	}

	shell := eavesdrop.NewShell(1*time.Second, 2*time.Second)

	err := shell.Run(`trap "exit 0" SIGTERM; while true; do sleep 1; done`)
	if err != nil {
		t.Fatalf("Run failed: %v", err)
	}

	time.Sleep(300 * time.Millisecond)

	err = shell.Kill()
	if err != nil {
		t.Fatalf("Expected graceful shutdown, got error: %v", err)
	}
}

func TestShell_Run_And_Kill_Force(t *testing.T) {
	if OS == "windows" {
		t.Skip("I have no idea how to test this on windows, lmao")
	}

	shell := eavesdrop.NewShell(1*time.Second, 300*time.Millisecond)

	err := shell.Run(`trap "" SIGTERM; while true; do sleep 1; done`)
	if err != nil {
		t.Fatalf("Run failed: %v", err)
	}

	time.Sleep(200 * time.Millisecond)

	start := time.Now()
	err = shell.Kill()
	if err == nil {
		t.Fatalf("failed to kill shell: %v", err)
	}
	duration := time.Since(start)

	if duration < 300*time.Millisecond {
		t.Errorf("Expected kill to wait ~300ms, happened too fast: %v", duration)
	}
}

func TestShell_Run_And_Kill_Not_Graceful(t *testing.T) {
	if OS == "windows" {
		t.Skip("I have no idea how to test this on windows, lmao")
	}

	shell := eavesdrop.NewShell(1*time.Second, 2*time.Second)

	err := shell.Run("while true; do sleep 1; done")
	if err != nil {
		t.Fatalf("Run failed: %v", err)
	}

	time.Sleep(300 * time.Millisecond)

	err = shell.Kill()
	if err == nil {
		t.Fatal("Expected non-graceful shutdown, got nil")
	}
}
