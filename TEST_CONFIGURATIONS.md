# Test Configurations

This directory contains test configuration files for the ServerHealth modular notification system.

## Test Configuration Files

### `test-config.yaml`

**Purpose**: Full test configuration with all notification providers and monitoring settings enabled.

**Usage**:

```bash
cp test-config.yaml ~/.config/serverhealth/config.yaml
./serverhealth status
```

**Features**:

- All monitoring metrics enabled (Disk, CPU, Memory)
- All notification providers configured (Slack, Telegram, Discord)
- Custom thresholds and intervals for testing
- Debug logging enabled

### `test-config-minimal.yaml`

**Purpose**: Minimal test configuration with basic settings for quick testing.

**Usage**:

```bash
cp test-config-minimal.yaml ~/.config/serverhealth/config.yaml
./serverhealth status
```

**Features**:

- Single Slack notification provider
- Basic monitoring settings
- Quick verification of core functionality

## Testing Scenarios

### 1. Full Configuration Test

```bash
cp test-config.yaml ~/.config/serverhealth/config.yaml
./serverhealth status
```

**Expected**: Shows all monitoring metrics and notification providers configured.

### 2. Minimal Configuration Test

```bash
cp test-config-minimal.yaml ~/.config/serverhealth/config.yaml
./serverhealth status
```

**Expected**: Shows basic configuration with single notification provider.

### 3. Configuration Validation Test

```bash
./serverhealth configure
```

**Expected**: Interactive wizard to set up new configuration.

## Configuration Details

### Monitoring Settings

- **Disk**: 85% threshold, 6-hour intervals, 3 max daily alerts
- **CPU**: 90% threshold, 30-minute intervals, 8 max daily alerts
- **Memory**: 88% threshold, 45-minute intervals, 6 max daily alerts

### Notification Providers

- **Slack**: Webhook URL for Slack integration
- **Telegram**: Bot token and chat ID for Telegram integration
- **Discord**: Webhook URL for Discord integration

## Important Notes

⚠️ **Placeholder URLs**: The test configuration files contain placeholder URLs and tokens. Replace these with your actual credentials before testing notifications.

⚠️ **Test Environment**: These configurations are designed for testing only. Use proper credentials in production environments.

## File Locations

- **Configuration**: `~/.config/serverhealth/config.yaml`
- **Logs**: `~/.local/log/serverhealth.log`
- **PID**: `~/.local/run/serverhealth.pid`
