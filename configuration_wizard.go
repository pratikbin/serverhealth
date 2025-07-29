package main

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/manifoldco/promptui"
)

// ConfigurationWizard handles the interactive configuration process
type ConfigurationWizard struct{}

// NewConfigurationWizard creates a new configuration wizard
func NewConfigurationWizard() *ConfigurationWizard {
	return &ConfigurationWizard{}
}

// Run executes the configuration wizard
func (w *ConfigurationWizard) Run(config *Config) error {
	steps := []func(*Config) error{
		w.configureMonitoringOptions,
		w.configureSlackWebhooks,
		w.configureThresholds,
		w.configureIntervals,
	}

	for _, step := range steps {
		if err := step(config); err != nil {
			return err
		}
	}

	// Set defaults for missing values
	w.setDefaults(config)

	return config.Validate()
}

func (w *ConfigurationWizard) setDefaults(config *Config) {
	if config.LogLevel == "" {
		config.LogLevel = "info"
	}
	if config.ServiceName == "" {
		config.ServiceName = appName
	}
}

func (w *ConfigurationWizard) configureMonitoringOptions(config *Config) error {
	fmt.Println(bold("📊 Monitoring Options"))
	fmt.Println("Choose which metrics you want to monitor:")
	fmt.Println()

	// Define available metrics
	metrics := []struct {
		Name        string
		Description string
		Key         string
		Default     bool
	}{
		{"Disk Usage", "Monitor disk space usage (/)", "disk", true},
		{"CPU Usage", "Monitor CPU utilization", "cpu", true},
		{"Memory Usage", "Monitor RAM utilization", "memory", true},
	}

	// Show instructions
	fmt.Println(blue("📝 Instructions:"))
	fmt.Println("  • Use ↑/↓ arrows to navigate")
	fmt.Println("  • Press Enter to toggle selection")
	fmt.Println("  • At least one metric must be selected")
	fmt.Println()

	selectedMetrics := make(map[string]bool)
	// Set current selections
	selectedMetrics["disk"] = config.Disk.Enabled
	selectedMetrics["cpu"] = config.CPU.Enabled
	selectedMetrics["memory"] = config.Memory.Enabled

	// Multi-select loop with proper cursor handling
	for {
		// Show current selections
		fmt.Println()
		fmt.Println(bold("Current selections: "))
		hasSelections := false
		for _, metric := range metrics {
			if selectedMetrics[metric.Key] {
				fmt.Printf(green("✓ %s "), metric.Name)
				hasSelections = true
			}
		}
		if !hasSelections {
			fmt.Print(yellow("None selected"))
		}
		fmt.Println()

		// Create menu items with proper formatting
		var items []string
		for _, metric := range metrics {
			checkbox := "[ ]"
			if selectedMetrics[metric.Key] {
				checkbox = "[✓]"
			}
			items = append(items, fmt.Sprintf("%s %s - %s", checkbox, metric.Name, metric.Description))
		}
		items = append(items, "── Continue with current selections ──")

		// Better templates with proper highlighting
		templates := &promptui.SelectTemplates{
			Label:    "{{ . }}",
			Active:   "👉 {{ . | cyan | bold }}",
			Inactive: "   {{ . | white }}",
			Selected: "{{ . | green }}",
		}

		prompt := promptui.Select{
			Label:     "Select metric to toggle (or continue)",
			Items:     items,
			Templates: templates,
			Size:      len(items),
			HideHelp:  true,
		}

		index, _, err := prompt.Run()
		if err != nil {
			return err
		}

		// If user selected "Continue"
		if index == len(metrics) {
			break
		}

		// Toggle selection
		metric := metrics[index]
		selectedMetrics[metric.Key] = !selectedMetrics[metric.Key]
	}

	// Validate at least one selection
	hasAnySelection := false
	for _, selected := range selectedMetrics {
		if selected {
			hasAnySelection = true
			break
		}
	}

	if !hasAnySelection {
		return fmt.Errorf("at least one monitoring option must be enabled")
	}

	// Apply selections to config
	config.Disk.Enabled = selectedMetrics["disk"]
	config.CPU.Enabled = selectedMetrics["cpu"]
	config.Memory.Enabled = selectedMetrics["memory"]

	fmt.Println()
	fmt.Println(green("✅ Monitoring options configured!"))
	return nil
}

func (w *ConfigurationWizard) configureSlackWebhooks(config *Config) error {
	fmt.Println()
	fmt.Println(bold("🔗 Notification Providers"))
	fmt.Println("Configure notification providers for alerts:")
	fmt.Println()

	// Define available notification providers
	providers := []struct {
		Name        string
		Description string
		Type        string
	}{
		{"Slack", "Send notifications to Slack channels", "slack"},
		{"Telegram", "Send notifications to Telegram chat", "telegram"},
		{"Discord", "Send notifications to Discord channels", "discord"},
	}

	fmt.Println(blue("📝 Instructions:"))
	fmt.Println("  • Use ↑/↓ arrows to navigate")
	fmt.Println("  • Press Enter to configure selected provider")
	fmt.Println("  • You can configure multiple providers")
	fmt.Println("  • At least one provider is recommended")
	fmt.Println()

	for {
		// Show current notification status
		fmt.Println(bold("Notification Providers:"))
		var items []string

		for _, provider := range providers {
			enabled := false
			for _, notification := range config.Notifications {
				if notification.Type == provider.Type && notification.Enabled {
					enabled = true
					break
				}
			}

			status := red("❌ Not configured")
			if enabled {
				status = green("✅ Configured")
			}

			display := fmt.Sprintf("%s - %s", provider.Name, status)
			items = append(items, display)
		}

		items = append(items, green("── Continue with current configuration ──"))

		templates := &promptui.SelectTemplates{
			Label:    "{{ . }}",
			Active:   "👉 {{ . | cyan | bold }}",
			Inactive: "   {{ . | white }}",
			Selected: "{{ . | green }}",
		}

		prompt := promptui.Select{
			Label:     "Select notification provider to configure",
			Items:     items,
			Templates: templates,
			Size:      len(items),
			HideHelp:  true,
		}

		index, _, err := prompt.Run()
		if err != nil {
			return err
		}

		// If user selected continue
		if index >= len(providers) {
			break
		}

		// Configure selected provider
		provider := &providers[index]
		if err := w.configureNotificationProvider(config, provider.Type); err != nil {
			return err
		}
	}

	// Check if at least one notification is configured
	enabledCount := 0
	for _, notification := range config.Notifications {
		if notification.Enabled {
			enabledCount++
		}
	}

	if enabledCount == 0 {
		fmt.Println(yellow("⚠️  No notification providers configured"))
		fmt.Println("You can add notifications later by running 'serverhealth configure' again")
	}

	fmt.Println(green("🎉 Notification configuration complete!"))
	return nil
}

func (w *ConfigurationWizard) configureNotificationProvider(config *Config, providerType string) error {
	fmt.Printf("\n%s Configuration\n", strings.ToUpper(providerType[:1])+providerType[1:])
	fmt.Println("═══════════════════════════════════")

	// Find existing notification or create new one
	var notification *NotificationConfig
	for i := range config.Notifications {
		if config.Notifications[i].Type == providerType {
			notification = &config.Notifications[i]
			break
		}
	}

	if notification == nil {
		// Create new notification
		config.Notifications = append(config.Notifications, NotificationConfig{
			Type:    providerType,
			Enabled: true,
		})
		notification = &config.Notifications[len(config.Notifications)-1]
	}

	// Configure based on provider type
	switch providerType {
	case "slack":
		return w.configureSlackProvider(notification)
	case "telegram":
		return w.configureTelegramProvider(notification)
	case "discord":
		return w.configureDiscordProvider(notification)
	default:
		return fmt.Errorf("unsupported provider type: %s", providerType)
	}
}

func (w *ConfigurationWizard) configureSlackProvider(notification *NotificationConfig) error {
	fmt.Println("Slack uses webhook URLs to send notifications.")
	fmt.Println("To create a webhook:")
	fmt.Println("1. Go to your Slack workspace")
	fmt.Println("2. Create a new app or use an existing one")
	fmt.Println("3. Enable Incoming Webhooks")
	fmt.Println("4. Create a webhook URL")
	fmt.Println()

	prompt := promptui.Prompt{
		Label:   "Enter Slack webhook URL",
		Default: notification.WebhookURL,
		Validate: func(input string) error {
			if input == "" {
				return fmt.Errorf("webhook URL cannot be empty")
			}
			if !strings.HasPrefix(input, "https://hooks.slack.com/") {
				return fmt.Errorf("invalid Slack webhook URL format")
			}
			return nil
		},
	}

	result, err := prompt.Run()
	if err != nil {
		return err
	}

	notification.WebhookURL = result
	notification.Enabled = true

	fmt.Println(green("✅ Slack notification configured successfully!"))
	return nil
}

func (w *ConfigurationWizard) configureTelegramProvider(notification *NotificationConfig) error {
	fmt.Println("Telegram uses bot tokens and chat IDs to send notifications.")
	fmt.Println("To set up:")
	fmt.Println("1. Create a bot with @BotFather")
	fmt.Println("2. Get your bot token")
	fmt.Println("3. Get your chat ID (send a message to your bot and check @userinfobot)")
	fmt.Println()

	// Bot Token
	prompt := promptui.Prompt{
		Label:   "Enter Telegram bot token",
		Default: notification.BotToken,
		Validate: func(input string) error {
			if input == "" {
				return fmt.Errorf("bot token cannot be empty")
			}
			return nil
		},
	}

	result, err := prompt.Run()
	if err != nil {
		return err
	}
	notification.BotToken = result

	// Chat ID
	prompt2 := promptui.Prompt{
		Label:   "Enter Telegram chat ID",
		Default: notification.ChatID,
		Validate: func(input string) error {
			if input == "" {
				return fmt.Errorf("chat ID cannot be empty")
			}
			return nil
		},
	}

	result2, err := prompt2.Run()
	if err != nil {
		return err
	}
	notification.ChatID = result2
	notification.Enabled = true

	fmt.Println(green("✅ Telegram notification configured successfully!"))
	return nil
}

func (w *ConfigurationWizard) configureDiscordProvider(notification *NotificationConfig) error {
	fmt.Println("Discord uses webhook URLs to send notifications.")
	fmt.Println("To create a webhook:")
	fmt.Println("1. Go to your Discord server")
	fmt.Println("2. Edit a channel")
	fmt.Println("3. Go to Integrations > Webhooks")
	fmt.Println("4. Create a new webhook")
	fmt.Println("5. Copy the webhook URL")
	fmt.Println()

	prompt := promptui.Prompt{
		Label:   "Enter Discord webhook URL",
		Default: notification.WebhookURL,
		Validate: func(input string) error {
			if input == "" {
				return fmt.Errorf("webhook URL cannot be empty")
			}
			if !strings.Contains(input, "discord.com") && !strings.Contains(input, "discordapp.com") {
				return fmt.Errorf("invalid Discord webhook URL format")
			}
			return nil
		},
	}

	result, err := prompt.Run()
	if err != nil {
		return err
	}

	notification.WebhookURL = result
	notification.Enabled = true

	fmt.Println(green("✅ Discord notification configured successfully!"))
	return nil
}

func (w *ConfigurationWizard) configureThresholds(config *Config) error {
	fmt.Println()
	fmt.Println(bold("⚠️ Alert Thresholds"))
	fmt.Println("Configure when you want to receive alerts:")
	fmt.Println()

	// Collect enabled metrics that need thresholds
	var thresholdNeeds []struct {
		Name        string
		Current     int
		Default     int
		Target      *int
		Unit        string
		Description string
	}

	if config.Disk.Enabled {
		if config.Disk.Threshold == 0 {
			config.Disk.Threshold = 80
		}
		thresholdNeeds = append(thresholdNeeds, struct {
			Name        string
			Current     int
			Default     int
			Target      *int
			Unit        string
			Description string
		}{
			Name:        "Disk Usage",
			Current:     config.Disk.Threshold,
			Default:     80,
			Target:      &config.Disk.Threshold,
			Unit:        "%",
			Description: "Alert when disk usage exceeds this percentage",
		})
	}

	if config.CPU.Enabled {
		if config.CPU.Threshold == 0 {
			config.CPU.Threshold = 85
		}
		thresholdNeeds = append(thresholdNeeds, struct {
			Name        string
			Current     int
			Default     int
			Target      *int
			Unit        string
			Description string
		}{
			Name:        "CPU Usage",
			Current:     config.CPU.Threshold,
			Default:     85,
			Target:      &config.CPU.Threshold,
			Unit:        "%",
			Description: "Alert when CPU usage exceeds this percentage",
		})
	}

	if config.Memory.Enabled {
		if config.Memory.Threshold == 0 {
			config.Memory.Threshold = 85
		}
		thresholdNeeds = append(thresholdNeeds, struct {
			Name        string
			Current     int
			Default     int
			Target      *int
			Unit        string
			Description string
		}{
			Name:        "Memory Usage",
			Current:     config.Memory.Threshold,
			Default:     85,
			Target:      &config.Memory.Threshold,
			Unit:        "%",
			Description: "Alert when memory usage exceeds this percentage",
		})
	}

	if len(thresholdNeeds) == 0 {
		return fmt.Errorf("no monitoring enabled - cannot configure thresholds")
	}

	fmt.Println(blue("📝 Instructions:"))
	fmt.Println("  • Use ↑/↓ arrows to navigate")
	fmt.Println("  • Press Enter to configure selected threshold")
	fmt.Println("  • Enter values between 1-100")
	fmt.Println("  • Recommended: Disk=80%, CPU/Memory=85%")
	fmt.Println()

	for {
		// Show current threshold status
		fmt.Println(bold("Threshold Configuration:"))
		var items []string

		for _, threshold := range thresholdNeeds {
			status := ""
			if threshold.Current > 0 {
				status = green(fmt.Sprintf("✅ %d%s", threshold.Current, threshold.Unit))
				if threshold.Current >= 95 {
					status += red(" (Very High!)")
				} else if threshold.Current >= 90 {
					status += yellow(" (High)")
				}
			} else {
				status = red("❌ Not set")
			}

			display := fmt.Sprintf("%s - %s", threshold.Name, status)
			items = append(items, display)
		}

		items = append(items, green("── All thresholds configured - Continue ──"))

		templates := &promptui.SelectTemplates{
			Label:    "{{ . }}",
			Active:   "👉 {{ . | cyan | bold }}",
			Inactive: "   {{ . | white }}",
			Selected: "{{ . | green }}",
		}

		prompt := promptui.Select{
			Label:     "Select threshold to configure",
			Items:     items,
			Templates: templates,
			Size:      len(items),
			HideHelp:  true,
		}

		index, _, err := prompt.Run()
		if err != nil {
			return err
		}

		// If user selected continue
		if index >= len(thresholdNeeds) {
			break
		}

		// Configure selected threshold
		threshold := &thresholdNeeds[index]

		fmt.Printf("\n%s\n", threshold.Description)
		fmt.Printf("Recommended: %d%s\n\n", threshold.Default, threshold.Unit)

		prompt2 := promptui.Prompt{
			Label:   fmt.Sprintf("%s threshold (1-100)", threshold.Name),
			Default: strconv.Itoa(threshold.Current),
			Validate: func(input string) error {
				val, err := strconv.Atoi(input)
				if err != nil {
					return fmt.Errorf("please enter a valid number")
				}
				if val < 1 || val > 100 {
					return fmt.Errorf("threshold must be between 1 and 100")
				}
				if val >= 95 {
					fmt.Printf(yellow("⚠️  Warning: %d%% is very high and may cause frequent alerts\n"), val)
				}
				return nil
			},
		}

		result, err := prompt2.Run()
		if err != nil {
			return err
		}

		newValue, _ := strconv.Atoi(result)
		*threshold.Target = newValue
		threshold.Current = newValue

		fmt.Println(green("✅ Threshold configured successfully!"))
		fmt.Println()
	}

	fmt.Println(green("🎯 Alert thresholds configuration complete!"))
	return nil
}

func (w *ConfigurationWizard) configureIntervals(config *Config) error {
	fmt.Println()
	fmt.Println(bold("⏰ Check Intervals"))
	fmt.Println("Configure how often to check each metric:")
	fmt.Println()

	// Collect enabled metrics that need intervals
	var intervalNeeds []struct {
		Name        string
		Current     int
		Default     int
		Target      *int
		Unit        string
		Description string
		Min         int
		Max         int
	}

	// Add CPU interval if enabled
	if config.CPU.Enabled {
		if config.CPU.CheckInterval == 0 {
			config.CPU.CheckInterval = 60
		}
		intervalNeeds = append(intervalNeeds, struct {
			Name        string
			Current     int
			Default     int
			Target      *int
			Unit        string
			Description string
			Min         int
			Max         int
		}{
			Name:        "CPU Check Interval",
			Current:     config.CPU.CheckInterval,
			Default:     60,
			Target:      &config.CPU.CheckInterval,
			Unit:        "minutes",
			Description: "How often to check CPU usage",
			Min:         1,
			Max:         1440, // 24 hours
		})
	}

	// Add Memory interval if enabled
	if config.Memory.Enabled {
		if config.Memory.CheckInterval == 0 {
			config.Memory.CheckInterval = 60
		}
		intervalNeeds = append(intervalNeeds, struct {
			Name        string
			Current     int
			Default     int
			Target      *int
			Unit        string
			Description string
			Min         int
			Max         int
		}{
			Name:        "Memory Check Interval",
			Current:     config.Memory.CheckInterval,
			Default:     60,
			Target:      &config.Memory.CheckInterval,
			Unit:        "minutes",
			Description: "How often to check memory usage",
			Min:         1,
			Max:         1440, // 24 hours
		})
	}

	// Add Disk interval if enabled
	if config.Disk.Enabled {
		if config.Disk.CheckInterval == 0 {
			config.Disk.CheckInterval = 12
		}
		intervalNeeds = append(intervalNeeds, struct {
			Name        string
			Current     int
			Default     int
			Target      *int
			Unit        string
			Description string
			Min         int
			Max         int
		}{
			Name:        "Disk Check Interval",
			Current:     config.Disk.CheckInterval,
			Default:     12,
			Target:      &config.Disk.CheckInterval,
			Unit:        "hours",
			Description: "How often to check disk usage",
			Min:         1,
			Max:         168, // 1 week
		})
	}

	if len(intervalNeeds) == 0 {
		return fmt.Errorf("no monitoring enabled - cannot configure intervals")
	}

	fmt.Println(blue("📝 Instructions:"))
	fmt.Println("  • Use ↑/↓ arrows to navigate")
	fmt.Println("  • Press Enter to configure selected interval")
	fmt.Println("  • Shorter intervals = more frequent checks")
	fmt.Println("  • Recommended: CPU/Memory=60min, Disk=12hours")
	fmt.Println()

	for {
		// Show current interval status
		fmt.Println(bold("Interval Configuration:"))
		var items []string

		for _, interval := range intervalNeeds {
			status := ""
			if interval.Current > 0 {
				status = green(fmt.Sprintf("✅ Every %d %s", interval.Current, interval.Unit))

				switch interval.Unit {
				case "minutes":
					if interval.Current <= 5 {
						status += red(" (Very Frequent!)")
					} else if interval.Current <= 15 {
						status += yellow(" (Frequent)")
					} else if interval.Current >= 240 {
						status += yellow(" (Infrequent)")
					}
				case "hours":
					if interval.Current <= 1 {
						status += yellow(" (Frequent)")
					} else if interval.Current >= 48 {
						status += yellow(" (Infrequent)")
					}
				}
			} else {
				status = red("❌ Not set")
			}

			display := fmt.Sprintf("%s - %s", interval.Name, status)
			items = append(items, display)
		}

		items = append(items, green("── All intervals configured - Continue ──"))

		templates := &promptui.SelectTemplates{
			Label:    "{{ . }}",
			Active:   "👉 {{ . | cyan | bold }}",
			Inactive: "   {{ . | white }}",
			Selected: "{{ . | green }}",
		}

		prompt := promptui.Select{
			Label:     "Select interval to configure",
			Items:     items,
			Templates: templates,
			Size:      len(items),
			HideHelp:  true,
		}

		index, _, err := prompt.Run()
		if err != nil {
			return err
		}

		// If user selected continue
		if index >= len(intervalNeeds) {
			break
		}

		// Configure selected interval
		interval := &intervalNeeds[index]

		fmt.Printf("\n%s\n", interval.Description)
		fmt.Printf("Recommended: %d %s\n", interval.Default, interval.Unit)
		fmt.Printf("Range: %d-%d %s\n\n", interval.Min, interval.Max, interval.Unit)

		prompt2 := promptui.Prompt{
			Label:   fmt.Sprintf("%s (%s)", interval.Name, interval.Unit),
			Default: strconv.Itoa(interval.Current),
			Validate: func(input string) error {
				val, err := strconv.Atoi(input)
				if err != nil {
					return fmt.Errorf("please enter a valid number")
				}
				if val < interval.Min || val > interval.Max {
					return fmt.Errorf("interval must be between %d and %d %s", interval.Min, interval.Max, interval.Unit)
				}

				// Warnings for extreme values
				if interval.Unit == "minutes" && val <= 5 {
					fmt.Printf(yellow("⚠️  Warning: %d minutes is very frequent and may impact performance\n"), val)
				}
				if interval.Unit == "hours" && val <= 1 {
					fmt.Printf(yellow("⚠️  Note: %d hour is quite frequent for disk checks\n"), val)
				}

				return nil
			},
		}

		result, err := prompt2.Run()
		if err != nil {
			return err
		}

		newValue, _ := strconv.Atoi(result)
		*interval.Target = newValue
		interval.Current = newValue

		fmt.Println(green("✅ Interval configured successfully!"))
		fmt.Println()
	}

	fmt.Println(green("⏱️ Check intervals configuration complete!"))
	return nil
}

// Note: This file assumes the following functions and types are defined elsewhere:
// - bold(), green(), red(), yellow(), blue(), faint() - console color functions
// - Config struct with all required fields
// - appName constant
// - Config.Validate() method
