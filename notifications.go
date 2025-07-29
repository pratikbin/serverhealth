package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// NotificationType represents the type of notification
type NotificationType string

const (
	NotificationTypeSlack    NotificationType = "slack"
	NotificationTypeTelegram NotificationType = "telegram"
	NotificationTypeDiscord  NotificationType = "discord"
)

// NotificationLevel represents the severity level of a notification
type NotificationLevel string

const (
	NotificationLevelInfo    NotificationLevel = "info"
	NotificationLevelWarning NotificationLevel = "warning"
	NotificationLevelError   NotificationLevel = "error"
)

// NotificationMessage represents a notification message
type NotificationMessage struct {
	Type      NotificationType  `json:"type"`
	Level     NotificationLevel `json:"level"`
	Title     string            `json:"title"`
	Message   string            `json:"message"`
	Hostname  string            `json:"hostname"`
	IP        string            `json:"ip"`
	Timestamp time.Time         `json:"timestamp"`
	Metric    string            `json:"metric,omitempty"`
	Value     interface{}       `json:"value,omitempty"`
	Threshold interface{}       `json:"threshold,omitempty"`
}

// NotificationProvider defines the interface for notification providers
type NotificationProvider interface {
	Send(ctx context.Context, message *NotificationMessage) error
	Validate() error
	GetType() NotificationType
}

// NotificationManager manages multiple notification providers
type NotificationManager struct {
	providers []NotificationProvider
	logger    *log.Logger
	client    *http.Client
}

// NewNotificationManager creates a new notification manager
func NewNotificationManager(logger *log.Logger) *NotificationManager {
	client := &http.Client{
		Timeout: 30 * time.Second,
		Transport: &http.Transport{
			MaxIdleConns:        10,
			MaxIdleConnsPerHost: 5,
			IdleConnTimeout:     90 * time.Second,
		},
	}

	return &NotificationManager{
		providers: make([]NotificationProvider, 0),
		logger:    logger,
		client:    client,
	}
}

// AddProvider adds a notification provider
func (nm *NotificationManager) AddProvider(provider NotificationProvider) error {
	if err := provider.Validate(); err != nil {
		return fmt.Errorf("invalid provider %s: %w", provider.GetType(), err)
	}
	nm.providers = append(nm.providers, provider)
	return nil
}

// Send sends a notification to all providers
func (nm *NotificationManager) Send(ctx context.Context, message *NotificationMessage) {
	if len(nm.providers) == 0 {
		nm.logger.Println("No notification providers configured")
		return
	}

	// Set timestamp if not set
	if message.Timestamp.IsZero() {
		message.Timestamp = time.Now()
	}

	for _, provider := range nm.providers {
		go func(p NotificationProvider) {
			if err := p.Send(ctx, message); err != nil {
				nm.logger.Printf("Failed to send notification via %s: %v", p.GetType(), err)
			} else {
				nm.logger.Printf("Notification sent successfully via %s", p.GetType())
			}
		}(provider)
	}
}

// SlackProvider implements Slack notifications
type SlackProvider struct {
	WebhookURL string
	client     *http.Client
}

// NewSlackProvider creates a new Slack notification provider
func NewSlackProvider(webhookURL string, client *http.Client) *SlackProvider {
	return &SlackProvider{
		WebhookURL: webhookURL,
		client:     client,
	}
}

// GetType returns the provider type
func (sp *SlackProvider) GetType() NotificationType {
	return NotificationTypeSlack
}

// Validate validates the Slack provider configuration
func (sp *SlackProvider) Validate() error {
	if sp.WebhookURL == "" {
		return fmt.Errorf("webhook URL is required")
	}

	parsedURL, err := url.Parse(sp.WebhookURL)
	if err != nil {
		return fmt.Errorf("invalid URL format: %w", err)
	}

	if parsedURL.Scheme != "https" {
		return fmt.Errorf("webhook URL must use HTTPS")
	}

	if !strings.Contains(parsedURL.Host, "hooks.slack.com") {
		return fmt.Errorf("webhook URL must be from hooks.slack.com")
	}

	return nil
}

// Send sends a notification to Slack
func (sp *SlackProvider) Send(ctx context.Context, message *NotificationMessage) error {
	// Create Slack message
	emoji := ":information_source:"
	switch message.Level {
	case NotificationLevelWarning:
		emoji = ":warning:"
	case NotificationLevelError:
		emoji = ":x:"
	}

	slackText := fmt.Sprintf("%s *%s*\n%s\n\n*Server:* %s (%s)\n*Time:* %s",
		emoji, message.Title, message.Message, message.Hostname, message.IP,
		message.Timestamp.Format("2006-01-02 15:04:05"))

	if message.Metric != "" && message.Value != nil {
		slackText += fmt.Sprintf("\n*Metric:* %s = %v", message.Metric, message.Value)
		if message.Threshold != nil {
			slackText += fmt.Sprintf(" (threshold: %v)", message.Threshold)
		}
	}

	payload := map[string]string{
		"text": slackText,
	}

	return sp.sendHTTPRequest(ctx, payload)
}

// sendHTTPRequest sends an HTTP request with retry logic
func (sp *SlackProvider) sendHTTPRequest(ctx context.Context, payload interface{}) error {
	jsonData, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal JSON: %w", err)
	}

	const maxRetries = 3
	const retryDelay = 5 * time.Second

	for attempt := 1; attempt <= maxRetries; attempt++ {
		req, err := http.NewRequestWithContext(ctx, "POST", sp.WebhookURL, bytes.NewBuffer(jsonData))
		if err != nil {
			return fmt.Errorf("failed to create request: %w", err)
		}

		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("User-Agent", "ServerHealth/1.0")

		resp, err := sp.client.Do(req)
		if err != nil {
			if attempt < maxRetries {
				time.Sleep(retryDelay)
				continue
			}
			return fmt.Errorf("failed to send request: %w", err)
		}

		defer resp.Body.Close()

		if resp.StatusCode == http.StatusOK {
			return nil
		}

		if attempt < maxRetries {
			time.Sleep(retryDelay)
		}
	}

	return fmt.Errorf("failed to send notification after %d attempts", maxRetries)
}

// TelegramProvider implements Telegram notifications
type TelegramProvider struct {
	BotToken string
	ChatID   string
	client   *http.Client
}

// NewTelegramProvider creates a new Telegram notification provider
func NewTelegramProvider(botToken, chatID string, client *http.Client) *TelegramProvider {
	return &TelegramProvider{
		BotToken: botToken,
		ChatID:   chatID,
		client:   client,
	}
}

// GetType returns the provider type
func (tp *TelegramProvider) GetType() NotificationType {
	return NotificationTypeTelegram
}

// Validate validates the Telegram provider configuration
func (tp *TelegramProvider) Validate() error {
	if tp.BotToken == "" {
		return fmt.Errorf("bot token is required")
	}
	if tp.ChatID == "" {
		return fmt.Errorf("chat ID is required")
	}
	return nil
}

// Send sends a notification to Telegram
func (tp *TelegramProvider) Send(ctx context.Context, message *NotificationMessage) error {
	// Create Telegram message
	emoji := "ℹ️"
	switch message.Level {
	case NotificationLevelWarning:
		emoji = "⚠️"
	case NotificationLevelError:
		emoji = "❌"
	}

	telegramText := fmt.Sprintf("%s *%s*\n\n%s\n\n*Server:* %s (%s)\n*Time:* %s",
		emoji, message.Title, message.Message, message.Hostname, message.IP,
		message.Timestamp.Format("2006-01-02 15:04:05"))

	if message.Metric != "" && message.Value != nil {
		telegramText += fmt.Sprintf("\n*Metric:* %s = %v", message.Metric, message.Value)
		if message.Threshold != nil {
			telegramText += fmt.Sprintf(" (threshold: %v)", message.Threshold)
		}
	}

	// Telegram API endpoint
	apiURL := fmt.Sprintf("https://api.telegram.org/bot%s/sendMessage", tp.BotToken)

	payload := map[string]interface{}{
		"chat_id":    tp.ChatID,
		"text":       telegramText,
		"parse_mode": "Markdown",
	}

	return tp.sendHTTPRequest(ctx, apiURL, payload)
}

// sendHTTPRequest sends an HTTP request with retry logic
func (tp *TelegramProvider) sendHTTPRequest(ctx context.Context, apiURL string, payload interface{}) error {
	jsonData, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal JSON: %w", err)
	}

	const maxRetries = 3
	const retryDelay = 5 * time.Second

	for attempt := 1; attempt <= maxRetries; attempt++ {
		req, err := http.NewRequestWithContext(ctx, "POST", apiURL, bytes.NewBuffer(jsonData))
		if err != nil {
			return fmt.Errorf("failed to create request: %w", err)
		}

		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("User-Agent", "ServerHealth/1.0")

		resp, err := tp.client.Do(req)
		if err != nil {
			if attempt < maxRetries {
				time.Sleep(retryDelay)
				continue
			}
			return fmt.Errorf("failed to send request: %w", err)
		}

		defer resp.Body.Close()

		if resp.StatusCode == http.StatusOK {
			return nil
		}

		if attempt < maxRetries {
			time.Sleep(retryDelay)
		}
	}

	return fmt.Errorf("failed to send notification after %d attempts", maxRetries)
}

// DiscordProvider implements Discord notifications
type DiscordProvider struct {
	WebhookURL string
	client     *http.Client
}

// NewDiscordProvider creates a new Discord notification provider
func NewDiscordProvider(webhookURL string, client *http.Client) *DiscordProvider {
	return &DiscordProvider{
		WebhookURL: webhookURL,
		client:     client,
	}
}

// GetType returns the provider type
func (dp *DiscordProvider) GetType() NotificationType {
	return NotificationTypeDiscord
}

// Validate validates the Discord provider configuration
func (dp *DiscordProvider) Validate() error {
	if dp.WebhookURL == "" {
		return fmt.Errorf("webhook URL is required")
	}

	parsedURL, err := url.Parse(dp.WebhookURL)
	if err != nil {
		return fmt.Errorf("invalid URL format: %w", err)
	}

	if parsedURL.Scheme != "https" {
		return fmt.Errorf("webhook URL must use HTTPS")
	}

	if !strings.Contains(parsedURL.Host, "discord.com") && !strings.Contains(parsedURL.Host, "discordapp.com") {
		return fmt.Errorf("webhook URL must be from discord.com or discordapp.com")
	}

	return nil
}

// Send sends a notification to Discord
func (dp *DiscordProvider) Send(ctx context.Context, message *NotificationMessage) error {
	// Create Discord embed
	color := 0x00ff00 // Green for info
	switch message.Level {
	case NotificationLevelWarning:
		color = 0xffff00 // Yellow for warning
	case NotificationLevelError:
		color = 0xff0000 // Red for error
	}

	// Create embed fields
	fields := []map[string]interface{}{
		{
			"name":   "Server",
			"value":  fmt.Sprintf("%s (%s)", message.Hostname, message.IP),
			"inline": true,
		},
		{
			"name":   "Time",
			"value":  message.Timestamp.Format("2006-01-02 15:04:05"),
			"inline": true,
		},
	}

	if message.Metric != "" && message.Value != nil {
		fieldValue := fmt.Sprintf("%v", message.Value)
		if message.Threshold != nil {
			fieldValue += fmt.Sprintf(" (threshold: %v)", message.Threshold)
		}
		fields = append(fields, map[string]interface{}{
			"name":   "Metric",
			"value":  fieldValue,
			"inline": true,
		})
	}

	embed := map[string]interface{}{
		"title":       message.Title,
		"description": message.Message,
		"color":       color,
		"fields":      fields,
		"timestamp":   message.Timestamp.Format(time.RFC3339),
	}

	payload := map[string]interface{}{
		"embeds": []map[string]interface{}{embed},
	}

	return dp.sendHTTPRequest(ctx, payload)
}

// sendHTTPRequest sends an HTTP request with retry logic
func (dp *DiscordProvider) sendHTTPRequest(ctx context.Context, payload interface{}) error {
	jsonData, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal JSON: %w", err)
	}

	const maxRetries = 3
	const retryDelay = 5 * time.Second

	for attempt := 1; attempt <= maxRetries; attempt++ {
		req, err := http.NewRequestWithContext(ctx, "POST", dp.WebhookURL, bytes.NewBuffer(jsonData))
		if err != nil {
			return fmt.Errorf("failed to create request: %w", err)
		}

		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("User-Agent", "ServerHealth/1.0")

		resp, err := dp.client.Do(req)
		if err != nil {
			if attempt < maxRetries {
				time.Sleep(retryDelay)
				continue
			}
			return fmt.Errorf("failed to send request: %w", err)
		}

		defer resp.Body.Close()

		if resp.StatusCode == http.StatusOK {
			return nil
		}

		if attempt < maxRetries {
			time.Sleep(retryDelay)
		}
	}

	return fmt.Errorf("failed to send notification after %d attempts", maxRetries)
}
