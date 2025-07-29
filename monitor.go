package main

import (
	"context"
	"fmt"
	"log"
	"time"
)

// Monitor represents the monitoring system
type Monitor struct {
	config              *Config
	logger              *log.Logger
	ctx                 context.Context
	cancel              context.CancelFunc
	notificationManager *NotificationManager
	notificationCounts  map[string]int
	lastResetDate       string
}

// NewMonitor creates a new monitor instance
func NewMonitor(config *Config, logger *log.Logger) *Monitor {
	ctx, cancel := context.WithCancel(context.Background())

	notificationManager := NewNotificationManager(logger)

	// Add notification providers based on configuration
	for _, notification := range config.GetEnabledNotifications() {
		var provider NotificationProvider

		switch notification.Type {
		case string(NotificationTypeSlack):
			provider = NewSlackProvider(notification.WebhookURL, notificationManager.client)
		case string(NotificationTypeTelegram):
			provider = NewTelegramProvider(notification.BotToken, notification.ChatID, notificationManager.client)
		case string(NotificationTypeDiscord):
			provider = NewDiscordProvider(notification.WebhookURL, notificationManager.client)
		}

		if provider != nil {
			if err := notificationManager.AddProvider(provider); err != nil {
				logger.Printf("Failed to add notification provider %s: %v", notification.Type, err)
			}
		}
	}

	return &Monitor{
		config:              config,
		logger:              logger,
		ctx:                 ctx,
		cancel:              cancel,
		notificationManager: notificationManager,
		notificationCounts:  make(map[string]int),
		lastResetDate:       time.Now().Format("2006-01-02"),
	}
}

// Start begins the monitoring process
func (m *Monitor) Start() {
	m.logger.Println("Starting ServerHealth monitoring...")

	hostname, serverIP := GetServerInfo()

	// Start monitoring goroutines
	if m.config.Disk.Enabled {
		go m.monitorDiskUsage(hostname, serverIP)
	}

	if m.config.CPU.Enabled {
		go m.monitorCPUUsage(hostname, serverIP)
	}

	if m.config.Memory.Enabled {
		go m.monitorMemoryUsage(hostname, serverIP)
	}

	// Keep the main goroutine alive
	<-m.ctx.Done()
}

// Stop stops the monitoring process
func (m *Monitor) Stop() {
	m.logger.Println("Stopping ServerHealth monitoring...")
	m.cancel()
}

// resetDailyCounts resets notification counts daily
func (m *Monitor) resetDailyCounts() {
	currentDate := time.Now().Format("2006-01-02")
	if m.lastResetDate != currentDate {
		m.notificationCounts = make(map[string]int)
		m.lastResetDate = currentDate
		m.logger.Println("Daily notification counts reset")
	}
}

// getDiskUsageFloat64 wraps GetDiskUsage to return float64
func getDiskUsageFloat64() (float64, error) {
	usage, err := GetDiskUsage()
	return float64(usage), err
}

// checkMetricUsage is a shared function for checking metric usage and sending notifications
func (m *Monitor) checkMetricUsage(metricKey string, config MonitoringConfig, getUsage func() (float64, error), metricName string) {
	// Reset daily counts if needed
	m.resetDailyCounts()

	// Check if we've exceeded daily alert limit
	if m.notificationCounts[metricKey] >= config.MaxDailyAlerts {
		return
	}

	usage, err := getUsage()
	if err != nil {
		m.logger.Printf("Error checking %s usage: %v", metricName, err)
		return
	}

	if usage >= float64(config.Threshold) {
		level := NotificationLevelWarning
		if usage >= 95 {
			level = NotificationLevelError
		}

		hostname, serverIP := GetServerInfo()
		message := &NotificationMessage{
			Type:      NotificationTypeSlack, // Will be overridden by providers
			Level:     level,
			Title:     fmt.Sprintf("%s Usage Alert", metricName),
			Message:   fmt.Sprintf("%s usage has exceeded the threshold of %d%%", metricName, config.Threshold),
			Hostname:  hostname,
			IP:        serverIP,
			Timestamp: time.Now(),
			Metric:    fmt.Sprintf("%s Usage", metricName),
			Value:     fmt.Sprintf("%.2f%%", usage),
			Threshold: fmt.Sprintf("%d%%", config.Threshold),
		}

		m.notificationManager.Send(m.ctx, message)
		m.notificationCounts[metricKey]++
	}
}

// monitorDiskUsage monitors disk usage
func (m *Monitor) monitorDiskUsage(hostname, serverIP string) {
	ticker := time.NewTicker(time.Duration(m.config.Disk.CheckInterval) * time.Hour)
	defer ticker.Stop()

	// Initial check
	m.checkDiskUsage(hostname, serverIP)

	for {
		select {
		case <-ticker.C:
			m.checkDiskUsage(hostname, serverIP)
		case <-m.ctx.Done():
			return
		}
	}
}

// checkDiskUsage checks disk usage and sends notification if threshold is exceeded
func (m *Monitor) checkDiskUsage(hostname, serverIP string) {
	m.checkMetricUsage("disk", m.config.Disk, getDiskUsageFloat64, "Disk")
}

// monitorCPUUsage monitors CPU usage
func (m *Monitor) monitorCPUUsage(hostname, serverIP string) {
	ticker := time.NewTicker(time.Duration(m.config.CPU.CheckInterval) * time.Minute)
	defer ticker.Stop()

	// Initial check
	m.checkCPUUsage(hostname, serverIP)

	for {
		select {
		case <-ticker.C:
			m.checkCPUUsage(hostname, serverIP)
		case <-m.ctx.Done():
			return
		}
	}
}

// checkCPUUsage checks CPU usage and sends notification if threshold is exceeded
func (m *Monitor) checkCPUUsage(hostname, serverIP string) {
	m.checkMetricUsage("cpu", m.config.CPU, GetCPUUsage, "CPU")
}

// monitorMemoryUsage monitors memory usage
func (m *Monitor) monitorMemoryUsage(hostname, serverIP string) {
	ticker := time.NewTicker(time.Duration(m.config.Memory.CheckInterval) * time.Minute)
	defer ticker.Stop()

	// Initial check
	m.checkMemoryUsage(hostname, serverIP)

	for {
		select {
		case <-ticker.C:
			m.checkMemoryUsage(hostname, serverIP)
		case <-m.ctx.Done():
			return
		}
	}
}

// checkMemoryUsage checks memory usage and sends notification if threshold is exceeded
func (m *Monitor) checkMemoryUsage(hostname, serverIP string) {
	m.checkMetricUsage("memory", m.config.Memory, GetMemoryUsage, "Memory")
}
