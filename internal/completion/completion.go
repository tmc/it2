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

// WindowIDCompletion provides completion for window IDs (both indices and UUIDs)
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

	// Get sessions to build window mapping
	sessions, err := c.ListSessions(ctx)
	if err != nil {
		return nil, cobra.ShellCompDirectiveError
	}

	// Build unique window info
	windowMap := make(map[string]*WindowCompletionInfo)
	for _, session := range sessions {
		if _, exists := windowMap[session.WindowID]; !exists {
			windowMap[session.WindowID] = &WindowCompletionInfo{
				UUID:  session.WindowID,
				Index: fmt.Sprintf("%d", session.WindowNumber),
				Title: session.WindowTitle,
			}
		}
	}

	var completions []string
	for _, winInfo := range windowMap {
		// Add index completion (preferred for hierarchical commands)
		if strings.HasPrefix(winInfo.Index, toComplete) {
			desc := winInfo.Index
			if winInfo.Title != "" {
				desc = fmt.Sprintf("%s\tWindow %s: %s", winInfo.Index, winInfo.Index, winInfo.Title)
			}
			completions = append(completions, desc)
		}
		// Add UUID completion (for advanced users)
		if strings.HasPrefix(winInfo.UUID, toComplete) {
			desc := winInfo.UUID
			if winInfo.Title != "" {
				desc = fmt.Sprintf("%s\t%s", winInfo.UUID, winInfo.Title)
			}
			completions = append(completions, desc)
		}
	}

	return completions, cobra.ShellCompDirectiveNoFileComp
}

// WindowCompletionInfo holds window completion data
type WindowCompletionInfo struct {
	UUID  string
	Index string
	Title string
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

// HierarchicalWindowCompletion provides context-aware completion for hierarchical window commands
func HierarchicalWindowCompletion(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	// args[0] would be the window ID, we're completing the subcommand
	if len(args) == 0 {
		// Completing window ID
		return WindowIDCompletion(cmd, args, toComplete)
	}

	if len(args) == 1 {
		// Completing window subcommand
		subcommands := []string{"tab", "session", "create"}
		var completions []string
		for _, subcommand := range subcommands {
			if strings.HasPrefix(subcommand, toComplete) {
				var desc string
				switch subcommand {
				case "tab":
					desc = fmt.Sprintf("%s\tNavigate to tab or list tabs", subcommand)
				case "session":
					desc = fmt.Sprintf("%s\tList sessions in this window", subcommand)
				case "create":
					desc = fmt.Sprintf("%s\tCreate new tab in this window", subcommand)
				}
				completions = append(completions, desc)
			}
		}
		return completions, cobra.ShellCompDirectiveNoFileComp
	}

	// Handle deeper hierarchical completion
	if len(args) >= 2 && args[1] == "tab" {
		return HierarchicalTabCompletion(cmd, args, toComplete)
	}

	return nil, cobra.ShellCompDirectiveNoFileComp
}

// HierarchicalTabCompletion provides context-aware completion for tab operations in window context
func HierarchicalTabCompletion(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	// args: [windowID, "tab", ...]
	if len(args) == 2 {
		// Completing: it2 window <id> tab <TAB>
		// Could be tab ID or "list" command
		completions := []string{"list\tList tabs in this window"}

		// Add actual tab IDs for this window
		windowID := args[0]
		tabIDs := getTabIDsForWindow(cmd, windowID)
		for _, tabID := range tabIDs {
			if strings.HasPrefix(tabID, toComplete) {
				completions = append(completions, fmt.Sprintf("%s\tNavigate to tab %s", tabID, tabID))
			}
		}

		return completions, cobra.ShellCompDirectiveNoFileComp
	}

	if len(args) == 3 {
		// Completing: it2 window <id> tab <tabID> <TAB>
		// Tab subcommands
		subcommands := []string{"session", "split"}
		var completions []string
		for _, subcommand := range subcommands {
			if strings.HasPrefix(subcommand, toComplete) {
				var desc string
				switch subcommand {
				case "session":
					desc = fmt.Sprintf("%s\tList sessions in this tab", subcommand)
				case "split":
					desc = fmt.Sprintf("%s\tSplit current session in this tab", subcommand)
				}
				completions = append(completions, desc)
			}
		}
		return completions, cobra.ShellCompDirectiveNoFileComp
	}

	if len(args) == 4 && args[3] == "session" {
		// Completing: it2 window <id> tab <tabID> session <TAB>
		subcommands := []string{
			"list\tList sessions in this tab",
			"focus\tFocus a specific session",
			"close\tClose a specific session",
			"send-text\tSend text to a session",
			"split\tSplit a session",
		}
		var completions []string
		for _, subcommand := range subcommands {
			if strings.HasPrefix(strings.Split(subcommand, "\t")[0], toComplete) {
				completions = append(completions, subcommand)
			}
		}
		return completions, cobra.ShellCompDirectiveNoFileComp
	}

	if len(args) == 5 && args[3] == "session" {
		// Completing: it2 window <id> tab <tabID> session <operation> <TAB>
		sessionOp := args[4]
		windowID := args[0]
		tabID := args[2]

		switch sessionOp {
		case "focus", "close", "send-text", "split":
			// These operations need a session ID - provide context-filtered session completion
			return getSessionIDsForWindowTab(cmd, windowID, tabID, toComplete)
		case "list":
			// list doesn't need additional args
			return nil, cobra.ShellCompDirectiveNoFileComp
		}
	}

	return nil, cobra.ShellCompDirectiveNoFileComp
}

// getTabIDsForWindow returns tab IDs for a specific window
func getTabIDsForWindow(cmd *cobra.Command, windowID string) []string {
	wsURL := cmd.Root().PersistentFlags().Lookup("url").Value.String()
	timeout, _ := time.ParseDuration(cmd.Root().PersistentFlags().Lookup("timeout").Value.String())

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	c := client.New(wsURL)
	if err := c.Connect(ctx); err != nil {
		return nil
	}
	defer c.Close()

	sessions, err := c.ListSessions(ctx)
	if err != nil {
		return nil
	}

	// Resolve window ID if it's an index
	resolvedWindowID := resolveWindowID(sessions, windowID)

	// Get unique tab IDs for this window
	seen := make(map[string]bool)
	var tabIDs []string
	for _, session := range sessions {
		if session.WindowID == resolvedWindowID && !seen[session.TabID] {
			seen[session.TabID] = true
			tabIDs = append(tabIDs, session.TabID)
		}
	}

	return tabIDs
}

// getSessionIDsForWindowTab returns session IDs for a specific window and tab combination
func getSessionIDsForWindowTab(cmd *cobra.Command, windowID, tabID, toComplete string) ([]string, cobra.ShellCompDirective) {
	wsURL := cmd.Root().PersistentFlags().Lookup("url").Value.String()
	timeout, _ := time.ParseDuration(cmd.Root().PersistentFlags().Lookup("timeout").Value.String())

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	c := client.New(wsURL)
	if err := c.Connect(ctx); err != nil {
		return nil, cobra.ShellCompDirectiveError
	}
	defer c.Close()

	sessions, err := c.ListSessions(ctx)
	if err != nil {
		return nil, cobra.ShellCompDirectiveError
	}

	// Resolve window ID if it's an index
	resolvedWindowID := resolveWindowID(sessions, windowID)

	var completions []string
	for _, session := range sessions {
		// Filter by window and tab
		if session.WindowID == resolvedWindowID && session.TabID == tabID {
			if strings.HasPrefix(session.SessionID, toComplete) {
				desc := session.SessionID
				if session.SessionName != "" {
					desc = fmt.Sprintf("%s\t%s", session.SessionID, session.SessionName)
				}
				completions = append(completions, desc)
			}
		}
	}

	return completions, cobra.ShellCompDirectiveNoFileComp
}

// resolveWindowID converts window index to UUID if needed
func resolveWindowID(sessions []*client.SessionInfo, windowIDOrIndex string) string {
	// If already a UUID, return as-is
	if strings.HasPrefix(windowIDOrIndex, "pty-") {
		return windowIDOrIndex
	}

	// Try to resolve index to UUID
	for _, session := range sessions {
		if fmt.Sprintf("%d", session.WindowNumber) == windowIDOrIndex {
			return session.WindowID
		}
	}

	return windowIDOrIndex
}
