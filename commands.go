package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/spf13/cobra"
)

// NewConfigureCmd creates the configure command
func NewConfigureCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "configure",
		Short: "Configure the monitoring settings",
		Run:   runConfigure,
	}
}

// NewStartCmd creates the start command with background support
func NewStartCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "start",
		Short: "Start the monitoring service",
		Run:   runStart,
	}

	// Add flags for background operation
	cmd.Flags().BoolP("background", "b", false, "Run in background mode")
	cmd.Flags().BoolP("daemon", "d", false, "Run as daemon (background)")

	return cmd
}

// NewStatusCmd creates the status command
func NewStatusCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "status",
		Short: "Check the status of the monitoring service",
		Run:   runStatus,
	}
}

// NewStopCmd creates the stop command
func NewStopCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "stop",
		Short: "Stop the monitoring service",
		Run:   runStop,
	}
}

// NewInstallCmd creates the install command
func NewInstallCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "install",
		Short: "Install the service to start automatically",
		Run:   runInstall,
	}
}

// NewUninstallCmd creates the uninstall command
func NewUninstallCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "uninstall",
		Short: "Uninstall the service",
		Run:   runUninstall,
	}
}

// NewLogsCmd creates the logs command
func NewLogsCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "logs",
		Short: "View service logs",
		Run:   runLogs,
	}
}

// NewDaemonCmd creates the daemon command
func NewDaemonCmd() *cobra.Command {
	return &cobra.Command{
		Use:    "daemon",
		Short:  "Run as background daemon (used by service)",
		Run:    runDaemon,
		Hidden: true, // Hide from help as it's internal
	}
}

// runConfigure runs the configuration wizard
func runConfigure(_ *cobra.Command, _ []string) {
	fmt.Println(bold("🔧 ServerHealth Configuration"))
	fmt.Println("═══════════════════════════════════════")

	config := NewConfig()

	// Check if configuration already exists
	configDir := getConfigDir()
	configFile := filepath.Join(configDir, configFileName+".yaml")
	if _, err := os.Stat(configFile); err == nil {
		fmt.Println(yellow("Existing configuration found. Loading current settings..."))
		if err := LoadConfig(config); err != nil {
			fmt.Println(red("Failed to load existing configuration. Creating new configuration..."))
		}
	} else {
		fmt.Println(yellow("No existing configuration found. Creating new configuration..."))
	}

	wizard := &ConfigurationWizard{}
	if err := wizard.Run(config); err != nil {
		fmt.Println(red("Configuration failed: " + err.Error()))
		os.Exit(1)
	}

	fmt.Println(green("Configuration saved successfully!"))
}

// runStart starts the monitoring service
func runStart(_ *cobra.Command, _ []string) {
	config := NewConfig()
	if err := LoadConfig(config); err != nil {
		fmt.Println(red("Failed to load configuration: " + err.Error()))
		os.Exit(1)
	}

	if err := config.Validate(); err != nil {
		fmt.Println(red("Configuration validation failed: " + err.Error()))
		os.Exit(1)
	}

	// Check if already running
	pidFile := getPIDFile()
	if checkPIDFile(pidFile) {
		fmt.Println(yellow("ServerHealth is already running."))
		return
	}

	// Create logger
	logger := log.New(os.Stdout, "", log.LstdFlags)

	// Create and start monitor
	monitor := NewMonitor(config, logger)
	monitor.Start()
}

// runStatus shows the current status and configuration
func runStatus(_ *cobra.Command, _ []string) {
	config := NewConfig()
	if err := LoadConfig(config); err != nil {
		fmt.Println(red("Failed to load configuration: " + err.Error()))
		os.Exit(1)
	}

	showStatus(config)
}

// showStatus displays the current status and configuration
func showStatus(config *Config) {
	fmt.Println(bold("📊 ServerHealth Status"))
	fmt.Println("═══════════════════════════════════════")

	// Check if service is running
	pidFile := getPIDFile()
	isRunning := false
	if checkPIDFile(pidFile) {
		isRunning = true
	}

	if isRunning {
		fmt.Println(green("✅ Service Status: Running"))
	} else {
		fmt.Println(red("❌ Service Status: Stopped"))
	}

	fmt.Println("\n📋 Configuration:")
	fmt.Printf("  • Log Level: %s\n", config.LogLevel)
	fmt.Printf("  • Service Name: %s\n", config.ServiceName)

	// Show monitoring configuration
	fmt.Println("\n🔍 Monitoring Configuration:")
	if config.Disk.Enabled {
		fmt.Printf("  • Disk usage (threshold: %d%%, check every %d hours, max alerts: %d/day)\n",
			config.Disk.Threshold, config.Disk.CheckInterval, config.Disk.MaxDailyAlerts)
	}
	if config.CPU.Enabled {
		fmt.Printf("  • CPU usage (threshold: %d%%, check every %d minutes, max alerts: %d/day)\n",
			config.CPU.Threshold, config.CPU.CheckInterval, config.CPU.MaxDailyAlerts)
	}
	if config.Memory.Enabled {
		fmt.Printf("  • Memory usage (threshold: %d%%, check every %d minutes, max alerts: %d/day)\n",
			config.Memory.Threshold, config.Memory.CheckInterval, config.Memory.MaxDailyAlerts)
	}

	// Show notification providers
	fmt.Println("\n🔔 Notification Providers:")
	if len(config.Notifications) == 0 {
		fmt.Println("  • No notification providers configured")
	} else {
		for _, notification := range config.Notifications {
			if notification.Enabled {
				switch notification.Type {
				case string(NotificationTypeSlack):
					fmt.Printf("  • Slack: %s\n", notification.WebhookURL)
				case string(NotificationTypeTelegram):
					fmt.Printf("  • Telegram: %s\n", notification.ChatID)
				case string(NotificationTypeDiscord):
					fmt.Printf("  • Discord: %s\n", notification.WebhookURL)
				}
			}
		}
	}
}

// runStop stops the monitoring service
func runStop(_ *cobra.Command, _ []string) {
	pidFile := getPIDFile()
	if !checkPIDFile(pidFile) {
		fmt.Println(yellow("ServerHealth is not running."))
		return
	}

	fmt.Println("Stopping ServerHealth...")
	if err := stopDaemonProcess(pidFile); err != nil {
		fmt.Println(red("Failed to stop ServerHealth: " + err.Error()))
		os.Exit(1)
	}

	fmt.Println(green("ServerHealth stopped successfully."))
}

// Install command implementation
func runInstall(cmd *cobra.Command, args []string) {
	fmt.Println(bold("📦 Installing ServerHealth Service"))
	fmt.Println("════════════════════════════════════")

	config := NewConfig()
	if err := LoadConfig(config); err != nil {
		fmt.Println(red("No configuration found. Please run:"), bold(appName+" configure"))
		os.Exit(1)
	}

	if err := InstallService(config); err != nil {
		fmt.Println(red("Failed to install service:"), err)
		os.Exit(1)
	}

	fmt.Println(green("✅ Service installed successfully!"))
	fmt.Println("The service will now start automatically on boot.")
}

// Uninstall command implementation
func runUninstall(cmd *cobra.Command, args []string) {
	fmt.Println(bold("🗑️ Uninstalling ServerHealth Service"))
	fmt.Println("═════════════════════════════════════")

	config := NewConfig()
	if err := LoadConfig(config); err != nil {
		fmt.Println(yellow("No configuration found, proceeding anyway..."))
		config.ServiceName = appName
	}

	if err := UninstallService(config.ServiceName); err != nil {
		fmt.Println(red("Failed to uninstall service:"), err)
		os.Exit(1)
	}

	fmt.Println(green("✅ Service uninstalled successfully!"))
}

// Logs command implementation
func runLogs(cmd *cobra.Command, args []string) {
	fmt.Println(bold("📋 ServerHealth Logs"))
	fmt.Println("═════════════════════════")

	config := NewConfig()
	if err := LoadConfig(config); err != nil {
		config.ServiceName = appName
	}

	// Check if running as daemon and show daemon logs
	pidFile := filepath.Join(getPIDDir(), appName+".pid")
	if checkPIDFile(pidFile) {
		logFile := filepath.Join(getLogDir(), appName+".log")
		if _, err := os.Stat(logFile); err == nil {
			fmt.Println("Showing daemon logs from:", logFile)
			fmt.Println("═════════════════════════════════════")

			// Use tail -f equivalent for live logs
			cmd := exec.Command("tail", "-f", logFile)
			cmd.Stdout = os.Stdout
			cmd.Stderr = os.Stderr

			if err := cmd.Run(); err != nil {
				fmt.Println(red("Failed to show daemon logs:"), err)
			}
			return
		}
	}

	// Fall back to system service logs
	if err := ShowLogs(config.ServiceName); err != nil {
		fmt.Println(red("Failed to show logs:"), err)
		return
	}
}

// Daemon command implementation
func runDaemon(cmd *cobra.Command, args []string) {
	config := NewConfig()
	if err := LoadConfig(config); err != nil {
		fmt.Println(red("Failed to load configuration: " + err.Error()))
		os.Exit(1)
	}

	if err := config.Validate(); err != nil {
		fmt.Println(red("Configuration validation failed: " + err.Error()))
		os.Exit(1)
	}

	// Create log file
	logDir := getLogDir()
	logFile := filepath.Join(logDir, appName+".log")

	f, err := os.OpenFile(logFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o666)
	if err != nil {
		fmt.Println(red("Failed to create log file: " + err.Error()))
		os.Exit(1)
	}

	defer func() {
		if err := f.Close(); err != nil {
			fmt.Println(red("Failed to close log file: " + err.Error()))
		}
	}()

	// Create logger for daemon mode
	logger := log.New(f, "[ServerHealth] ", log.LstdFlags)
	logger.Println("Starting ServerHealth daemon...")

	monitor := NewMonitor(config, logger)

	// Setup PID file cleanup
	pidFile := filepath.Join(getPIDDir(), appName+".pid")
	defer func() {
		if err := os.Remove(pidFile); err != nil && !os.IsNotExist(err) {
			logger.Printf("Error removing PID file: %v", err)
		}
		logger.Println("Cleaned up PID file")
	}()

	// Setup signal handling for graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// Start monitoring in background
	go monitor.Start()

	logger.Println("ServerHealth daemon started successfully")

	// Wait for shutdown signal
	<-sigChan
	logger.Println("Received shutdown signal, stopping daemon...")
	monitor.Stop()
	logger.Println("ServerHealth daemon stopped")
}

// Helper function to stop daemon process
func stopDaemonProcess(pidFile string) error {
	// Read PID from file
	data, err := os.ReadFile(pidFile)
	if err != nil {
		return fmt.Errorf("failed to read PID file: %v", err)
	}

	pidStr := strings.TrimSpace(string(data))
	pid, err := strconv.Atoi(pidStr)
	if err != nil {
		return fmt.Errorf("invalid PID in file: %v", err)
	}

	// Find the process
	process, err := os.FindProcess(pid)
	if err != nil {
		// Remove stale PID file
		if removeErr := os.Remove(pidFile); removeErr != nil && !os.IsNotExist(removeErr) {
			fmt.Printf("Error removing PID file: %v\n", removeErr)
		}
		return fmt.Errorf("failed to find process: %v", err)
	}

	// Check if process is still running
	if err := process.Signal(syscall.Signal(0)); err != nil {
		// Remove stale PID file
		if removeErr := os.Remove(pidFile); removeErr != nil && !os.IsNotExist(removeErr) {
			fmt.Printf("Error removing PID file: %v\n", removeErr)
		}
		return fmt.Errorf("process is not running (stale PID file)")
	}

	fmt.Printf("Stopping process (PID: %d)...\n", pid)

	// Send SIGTERM for graceful shutdown
	if err := process.Signal(syscall.SIGTERM); err != nil {
		return fmt.Errorf("failed to send SIGTERM: %v", err)
	}

	// Wait for process to exit gracefully
	for i := 0; i < 30; i++ { // Wait up to 30 seconds
		if err := process.Signal(syscall.Signal(0)); err != nil {
			// Process has exited
			if removeErr := os.Remove(pidFile); removeErr != nil && !os.IsNotExist(removeErr) {
				fmt.Printf("Error removing PID file: %v\n", removeErr)
			}
			return nil
		}
		time.Sleep(1 * time.Second)
		if i == 10 {
			fmt.Println("Waiting for graceful shutdown...")
		}
	}

	// Force kill if still running
	fmt.Println("Force killing process...")
	if err := process.Signal(syscall.SIGKILL); err != nil {
		return fmt.Errorf("failed to send SIGKILL: %v", err)
	}

	// Wait a bit more
	time.Sleep(2 * time.Second)
	if err := os.Remove(pidFile); err != nil && !os.IsNotExist(err) {
		fmt.Printf("Error removing PID file: %v\n", err)
	}
	return nil
}

// Helper function to check PID file with improved validation
func checkPIDFile(pidFile string) bool {
	data, err := os.ReadFile(pidFile)
	if err != nil {
		return false
	}

	pidStr := strings.TrimSpace(string(data))
	pid, err := strconv.Atoi(pidStr)
	if err != nil {
		// Remove invalid PID file
		os.Remove(pidFile)
		return false
	}

	// Additional validation: check if PID is reasonable
	if pid <= 0 || pid > 999999 {
		os.Remove(pidFile)
		return false
	}

	return isProcessRunning(pid)
}

// Helper function to check if process is running with improved validation
func isProcessRunning(pid int) bool {
	process, err := os.FindProcess(pid)
	if err != nil {
		return false
	}

	// Send signal 0 to check if process exists
	err = process.Signal(syscall.Signal(0))
	return err == nil
}

// Helper function to get PID directory with improved path handling
func getPIDDir() string {
	if os.Geteuid() == 0 {
		return "/var/run"
	}

	// Use XDG_RUNTIME_DIR if available
	if runtimeDir := os.Getenv("XDG_RUNTIME_DIR"); runtimeDir != "" {
		return filepath.Join(runtimeDir, appName)
	}

	return filepath.Join(os.Getenv("HOME"), ".local", "run")
}

// Helper function to get log directory with improved path handling
func getLogDir() string {
	if os.Geteuid() == 0 {
		return "/var/log"
	}

	// Use XDG_DATA_HOME if available
	if dataHome := os.Getenv("XDG_DATA_HOME"); dataHome != "" {
		return filepath.Join(dataHome, appName, "logs")
	}

	return filepath.Join(os.Getenv("HOME"), ".local", "log")
}

// Helper function to get PID file with improved path handling
func getPIDFile() string {
	return filepath.Join(getPIDDir(), appName+".pid")
}
