package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

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

// NewStartCmd creates the start command
func NewStartCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "start",
		Short: "Start the monitoring service",
		Run:   runStart,
	}
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

func runStart(cmd *cobra.Command, args []string) {
	config := NewConfig()
	if err := LoadConfig(config); err != nil {
		fmt.Println(red("No configuration found. Please run:"), bold(appName+" configure"))
		os.Exit(1)
	}

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

func runStatus(cmd *cobra.Command, args []string) {
	fmt.Println(bold("📊 ServerHealth Status"))
	fmt.Println("═══════════════════════════════")

	config := NewConfig()
	if err := LoadConfig(config); err != nil {
		fmt.Println(red("❌ Configuration not found"))
		return
	}

	fmt.Println(green("✅ Configuration found"))
	fmt.Println("Monitoring enabled for:")

	if config.DiskEnabled {
		fmt.Printf("  • Disk usage (threshold: %d%%)\n", config.DiskThreshold)
	}
	if config.CPUEnabled {
		fmt.Printf("  • CPU usage (threshold: %d%%)\n", config.CPUThreshold)
	}
	if config.MemoryEnabled {
		fmt.Printf("  • Memory usage (threshold: %d%%)\n", config.MemoryThreshold)
	}

	// Check if service is running
	if IsServiceRunning(config.ServiceName) {
		fmt.Println(green("✅ Service is running"))
	} else {
		fmt.Println(yellow("⚠️ Service is not running"))
	}
}

func runStop(cmd *cobra.Command, args []string) {
	fmt.Println(bold("🛑 Stopping ServerHealth"))
	fmt.Println("══════════════════════════════")

	config := NewConfig()
	if err := LoadConfig(config); err != nil {
		fmt.Println(red("Configuration not found"))
		return
	}

	if err := StopService(config.ServiceName); err != nil {
		fmt.Println(red("Failed to stop service:"), err)
		return
	}

	fmt.Println(green("✅ Service stopped successfully"))
}

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

func runUninstall(cmd *cobra.Command, args []string) {
	fmt.Println(bold("🗑️ Uninstalling ServerHealth Service"))
	fmt.Println("═════════════════════════════════════")

	config := NewConfig()
	if err := LoadConfig(config); err != nil {
		fmt.Println(yellow("No configuration found, proceeding anyway..."))
	}

	if err := UninstallService(config.ServiceName); err != nil {
		fmt.Println(red("Failed to uninstall service:"), err)
		os.Exit(1)
	}

	fmt.Println(green("✅ Service uninstalled successfully!"))
}

func runLogs(cmd *cobra.Command, args []string) {
	fmt.Println(bold("📋 ServerHealth Logs"))
	fmt.Println("═════════════════════════")

	config := NewConfig()
	if err := LoadConfig(config); err != nil {
		config.ServiceName = appName
	}

	if err := ShowLogs(config.ServiceName); err != nil {
		fmt.Println(red("Failed to show logs:"), err)
		return
	}
}

func runDaemon(cmd *cobra.Command, args []string) {
	config := NewConfig()
	if err := LoadConfig(config); err != nil {
		log.Fatal("No configuration found. Please run 'serverhealth configure' first.")
	}

	// Create logger for daemon mode
	logger := log.New(os.Stdout, "[ServerHealth] ", log.LstdFlags)
	logger.Println("Starting ServerHealth daemon...")

	monitor := NewMonitor(config)
	monitor.SetLogger(logger)

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
