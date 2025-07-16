package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/spf13/viper"
)

const (
	configFileName = "config"
)

// Config represents the application configuration
type Config struct {
	SlackDiskWebhookURL      string `mapstructure:"slack_disk_webhook_url"`
	SlackCPUMemoryWebhookURL string `mapstructure:"slack_cpu_memory_webhook_url"`
	DiskEnabled              bool   `mapstructure:"disk_enabled"`
	CPUEnabled               bool   `mapstructure:"cpu_enabled"`
	MemoryEnabled            bool   `mapstructure:"memory_enabled"`
	DiskThreshold            int    `mapstructure:"disk_threshold"`
	CPUThreshold             int    `mapstructure:"cpu_threshold"`
	MemoryThreshold          int    `mapstructure:"memory_threshold"`
	CheckInterval            int    `mapstructure:"check_interval_minutes"`
	DiskCheckInterval        int    `mapstructure:"disk_check_interval_hours"`
	LogLevel                 string `mapstructure:"log_level"`
	ServiceName              string `mapstructure:"service_name"`
}

// NewConfig creates a new configuration with default values
func NewConfig() *Config {
	return &Config{
		DiskThreshold:     80,
		CPUThreshold:      85,
		MemoryThreshold:   85,
		CheckInterval:     60,
		DiskCheckInterval: 12,
		LogLevel:          "info",
		ServiceName:       appName,
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

// LoadConfig loads configuration from file
func LoadConfig(config *Config) error {
	configDir := getConfigDir()
	viper.SetConfigName(configFileName)
	viper.SetConfigType("yaml")
	viper.AddConfigPath(configDir)

	// Set defaults
	viper.SetDefault("disk_threshold", 80)
	viper.SetDefault("cpu_threshold", 85)
	viper.SetDefault("memory_threshold", 85)
	viper.SetDefault("check_interval_minutes", 60)
	viper.SetDefault("disk_check_interval_hours", 12)
	viper.SetDefault("log_level", "info")
	viper.SetDefault("service_name", appName)

	if err := viper.ReadInConfig(); err != nil {
		return err
	}

	return viper.Unmarshal(config)
}

// SaveConfig saves configuration to file
func SaveConfig(config *Config) error {
	configDir := getConfigDir()
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return err
	}

	viper.Set("slack_disk_webhook_url", config.SlackDiskWebhookURL)
	viper.Set("slack_cpu_memory_webhook_url", config.SlackCPUMemoryWebhookURL)
	viper.Set("disk_enabled", config.DiskEnabled)
	viper.Set("cpu_enabled", config.CPUEnabled)
	viper.Set("memory_enabled", config.MemoryEnabled)
	viper.Set("disk_threshold", config.DiskThreshold)
	viper.Set("cpu_threshold", config.CPUThreshold)
	viper.Set("memory_threshold", config.MemoryThreshold)
	viper.Set("check_interval_minutes", config.CheckInterval)
	viper.Set("disk_check_interval_hours", config.DiskCheckInterval)
	viper.Set("log_level", config.LogLevel)
	viper.Set("service_name", config.ServiceName)

	configFile := filepath.Join(configDir, configFileName+".yaml")
	return viper.WriteConfigAs(configFile)
}

// Validate validates the configuration
func (c *Config) Validate() error {
	if !c.DiskEnabled && !c.CPUEnabled && !c.MemoryEnabled {
		return fmt.Errorf("at least one monitoring option must be enabled")
	}

	if c.DiskEnabled && c.SlackDiskWebhookURL == "" {
		return fmt.Errorf("disk monitoring enabled but no disk webhook URL configured")
	}

	if (c.CPUEnabled || c.MemoryEnabled) && c.SlackCPUMemoryWebhookURL == "" {
		return fmt.Errorf("CPU/Memory monitoring enabled but no CPU/Memory webhook URL configured")
	}

	return nil
}
