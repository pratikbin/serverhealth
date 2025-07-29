//go:build linux
// +build linux

package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

// isSystemdAvailable checks if systemd is available and running
func isSystemdAvailable() bool {
	// Check if systemd is running
	cmd := exec.Command("systemctl", "--version")
	if err := cmd.Run(); err != nil {
		return false
	}

	// Check if we're running under systemd
	cmd = exec.Command("systemctl", "is-system-running")
	if err := cmd.Run(); err != nil {
		return false
	}

	return true
}

// isServiceRunning checks if the systemd service is running
func isServiceRunning(serviceName string) bool {
	if !isSystemdAvailable() {
		return false
	}

	cmd := exec.Command("systemctl", "is-active", serviceName)
	err := cmd.Run()
	return err == nil
}

// stopService stops the systemd service
func stopService(serviceName string) error {
	if !isSystemdAvailable() {
		return fmt.Errorf("systemd is not available")
	}

	cmd := exec.Command("systemctl", "stop", serviceName)
	return cmd.Run()
}

// installService installs the service as a systemd unit
func installService(config *Config) error {
	serviceName := config.ServiceName
	if serviceName == "" {
		serviceName = appName
	}

	// Check if systemd is available
	if !isSystemdAvailable() {
		return fmt.Errorf("systemd is not available on this system. ServerHealth requires systemd for service installation")
	}

	// Get the current executable path
	execPath, err := os.Executable()
	if err != nil {
		return fmt.Errorf("failed to get executable path: %v", err)
	}

	// Ensure executable path is absolute
	if !filepath.IsAbs(execPath) {
		absPath, err := filepath.Abs(execPath)
		if err != nil {
			return fmt.Errorf("failed to get absolute executable path: %v", err)
		}
		execPath = absPath
	}

	// Create systemd service file content with improved configuration
	serviceContent := fmt.Sprintf(`[Unit]
Description=Server Health Monitor
Documentation=https://github.com/pratikbin/serverhealth
After=network-online.target
Wants=network-online.target

[Service]
Type=simple
User=serverhealth
Group=serverhealth
ExecStart=%s start
Restart=always
RestartSec=30
StandardOutput=journal
StandardError=journal
SyslogIdentifier=%s

# Security settings
NoNewPrivileges=true
PrivateTmp=true
ProtectSystem=strict
ProtectHome=true
ReadWritePaths=/var/log /var/run /etc/serverhealth

# Resource limits
LimitNOFILE=65536
LimitNPROC=4096

[Install]
WantedBy=multi-user.target
`, execPath, serviceName)

	// Write service file
	serviceFile := fmt.Sprintf("/etc/systemd/system/%s.service", serviceName)
	err = os.WriteFile(serviceFile, []byte(serviceContent), 0o644)
	if err != nil {
		return fmt.Errorf("failed to create service file: %v", err)
	}

	// Create serverhealth user if it doesn't exist
	if err := createServiceUser(); err != nil {
		return fmt.Errorf("failed to create service user: %v", err)
	}

	// Create config directory with proper permissions
	configDir := "/etc/serverhealth"
	if err := os.MkdirAll(configDir, 0o750); err != nil {
		return fmt.Errorf("failed to create config directory: %v", err)
	}

	// Set ownership of config directory
	cmd := exec.Command("chown", "serverhealth:serverhealth", configDir)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to set config directory ownership: %v", err)
	}

	// Set proper permissions
	cmd = exec.Command("chmod", "750", configDir)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to set config directory permissions: %v", err)
	}

	// Reload systemd
	cmd = exec.Command("systemctl", "daemon-reload")
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to reload systemd: %v", err)
	}

	fmt.Printf("✅ Service installed successfully!\n")
	fmt.Printf("Service file: %s\n", serviceFile)
	fmt.Printf("Configuration directory: %s\n", configDir)
	fmt.Printf("Run 'sudo systemctl enable %s' to start on boot\n", serviceName)
	fmt.Printf("Run 'sudo systemctl start %s' to start now\n", serviceName)
	fmt.Printf("Run 'sudo systemctl status %s' to check status\n", serviceName)

	return nil
}

// uninstallService removes the systemd service
func uninstallService(serviceName string) error {
	if !isSystemdAvailable() {
		return fmt.Errorf("systemd is not available")
	}

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

// createServiceUser creates the serverhealth system user with improved error handling
func createServiceUser() error {
	// Check if user already exists
	cmd := exec.Command("id", "serverhealth")
	if err := cmd.Run(); err == nil {
		fmt.Println("✅ Service user 'serverhealth' already exists")
		return nil // User already exists
	}

	// Check if we have sufficient privileges
	if os.Geteuid() != 0 {
		return fmt.Errorf("service installation requires root privileges")
	}

	// Create system user with proper settings
	cmd = exec.Command("useradd",
		"--system",              // System user
		"--no-create-home",      // Don't create home directory
		"--shell", "/bin/false", // No login shell
		"--comment", "ServerHealth monitoring service user", // Description
		"serverhealth")

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to create user: %v", err)
	}

	fmt.Println("✅ Created service user 'serverhealth'")
	return nil
}
