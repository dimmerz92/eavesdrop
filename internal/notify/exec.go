package notify

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"syscall"
	"time"
)

type Exec struct {
	*exec.Cmd
	ctx    context.Context
	cancel context.CancelFunc
}

// Build runs the given command and waits for the combined output.
// If the command takes longer than 30 seconds to execute, it will be cancelled.
func (e *Exec) Build(command string) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	out, err := exec.CommandContext(ctx, "sh", "-c", command).CombinedOutput()
	return strings.TrimSpace(string(out)), err
}

// Run runs the given long running command in a separate process group without
// waiting for it to finish.
// Kill the process using the Kill() method.
func (e *Exec) Run(command string) error {
	ctx, cancel := context.WithCancel(context.Background())
	cmd := exec.CommandContext(ctx, "sh", "-c", command)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}

	err := cmd.Start()
	if err != nil {
		cancel()
		return fmt.Errorf("failed to execute run command: %w", err)
	}

	e.Cmd = cmd
	e.ctx = ctx
	e.cancel = cancel

	return nil
}

// Kill signals the process with SIGTERM for a graceful shutdown with 5 seconds
// grace. On error, an explicit kill signal is sent.
func (e *Exec) Kill() error {
	if e.Cmd == nil {
		return nil
	}

	defer e.cancel()

	var err error

	done := make(chan error)
	defer close(done)
	go func() { done <- e.Cmd.Wait() }()
	e.Cmd.Process.Signal(syscall.SIGTERM)

	select {
	case err = <-done:
		if err != nil && err.Error() == "signal: terminated" {
			err = nil
		}

	case <-time.After(5 * time.Second):
		err = e.Cmd.Process.Kill()
	}

	e.Cmd = nil
	e.ctx = nil
	e.cancel = nil

	return err
}
