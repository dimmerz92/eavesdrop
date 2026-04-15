//go:build windows

package eavesdrop

import (
	"errors"
	"fmt"
	"os/exec"
	"strconv"
	"syscall"

	"golang.org/x/sys/windows"
)

const PROCESS_NOT_FOUND = 128

func DetectShell() string {
	prefix := "powershell.exe"

	_, err := exec.LookPath(prefix)
	if err != nil {
		return "cmd.exe"
	}

	return prefix
}

func ShellFlag(prefix string) string {
	if prefix == "powershell.exe" {
		return "-Command"
	}
	return "/C"
}

func (s *Shell) ToProcessGroup() error {
	if s.cmd == nil {
		return fmt.Errorf("nil shell")
	}

	s.cmd.SysProcAttr = &syscall.SysProcAttr{CreationFlags: syscall.CREATE_NEW_PROCESS_GROUP}

	return nil
}

func (s *Shell) TerminateProcessGroup() error {
	return windows.GenerateConsoleCtrlEvent(windows.CTRL_BREAK_EVENT, uint32(s.pid))
}

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
