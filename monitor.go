package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"
)

const (
	resetNotificationHour = 0
)

// Monitor represents the main monitoring service
type Monitor struct {
	config              *Config
	notificationManager *NotificationManager
	notificationCounts  map[string]int // Track counts per metric
	logger              *log.Logger
	ctx                 context.Context
	cancel              context.CancelFunc
}

// NewMonitor creates a new monitoring instance
func NewMonitor(config *Config) *Monitor {
	ctx, cancel := context.WithCancel(context.Background())
	logger := log.New(os.Stdout, "", log.LstdFlags)

	// Create notification manager
	notificationManager := NewNotificationManager(logger)

	// Add notification providers based on configuration
	for _, notification := range config.GetEnabledNotifications() {
		var provider NotificationProvider

		switch notification.Type {
		case "slack":
			provider = NewSlackProvider(notification.WebhookURL, notificationManager.client)
		case "telegram":
			provider = NewTelegramProvider(notification.BotToken, notification.ChatID, notificationManager.client)
		case "discord":
			provider = NewDiscordProvider(notification.WebhookURL, notificationManager.client)
		default:
			logger.Printf("Unknown notification type: %s", notification.Type)
			continue
		}

		if err := notificationManager.AddProvider(provider); err != nil {
			logger.Printf("Failed to add notification provider %s: %v", notification.Type, err)
		} else {
			logger.Printf("Added notification provider: %s", notification.Type)
		}
	}

	return &Monitor{
		config:              config,
		notificationManager: notificationManager,
		notificationCounts:  make(map[string]int),
		logger:              logger,
		ctx:                 ctx,
		cancel:              cancel,
	}
}

// SetLogger sets a custom logger for the monitor
func (m *Monitor) SetLogger(logger *log.Logger) {
	m.logger = logger
}

// Start begins the monitoring process
func (m *Monitor) Start() {
	m.logger.Println("Starting monitoring service...")

	hostname, ip := GetServerInfo()
	m.logger.Printf("Monitoring server: %s (%s)", hostname, ip)

	// Start notification count reset routine
	go m.resetNotificationCounts()

	// Start monitoring routines based on configuration
	if m.config.Disk.Enabled {
		go m.monitorDiskUsage(hostname, ip)
	}

	if m.config.CPU.Enabled {
		go m.monitorCPUUsage(hostname, ip)
	}

	if m.config.Memory.Enabled {
		go m.monitorMemoryUsage(hostname, ip)
	}

	// Keep the service running
	<-m.ctx.Done()
	m.logger.Println("Monitoring service stopped.")
}

// Stop terminates the monitoring process
func (m *Monitor) Stop() {
	m.cancel()
}

// resetNotificationCounts resets notification counts daily
func (m *Monitor) resetNotificationCounts() {
	ticker := time.NewTicker(1 * time.Hour)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			now := time.Now()
			if now.Hour() == resetNotificationHour {
				m.notificationCounts = make(map[string]int)
				m.logger.Println("Notification counts reset.")
			}
		case <-m.ctx.Done():
			return
		}
	}
}

// monitorDiskUsage monitors disk usage
func (m *Monitor) monitorDiskUsage(hostname, ip string) {
	ticker := time.NewTicker(time.Duration(m.config.Disk.CheckInterval) * time.Hour)
	defer ticker.Stop()

	// Check immediately on start
	m.checkDiskUsage(hostname, ip)

	for {
		select {
		case <-ticker.C:
			m.checkDiskUsage(hostname, ip)
		case <-m.ctx.Done():
			return
		}
	}
}

// monitorCPUUsage monitors CPU usage
func (m *Monitor) monitorCPUUsage(hostname, ip string) {
	ticker := time.NewTicker(time.Duration(m.config.CPU.CheckInterval) * time.Minute)
	defer ticker.Stop()

	// Check immediately on start
	m.checkCPUUsage(hostname, ip)

	for {
		select {
		case <-ticker.C:
			m.checkCPUUsage(hostname, ip)
		case <-m.ctx.Done():
			return
		}
	}
}

// monitorMemoryUsage monitors memory usage
func (m *Monitor) monitorMemoryUsage(hostname, ip string) {
	ticker := time.NewTicker(time.Duration(m.config.Memory.CheckInterval) * time.Minute)
	defer ticker.Stop()

	// Check immediately on start
	m.checkMemoryUsage(hostname, ip)

	for {
		select {
		case <-ticker.C:
			m.checkMemoryUsage(hostname, ip)
		case <-m.ctx.Done():
			return
		}
	}
}

// checkDiskUsage checks disk usage and sends alerts if needed
func (m *Monitor) checkDiskUsage(hostname, ip string) {
	metricKey := "disk"
	if m.notificationCounts[metricKey] >= m.config.Disk.MaxDailyAlerts {
		return
	}

	usage, err := GetDiskUsage()
	if err != nil {
		m.logger.Printf("Error checking disk usage: %v", err)
		return
	}

	if usage >= m.config.Disk.Threshold {
		level := NotificationLevelWarning
		if usage >= 95 {
			level = NotificationLevelError
		}

		message := &NotificationMessage{
			Type:      NotificationTypeSlack, // Will be overridden by providers
			Level:     level,
			Title:     "Disk Usage Alert",
			Message:   fmt.Sprintf("Disk usage has exceeded the threshold of %d%%", m.config.Disk.Threshold),
			Hostname:  hostname,
			IP:        ip,
			Timestamp: time.Now(),
			Metric:    "Disk Usage",
			Value:     fmt.Sprintf("%d%%", usage),
			Threshold: fmt.Sprintf("%d%%", m.config.Disk.Threshold),
		}

		m.notificationManager.Send(m.ctx, message)
		m.notificationCounts[metricKey]++
	}
}

// checkCPUUsage checks CPU usage and sends alerts if needed
func (m *Monitor) checkCPUUsage(hostname, ip string) {
	metricKey := "cpu"
	if m.notificationCounts[metricKey] >= m.config.CPU.MaxDailyAlerts {
		return
	}

	usage, err := GetCPUUsage()
	if err != nil {
		m.logger.Printf("Error checking CPU usage: %v", err)
		return
	}

	if usage >= float64(m.config.CPU.Threshold) {
		level := NotificationLevelWarning
		if usage >= 95 {
			level = NotificationLevelError
		}

		message := &NotificationMessage{
			Type:      NotificationTypeSlack, // Will be overridden by providers
			Level:     level,
			Title:     "CPU Usage Alert",
			Message:   fmt.Sprintf("CPU usage has exceeded the threshold of %d%%", m.config.CPU.Threshold),
			Hostname:  hostname,
			IP:        ip,
			Timestamp: time.Now(),
			Metric:    "CPU Usage",
			Value:     fmt.Sprintf("%.2f%%", usage),
			Threshold: fmt.Sprintf("%d%%", m.config.CPU.Threshold),
		}

		m.notificationManager.Send(m.ctx, message)
		m.notificationCounts[metricKey]++
	}
}

// checkMemoryUsage checks memory usage and sends alerts if needed
func (m *Monitor) checkMemoryUsage(hostname, ip string) {
	metricKey := "memory"
	if m.notificationCounts[metricKey] >= m.config.Memory.MaxDailyAlerts {
		return
	}

	usage, err := GetMemoryUsage()
	if err != nil {
		m.logger.Printf("Error checking memory usage: %v", err)
		return
	}

	if usage >= float64(m.config.Memory.Threshold) {
		level := NotificationLevelWarning
		if usage >= 95 {
			level = NotificationLevelError
		}

		message := &NotificationMessage{
			Type:      NotificationTypeSlack, // Will be overridden by providers
			Level:     level,
			Title:     "Memory Usage Alert",
			Message:   fmt.Sprintf("Memory usage has exceeded the threshold of %d%%", m.config.Memory.Threshold),
			Hostname:  hostname,
			IP:        ip,
			Timestamp: time.Now(),
			Metric:    "Memory Usage",
			Value:     fmt.Sprintf("%.2f%%", usage),
			Threshold: fmt.Sprintf("%d%%", m.config.Memory.Threshold),
		}

		m.notificationManager.Send(m.ctx, message)
		m.notificationCounts[metricKey]++
	}
}
