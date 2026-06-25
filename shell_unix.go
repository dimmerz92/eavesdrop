//go:build !windows

package ev

import (
	"errors"
	"fmt"
	"os"
	"syscall"
)

var defaultShell = "/bin/sh"

// DetectShell attempts to determine the shell environment from the $SHELL
// environment variable, otherwise fallse back to /bin/sh
func DetectShell() string {
	prefix := os.Getenv("SHELL")
	if prefix == "" {
		prefix = defaultShell
	}

	return prefix
}

// ShellFlag always returns -c.
func ShellFlag(prefix string) string {
	_ = prefix
	return "-c"
}

// ToProcessGroup sets the shell with a flag to spawn a new process group.
func (s *Shell) ToProcessGroup() error {
	if s.cmd == nil {
		return fmt.Errorf("nil shell")
	}
	s.cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	return nil
}

// SignalProcessGroup sends the given signal to the shell.
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

// TerminateProcessGroup sends a SIGTERM to the shell.
func (s *Shell) TerminateProcessGroup() error {
	return s.SignalProcessGroup(syscall.SIGTERM)
}

// KillProcessGroup sends a SIGKILL to the shell.
func (s *Shell) KillProcessGroup() error {
	return s.SignalProcessGroup(syscall.SIGKILL)
}
