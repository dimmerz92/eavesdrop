//go:build windows

package ev

import (
	"errors"
	"fmt"
	"os/exec"
	"strconv"
	"syscall"

	"golang.org/x/sys/windows"
)

const PROCESS_NOT_FOUND = 128

// DetectShell checks if powershell exists otherwise falls back to cmd.
func DetectShell() string {
	prefix := "powershell.exe"

	_, err := exec.LookPath(prefix)
	if err != nil {
		return "cmd.exe"
	}

	return prefix
}

// ShellFlag returns -Command if the detected shell is powershell, otherwise
// /C is returned.
func ShellFlag(prefix string) string {
	if prefix == "powershell.exe" {
		return "-Command"
	}
	return "/C"
}

// ToProcessGroup sets the shell with a flag to spawn a new process group.
func (s *Shell) ToProcessGroup() error {
	if s.cmd == nil {
		return fmt.Errorf("nil shell")
	}

	s.cmd.SysProcAttr = &syscall.SysProcAttr{
		CreationFlags: syscall.CREATE_NEW_PROCESS_GROUP,
	}

	return nil
}

// TerminateProcessGroup sends a CTRL_BREAK_EVENT to the shell.
func (s *Shell) TerminateProcessGroup() error {
	return windows.GenerateConsoleCtrlEvent(windows.CTRL_BREAK_EVENT, uint32(s.pid))
}

// KillProcessGroup uses taskkill to kill the underlying process group.
func (s *Shell) KillProcessGroup() error {
	pid := strconv.Itoa(s.pid)
	cmd := exec.Command("taskkill", "/F", "/T", "/PID", pid)
	cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}

	err := cmd.Run()
	if err != nil {
		var exitErr *exec.ExitError
		if errors.As(err, &exitErr) && exitErr.ExitCode() == PROCESS_NOT_FOUND {
			err = nil
		}
		return err
	}

	return nil
}
