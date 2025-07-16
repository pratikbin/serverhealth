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
	selectedMetrics["disk"] = config.DiskEnabled
	selectedMetrics["cpu"] = config.CPUEnabled
	selectedMetrics["memory"] = config.MemoryEnabled

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
	config.DiskEnabled = selectedMetrics["disk"]
	config.CPUEnabled = selectedMetrics["cpu"]
	config.MemoryEnabled = selectedMetrics["memory"]

	fmt.Println()
	fmt.Println(green("✅ Monitoring options configured!"))
	return nil
}

func (w *ConfigurationWizard) configureSlackWebhooks(config *Config) error {
	fmt.Println()
	fmt.Println(bold("🔗 Slack Webhook Configuration"))
	fmt.Println("Enter your Slack webhook URLs for notifications:")
	fmt.Println()

	// Collect enabled metrics that need webhooks
	var webhookNeeds []struct {
		Name     string
		Current  string
		Required []string
		Target   *string
	}

	if config.DiskEnabled {
		webhookNeeds = append(webhookNeeds, struct {
			Name     string
			Current  string
			Required []string
			Target   *string
		}{
			Name:     "Disk Usage Notifications",
			Current:  config.SlackDiskWebhookURL,
			Required: []string{"disk"},
			Target:   &config.SlackDiskWebhookURL,
		})
	}

	if config.CPUEnabled || config.MemoryEnabled {
		var required []string
		if config.CPUEnabled {
			required = append(required, "CPU")
		}
		if config.MemoryEnabled {
			required = append(required, "memory")
		}

		webhookNeeds = append(webhookNeeds, struct {
			Name     string
			Current  string
			Required []string
			Target   *string
		}{
			Name:     fmt.Sprintf("CPU/Memory Notifications (%s)", strings.Join(required, ", ")),
			Current:  config.SlackCPUMemoryWebhookURL,
			Required: required,
			Target:   &config.SlackCPUMemoryWebhookURL,
		})
	}

	if len(webhookNeeds) == 0 {
		return fmt.Errorf("no monitoring enabled - cannot configure webhooks")
	}

	fmt.Println(blue("📝 Instructions:"))
	fmt.Println("  • Use ↑/↓ arrows to navigate")
	fmt.Println("  • Press Enter to configure selected webhook")
	fmt.Println("  • Enter valid Slack webhook URLs")
	fmt.Println("  • URLs must start with 'https://hooks.slack.com/'")
	fmt.Println()

	for {
		// Show current webhook status
		fmt.Println(bold("Webhook Configuration Status:"))
		allConfigured := true
		var items []string

		for i, webhook := range webhookNeeds {
			status := red("❌ Not configured")
			if webhook.Current != "" {
				status = green("✅ Configured")
			} else {
				allConfigured = false
			}

			display := fmt.Sprintf("%s - %s", webhook.Name, status)
			if webhook.Current != "" {
				// Show partial URL for confirmation
				partial := webhook.Current
				if len(partial) > 50 {
					partial = partial[:30] + "..." + partial[len(partial)-15:]
				}
				display += fmt.Sprintf(" (%s)", faint(partial))
			}
			items = append(items, display)
			_ = i
		}

		if allConfigured {
			items = append(items, green("── All webhooks configured - Continue ──"))
		} else {
			items = append(items, yellow("── Skip remaining (not recommended) ──"))
		}

		templates := &promptui.SelectTemplates{
			Label:    "{{ . }}",
			Active:   "👉 {{ . | cyan | bold }}",
			Inactive: "   {{ . | white }}",
			Selected: "{{ . | green }}",
		}

		prompt := promptui.Select{
			Label:     "Select webhook to configure",
			Items:     items,
			Templates: templates,
			Size:      len(items),
			HideHelp:  true,
		}

		index, _, err := prompt.Run()
		if err != nil {
			return err
		}

		// If user selected continue/skip
		if index >= len(webhookNeeds) {
			if allConfigured {
				break
			} else {
				// Confirm skipping
				confirmPrompt := promptui.Select{
					Label: "Some webhooks are not configured. Continue anyway?",
					Items: []string{"No, let me configure them", "Yes, skip for now"},
					Templates: &promptui.SelectTemplates{
						Label:    "{{ . }}",
						Active:   "👉 {{ . | cyan | bold }}",
						Inactive: "   {{ . | white }}",
						Selected: "{{ . | green }}",
					},
					HideHelp: true,
				}
				_, result, err := confirmPrompt.Run()
				if err != nil {
					return err
				}
				if result == "Yes, skip for now" {
					break
				}
				continue
			}
		}

		// Configure selected webhook
		webhook := &webhookNeeds[index]

		prompt2 := promptui.Prompt{
			Label:   fmt.Sprintf("Enter Slack webhook URL for %s", webhook.Name),
			Default: webhook.Current,
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

		result, err := prompt2.Run()
		if err != nil {
			return err
		}

		*webhook.Target = result
		webhook.Current = result

		fmt.Println(green("✅ Webhook configured successfully!"))
		fmt.Println()
	}

	fmt.Println(green("🎉 Slack webhook configuration complete!"))
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

	if config.DiskEnabled {
		if config.DiskThreshold == 0 {
			config.DiskThreshold = 80
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
			Current:     config.DiskThreshold,
			Default:     80,
			Target:      &config.DiskThreshold,
			Unit:        "%",
			Description: "Alert when disk usage exceeds this percentage",
		})
	}

	if config.CPUEnabled {
		if config.CPUThreshold == 0 {
			config.CPUThreshold = 85
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
			Current:     config.CPUThreshold,
			Default:     85,
			Target:      &config.CPUThreshold,
			Unit:        "%",
			Description: "Alert when CPU usage exceeds this percentage",
		})
	}

	if config.MemoryEnabled {
		if config.MemoryThreshold == 0 {
			config.MemoryThreshold = 85
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
			Current:     config.MemoryThreshold,
			Default:     85,
			Target:      &config.MemoryThreshold,
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

	// Always add CPU/Memory interval if either is enabled
	if config.CPUEnabled || config.MemoryEnabled {
		if config.CheckInterval == 0 {
			config.CheckInterval = 60
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
			Name:        "CPU/Memory Check Interval",
			Current:     config.CheckInterval,
			Default:     60,
			Target:      &config.CheckInterval,
			Unit:        "minutes",
			Description: "How often to check CPU and memory usage",
			Min:         1,
			Max:         1440, // 24 hours
		})
	}

	if config.DiskEnabled {
		if config.DiskCheckInterval == 0 {
			config.DiskCheckInterval = 12
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
			Current:     config.DiskCheckInterval,
			Default:     12,
			Target:      &config.DiskCheckInterval,
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
