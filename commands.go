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

// Configure command implementation
func runConfigure(cmd *cobra.Command, args []string) {
	fmt.Println(bold("🔧 ServerHealth Configuration"))
	fmt.Println("═══════════════════════════════════════")

	config := NewConfig()

	// Load existing config if it exists
	if err := LoadConfig(config); err != nil {
		fmt.Println(yellow("No existing configuration found. Creating new configuration..."))
	}

	// Welcome message
	fmt.Println()
	fmt.Println(blue("Welcome to ServerHealth!"))
	fmt.Println("This tool will help you monitor your server's health and send notifications to Slack.")
	fmt.Println()
	fmt.Println("Let's configure your monitoring preferences:")
	fmt.Println()

	// Run configuration wizard
	wizard := NewConfigurationWizard()
	if err := wizard.Run(config); err != nil {
		fmt.Println(red("Configuration failed:"), err)
		os.Exit(1)
	}

	// Save configuration
	if err := SaveConfig(config); err != nil {
		fmt.Println(red("Failed to save configuration:"), err)
		os.Exit(1)
	}

	fmt.Println()
	fmt.Println(green("✅ Configuration saved successfully!"))
	fmt.Println("Run '" + bold(appName+" start") + "' to begin monitoring.")
}

// Start command implementation with background support
func runStart(cmd *cobra.Command, args []string) {
	config := NewConfig()
	if err := LoadConfig(config); err != nil {
		fmt.Println(red("No configuration found. Please run:"), bold(appName+" configure"))
		os.Exit(1)
	}

	// Check for background/daemon flags
	background, _ := cmd.Flags().GetBool("background")
	daemon, _ := cmd.Flags().GetBool("daemon")

	if background || daemon {
		startInBackground()
		return
	}

	// Check if already running
	if isAlreadyRunning() {
		fmt.Println(yellow("⚠️  ServerHealth is already running"))
		fmt.Println("Use '" + bold(appName+" status") + "' to check status")
		fmt.Println("Use '" + bold(appName+" stop") + "' to stop")
		return
	}

	// Normal foreground start
	startForeground(config)
}

// Status command implementation
func runStatus(cmd *cobra.Command, args []string) {
	fmt.Println(bold("📊 ServerHealth Status"))
	fmt.Println("═══════════════════════════════")

	config := NewConfig()
	if err := LoadConfig(config); err != nil {
		fmt.Println(red("❌ Configuration not found"))
		fmt.Println("Run '" + bold(appName+" configure") + "' to set up monitoring")
		return
	}

	fmt.Println(green("✅ Configuration found"))
	fmt.Println()
	fmt.Println(bold("Monitoring Configuration:"))

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

	fmt.Println()
	fmt.Println(bold("Runtime Status:"))

	isRunning := false
	runningAs := ""

	// Check if service is running
	if IsServiceRunning(config.ServiceName) {
		fmt.Println(green("✅ Running as system service"))
		isRunning = true
		runningAs = "system service"
	}

	// Check if daemon process is running
	pidFile := filepath.Join(getPIDDir(), appName+".pid")
	if checkPIDFile(pidFile) {
		if isRunning {
			fmt.Println(yellow("⚠️  Also running as daemon process (this might be a problem)"))
		} else {
			fmt.Println(green("✅ Running as daemon process"))
			isRunning = true
			runningAs = "daemon process"
		}

		// Show PID
		if data, err := os.ReadFile(pidFile); err == nil {
			pidStr := string(data)
			if pid, err := strconv.Atoi(pidStr[:len(pidStr)-1]); err == nil {
				fmt.Printf("   PID: %d\n", pid)
			}
		}
	}

	if !isRunning {
		fmt.Println(red("❌ ServerHealth is not running"))
		fmt.Println()
		fmt.Println(bold("To start ServerHealth:"))
		fmt.Println("  • Foreground: " + bold(appName+" start"))
		fmt.Println("  • Background: " + bold(appName+" start --background"))
		fmt.Println("  • As service: " + bold(appName+" install") + " then " + bold("systemctl start "+appName))
		return
	}

	// Show additional info if running
	fmt.Println()
	fmt.Println(bold("Additional Info:"))

	if runningAs == "daemon process" {
		logFile := filepath.Join(getLogDir(), appName+".log")
		fmt.Printf("  • Log file: %s\n", logFile)
		fmt.Printf("  • PID file: %s\n", pidFile)
	}

	// Show notification providers
	if len(config.Notifications) > 0 {
		fmt.Println("  • Notification providers:")
		for _, notification := range config.Notifications {
			if notification.Enabled {
				switch notification.Type {
				case "slack":
					fmt.Printf("    - Slack: %s\n", notification.WebhookURL[:30]+"...")
				case "telegram":
					fmt.Printf("    - Telegram: Bot token configured\n")
				case "discord":
					fmt.Printf("    - Discord: %s\n", notification.WebhookURL[:30]+"...")
				}
			}
		}
	} else {
		fmt.Println("  • No notification providers configured")
	}

	fmt.Println()
	fmt.Println(bold("Commands:"))
	fmt.Println("  • Stop: " + bold(appName+" stop"))
	fmt.Println("  • View logs: " + bold(appName+" logs"))
	fmt.Println("  • Reconfigure: " + bold(appName+" configure"))
}

// Stop command implementation
func runStop(cmd *cobra.Command, args []string) {
	fmt.Println(bold("🛑 Stopping ServerHealth"))
	fmt.Println("══════════════════════════════")

	config := NewConfig()
	if err := LoadConfig(config); err != nil {
		config.ServiceName = appName // Use default if config not found
	}

	stopped := false

	// First, try to stop system service
	if IsServiceRunning(config.ServiceName) {
		fmt.Println("Stopping system service...")
		if err := StopService(config.ServiceName); err != nil {
			fmt.Println(red("Failed to stop service:"), err)
		} else {
			fmt.Println(green("✅ System service stopped"))
			stopped = true
		}
	}

	// Then, try to stop daemon process
	pidFile := filepath.Join(getPIDDir(), appName+".pid")
	if checkPIDFile(pidFile) {
		fmt.Println("Stopping daemon process...")
		if err := stopDaemonProcess(pidFile); err != nil {
			fmt.Println(red("Failed to stop daemon:"), err)
		} else {
			fmt.Println(green("✅ Daemon process stopped"))
			stopped = true
		}
	}

	if !stopped {
		fmt.Println(yellow("⚠️  ServerHealth is not running"))
	}
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
		log.Fatal("No configuration found. Please run 'serverhealth configure' first.")
	}

	// Setup logging to file
	logFile := filepath.Join(getLogDir(), appName+".log")
	f, err := os.OpenFile(logFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o666)
	if err != nil {
		log.Fatal("Failed to open log file:", err)
	}
	defer func() {
		if err := f.Close(); err != nil {
			log.Printf("Error closing log file: %v", err)
		}
	}()

	// Create logger for daemon mode
	logger := log.New(f, "[ServerHealth] ", log.LstdFlags)
	logger.Println("Starting ServerHealth daemon...")

	monitor := NewMonitor(config)
	monitor.SetLogger(logger)

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

// Background start implementation
func startInBackground() {
	// Check if already running
	if isAlreadyRunning() {
		fmt.Println(yellow("⚠️  ServerHealth is already running"))
		return
	}

	fmt.Println(bold("🚀 Starting ServerHealth in background"))
	fmt.Println("═══════════════════════════════════════")

	// Get current executable path
	execPath, err := os.Executable()
	if err != nil {
		fmt.Println(red("Failed to get executable path:"), err)
		os.Exit(1)
	}

	// Ensure directories exist
	if err := ensureDirectories(); err != nil {
		fmt.Println(red("Failed to create directories:"), err)
		os.Exit(1)
	}

	pidDir := getPIDDir()
	logDir := getLogDir()
	pidFile := filepath.Join(pidDir, appName+".pid")
	logFile := filepath.Join(logDir, appName+".log")

	// Start daemon process
	cmd := exec.Command(execPath, "daemon")

	// Setup logging
	f, err := os.OpenFile(logFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o666)
	if err != nil {
		fmt.Println(red("Failed to open log file:"), err)
		os.Exit(1)
	}

	cmd.Stdout = f
	cmd.Stderr = f

	// Detach from parent process (platform-specific)
	setPlatformSysProcAttr(cmd)

	// Start the process
	if err := cmd.Start(); err != nil {
		fmt.Println(red("Failed to start daemon:"), err)
		if closeErr := f.Close(); closeErr != nil {
			fmt.Printf("Error closing file: %v\n", closeErr)
		}
		os.Exit(1)
	}

	// Write PID file
	pidContent := fmt.Sprintf("%d\n", cmd.Process.Pid)
	if err := os.WriteFile(pidFile, []byte(pidContent), 0o644); err != nil {
		fmt.Println(red("Failed to write PID file:"), err)
		if killErr := cmd.Process.Kill(); killErr != nil {
			fmt.Printf("Error killing process: %v\n", killErr)
		}
		if closeErr := f.Close(); closeErr != nil {
			fmt.Printf("Error closing file: %v\n", closeErr)
		}
		os.Exit(1)
	}

	if err := f.Close(); err != nil {
		fmt.Printf("Error closing file: %v\n", err)
	}

	// Wait a moment to ensure process started
	time.Sleep(200 * time.Millisecond)

	// Verify it's running
	if !isProcessRunning(cmd.Process.Pid) {
		fmt.Println(red("❌ Failed to start daemon"))
		if err := os.Remove(pidFile); err != nil && !os.IsNotExist(err) {
			fmt.Printf("Error removing PID file: %v\n", err)
		}
		os.Exit(1)
	}

	fmt.Println(green("✅ ServerHealth started in background!"))
	fmt.Printf("PID: %d\n", cmd.Process.Pid)
	fmt.Printf("Log file: %s\n", logFile)
	fmt.Println("Use '" + bold(appName+" status") + "' to check status")
	fmt.Println("Use '" + bold(appName+" stop") + "' to stop")
	fmt.Println("Use '" + bold(appName+" logs") + "' to view logs")
}

// Foreground start implementation
func startForeground(config *Config) {
	fmt.Println(bold("🚀 Starting ServerHealth"))
	fmt.Println("═══════════════════════════════════")

	monitor := NewMonitor(config)

	// Setup signal handling
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// Start monitoring
	go monitor.Start()

	fmt.Println(green("✅ ServerHealth started successfully!"))
	fmt.Println("Press Ctrl+C to stop...")

	// Wait for shutdown signal
	<-sigChan
	fmt.Println(yellow("\n🛑 Stopping ServerHealth..."))
	monitor.Stop()
	fmt.Println(green("✅ ServerHealth stopped successfully!"))
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

// Helper function to check if already running
func isAlreadyRunning() bool {
	// Check if running as system service
	if IsServiceRunning(appName) {
		return true
	}

	// Check if daemon process is running
	pidFile := filepath.Join(getPIDDir(), appName+".pid")
	return checkPIDFile(pidFile)
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

// Helper function to ensure directories exist with proper permissions
func ensureDirectories() error {
	pidDir := getPIDDir()
	logDir := getLogDir()

	// Create PID directory
	if err := os.MkdirAll(pidDir, 0o755); err != nil {
		return fmt.Errorf("failed to create PID directory: %w", err)
	}

	// Create log directory
	if err := os.MkdirAll(logDir, 0o755); err != nil {
		return fmt.Errorf("failed to create log directory: %w", err)
	}

	return nil
}
