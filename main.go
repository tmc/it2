package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/spf13/cobra"
	"github.com/tmc/it2/internal/auth"
	"github.com/tmc/it2/internal/client"
	"github.com/tmc/it2/internal/formatting"
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

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List all iTerm2 sessions",
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx, cancel := context.WithTimeout(context.Background(), timeout)
		defer cancel()

		c := client.New(wsURL)
		if err := c.Connect(ctx); err != nil {
			return fmt.Errorf("failed to connect: %w", err)
		}
		defer c.Close()

		sessions, err := c.ListSessions(ctx)
		if err != nil {
			return fmt.Errorf("failed to list sessions: %w", err)
		}

		formatter := formatting.New(format)
		return formatter.FormatSessions(sessions)
	},
}

var sendCmd = &cobra.Command{
	Use:   "send <session-id> <text>",
	Short: "Send text to a session",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx, cancel := context.WithTimeout(context.Background(), timeout)
		defer cancel()

		c := client.New(wsURL)
		if err := c.Connect(ctx); err != nil {
			return fmt.Errorf("failed to connect: %w", err)
		}
		defer c.Close()

		if err := c.SendText(ctx, args[0], args[1]); err != nil {
			return fmt.Errorf("failed to send text: %w", err)
		}

		fmt.Println("Text sent successfully")
		return nil
	},
}

var createTabCmd = &cobra.Command{
	Use:   "create-tab",
	Short: "Create a new tab",
	RunE: func(cmd *cobra.Command, args []string) error {
		profileName, _ := cmd.Flags().GetString("profile")
		windowID, _ := cmd.Flags().GetString("window")

		ctx, cancel := context.WithTimeout(context.Background(), timeout)
		defer cancel()

		c := client.New(wsURL)
		if err := c.Connect(ctx); err != nil {
			return fmt.Errorf("failed to connect: %w", err)
		}
		defer c.Close()

		resp, err := c.CreateTab(ctx, profileName, windowID)
		if err != nil {
			return fmt.Errorf("failed to create tab: %w", err)
		}

		formatter := formatting.New(format)
		return formatter.FormatTabResponse(resp)
	},
}

var bufferCmd = &cobra.Command{
	Use:   "buffer <session-id>",
	Short: "Get the buffer contents of a session",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		lines, _ := cmd.Flags().GetInt32("lines")

		ctx, cancel := context.WithTimeout(context.Background(), timeout)
		defer cancel()

		c := client.New(wsURL)
		if err := c.Connect(ctx); err != nil {
			return fmt.Errorf("failed to connect: %w", err)
		}
		defer c.Close()

		resp, err := c.GetBuffer(ctx, args[0], lines)
		if err != nil {
			return fmt.Errorf("failed to get buffer: %w", err)
		}

		formatter := formatting.New(format)
		return formatter.FormatBuffer(resp)
	},
}

var jobsCmd = &cobra.Command{
	Use:   "jobs <session-id>",
	Short: "Get running jobs in a session",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx, cancel := context.WithTimeout(context.Background(), timeout)
		defer cancel()

		c := client.New(wsURL)
		if err := c.Connect(ctx); err != nil {
			return fmt.Errorf("failed to connect: %w", err)
		}
		defer c.Close()

		jobs, err := c.GetJobs(ctx, args[0])
		if err != nil {
			return fmt.Errorf("failed to get jobs: %w", err)
		}

		formatter := formatting.New(format)
		return formatter.FormatJobs(jobs)
	},
}

var authCmd = &cobra.Command{
	Use:   "auth",
	Short: "Manage iTerm2 API authentication",
}

var authRequestCmd = &cobra.Command{
	Use:   "request",
	Short: "Request authentication from iTerm2",
	RunE: func(cmd *cobra.Command, args []string) error {
		return auth.RequestAuthentication()
	},
}

var authCheckCmd = &cobra.Command{
	Use:   "check",
	Short: "Check if authentication is configured",
	Run: func(cmd *cobra.Command, args []string) {
		if auth.HasAuthentication() {
			fmt.Println("Authentication is configured")
			if cookie := os.Getenv("ITERM2_COOKIE"); cookie != "" {
				fmt.Printf("Cookie: %s...\n", cookie[:min(len(cookie), 10)])
			}
			if key := os.Getenv("ITERM2_KEY"); key != "" {
				fmt.Printf("Key: %s...\n", key[:min(len(key), 10)])
			}
		} else {
			fmt.Println("Authentication is not configured")
			fmt.Println("Run 'it2 auth request' to request authentication")
		}
	},
}

func init() {
	rootCmd.PersistentFlags().StringVar(&wsURL, "url", "ws://localhost:1912", "WebSocket URL for iTerm2 API")
	rootCmd.PersistentFlags().DurationVar(&timeout, "timeout", 5*time.Second, "Connection timeout")
	rootCmd.PersistentFlags().StringVar(&format, "format", "text", "Output format (text, json, yaml)")

	createTabCmd.Flags().String("profile", "Default", "Profile name to use")
	createTabCmd.Flags().String("window", "", "Window ID (empty for current window)")

	bufferCmd.Flags().Int32("lines", 100, "Number of lines to retrieve")

	authCmd.AddCommand(authRequestCmd)
	authCmd.AddCommand(authCheckCmd)

	rootCmd.AddCommand(listCmd)
	rootCmd.AddCommand(sendCmd)
	rootCmd.AddCommand(createTabCmd)
	rootCmd.AddCommand(bufferCmd)
	rootCmd.AddCommand(jobsCmd)
	rootCmd.AddCommand(authCmd)
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		log.Fatal(err)
	}
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}