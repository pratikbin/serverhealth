package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/viper"
)

const (
	configFileName = "config"
)

// NotificationConfig represents notification provider configuration
type NotificationConfig struct {
	Type       string `mapstructure:"type" yaml:"type"`
	Enabled    bool   `mapstructure:"enabled" yaml:"enabled"`
	WebhookURL string `mapstructure:"webhook_url" yaml:"webhook_url,omitempty"`
	BotToken   string `mapstructure:"bot_token" yaml:"bot_token,omitempty"`
	ChatID     string `mapstructure:"chat_id" yaml:"chat_id,omitempty"`
}

// MonitoringConfig represents monitoring configuration
type MonitoringConfig struct {
	Enabled        bool `mapstructure:"enabled" yaml:"enabled"`
	Threshold      int  `mapstructure:"threshold" yaml:"threshold"`
	CheckInterval  int  `mapstructure:"check_interval" yaml:"check_interval"`
	MaxDailyAlerts int  `mapstructure:"max_daily_alerts" yaml:"max_daily_alerts"`
}

// Config represents the application configuration
type Config struct {
	// Monitoring settings
	Disk   MonitoringConfig `mapstructure:"disk" yaml:"disk"`
	CPU    MonitoringConfig `mapstructure:"cpu" yaml:"cpu"`
	Memory MonitoringConfig `mapstructure:"memory" yaml:"memory"`

	// Notification settings
	Notifications []NotificationConfig `mapstructure:"notifications" yaml:"notifications"`

	// General settings
	LogLevel    string `mapstructure:"log_level" yaml:"log_level"`
	ServiceName string `mapstructure:"service_name" yaml:"service_name"`

	// Legacy support (deprecated)
	SlackDiskWebhookURL      string `mapstructure:"slack_disk_webhook_url" yaml:"slack_disk_webhook_url,omitempty"`
	SlackCPUMemoryWebhookURL string `mapstructure:"slack_cpu_memory_webhook_url" yaml:"slack_cpu_memory_webhook_url,omitempty"`
	DiskEnabled              bool   `mapstructure:"disk_enabled" yaml:"disk_enabled,omitempty"`
	CPUEnabled               bool   `mapstructure:"cpu_enabled" yaml:"cpu_enabled,omitempty"`
	MemoryEnabled            bool   `mapstructure:"memory_enabled" yaml:"memory_enabled,omitempty"`
	DiskThreshold            int    `mapstructure:"disk_threshold" yaml:"disk_threshold,omitempty"`
	CPUThreshold             int    `mapstructure:"cpu_threshold" yaml:"cpu_threshold,omitempty"`
	MemoryThreshold          int    `mapstructure:"memory_threshold" yaml:"memory_threshold,omitempty"`
	CheckInterval            int    `mapstructure:"check_interval_minutes" yaml:"check_interval_minutes,omitempty"`
	DiskCheckInterval        int    `mapstructure:"disk_check_interval_hours" yaml:"disk_check_interval_hours,omitempty"`
}

// NewConfig creates a new configuration with default values
func NewConfig() *Config {
	return &Config{
		Disk: MonitoringConfig{
			Enabled:        true,
			Threshold:      80,
			CheckInterval:  1, // hours
			MaxDailyAlerts: 5,
		},
		CPU: MonitoringConfig{
			Enabled:        true,
			Threshold:      90,
			CheckInterval:  1, // minutes
			MaxDailyAlerts: 5,
		},
		Memory: MonitoringConfig{
			Enabled:        true,
			Threshold:      90,
			CheckInterval:  1, // minutes
			MaxDailyAlerts: 5,
		},
		Notifications: []NotificationConfig{},
		LogLevel:      "info",
		ServiceName:   appName,
	}
}

// getConfigDir returns the configuration directory path
func getConfigDir() string {
	configDir := os.Getenv("XDG_CONFIG_HOME")
	if configDir == "" {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			log.Fatal("Could not determine user home directory")
		}
		configDir = filepath.Join(homeDir, ".config")
	}
	return filepath.Join(configDir, appName)
}

// LoadConfig loads configuration from file and environment variables
func LoadConfig(config *Config) error {
	configDir := getConfigDir()
	viper.SetConfigName(configFileName)
	viper.SetConfigType("yaml")
	viper.AddConfigPath(configDir)

	// Set defaults
	viper.SetDefault("disk.enabled", true)
	viper.SetDefault("disk.threshold", 80)
	viper.SetDefault("disk.check_interval", 12)
	viper.SetDefault("disk.max_daily_alerts", 5)

	viper.SetDefault("cpu.enabled", true)
	viper.SetDefault("cpu.threshold", 85)
	viper.SetDefault("cpu.check_interval", 60)
	viper.SetDefault("cpu.max_daily_alerts", 5)

	viper.SetDefault("memory.enabled", true)
	viper.SetDefault("memory.threshold", 85)
	viper.SetDefault("memory.check_interval", 60)
	viper.SetDefault("memory.max_daily_alerts", 5)

	viper.SetDefault("log_level", "info")
	viper.SetDefault("service_name", appName)

	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return fmt.Errorf("failed to read config file: %w", err)
		}
		// Config file not found, use defaults
	}

	// Unmarshal configuration
	if err := viper.Unmarshal(config); err != nil {
		return fmt.Errorf("failed to unmarshal config: %w", err)
	}

	// Migrate legacy configuration
	config.migrateLegacyConfig()

	return nil
}

// migrateLegacyConfig migrates legacy configuration to new structure
func (c *Config) migrateLegacyConfig() {
	// Migrate legacy disk configuration
	if c.DiskEnabled {
		c.Disk.Enabled = true
		if c.DiskThreshold > 0 {
			c.Disk.Threshold = c.DiskThreshold
		}
		if c.DiskCheckInterval > 0 {
			c.Disk.CheckInterval = c.DiskCheckInterval
		}
	}

	// Migrate legacy CPU configuration
	if c.CPUEnabled {
		c.CPU.Enabled = true
		if c.CPUThreshold > 0 {
			c.CPU.Threshold = c.CPUThreshold
		}
		if c.CheckInterval > 0 {
			c.CPU.CheckInterval = c.CheckInterval
		}
	}

	// Migrate legacy memory configuration
	if c.MemoryEnabled {
		c.Memory.Enabled = true
		if c.MemoryThreshold > 0 {
			c.Memory.Threshold = c.MemoryThreshold
		}
		if c.CheckInterval > 0 {
			c.Memory.CheckInterval = c.CheckInterval
		}
	}

	// Migrate legacy Slack notifications
	if c.SlackDiskWebhookURL != "" || c.SlackCPUMemoryWebhookURL != "" {
		// Add Slack notification for disk
		if c.SlackDiskWebhookURL != "" {
			c.Notifications = append(c.Notifications, NotificationConfig{
				Type:       "slack",
				Enabled:    true,
				WebhookURL: c.SlackDiskWebhookURL,
			})
		}

		// Add Slack notification for CPU/Memory
		if c.SlackCPUMemoryWebhookURL != "" {
			c.Notifications = append(c.Notifications, NotificationConfig{
				Type:       "slack",
				Enabled:    true,
				WebhookURL: c.SlackCPUMemoryWebhookURL,
			})
		}
	}
}

// SaveConfig saves configuration to file
func SaveConfig(config *Config) error {
	configDir := getConfigDir()
	if err := os.MkdirAll(configDir, 0o755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	// Set new configuration structure
	viper.Set("disk", config.Disk)
	viper.Set("cpu", config.CPU)
	viper.Set("memory", config.Memory)
	viper.Set("notifications", config.Notifications)
	viper.Set("log_level", config.LogLevel)
	viper.Set("service_name", config.ServiceName)

	configFile := filepath.Join(configDir, configFileName+".yaml")
	return viper.WriteConfigAs(configFile)
}

// Validate validates the configuration with comprehensive checks
func (c *Config) Validate() error {
	var errors []string

	// Check if at least one monitoring option is enabled
	if !c.Disk.Enabled && !c.CPU.Enabled && !c.Memory.Enabled {
		errors = append(errors, "at least one monitoring option must be enabled")
	}

	// Validate disk monitoring configuration
	if c.Disk.Enabled {
		if c.Disk.Threshold < 1 || c.Disk.Threshold > 100 {
			errors = append(errors, "disk threshold must be between 1 and 100")
		}
		if c.Disk.CheckInterval < 1 || c.Disk.CheckInterval > 168 {
			errors = append(errors, "disk check interval must be between 1 and 168 hours")
		}
		if c.Disk.MaxDailyAlerts < 1 || c.Disk.MaxDailyAlerts > 100 {
			errors = append(errors, "disk max daily alerts must be between 1 and 100")
		}
	}

	// Validate CPU monitoring configuration
	if c.CPU.Enabled {
		if c.CPU.Threshold < 1 || c.CPU.Threshold > 100 {
			errors = append(errors, "CPU threshold must be between 1 and 100")
		}
		if c.CPU.CheckInterval < 1 || c.CPU.CheckInterval > 1440 {
			errors = append(errors, "CPU check interval must be between 1 and 1440 minutes")
		}
		if c.CPU.MaxDailyAlerts < 1 || c.CPU.MaxDailyAlerts > 100 {
			errors = append(errors, "CPU max daily alerts must be between 1 and 100")
		}
	}

	// Validate memory monitoring configuration
	if c.Memory.Enabled {
		if c.Memory.Threshold < 1 || c.Memory.Threshold > 100 {
			errors = append(errors, "memory threshold must be between 1 and 100")
		}
		if c.Memory.CheckInterval < 1 || c.Memory.CheckInterval > 1440 {
			errors = append(errors, "memory check interval must be between 1 and 1440 minutes")
		}
		if c.Memory.MaxDailyAlerts < 1 || c.Memory.MaxDailyAlerts > 100 {
			errors = append(errors, "memory max daily alerts must be between 1 and 100")
		}
	}

	// Validate log level
	validLogLevels := map[string]bool{
		"debug": true,
		"info":  true,
		"warn":  true,
		"error": true,
	}
	if !validLogLevels[c.LogLevel] {
		errors = append(errors, "log level must be one of: debug, info, warn, error")
	}

	// Validate service name
	if c.ServiceName == "" {
		errors = append(errors, "service name cannot be empty")
	}

	// Validate notifications
	enabledNotifications := 0
	for i, notification := range c.Notifications {
		if notification.Enabled {
			enabledNotifications++
			if err := c.validateNotification(&notification); err != nil {
				errors = append(errors, fmt.Sprintf("notification %d (%s): %v", i+1, notification.Type, err))
			}
		}
	}

	// Check if we have at least one enabled notification if monitoring is enabled
	if (c.Disk.Enabled || c.CPU.Enabled || c.Memory.Enabled) && enabledNotifications == 0 {
		errors = append(errors, "at least one notification provider must be enabled when monitoring is enabled")
	}

	if len(errors) > 0 {
		return fmt.Errorf("configuration validation failed:\n  • %s", strings.Join(errors, "\n  • "))
	}

	return nil
}

// validateNotification validates a single notification configuration
func (c *Config) validateNotification(notification *NotificationConfig) error {
	switch notification.Type {
	case "slack":
		if notification.WebhookURL == "" {
			return fmt.Errorf("webhook URL is required for Slack notifications")
		}
		if !strings.HasPrefix(notification.WebhookURL, "https://hooks.slack.com/") {
			return fmt.Errorf("webhook URL must be from hooks.slack.com")
		}
	case "telegram":
		if notification.BotToken == "" {
			return fmt.Errorf("bot token is required for Telegram notifications")
		}
		if notification.ChatID == "" {
			return fmt.Errorf("chat ID is required for Telegram notifications")
		}
	case "discord":
		if notification.WebhookURL == "" {
			return fmt.Errorf("webhook URL is required for Discord notifications")
		}
		if !strings.Contains(notification.WebhookURL, "discord.com") && !strings.Contains(notification.WebhookURL, "discordapp.com") {
			return fmt.Errorf("webhook URL must be from discord.com or discordapp.com")
		}
	default:
		return fmt.Errorf("unsupported notification type: %s", notification.Type)
	}

	return nil
}

// GetEnabledNotifications returns all enabled notification providers
func (c *Config) GetEnabledNotifications() []NotificationConfig {
	var enabled []NotificationConfig
	for _, notification := range c.Notifications {
		if notification.Enabled {
			enabled = append(enabled, notification)
		}
	}
	return enabled
}
