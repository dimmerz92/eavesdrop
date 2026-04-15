package eavesdrop

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"
)

const (
	DefaultTaskRunTimeout         = 2 * time.Second
	DefaultServiceShutdownTimeout = 5 * time.Second
)

// var (
// 	ErrShellBlankTask      = errors.New("provided task command was a blank string")
// 	ErrShellBlankService   = errors.New("provided service command was a blank string")
// 	ErrShellInvalidTimeout = errors.New("shell requires positive nonzero task and service timeout")
// 	ErrNilShell            = errors.New("nil shell")
// )

// type Shell interface {
// 	RunTask(task string) error
// 	RunService(service string) error
// 	StopService() error
// 	SignalProcessGroup(signal syscall.Signal) error
// 	TerminateProcessGroup() error
// 	KillProcessGroup() error
// }

type Shell struct {
	ctx            context.Context
	cmd            *exec.Cmd
	pid            int
	prefix         string
	flag           string
	taskTimeout    time.Duration
	serviceTimeout time.Duration
}

type ShellOption func(*Shell)

func WithTaskTimeout(d time.Duration) ShellOption {
	return func(s *Shell) {
		if d > 0 {
			s.taskTimeout = d
		}
	}
}

func WithServiceTimeout(d time.Duration) ShellOption {
	return func(s *Shell) {
		if d > 0 {
			s.serviceTimeout = d
		}
	}
}

func NewShell(ctx context.Context, opts ...ShellOption) *Shell {
	prefix := DetectShell()
	shell := &Shell{
		ctx:            ctx,
		prefix:         prefix,
		flag:           ShellFlag(prefix),
		taskTimeout:    DefaultTaskRunTimeout,
		serviceTimeout: DefaultServiceShutdownTimeout,
	}

	for _, opt := range opts {
		opt(shell)
	}

	return shell
}

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

func (s *Shell) Stop() error {
	if s.cmd == nil || s.cmd.Process == nil {
		return nil
	}

	done := make(chan error, 1)
	go func() { done <- s.cmd.Wait() }()

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
