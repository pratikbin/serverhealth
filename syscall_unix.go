//go:build !windows
// +build !windows

package main

import (
	"os/exec"
	"syscall"
)

func setPlatformSysProcAttr(cmd *exec.Cmd) {
	cmd.SysProcAttr = &syscall.SysProcAttr{
		Setpgid: true,
	}
}
