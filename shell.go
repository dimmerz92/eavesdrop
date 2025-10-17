//go:build !windows

package eavesdrop

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"syscall"
	"time"
)

type Shell struct {
	cmd            *exec.Cmd
	ctx            context.Context
	cancel         context.CancelFunc
	taskTimeout    time.Duration
	serviceTimeout time.Duration
}

// NewShell returns a os specific shell ready for executing commands.
// args:
// - taskTimeout is used to give a max time for a single task to run before it is cancelled.
// - serviceTimeout is used to give a max wait for a service to gracefully exit.
func NewShell(taskTimeout, serviceTimeout time.Duration) *Shell {
	return &Shell{
		taskTimeout:    taskTimeout,
		serviceTimeout: serviceTimeout,
	}
}

// Exec runs the given command and waits for the combined output.
// args:
// - command specifies the shell command to be run.
func (s *Shell) Exec(command string) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), s.taskTimeout)
	defer cancel()

	out, err := exec.CommandContext(ctx, "sh", "-c", command).CombinedOutput()
	return strings.TrimSpace(string(out)), err
}

// Run runs the given command in a separate process group without waiting for it to finish.
// Kill the process using the Kill() method.
// args:
// - command specifies the shell command to be run.
func (s *Shell) Run(command string) error {
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

	s.cmd = cmd
	s.ctx = ctx
	s.cancel = cancel

	return nil
}

// Kill signals the process for a graceful shutdown.
func (s *Shell) Kill() error {
	if s.cmd == nil || s.cmd.Process == nil {
		return nil
	}
	defer s.cancel()

	done := make(chan error, 1)
	go func() { done <- s.cmd.Wait() }()

	pgid, err := syscall.Getpgid(s.cmd.Process.Pid)
	if err == nil {
		_ = syscall.Kill(-pgid, syscall.SIGTERM)
	} else {
		_ = s.cmd.Process.Signal(syscall.SIGTERM)
	}

	select {
	case err = <-done:
		if err != nil && strings.Contains(err.Error(), "terminated") {
			err = nil
		}

	case <-time.After(s.serviceTimeout):
		_ = s.cmd.Process.Kill()
		err = <-done
	}

	s.cmd = nil
	s.ctx = nil
	s.cancel = nil

	return err
}
