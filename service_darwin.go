//go:build darwin

package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"text/template"
)

const launchdPlistTemplate = `<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
	<key>Label</key>
	<string>{{.ServiceName}}</string>
	<key>ProgramArguments</key>
	<array>
		<string>{{.ExecPath}}</string>
		<string>daemon</string>
	</array>
	<key>RunAtLoad</key>
	<true/>
	<key>KeepAlive</key>
	<true/>
	<key>WorkingDirectory</key>
	<string>{{.WorkingDir}}</string>
	<key>StandardOutPath</key>
	<string>/tmp/{{.ServiceName}}.log</string>
	<key>StandardErrorPath</key>
	<string>/tmp/{{.ServiceName}}.log</string>
	<key>ProcessType</key>
	<string>Background</string>
</dict>
</plist>
`

type LaunchdConfig struct {
	ServiceName string
	ExecPath    string
	WorkingDir  string
}

func isServiceRunning(serviceName string) bool {
	cmd := exec.Command("launchctl", "list", serviceName)
	return cmd.Run() == nil
}

func stopService(serviceName string) error {
	cmd := exec.Command("launchctl", "unload", getPlistPath(serviceName))
	return cmd.Run()
}

func installService(config *Config) error {
	// Get current executable path
	execPath, err := os.Executable()
	if err != nil {
		return fmt.Errorf("failed to get executable path: %w", err)
	}

	// Create launchd configuration
	launchdConfig := LaunchdConfig{
		ServiceName: config.ServiceName,
		ExecPath:    execPath,
		WorkingDir:  filepath.Dir(execPath),
	}

	// Create plist file
	plistPath := getPlistPath(config.ServiceName)

	tmpl, err := template.New("plist").Parse(launchdPlistTemplate)
	if err != nil {
		return fmt.Errorf("failed to parse plist template: %w", err)
	}

	// Ensure directory exists
	if err := os.MkdirAll(filepath.Dir(plistPath), 0755); err != nil {
		return fmt.Errorf("failed to create plist directory: %w", err)
	}

	file, err := os.Create(plistPath)
	if err != nil {
		return fmt.Errorf("failed to create plist file: %w", err)
	}
	defer file.Close()

	if err := tmpl.Execute(file, launchdConfig); err != nil {
		return fmt.Errorf("failed to write plist file: %w", err)
	}

	// Set correct permissions
	if err := os.Chmod(plistPath, 0644); err != nil {
		return fmt.Errorf("failed to set plist file permissions: %w", err)
	}

	// Load service
	if err := exec.Command("launchctl", "load", plistPath).Run(); err != nil {
		return fmt.Errorf("failed to load service: %w", err)
	}

	return nil
}

func uninstallService(serviceName string) error {
	plistPath := getPlistPath(serviceName)

	// Unload service
	exec.Command("launchctl", "unload", plistPath).Run()

	// Remove plist file
	if err := os.Remove(plistPath); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to remove plist file: %w", err)
	}

	return nil
}

func getPlistPath(serviceName string) string {
	homeDir, _ := os.UserHomeDir()
	return filepath.Join(homeDir, "Library", "LaunchAgents", fmt.Sprintf("com.%s.plist", serviceName))
}
