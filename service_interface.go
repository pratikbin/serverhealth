package main

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"
)

// Service management interface functions
// These are implemented in platform-specific files

// IsServiceRunning checks if the service is currently running
func IsServiceRunning(serviceName string) bool {
	return isServiceRunning(serviceName)
}

// StopService stops the specified service
func StopService(serviceName string) error {
	return stopService(serviceName)
}

// InstallService installs the service for automatic startup
func InstallService(config *Config) error {
	return installService(config)
}

// UninstallService removes the service
func UninstallService(serviceName string) error {
	return uninstallService(serviceName)
}

// ShowLogs displays service logs
func ShowLogs(serviceName string) error {
	return showLogs(serviceName)
}

// Default implementation for showLogs (can be overridden by platform-specific files)
func showLogs(serviceName string) error {
	switch runtime.GOOS {
	case "linux":
		cmd := exec.Command("journalctl", "-u", serviceName, "-f", "--no-pager")
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		return cmd.Run()
	case "windows":
		fmt.Println("Check Windows Event Viewer for service logs")
		return nil
	case "darwin":
		cmd := exec.Command("log", "stream", "--predicate", fmt.Sprintf("subsystem == '%s'", serviceName))
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		return cmd.Run()
	default:
		return fmt.Errorf("unsupported platform: %s", runtime.GOOS)
	}
}
