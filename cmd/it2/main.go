package main

import (
	"log"
	"time"

	"github.com/spf13/cobra"
	"github.com/tmc/it2/internal/cmd/app"
	"github.com/tmc/it2/internal/cmd/auth"
	"github.com/tmc/it2/internal/cmd/job"
	"github.com/tmc/it2/internal/cmd/notification"
	"github.com/tmc/it2/internal/cmd/profile"
	"github.com/tmc/it2/internal/cmd/session"
	"github.com/tmc/it2/internal/cmd/tab"
	"github.com/tmc/it2/internal/cmd/text"
	"github.com/tmc/it2/internal/cmd/variable"
	"github.com/tmc/it2/internal/cmd/window"
)

var (
	wsURL   string
	timeout time.Duration
	format  string
)

var rootCmd = &cobra.Command{
	Use:   "it2",
	Short: "iTerm2 API command-line client",
	Long:  `A command-line interface for interacting with iTerm2's API`,
}


func init() {
	rootCmd.PersistentFlags().StringVar(&wsURL, "url", "ws://localhost:1912", "WebSocket URL for iTerm2 API")
	rootCmd.PersistentFlags().DurationVar(&timeout, "timeout", 5*time.Second, "Connection timeout")
	rootCmd.PersistentFlags().StringVar(&format, "format", "text", "Output format (text, json, yaml)")

	// Add organized command groups
	rootCmd.AddCommand(app.NewCommand())
	rootCmd.AddCommand(session.NewCommand())
	rootCmd.AddCommand(tab.NewCommand())
	rootCmd.AddCommand(window.NewCommand())
	rootCmd.AddCommand(text.NewCommand())
	rootCmd.AddCommand(profile.NewCommand())
	rootCmd.AddCommand(variable.NewCommand())
	rootCmd.AddCommand(job.NewCommand())
	rootCmd.AddCommand(notification.NewCommand())
	rootCmd.AddCommand(auth.NewCommand())
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		log.Fatal(err)
	}
}