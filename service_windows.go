//go:build windows
// +build windows

package main

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
	"syscall"
)

// isServiceRunning checks if the Windows service is running
func isServiceRunning(serviceName string) bool {
	cmd := exec.Command("sc", "query", serviceName)
	cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}
	output, err := cmd.Output()
	if err != nil {
		return false
	}
	return strings.Contains(string(output), "RUNNING")
}

// stopService stops the Windows service
func stopService(serviceName string) error {
	cmd := exec.Command("sc", "stop", serviceName)
	cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}
	return cmd.Run()
}

// installService installs the service using Windows Service Control Manager
func installService(config *Config) error {
	serviceName := config.ServiceName
	if serviceName == "" {
		serviceName = appName
	}

	// Get the current executable path
	execPath, err := os.Executable()
	if err != nil {
		return fmt.Errorf("failed to get executable path: %v", err)
	}

	// Create service using sc command
	cmd := exec.Command("sc", "create", serviceName,
		"binPath=", fmt.Sprintf("\"%s\" daemon", execPath),
		"DisplayName=", "Server Health Monitor",
		"start=", "auto",
		"depend=", "tcpip")
	cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to create service: %v", err)
	}

	// Set service description
	cmd = exec.Command("sc", "description", serviceName, "Monitors server health metrics and sends Slack notifications")
	cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}
	cmd.Run() // Don't fail if this doesn't work

	// Set recovery options
	cmd = exec.Command("sc", "failure", serviceName, "reset=", "86400", "actions=", "restart/30000/restart/30000/restart/30000")
	cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}
	cmd.Run() // Don't fail if this doesn't work

	fmt.Printf("✅ Service installed successfully!\n")
	fmt.Printf("Run 'sc start %s' to start the service\n", serviceName)
	fmt.Printf("The service will start automatically on boot\n")

	return nil
}

// uninstallService removes the Windows service
func uninstallService(serviceName string) error {
	// Stop service first
	stopService(serviceName) // Don't fail if already stopped

	// Delete service
	cmd := exec.Command("sc", "delete", serviceName)
	cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to delete service: %v", err)
	}

	fmt.Printf("✅ Service uninstalled successfully!\n")
	return nil
}
