# Modular Notification System

ServerHealth now supports a modular notification system that allows you to configure multiple notification providers simultaneously. This system supports Slack, Telegram, and Discord notifications with a unified interface.

## 🚀 Features

- **Multiple Providers**: Configure Slack, Telegram, and Discord simultaneously
- **Unified Interface**: All providers use the same notification format
- **Retry Logic**: Automatic retry with exponential backoff
- **Rate Limiting**: Per-metric daily alert limits
- **Rich Notifications**: Structured messages with metadata
- **Environment Variables**: Support for containerized deployments

## 📋 Supported Providers

### 1. Slack Notifications

**Setup:**

1. Go to your Slack workspace
2. Create a new app or use an existing one
3. Enable "Incoming Webhooks"
4. Create a webhook URL
5. Copy the webhook URL

**Configuration:**

```yaml
notifications:
  - type: slack
    enabled: true
    webhook_url: "https://hooks.slack.com/services/YOUR/SLACK/WEBHOOK"
```

**Environment Variable:**

```bash
export SERVERHEALTH_NOTIFICATIONS_0_TYPE=slack
export SERVERHEALTH_NOTIFICATIONS_0_ENABLED=true
export SERVERHEALTH_NOTIFICATIONS_0_WEBHOOK_URL="https://hooks.slack.com/services/YOUR/SLACK/WEBHOOK"
```

### 2. Telegram Notifications

**Setup:**

1. Create a bot with [@BotFather](https://t.me/botfather)
2. Get your bot token
3. Get your chat ID by sending a message to your bot and checking [@userinfobot](https://t.me/userinfobot)

**Configuration:**

```yaml
notifications:
  - type: telegram
    enabled: true
    bot_token: "YOUR_BOT_TOKEN_HERE"
    chat_id: "YOUR_CHAT_ID_HERE"
```

**Environment Variable:**

```bash
export SERVERHEALTH_NOTIFICATIONS_0_TYPE=telegram
export SERVERHEALTH_NOTIFICATIONS_0_ENABLED=true
export SERVERHEALTH_NOTIFICATIONS_0_BOT_TOKEN="YOUR_BOT_TOKEN_HERE"
export SERVERHEALTH_NOTIFICATIONS_0_CHAT_ID="YOUR_CHAT_ID_HERE"
```

### 3. Discord Notifications

**Setup:**

1. Go to your Discord server
2. Edit a channel
3. Go to Integrations > Webhooks
4. Create a new webhook
5. Copy the webhook URL

**Configuration:**

```yaml
notifications:
  - type: discord
    enabled: true
    webhook_url: "https://discord.com/api/webhooks/YOUR/DISCORD/WEBHOOK"
```

**Environment Variable:**

```bash
export SERVERHEALTH_NOTIFICATIONS_0_TYPE=discord
export SERVERHEALTH_NOTIFICATIONS_0_ENABLED=true
export SERVERHEALTH_NOTIFICATIONS_0_WEBHOOK_URL="https://discord.com/api/webhooks/YOUR/DISCORD/WEBHOOK"
```

## 🔧 Configuration Examples

### Multiple Providers

You can configure multiple notification providers simultaneously:

```yaml
notifications:
  # Slack for critical alerts
  - type: slack
    enabled: true
    webhook_url: "https://hooks.slack.com/services/CRITICAL/ALERTS"

  # Telegram for general monitoring
  - type: telegram
    enabled: true
    bot_token: "YOUR_BOT_TOKEN"
    chat_id: "YOUR_CHAT_ID"

  # Discord for team notifications
  - type: discord
    enabled: true
    webhook_url: "https://discord.com/api/webhooks/TEAM/ALERTS"
```

### Environment Variables for Multiple Providers

```bash
# First provider (Slack)
export SERVERHEALTH_NOTIFICATIONS_0_TYPE=slack
export SERVERHEALTH_NOTIFICATIONS_0_ENABLED=true
export SERVERHEALTH_NOTIFICATIONS_0_WEBHOOK_URL="https://hooks.slack.com/services/YOUR/SLACK/WEBHOOK"

# Second provider (Telegram)
export SERVERHEALTH_NOTIFICATIONS_1_TYPE=telegram
export SERVERHEALTH_NOTIFICATIONS_1_ENABLED=true
export SERVERHEALTH_NOTIFICATIONS_1_BOT_TOKEN="YOUR_BOT_TOKEN"
export SERVERHEALTH_NOTIFICATIONS_1_CHAT_ID="YOUR_CHAT_ID"

# Third provider (Discord)
export SERVERHEALTH_NOTIFICATIONS_2_TYPE=discord
export SERVERHEALTH_NOTIFICATIONS_2_ENABLED=true
export SERVERHEALTH_NOTIFICATIONS_2_WEBHOOK_URL="https://discord.com/api/webhooks/YOUR/DISCORD/WEBHOOK"
```

## 📊 Notification Format

All providers send structured notifications with the following information:

### Slack Format

```
⚠️ Disk Usage Alert
Disk usage has exceeded the threshold of 80%

Server: server01 (192.168.1.100)
Time: 2024-01-15 14:30:25
Metric: Disk Usage = 85% (threshold: 80%)
```

### Telegram Format

```
⚠️ Disk Usage Alert

Disk usage has exceeded the threshold of 80%

Server: server01 (192.168.1.100)
Time: 2024-01-15 14:30:25
Metric: Disk Usage = 85% (threshold: 80%)
```

### Discord Format

```
Embed with:
- Title: "Disk Usage Alert"
- Description: "Disk usage has exceeded the threshold of 80%"
- Fields: Server, Time, Metric
- Color: Yellow for warning, Red for error
```

## 🎯 Notification Levels

The system supports three notification levels:

- **Info** (ℹ️): General information
- **Warning** (⚠️): Threshold exceeded but not critical
- **Error** (❌): Critical threshold exceeded (95%+)

## ⚙️ Advanced Configuration

### Per-Metric Alert Limits

You can configure different alert limits for each metric:

```yaml
disk:
  enabled: true
  threshold: 80
  check_interval: 12 # hours
  max_daily_alerts: 3 # Only 3 disk alerts per day

cpu:
  enabled: true
  threshold: 85
  check_interval: 60 # minutes
  max_daily_alerts: 10 # Up to 10 CPU alerts per day

memory:
  enabled: true
  threshold: 85
  check_interval: 60 # minutes
  max_daily_alerts: 5 # 5 memory alerts per day
```

### Retry Logic

All providers include automatic retry logic:

- **Max Retries**: 3 attempts
- **Retry Delay**: 5 seconds between attempts
- **Timeout**: 30 seconds per request
- **Connection Pooling**: 10 max idle connections, 5 per host

## 🔍 Troubleshooting

### Common Issues

1. **Slack Webhook Not Working**

   ```
   Error: webhook URL must be from hooks.slack.com
   ```

   **Solution**: Ensure the webhook URL starts with `https://hooks.slack.com/`

2. **Telegram Bot Not Responding**

   ```
   Error: bot token is required for Telegram notifications
   ```

   **Solution**: Verify your bot token and chat ID are correct

3. **Discord Webhook Invalid**
   ```
   Error: webhook URL must be from discord.com or discordapp.com
   ```
   **Solution**: Ensure the webhook URL is from discord.com or discordapp.com

### Testing Notifications

You can test your notification configuration:

```bash
# Test with debug logging
export SERVERHEALTH_LOG_LEVEL=debug
./serverhealth start
```

### Debug Mode

Enable debug logging to see detailed notification information:

```bash
export SERVERHEALTH_LOG_LEVEL=debug
./serverhealth start
```

## 🔄 Migration from Legacy Configuration

The system automatically migrates legacy configuration:

### Legacy Format

```yaml
slack_disk_webhook_url: "https://hooks.slack.com/services/OLD/WEBHOOK"
slack_cpu_memory_webhook_url: "https://hooks.slack.com/services/OLD/WEBHOOK"
disk_enabled: true
cpu_enabled: true
memory_enabled: true
```

### Migrated to New Format

```yaml
notifications:
  - type: slack
    enabled: true
    webhook_url: "https://hooks.slack.com/services/OLD/WEBHOOK"
  - type: slack
    enabled: true
    webhook_url: "https://hooks.slack.com/services/OLD/WEBHOOK"
disk:
  enabled: true
  threshold: 80
  check_interval: 12
  max_daily_alerts: 5
cpu:
  enabled: true
  threshold: 85
  check_interval: 60
  max_daily_alerts: 5
memory:
  enabled: true
  threshold: 85
  check_interval: 60
  max_daily_alerts: 5
```

## 🚀 Best Practices

1. **Use Multiple Providers**: Configure at least two providers for redundancy
2. **Set Appropriate Limits**: Configure `max_daily_alerts` to prevent spam
3. **Monitor Logs**: Check logs for notification delivery status
4. **Test Configuration**: Test notifications before deploying to production
5. **Use Environment Variables**: For containerized deployments

## 📈 Performance

- **Concurrent Notifications**: All providers send notifications concurrently
- **Non-Blocking**: Notification failures don't affect monitoring
- **Efficient**: Connection pooling and retry logic optimize performance
- **Scalable**: Easy to add new notification providers

## 🔮 Future Enhancements

Planned features for future releases:

- Email notifications
- PagerDuty integration
- OpsGenie integration
- Custom webhook support
- Notification templates
- Advanced filtering
- Notification history
