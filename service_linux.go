//go:build linux
// +build linux

package main

import (
	"fmt"
	"os"
	"os/exec"
)

// isServiceRunning checks if the systemd service is running
func isServiceRunning(serviceName string) bool {
	cmd := exec.Command("systemctl", "is-active", serviceName)
	err := cmd.Run()
	return err == nil
}

// stopService stops the systemd service
func stopService(serviceName string) error {
	cmd := exec.Command("systemctl", "stop", serviceName)
	return cmd.Run()
}

// installService installs the service as a systemd unit
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

	// Create systemd service file content
	serviceContent := fmt.Sprintf(`[Unit]
Description=Server Health Monitor
After=network.target

[Service]
Type=simple
User=serverhealth
Group=serverhealth
ExecStart=%s daemon
Restart=always
RestartSec=30
StandardOutput=journal
StandardError=journal

[Install]
WantedBy=multi-user.target
`, execPath)

	// Write service file
	serviceFile := fmt.Sprintf("/etc/systemd/system/%s.service", serviceName)
	err = os.WriteFile(serviceFile, []byte(serviceContent), 0644)
	if err != nil {
		return fmt.Errorf("failed to create service file: %v", err)
	}

	// Create serverhealth user if it doesn't exist
	if err := createServiceUser(); err != nil {
		return fmt.Errorf("failed to create service user: %v", err)
	}

	// Create config directory
	configDir := "/etc/serverhealth"
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %v", err)
	}

	// Set ownership of config directory
	cmd := exec.Command("chown", "serverhealth:serverhealth", configDir)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to set config directory ownership: %v", err)
	}

	// Reload systemd
	cmd = exec.Command("systemctl", "daemon-reload")
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to reload systemd: %v", err)
	}

	fmt.Printf("✅ Service installed successfully!\n")
	fmt.Printf("Run 'sudo systemctl enable %s' to start on boot\n", serviceName)
	fmt.Printf("Run 'sudo systemctl start %s' to start now\n", serviceName)

	return nil
}

// uninstallService removes the systemd service
func uninstallService(serviceName string) error {
	serviceFile := fmt.Sprintf("/etc/systemd/system/%s.service", serviceName)

	// Stop and disable service
	exec.Command("systemctl", "stop", serviceName).Run()
	exec.Command("systemctl", "disable", serviceName).Run()

	// Remove service file
	if err := os.Remove(serviceFile); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to remove service file: %v", err)
	}

	// Reload systemd
	cmd := exec.Command("systemctl", "daemon-reload")
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to reload systemd: %v", err)
	}

	fmt.Printf("✅ Service uninstalled successfully!\n")
	return nil
}

// createServiceUser creates the serverhealth system user
func createServiceUser() error {
	// Check if user already exists
	cmd := exec.Command("id", "serverhealth")
	if err := cmd.Run(); err == nil {
		return nil // User already exists
	}

	// Create system user
	cmd = exec.Command("useradd", "--system", "--no-create-home", "--shell", "/bin/false", "serverhealth")
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to create user: %v", err)
	}

	return nil
}
