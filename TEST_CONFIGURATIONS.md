# Test Configuration Files

This directory contains several test configuration files to help you test the modular notification system.

## 📁 Available Test Configurations

### 1. `test-config.yaml` - Full Configuration

**Purpose**: Complete test configuration with all notification providers
**Use Case**: Testing all features of the modular notification system

**Features:**

- All 3 notification providers (Slack, Telegram, Discord)
- Aggressive monitoring thresholds (85-90%)
- Short check intervals for testing
- Debug logging enabled
- Legacy configuration fields for migration testing

**Usage:**

```bash
cp test-config.yaml ~/.config/serverhealth/config.yaml
./serverhealth status
```

### 2. `test-config-minimal.yaml` - Minimal Configuration

**Purpose**: Simple configuration with basic settings
**Use Case**: Quick testing with minimal setup

**Features:**

- Single Slack notification provider
- Standard monitoring thresholds (80-85%)
- Default check intervals
- Info logging level

**Usage:**

```bash
cp test-config-minimal.yaml ~/.config/serverhealth/config.yaml
./serverhealth status
```

### 3. `test-config-env.yaml` - Environment Variable Example

**Purpose**: Demonstrates environment variable usage
**Use Case**: Containerized deployments and CI/CD

**Features:**

- All notification providers configured
- Environment variable override examples
- Comments showing how to use environment variables

**Usage:**

```bash
cp test-config-env.yaml ~/.config/serverhealth/config.yaml
# Set environment variables to override configuration
export SERVERHEALTH_DISK_THRESHOLD=90
export SERVERHEALTH_LOG_LEVEL=debug
./serverhealth status
```

## 🔧 Configuration Details

### Monitoring Settings

**Disk Monitoring:**

- Threshold: 80-85%
- Check Interval: 6-12 hours
- Max Daily Alerts: 3-5

**CPU Monitoring:**

- Threshold: 85-90%
- Check Interval: 30-60 minutes
- Max Daily Alerts: 5-8

**Memory Monitoring:**

- Threshold: 85-88%
- Check Interval: 45-60 minutes
- Max Daily Alerts: 5-6

### Notification Providers

**Slack:**

- Webhook URL format: `https://hooks.slack.com/services/T00000000/B00000000/XXXXXXXXXXXXXXXXXXXXXXXX`
- Rich text formatting with emojis
- Automatic retry logic

**Telegram:**

- Bot token format: `1234567890:ABCdefGHIjklMNOpqrsTUVwxyz`
- Chat ID format: `123456789`
- Markdown formatting support

**Discord:**

- Webhook URL format: `https://discord.com/api/webhooks/123456789012345678/abcdefghijklmnopqrstuvwxyz1234567890123456789012345678901234567890`
- Rich embed messages with colors
- Structured field display

## 🧪 Testing Scenarios

### 1. Basic Functionality Test

```bash
# Use minimal configuration
cp test-config-minimal.yaml ~/.config/serverhealth/config.yaml
./serverhealth status
./serverhealth configure
```

### 2. Full Feature Test

```bash
# Use full configuration
cp test-config.yaml ~/.config/serverhealth/config.yaml
./serverhealth status
./serverhealth start --background
```

### 3. Environment Variable Test

```bash
# Use environment variable configuration
cp test-config-env.yaml ~/.config/serverhealth/config.yaml
export SERVERHEALTH_DISK_THRESHOLD=95
export SERVERHEALTH_LOG_LEVEL=debug
./serverhealth status
```

### 4. Migration Test

```bash
# Test legacy configuration migration
# Create a legacy config file and test automatic migration
./serverhealth configure
```

## ⚠️ Important Notes

### Test Webhook URLs

The webhook URLs in these test files are **placeholder values** and will not work for actual notifications. To test real notifications:

1. **Replace Slack Webhook**: Get a real webhook URL from your Slack workspace
2. **Replace Telegram Credentials**: Create a bot with @BotFather and get your chat ID
3. **Replace Discord Webhook**: Create a webhook in your Discord server

### Example Real Configuration

```yaml
notifications:
  - type: slack
    enabled: true
    webhook_url: "https://hooks.slack.com/services/YOUR/ACTUAL/WEBHOOK"

  - type: telegram
    enabled: true
    bot_token: "YOUR_ACTUAL_BOT_TOKEN"
    chat_id: "YOUR_ACTUAL_CHAT_ID"

  - type: discord
    enabled: true
    webhook_url: "https://discord.com/api/webhooks/YOUR/ACTUAL/WEBHOOK"
```

## 🔍 Validation Testing

### Configuration Validation

```bash
# Test configuration loading
./serverhealth status

# Test configuration wizard
./serverhealth configure

# Test with invalid configuration
# Edit config.yaml with invalid values and test error handling
```

### Notification Testing

```bash
# Test notification system (requires real webhook URLs)
./serverhealth start --background

# Check logs for notification attempts
./serverhealth logs

# Stop the service
./serverhealth stop
```

## 🚀 Production Deployment

When ready for production:

1. **Replace Test Values**: Update all webhook URLs and credentials
2. **Adjust Thresholds**: Set appropriate thresholds for your environment
3. **Configure Intervals**: Set check intervals based on your needs
4. **Set Alert Limits**: Configure max daily alerts to prevent spam
5. **Test Thoroughly**: Verify all notification providers work correctly

## 📊 Expected Output

When using the test configurations, you should see:

**Status Command Output:**

```
📊 ServerHealth Status
═══════════════════════════════
✅ Configuration found

Monitoring Configuration:
  • Disk usage (threshold: 85%, check every 6 hours, max alerts: 3/day)
  • CPU usage (threshold: 90%, check every 30 minutes, max alerts: 8/day)
  • Memory usage (threshold: 88%, check every 45 minutes, max alerts: 6/day)

Runtime Status:
❌ ServerHealth is not running
```

**Configuration Wizard Output:**

```
🔧 ServerHealth Configuration
═══════════════════════════════

📊 Monitoring Options
Choose which metrics you want to monitor:

🔗 Notification Providers
Configure notification providers for alerts:
```

These test configurations provide a comprehensive way to test all features of the modular notification system before deploying to production.
