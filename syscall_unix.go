//go:build !windows
// +build !windows

package main

import (
	"os/exec"
	"syscall"
)

// setPlatformSysProcAttr sets platform-specific process attributes for Unix systems
// nolint:unused // This function is kept for potential future background process functionality
func setPlatformSysProcAttr(cmd *exec.Cmd) {
	cmd.SysProcAttr = &syscall.SysProcAttr{
		Setpgid: true,
	}
}
