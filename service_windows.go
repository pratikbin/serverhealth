//go:build windows

package main

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
	"syscall"
)

func isServiceRunning(serviceName string) bool {
	cmd := exec.Command("sc", "query", serviceName)
	cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}
	output, err := cmd.Output()
	if err != nil {
		return false
	}
	return strings.Contains(string(output), "RUNNING")
}

func stopService(serviceName string) error {
	cmd := exec.Command("sc", "stop", serviceName)
	cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}
	return cmd.Run()
}

func installService(config *Config) error {
	// Get current executable path
	execPath, err := os.Executable()
	if err != nil {
		return fmt.Errorf("failed to get executable path: %w", err)
	}

	// Create service
	cmd := exec.Command("sc", "create", config.ServiceName,
		"binPath=", fmt.Sprintf("\"%s\" daemon", execPath),
		"start=", "auto",
		"DisplayName=", "ServerHealth Service",
		"Description=", "Monitors server health and sends Slack notifications")

	cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to create service: %w", err)
	}

	// Start service
	cmd = exec.Command("sc", "start", config.ServiceName)
	cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to start service: %w", err)
	}

	return nil
}

func uninstallService(serviceName string) error {
	// Stop service
	stopCmd := exec.Command("sc", "stop", serviceName)
	stopCmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}
	stopCmd.Run() // Ignore errors

	// Delete service
	cmd := exec.Command("sc", "delete", serviceName)
	cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to delete service: %w", err)
	}

	return nil
}
