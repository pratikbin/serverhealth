package main

import (
	"fmt"
	"os"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

const (
	appName = "serverhealth"
	version = "1.0.4"
)

var (
	green  = color.New(color.FgGreen).SprintFunc()
	red    = color.New(color.FgRed).SprintFunc()
	yellow = color.New(color.FgYellow).SprintFunc()
	blue   = color.New(color.FgBlue).SprintFunc()
	bold   = color.New(color.Bold).SprintFunc()
	faint  = color.New(color.Faint).SprintFunc()
)

func main() {
	var rootCmd = &cobra.Command{
		Use:     appName,
		Short:   "A comprehensive server health monitoring tool",
		Long:    `ServerHealth is a CLI tool that monitors server health metrics and sends notifications to Slack.`,
		Version: version,
	}

	// Add all commands
	rootCmd.AddCommand(
		NewConfigureCmd(),
		NewStartCmd(),
		NewStatusCmd(),
		NewStopCmd(),
		NewInstallCmd(),
		NewUninstallCmd(),
		NewLogsCmd(),
		NewDaemonCmd(),
	)

	if err := rootCmd.Execute(); err != nil {
		fmt.Println(red("Error:"), err)
		os.Exit(1)
	}
}
