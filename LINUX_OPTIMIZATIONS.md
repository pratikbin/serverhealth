# Linux Optimizations for ServerHealth

This document outlines the Linux-specific optimizations and improvements made to the ServerHealth monitoring tool.

## 🐧 Linux-Specific Features

### Native System Calls

- **Direct syscalls** for disk usage monitoring
- **Optimized process management** with proper signal handling
- **Platform-specific service management** using systemd

### Performance Improvements

- **Reduced memory footprint** through efficient Go routines
- **Optimized file I/O** with proper buffering
- **Enhanced error handling** for Linux-specific edge cases

## 🔧 Installation on Linux

### Quick Install

```bash
# Download and install
curl -sSL https://github.com/yourusername/serverhealth/releases/latest/download/serverhealth-linux-amd64 | sudo tee /usr/local/bin/serverhealth
sudo chmod +x /usr/local/bin/serverhealth

# Create configuration directory
mkdir -p ~/.config/serverhealth
```

### System Service Installation

```bash
# Install as system service
sudo serverhealth install

# Start the service
sudo systemctl start serverhealth

# Enable auto-start
sudo systemctl enable serverhealth

# Check status
sudo systemctl status serverhealth
```

## 📊 Configuration

### YAML Configuration

```yaml
# Monitoring settings
disk:
  enabled: true
  threshold: 80
  check_interval: 12 # hours
  max_daily_alerts: 5

cpu:
  enabled: true
  threshold: 85
  check_interval: 60 # minutes
  max_daily_alerts: 5

memory:
  enabled: true
  threshold: 85
  check_interval: 60 # minutes
  max_daily_alerts: 5

# Notification providers
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

# General settings
log_level: info
service_name: serverhealth
```

## 🚀 Usage Examples

### Basic Monitoring

```bash
# Start monitoring in foreground
serverhealth start

# Start as background daemon
serverhealth start --background

# Check status
serverhealth status

# View logs
serverhealth logs
```

### Service Management

```bash
# Install as system service
sudo serverhealth install

# Start service
sudo systemctl start serverhealth

# Stop service
sudo systemctl stop serverhealth

# Check service status
sudo systemctl status serverhealth

# View service logs
sudo journalctl -u serverhealth -f
```

## 🔍 Monitoring Details

### Disk Usage Monitoring

- **Native Linux syscalls** for accurate disk usage
- **Support for all filesystem types** (ext4, xfs, btrfs, etc.)
- **Automatic threshold detection** based on filesystem
- **Detailed error reporting** for mount point issues

### CPU Usage Monitoring

- **Real-time CPU utilization** tracking
- **Multi-core support** with aggregate reporting
- **Load average monitoring** for system health
- **Process-specific CPU tracking** (optional)

### Memory Usage Monitoring

- **Physical and virtual memory** monitoring
- **Swap usage tracking** for memory pressure detection
- **Detailed memory breakdown** (used, cached, buffers)
- **Memory leak detection** capabilities

## 🛠️ Troubleshooting

### Common Issues

#### Permission Denied

```bash
# Ensure proper permissions
sudo chown -R $USER:$USER ~/.config/serverhealth
chmod 755 ~/.config/serverhealth
```

#### Service Won't Start

```bash
# Check service logs
sudo journalctl -u serverhealth -n 50

# Verify configuration
serverhealth status

# Test configuration
serverhealth configure
```

#### Disk Monitoring Issues

```bash
# Check filesystem mounts
df -h

# Verify disk permissions
ls -la /proc/mounts

# Test disk access
cat /proc/diskstats
```

## 📈 Performance Metrics

### Resource Usage

- **Memory**: ~5-10MB typical usage
- **CPU**: <1% during idle, spikes during checks
- **Disk I/O**: Minimal, only during log writes
- **Network**: Only during notification sends

### Optimization Tips

1. **Use SSD storage** for log files
2. **Configure appropriate check intervals** based on system load
3. **Set reasonable alert thresholds** to avoid false positives
4. **Monitor log file size** to prevent disk space issues

## 🔒 Security Considerations

### File Permissions

```bash
# Secure configuration directory
chmod 700 ~/.config/serverhealth
chmod 600 ~/.config/serverhealth/config.yaml

# Secure log directory
chmod 755 ~/.local/log
chmod 644 ~/.local/log/serverhealth.log
```

### Service Security

- **Run as non-root user** when possible
- **Use dedicated service account** for production
- **Restrict file system access** to necessary directories
- **Validate webhook URLs** before configuration

## 🐛 Debugging

### Enable Debug Logging

```yaml
# In config.yaml
log_level: debug
```

### Verbose Output

```bash
# Run with debug output
serverhealth start --debug

# Check detailed status
serverhealth status --verbose
```

### Log Analysis

```bash
# View real-time logs
tail -f ~/.local/log/serverhealth.log

# Search for errors
grep ERROR ~/.local/log/serverhealth.log

# Monitor notification attempts
grep "notification" ~/.local/log/serverhealth.log
```

## 📋 System Requirements

### Minimum Requirements

- **OS**: Linux kernel 3.10+
- **Architecture**: x86_64, ARM64
- **Memory**: 50MB available RAM
- **Storage**: 100MB free space
- **Network**: Internet access for notifications

### Recommended Requirements

- **OS**: Linux kernel 4.19+
- **Architecture**: x86_64 with SSE4.2
- **Memory**: 100MB available RAM
- **Storage**: 500MB free space (SSD preferred)
- **Network**: Stable internet connection

## 🔄 Updates and Maintenance

### Updating ServerHealth

```bash
# Download latest version
curl -sSL https://github.com/yourusername/serverhealth/releases/latest/download/serverhealth-linux-amd64 | sudo tee /usr/local/bin/serverhealth

# Restart service
sudo systemctl restart serverhealth
```

### Configuration Backup

```bash
# Backup configuration
cp ~/.config/serverhealth/config.yaml ~/.config/serverhealth/config.yaml.backup

# Restore configuration
cp ~/.config/serverhealth/config.yaml.backup ~/.config/serverhealth/config.yaml
```

### Log Rotation

```bash
# Create logrotate configuration
sudo tee /etc/logrotate.d/serverhealth << EOF
~/.local/log/serverhealth.log {
    daily
    missingok
    rotate 7
    compress
    delaycompress
    notifempty
    create 644 $USER $USER
}
EOF
```

This optimized version provides enhanced performance and reliability for Linux environments while maintaining full compatibility with the modular notification system.
