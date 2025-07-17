//go:build darwin
// +build darwin

package main

import (
	"fmt"
	"os"
	"os/exec"
	"os/user"
	"path/filepath"
)

// isServiceRunning checks if the launchd service is running
func isServiceRunning(serviceName string) bool {
	cmd := exec.Command("launchctl", "list", serviceName)
	err := cmd.Run()
	return err == nil
}

// stopService stops the launchd service
func stopService(serviceName string) error {
	cmd := exec.Command("launchctl", "unload", getPlistPath(serviceName))
	return cmd.Run()
}

// installService installs the service as a launchd daemon
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

	// Get current user
	currentUser, err := user.Current()
	if err != nil {
		return fmt.Errorf("failed to get current user: %v", err)
	}

	// Determine if this is a user or system service
	isSystemService := currentUser.Uid == "0"
	plistPath := getPlistPath(serviceName)

	// Create plist content
	plistContent := fmt.Sprintf(`<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
	<key>Label</key>
	<string>%s</string>
	<key>ProgramArguments</key>
	<array>
		<string>%s</string>
		<string>daemon</string>
	</array>
	<key>RunAtLoad</key>
	<true/>
	<key>KeepAlive</key>
	<true/>
	<key>StandardOutPath</key>
	<string>/var/log/%s.log</string>
	<key>StandardErrorPath</key>
	<string>/var/log/%s.log</string>
</dict>
</plist>
`, serviceName, execPath, serviceName, serviceName)

	// Create directory if needed
	plistDir := filepath.Dir(plistPath)
	if err := os.MkdirAll(plistDir, 0755); err != nil {
		return fmt.Errorf("failed to create plist directory: %v", err)
	}

	// Write plist file
	if err := os.WriteFile(plistPath, []byte(plistContent), 0644); err != nil {
		return fmt.Errorf("failed to create plist file: %v", err)
	}

	// Load the service
	cmd := exec.Command("launchctl", "load", plistPath)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to load service: %v", err)
	}

	serviceType := "user"
	if isSystemService {
		serviceType = "system"
	}

	fmt.Printf("✅ Service installed successfully as %s service!\n", serviceType)
	fmt.Printf("Service will start automatically on boot\n")
	fmt.Printf("To start now: launchctl start %s\n", serviceName)

	return nil
}

// uninstallService removes the launchd service
func uninstallService(serviceName string) error {
	plistPath := getPlistPath(serviceName)

	// Unload service
	cmd := exec.Command("launchctl", "unload", plistPath)
	if err := cmd.Run(); err != nil {
		// Don't fail if already unloaded, just log the error
		fmt.Printf("Warning: failed to unload service (may already be unloaded): %v\n", err)
	}

	// Remove plist file
	if err := os.Remove(plistPath); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to remove plist file: %v", err)
	}

	fmt.Printf("✅ Service uninstalled successfully!\n")
	return nil
}

// getPlistPath returns the appropriate plist path based on user privileges
func getPlistPath(serviceName string) string {
	currentUser, err := user.Current()
	if err != nil || currentUser.Uid == "0" {
		// System service (root)
		return fmt.Sprintf("/Library/LaunchDaemons/%s.plist", serviceName)
	}
	// User service
	return fmt.Sprintf("%s/Library/LaunchAgents/%s.plist", currentUser.HomeDir, serviceName)
}
