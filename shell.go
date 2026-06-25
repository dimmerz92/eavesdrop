package ev

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"
)

type Shell struct {
	ctx            context.Context
	cmd            *exec.Cmd
	pid            int
	prefix         string
	flag           string
	taskTimeout    time.Duration
	serviceTimeout time.Duration
}

// NewShell returns a Shell that invokes commands via the detected system shell
// (e.g. /bin/sh on Unix, powershell.exe or cmd.exe on Windows).
func NewShell(ctx context.Context, taskTimeoutMs, serviceTimeoutMs uint) *Shell {
	prefix := DetectShell()
	return &Shell{
		ctx:            ctx,
		prefix:         prefix,
		flag:           ShellFlag(prefix),
		taskTimeout:    time.Duration(taskTimeoutMs) * time.Millisecond,
		serviceTimeout: time.Duration(serviceTimeoutMs) * (time.Millisecond),
	}

}

// ExecAndWait runs task and blocks until it finishes or the task timeout
// elapses.
func (s *Shell) ExecAndWait(task string) error {
	if strings.TrimSpace(task) == "" {
		return fmt.Errorf("cannot run blank task")
	}

	ctx, cancel := context.WithTimeout(s.ctx, s.taskTimeout)
	defer cancel()

	s.cmd = exec.CommandContext(ctx, s.prefix, s.flag, task)

	s.ToProcessGroup()
	s.cmd.Stdout = os.Stdout
	s.cmd.Stderr = os.Stdout

	errCh := make(chan error, 1)
	go func() {
		errCh <- s.cmd.Start()
	}()

	select {
	case <-ctx.Done():
		return s.KillProcessGroup()
	case err := <-errCh:
		if err != nil {
			return err
		}
	}

	s.pid = s.cmd.Process.Pid

	return s.cmd.Wait()
}

// ExecAndReturn starts service as a background process and returns without
// waiting.
func (s *Shell) ExecAndReturn(service string) error {
	if strings.TrimSpace(service) == "" {
		return fmt.Errorf("cannot run blank task")
	}

	s.cmd = exec.CommandContext(s.ctx, s.prefix, s.flag, service)

	s.ToProcessGroup()
	s.cmd.Stdout = os.Stdout
	s.cmd.Stderr = os.Stdout

	err := s.cmd.Start()
	if err != nil {
		return err
	}

	s.pid = s.cmd.Process.Pid

	return nil
}

// Stop gracefully shuts down the running service. Sends SIGTERM and waits up
// to the service timeout before sending SIGKILL.
func (s *Shell) Stop() error {
	if s.cmd == nil || s.cmd.Process == nil {
		return nil
	}

	cmd := s.cmd
	done := make(chan error, 1)
	go func() { done <- cmd.Wait() }()

	s.TerminateProcessGroup()

	select {
	case err := <-done:
		if err != nil {
			var exitErr *exec.ExitError
			if !errors.As(err, &exitErr) {
				return err
			}
		}

	case <-time.After(s.serviceTimeout):
		err := s.KillProcessGroup()
		if err != nil {
			return err
		}
	}

	s.cmd = nil

	return nil
}
