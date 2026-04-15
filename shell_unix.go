//go:build !windows

package eavesdrop

import (
	"errors"
	"fmt"
	"os"
	"slices"
	"syscall"
)

// not exhaustive, covers the vast majority.
var (
	defaultShell = "/bin/sh"
	knownShells  = []string{"/bin/bash", "/bin/zsh", "/bin/tcsh", "/bin/dash", "/bin/fish"}
)

func DetectShell() string {
	prefix := os.Getenv("SHELL")

	if slices.Contains(knownShells, prefix) {
		return prefix
	}

	return defaultShell
}

func ShellFlag(prefix string) string {
	_ = prefix
	return "-c"
}

func (s *Shell) ToProcessGroup() error {
	if s.cmd == nil {
		return fmt.Errorf("nil shell")
	}

	s.cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}

	return nil
}

func (s *Shell) SignalProcessGroup(signal syscall.Signal) error {
	if s.cmd == nil || s.cmd.Process == nil {
		return nil
	}

	pgid, err := syscall.Getpgid(s.pid)
	if err != nil {
		if errors.Is(err, syscall.ESRCH) {
			err = nil
		}
		return err
	}

	return syscall.Kill(-pgid, signal)
}

func (s *Shell) TerminateProcessGroup() error {
	return s.SignalProcessGroup(syscall.SIGTERM)
}

func (s *Shell) KillProcessGroup() error {
	return s.SignalProcessGroup(syscall.SIGKILL)
}
