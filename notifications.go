package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"
)

// NotificationType represents the type of notification provider
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
	Metric    string            `json:"metric"`
	Value     string            `json:"value"`
	Threshold string            `json:"threshold"`
}

// NotificationProvider interface defines methods for notification providers
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

// AddProvider adds a notification provider to the manager
func (nm *NotificationManager) AddProvider(provider NotificationProvider) error {
	if err := provider.Validate(); err != nil {
		return fmt.Errorf("invalid provider %s: %w", provider.GetType(), err)
	}
	nm.providers = append(nm.providers, provider)
	return nil
}

// Send sends a notification message to all enabled providers concurrently
func (nm *NotificationManager) Send(ctx context.Context, message *NotificationMessage) {
	if len(nm.providers) == 0 {
		nm.logger.Println("No notification providers configured")
		return
	}

	// Send to all providers concurrently
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

// sendHTTPRequest is a shared function for sending HTTP requests with retry logic
func sendHTTPRequest(ctx context.Context, client *http.Client, url string, payload interface{}) error {
	jsonData, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal JSON: %w", err)
	}

	const maxRetries = 3
	const retryDelay = 5 * time.Second

	for attempt := 1; attempt <= maxRetries; attempt++ {
		req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewBuffer(jsonData))
		if err != nil {
			return fmt.Errorf("failed to create request: %w", err)
		}

		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("User-Agent", "ServerHealth/1.0")

		resp, err := client.Do(req)
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

// SlackProvider implements NotificationProvider for Slack
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

// Validate validates the Slack provider configuration
func (sp *SlackProvider) Validate() error {
	if sp.WebhookURL == "" {
		return fmt.Errorf("webhook URL is required")
	}

	if !strings.HasPrefix(sp.WebhookURL, "https://") {
		return fmt.Errorf("webhook URL must use HTTPS")
	}

	if !strings.Contains(sp.WebhookURL, "hooks.slack.com") {
		return fmt.Errorf("webhook URL must be from hooks.slack.com")
	}

	return nil
}

// GetType returns the notification type
func (sp *SlackProvider) GetType() NotificationType {
	return NotificationTypeSlack
}

// Send sends a notification to Slack
func (sp *SlackProvider) Send(ctx context.Context, message *NotificationMessage) error {
	// Determine emoji based on level
	var emoji string
	switch message.Level {
	case NotificationLevelInfo:
		emoji = "ℹ️"
	case NotificationLevelWarning:
		emoji = "⚠️"
	case NotificationLevelError:
		emoji = "❌"
	default:
		emoji = "ℹ️"
	}

	// Create Slack payload
	payload := map[string]interface{}{
		"text": fmt.Sprintf("%s *%s*\n%s\n\n*Server:* %s (%s)\n*Metric:* %s\n*Value:* %s\n*Threshold:* %s\n*Time:* %s",
			emoji, message.Title, message.Message, message.Hostname, message.IP,
			message.Metric, message.Value, message.Threshold, message.Timestamp.Format("2006-01-02 15:04:05")),
	}

	return sendHTTPRequest(ctx, sp.client, sp.WebhookURL, payload)
}

// TelegramProvider implements NotificationProvider for Telegram
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

// GetType returns the notification type
func (tp *TelegramProvider) GetType() NotificationType {
	return NotificationTypeTelegram
}

// Send sends a notification to Telegram
func (tp *TelegramProvider) Send(ctx context.Context, message *NotificationMessage) error {
	// Determine emoji based on level
	var emoji string
	switch message.Level {
	case NotificationLevelInfo:
		emoji = "ℹ️"
	case NotificationLevelWarning:
		emoji = "⚠️"
	case NotificationLevelError:
		emoji = "❌"
	default:
		emoji = "ℹ️"
	}

	// Create Telegram payload
	payload := map[string]interface{}{
		"chat_id": tp.ChatID,
		"text": fmt.Sprintf("%s *%s*\n%s\n\n*Server:* %s (%s)\n*Metric:* %s\n*Value:* %s\n*Threshold:* %s\n*Time:* %s",
			emoji, message.Title, message.Message, message.Hostname, message.IP,
			message.Metric, message.Value, message.Threshold, message.Timestamp.Format("2006-01-02 15:04:05")),
		"parse_mode": "Markdown",
	}

	apiURL := fmt.Sprintf("https://api.telegram.org/bot%s/sendMessage", tp.BotToken)

	return sendHTTPRequest(ctx, tp.client, apiURL, payload)
}

// DiscordProvider implements NotificationProvider for Discord
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

// Validate validates the Discord provider configuration
func (dp *DiscordProvider) Validate() error {
	if dp.WebhookURL == "" {
		return fmt.Errorf("webhook URL is required")
	}

	if !strings.HasPrefix(dp.WebhookURL, "https://") {
		return fmt.Errorf("webhook URL must use HTTPS")
	}

	if !strings.Contains(dp.WebhookURL, "discord.com") && !strings.Contains(dp.WebhookURL, "discordapp.com") {
		return fmt.Errorf("webhook URL must be from discord.com or discordapp.com")
	}

	return nil
}

// GetType returns the notification type
func (dp *DiscordProvider) GetType() NotificationType {
	return NotificationTypeDiscord
}

// Send sends a notification to Discord
func (dp *DiscordProvider) Send(ctx context.Context, message *NotificationMessage) error {
	// Determine color based on level
	var color int
	switch message.Level {
	case NotificationLevelInfo:
		color = 0x3498db // Blue
	case NotificationLevelWarning:
		color = 0xf39c12 // Orange
	case NotificationLevelError:
		color = 0xe74c3c // Red
	default:
		color = 0x3498db // Blue
	}

	// Create Discord embed
	embed := map[string]interface{}{
		"title":       message.Title,
		"description": message.Message,
		"color":       color,
		"fields": []map[string]interface{}{
			{
				"name":   "Server",
				"value":  fmt.Sprintf("%s (%s)", message.Hostname, message.IP),
				"inline": true,
			},
			{
				"name":   "Metric",
				"value":  message.Metric,
				"inline": true,
			},
			{
				"name":   "Value",
				"value":  message.Value,
				"inline": true,
			},
			{
				"name":   "Threshold",
				"value":  message.Threshold,
				"inline": true,
			},
		},
		"timestamp": message.Timestamp.Format(time.RFC3339),
	}

	payload := map[string]interface{}{
		"embeds": []map[string]interface{}{embed},
	}

	return sendHTTPRequest(ctx, dp.client, dp.WebhookURL, payload)
}
