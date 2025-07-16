//go:build linux

package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"text/template"
)

const systemdServiceTemplate = `[Unit]
Description={{.ServiceName}} - Server Health Monitoring Service
After=network.target
StartLimitIntervalSec=0

[Service]
Type=simple
Restart=always
RestartSec=5
User={{.User}}
Group={{.Group}}
ExecStart={{.ExecPath}} daemon
WorkingDirectory={{.WorkingDir}}
Environment=PATH=/usr/local/bin:/usr/bin:/bin
StandardOutput=journal
StandardError=journal
SyslogIdentifier={{.ServiceName}}

[Install]
WantedBy=multi-user.target
`

type ServiceConfig struct {
	ServiceName string
	User        string
	Group       string
	ExecPath    string
	WorkingDir  string
}

func isServiceRunning(serviceName string) bool {
	cmd := exec.Command("systemctl", "is-active", "--quiet", serviceName)
	return cmd.Run() == nil
}

func stopService(serviceName string) error {
	cmd := exec.Command("systemctl", "stop", serviceName)
	return cmd.Run()
}

func installService(config *Config) error {
	// Get current executable path
	execPath, err := os.Executable()
	if err != nil {
		return fmt.Errorf("failed to get executable path: %w", err)
	}

	// Get current user
	user := os.Getenv("USER")
	if user == "" {
		user = "root"
	}

	// Create service configuration
	serviceConfig := ServiceConfig{
		ServiceName: config.ServiceName,
		User:        user,
		Group:       user,
		ExecPath:    execPath,
		WorkingDir:  filepath.Dir(execPath),
	}

	// Create systemd service file
	servicePath := fmt.Sprintf("/etc/systemd/system/%s.service", config.ServiceName)

	tmpl, err := template.New("service").Parse(systemdServiceTemplate)
	if err != nil {
		return fmt.Errorf("failed to parse service template: %w", err)
	}

	file, err := os.Create(servicePath)
	if err != nil {
		return fmt.Errorf("failed to create service file (try running with sudo): %w", err)
	}
	defer file.Close()

	if err := tmpl.Execute(file, serviceConfig); err != nil {
		return fmt.Errorf("failed to write service file: %w", err)
	}

	// Set correct permissions
	if err := os.Chmod(servicePath, 0644); err != nil {
		return fmt.Errorf("failed to set service file permissions: %w", err)
	}

	// Reload systemd
	if err := exec.Command("systemctl", "daemon-reload").Run(); err != nil {
		return fmt.Errorf("failed to reload systemd: %w", err)
	}

	// Enable service
	if err := exec.Command("systemctl", "enable", config.ServiceName).Run(); err != nil {
		return fmt.Errorf("failed to enable service: %w", err)
	}

	// Start service
	if err := exec.Command("systemctl", "start", config.ServiceName).Run(); err != nil {
		return fmt.Errorf("failed to start service: %w", err)
	}

	return nil
}

func uninstallService(serviceName string) error {
	// Stop service
	exec.Command("systemctl", "stop", serviceName).Run()

	// Disable service
	exec.Command("systemctl", "disable", serviceName).Run()

	// Remove service file
	servicePath := fmt.Sprintf("/etc/systemd/system/%s.service", serviceName)
	if err := os.Remove(servicePath); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to remove service file: %w", err)
	}

	// Reload systemd
	if err := exec.Command("systemctl", "daemon-reload").Run(); err != nil {
		return fmt.Errorf("failed to reload systemd: %w", err)
	}

	return nil
}
