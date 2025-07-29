# Modular Notifications Implementation

This document outlines the implementation of the modular notification system for ServerHealth, supporting multiple notification providers including Slack, Telegram, and Discord.

## 🏗️ Architecture Overview

### Core Components

1. **NotificationProvider Interface** - Defines the contract for all notification providers
2. **NotificationManager** - Manages multiple providers and sends notifications concurrently
3. **NotificationMessage** - Structured message format for all notifications
4. **Provider Implementations** - Specific implementations for each platform

### Design Principles

- **Extensible**: Easy to add new notification providers
- **Concurrent**: Multiple providers can send notifications simultaneously
- **Reliable**: Built-in retry logic and error handling
- **Structured**: Consistent message format across all providers

## 🔧 Implementation Details

### NotificationProvider Interface

```go
type NotificationProvider interface {
    Send(ctx context.Context, message *NotificationMessage) error
    Validate() error
    GetType() string
}
```

### NotificationMessage Structure

```go
type NotificationMessage struct {
    Type      string    `json:"type"`
    Level     string    `json:"level"`
    Title     string    `json:"title"`
    Message   string    `json:"message"`
    Hostname  string    `json:"hostname"`
    IP        string    `json:"ip"`
    Timestamp time.Time `json:"timestamp"`
    Metric    string    `json:"metric"`
    Value     string    `json:"value"`
    Threshold string    `json:"threshold"`
}
```

## 📱 Supported Providers

### 1. Slack Integration

**Configuration:**

```yaml
notifications:
  - type: slack
    enabled: true
    webhook_url: "https://hooks.slack.com/services/YOUR/WEBHOOK"
```

**Features:**

- Rich text formatting with emojis
- Automatic retry logic (3 attempts)
- 30-second timeout
- Connection pooling
- Webhook URL validation

**Message Format:**

```json
{
  "text": "🚨 Disk Usage Alert",
  "attachments": [
    {
      "color": "#ff0000",
      "fields": [
        { "title": "Hostname", "value": "server-01", "short": true },
        { "title": "IP", "value": "192.168.1.100", "short": true },
        { "title": "Metric", "value": "Disk Usage", "short": true },
        { "title": "Value", "value": "95.2%", "short": true },
        { "title": "Threshold", "value": "80%", "short": true }
      ]
    }
  ]
}
```

### 2. Telegram Integration

**Configuration:**

```yaml
notifications:
  - type: telegram
    enabled: true
    bot_token: "YOUR_BOT_TOKEN"
    chat_id: "YOUR_CHAT_ID"
```

**Features:**

- Markdown formatting support
- Bot token validation
- Chat ID validation
- Automatic message formatting
- Retry logic with exponential backoff

**Message Format:**

```
🚨 *Disk Usage Alert*

*Hostname:* server-01
*IP:* 192.168.1.100
*Metric:* Disk Usage
*Value:* 95.2%
*Threshold:* 80%
*Time:* 2024-01-15 14:30:00
```

### 3. Discord Integration

**Configuration:**

```yaml
notifications:
  - type: discord
    enabled: true
    webhook_url: "https://discord.com/api/webhooks/YOUR/WEBHOOK"
```

**Features:**

- Rich embed messages
- Color-coded alerts (red for error, yellow for warning)
- Structured field display
- Webhook URL validation
- Automatic retry logic

**Message Format:**

```json
{
  "embeds": [
    {
      "title": "🚨 Disk Usage Alert",
      "color": 15158332,
      "fields": [
        { "name": "Hostname", "value": "server-01", "inline": true },
        { "name": "IP", "value": "192.168.1.100", "inline": true },
        { "name": "Metric", "value": "Disk Usage", "inline": true },
        { "name": "Value", "value": "95.2%", "inline": true },
        { "name": "Threshold", "value": "80%", "inline": true }
      ],
      "timestamp": "2024-01-15T14:30:00Z"
    }
  ]
}
```

## 🔄 NotificationManager

### Concurrent Processing

The NotificationManager processes notifications concurrently:

```go
func (nm *NotificationManager) Send(ctx context.Context, message *NotificationMessage) {
    var wg sync.WaitGroup

    for _, provider := range nm.providers {
        wg.Add(1)
        go func(p NotificationProvider) {
            defer wg.Done()
            if err := p.Send(ctx, message); err != nil {
                // Log error but don't fail other providers
                log.Printf("Failed to send notification via %s: %v", p.GetType(), err)
            }
        }(provider)
    }

    wg.Wait()
}
```

### Provider Management

```go
func (nm *NotificationManager) AddProvider(provider NotificationProvider) error {
    if err := provider.Validate(); err != nil {
        return fmt.Errorf("invalid provider %s: %w", provider.GetType(), err)
    }
    nm.providers = append(nm.providers, provider)
    return nil
}
```

## 🛡️ Error Handling & Reliability

### Retry Logic

All providers implement retry logic:

```go
const (
    maxRetries = 3
    retryDelay = 5 * time.Second
)

func sendWithRetry(client *http.Client, url string, payload interface{}) error {
    for attempt := 1; attempt <= maxRetries; attempt++ {
        if err := sendRequest(client, url, payload); err == nil {
            return nil
        }

        if attempt < maxRetries {
            time.Sleep(retryDelay)
        }
    }
    return fmt.Errorf("failed after %d attempts", maxRetries)
}
```

### Validation

Each provider validates its configuration:

```go
func (p *SlackProvider) Validate() error {
    if p.WebhookURL == "" {
        return errors.New("webhook URL is required")
    }
    if !strings.HasPrefix(p.WebhookURL, "https://hooks.slack.com/") {
        return errors.New("invalid Slack webhook URL")
    }
    return nil
}
```

## 📊 Configuration Examples

### Single Provider Setup

```yaml
# Slack only
notifications:
  - type: slack
    enabled: true
    webhook_url: "https://hooks.slack.com/services/YOUR/WEBHOOK"
```

### Multiple Providers Setup

```yaml
# Multiple providers
notifications:
  - type: slack
    enabled: true
    webhook_url: "https://hooks.slack.com/services/YOUR/WEBHOOK"

  - type: telegram
    enabled: true
    bot_token: "YOUR_BOT_TOKEN"
    chat_id: "YOUR_CHAT_ID"

  - type: discord
    enabled: true
    webhook_url: "https://discord.com/api/webhooks/YOUR/WEBHOOK"
```

### Mixed Configuration

```yaml
# Some enabled, some disabled
notifications:
  - type: slack
    enabled: true
    webhook_url: "https://hooks.slack.com/services/YOUR/WEBHOOK"

  - type: telegram
    enabled: false # Disabled but configured
    bot_token: "YOUR_BOT_TOKEN"
    chat_id: "YOUR_CHAT_ID"

  - type: discord
    enabled: true
    webhook_url: "https://discord.com/api/webhooks/YOUR/WEBHOOK"
```

## 🔧 HTTP Client Configuration

### Shared HTTP Client

All providers use a shared, optimized HTTP client:

```go
var httpClient = &http.Client{
    Timeout: 30 * time.Second,
    Transport: &http.Transport{
        MaxIdleConns:        10,
        MaxIdleConnsPerHost: 5,
        IdleConnTimeout:     90 * time.Second,
    },
}
```

### Request Headers

```go
req.Header.Set("Content-Type", "application/json")
req.Header.Set("User-Agent", "ServerHealth/1.0")
```

## 📈 Performance Characteristics

### Resource Usage

- **Memory**: ~2MB per provider
- **CPU**: Minimal overhead
- **Network**: Only during notification sends
- **Concurrency**: Up to 3 providers simultaneously

### Timeout Handling

- **Request Timeout**: 30 seconds
- **Retry Delay**: 5 seconds between attempts
- **Max Retries**: 3 attempts per notification
- **Connection Pooling**: 10 idle connections, 5 per host

## 🧪 Testing

### Unit Tests

```bash
# Run notification tests
go test -v ./notifications

# Run with coverage
go test -cover ./notifications
```

### Integration Tests

```bash
# Test with real webhooks (replace with actual URLs)
export SLACK_WEBHOOK="https://hooks.slack.com/services/YOUR/WEBHOOK"
export TELEGRAM_BOT_TOKEN="YOUR_BOT_TOKEN"
export TELEGRAM_CHAT_ID="YOUR_CHAT_ID"
export DISCORD_WEBHOOK="https://discord.com/api/webhooks/YOUR/WEBHOOK"

go test -v -tags=integration ./notifications
```

## 🔮 Future Enhancements

### Planned Providers

1. **Email Provider**

   - SMTP support
   - HTML email templates
   - Multiple recipients

2. **PagerDuty Provider**

   - Incident creation
   - Escalation policies
   - Incident management

3. **OpsGenie Provider**
   - Alert creation
   - Team routing
   - Escalation support

### Advanced Features

1. **Template System**

   - Custom message templates
   - Variable substitution
   - Provider-specific formatting

2. **Filtering System**

   - Alert filtering by severity
   - Time-based filtering
   - Custom filter rules

3. **Rate Limiting**
   - Per-provider rate limits
   - Global rate limiting
   - Burst protection

## 📝 Migration Guide

### From Legacy Configuration

The system automatically migrates legacy configurations:

```yaml
# Old format (automatically migrated)
slack_webhook_url: "https://hooks.slack.com/services/YOUR/WEBHOOK"

# New format
notifications:
  - type: slack
    enabled: true
    webhook_url: "https://hooks.slack.com/services/YOUR/WEBHOOK"
```

### Configuration Validation

```bash
# Validate configuration
./serverhealth status

# Interactive configuration
./serverhealth configure
```

This modular notification system provides a robust, extensible foundation for alerting with support for multiple providers, concurrent processing, and comprehensive error handling.
