//go:build !linux && !darwin && !windows
// +build !linux,!darwin,!windows

package main

import (
	"fmt"
	"runtime"
)

// isServiceRunning - fallback implementation for unsupported platforms
func isServiceRunning(serviceName string) bool {
	return false
}

// stopService - fallback implementation for unsupported platforms
func stopService(serviceName string) error {
	return fmt.Errorf("service management not supported on %s", runtime.GOOS)
}

// installService - fallback implementation for unsupported platforms
func installService(config *Config) error {
	return fmt.Errorf("service installation not supported on %s", runtime.GOOS)
}

// uninstallService - fallback implementation for unsupported platforms
func uninstallService(serviceName string) error {
	return fmt.Errorf("service management not supported on %s", runtime.GOOS)
}
