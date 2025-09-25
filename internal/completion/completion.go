package completion

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"github.com/tmc/it2/internal/client"
)

// SessionIDCompletion provides completion for session IDs
func SessionIDCompletion(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	wsURL := cmd.Root().PersistentFlags().Lookup("url").Value.String()
	timeout, _ := time.ParseDuration(cmd.Root().PersistentFlags().Lookup("timeout").Value.String())

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	c := client.New(wsURL)
	err := c.Connect(ctx)
	if err != nil {
		return nil, cobra.ShellCompDirectiveError
	}
	defer c.Close()

	sessions, err := c.ListSessions(ctx)
	if err != nil {
		return nil, cobra.ShellCompDirectiveError
	}

	var completions []string
	for _, session := range sessions {
		if strings.HasPrefix(session.SessionID, toComplete) {
			desc := session.SessionID
			if session.SessionName != "" {
				desc = fmt.Sprintf("%s\t%s", session.SessionID, session.SessionName)
			}
			completions = append(completions, desc)
		}
	}

	return completions, cobra.ShellCompDirectiveNoFileComp
}

// WindowIDCompletion provides completion for window IDs
func WindowIDCompletion(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	wsURL := cmd.Root().PersistentFlags().Lookup("url").Value.String()
	timeout, _ := time.ParseDuration(cmd.Root().PersistentFlags().Lookup("timeout").Value.String())

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	c := client.New(wsURL)
	err := c.Connect(ctx)
	if err != nil {
		return nil, cobra.ShellCompDirectiveError
	}
	defer c.Close()

	windows, err := c.ListWindows(ctx)
	if err != nil {
		return nil, cobra.ShellCompDirectiveError
	}

	var completions []string
	for _, window := range windows {
		if strings.HasPrefix(window.WindowID, toComplete) {
			desc := window.WindowID
			if window.Title != "" {
				desc = fmt.Sprintf("%s\t%s", window.WindowID, window.Title)
			}
			completions = append(completions, desc)
		}
	}

	return completions, cobra.ShellCompDirectiveNoFileComp
}

// TabIDCompletion provides completion for tab IDs
func TabIDCompletion(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	wsURL := cmd.Root().PersistentFlags().Lookup("url").Value.String()
	timeout, _ := time.ParseDuration(cmd.Root().PersistentFlags().Lookup("timeout").Value.String())

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	c := client.New(wsURL)
	err := c.Connect(ctx)
	if err != nil {
		return nil, cobra.ShellCompDirectiveError
	}
	defer c.Close()

	// Get sessions and extract unique tab IDs
	sessions, err := c.ListSessions(ctx)
	if err != nil {
		return nil, cobra.ShellCompDirectiveError
	}

	seen := make(map[string]bool)
	var completions []string
	for _, session := range sessions {
		if session.TabID != "" && !seen[session.TabID] && strings.HasPrefix(session.TabID, toComplete) {
			seen[session.TabID] = true
			completions = append(completions, session.TabID)
		}
	}

	return completions, cobra.ShellCompDirectiveNoFileComp
}

// ProfileCompletion provides completion for profile names
func ProfileCompletion(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	wsURL := cmd.Root().PersistentFlags().Lookup("url").Value.String()
	timeout, _ := time.ParseDuration(cmd.Root().PersistentFlags().Lookup("timeout").Value.String())

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	c := client.New(wsURL)
	err := c.Connect(ctx)
	if err != nil {
		return nil, cobra.ShellCompDirectiveError
	}
	defer c.Close()

	profiles, err := c.ListProfiles(ctx, false)
	if err != nil {
		return nil, cobra.ShellCompDirectiveError
	}

	var completions []string
	for _, profile := range profiles {
		if strings.HasPrefix(profile, toComplete) {
			completions = append(completions, profile)
		}
	}

	return completions, cobra.ShellCompDirectiveNoFileComp
}

// FormatCompletion provides completion for output formats
func FormatCompletion(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	formats := []string{"table", "text", "json", "yaml"}

	var completions []string
	for _, format := range formats {
		if strings.HasPrefix(format, toComplete) {
			completions = append(completions, format)
		}
	}

	return completions, cobra.ShellCompDirectiveNoFileComp
}