package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"
)

const (
	maxNotificationsPerDay = 5
	resetNotificationHour  = 0
)

// Monitor represents the main monitoring service
type Monitor struct {
	config                  *Config
	cpuNotificationCount    int
	memoryNotificationCount int
	diskNotificationCount   int
	logger                  *log.Logger
	ctx                     context.Context
	cancel                  context.CancelFunc
}

// Notification struct for Slack messages
type Notification struct {
	Text string `json:"text"`
}

// NewMonitor creates a new monitoring instance
func NewMonitor(config *Config) *Monitor {
	ctx, cancel := context.WithCancel(context.Background())
	logger := log.New(os.Stdout, "", log.LstdFlags)

	return &Monitor{
		config: config,
		logger: logger,
		ctx:    ctx,
		cancel: cancel,
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
	if m.config.DiskEnabled {
		go m.monitorDiskUsage(hostname, ip)
	}

	if m.config.CPUEnabled || m.config.MemoryEnabled {
		go m.monitorCPUAndMemory(hostname, ip)
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
				m.cpuNotificationCount = 0
				m.memoryNotificationCount = 0
				m.diskNotificationCount = 0
				m.logger.Println("Notification counts reset.")
			}
		case <-m.ctx.Done():
			return
		}
	}
}

// monitorDiskUsage monitors disk usage
func (m *Monitor) monitorDiskUsage(hostname, ip string) {
	ticker := time.NewTicker(time.Duration(m.config.DiskCheckInterval) * time.Hour)
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

// monitorCPUAndMemory monitors CPU and memory usage
func (m *Monitor) monitorCPUAndMemory(hostname, ip string) {
	ticker := time.NewTicker(time.Duration(m.config.CheckInterval) * time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			if m.config.CPUEnabled {
				m.checkCPUUsage(hostname, ip)
			}
			if m.config.MemoryEnabled {
				m.checkMemoryUsage(hostname, ip)
			}
		case <-m.ctx.Done():
			return
		}
	}
}

// checkDiskUsage checks disk usage and sends alerts if needed
func (m *Monitor) checkDiskUsage(hostname, ip string) {
	if m.diskNotificationCount >= maxNotificationsPerDay {
		return
	}

	usage, err := GetDiskUsage()
	if err != nil {
		m.logger.Printf("Error checking disk usage: %v", err)
		return
	}

	if usage >= m.config.DiskThreshold {
		emoji := ":warning:"
		if usage >= 95 {
			emoji = ":x:"
		}
		message := fmt.Sprintf("%s Server: %s (%s)\nDisk usage alert: %d%% used.", emoji, hostname, ip, usage)
		m.sendSlackNotification(m.config.SlackDiskWebhookURL, message)
		m.diskNotificationCount++
	}
}

// checkCPUUsage checks CPU usage and sends alerts if needed
func (m *Monitor) checkCPUUsage(hostname, ip string) {
	if m.cpuNotificationCount >= maxNotificationsPerDay {
		return
	}

	usage, err := GetCPUUsage()
	if err != nil {
		m.logger.Printf("Error checking CPU usage: %v", err)
		return
	}

	if usage >= float64(m.config.CPUThreshold) {
		emoji := ":warning:"
		if usage >= 95 {
			emoji = ":x:"
		}
		message := fmt.Sprintf("%s Server: %s (%s)\nCPU usage alert: %.2f%% used.", emoji, hostname, ip, usage)
		m.sendSlackNotification(m.config.SlackCPUMemoryWebhookURL, message)
		m.cpuNotificationCount++
	}
}

// checkMemoryUsage checks memory usage and sends alerts if needed
func (m *Monitor) checkMemoryUsage(hostname, ip string) {
	if m.memoryNotificationCount >= maxNotificationsPerDay {
		return
	}

	usage, err := GetMemoryUsage()
	if err != nil {
		m.logger.Printf("Error checking memory usage: %v", err)
		return
	}

	if usage >= float64(m.config.MemoryThreshold) {
		emoji := ":warning:"
		if usage >= 95 {
			emoji = ":x:"
		}
		message := fmt.Sprintf("%s Server: %s (%s)\nMemory usage alert: %.2f%% used.", emoji, hostname, ip, usage)
		m.sendSlackNotification(m.config.SlackCPUMemoryWebhookURL, message)
		m.memoryNotificationCount++
	}
}

// sendSlackNotification sends a notification to Slack
func (m *Monitor) sendSlackNotification(url, message string) {
	if url == "" {
		return
	}

	payload := Notification{Text: message}
	jsonData, err := json.Marshal(payload)
	if err != nil {
		m.logger.Printf("Error marshaling JSON: %v", err)
		return
	}

	resp, err := http.Post(url, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		m.logger.Printf("Error sending Slack notification: %v", err)
		return
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			m.logger.Printf("Error closing response body: %v", err)
		}
	}()

	if resp.StatusCode != http.StatusOK {
		m.logger.Printf("Non-OK HTTP status: %d", resp.StatusCode)
	} else {
		m.logger.Println("Slack notification sent successfully")
	}
}
