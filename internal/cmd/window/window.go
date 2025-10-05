package window

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/spf13/cobra"
	"github.com/tmc/it2/internal/cmdcore"
	"github.com/tmc/it2/internal/cmdutil"
	"github.com/tmc/it2/internal/completion"
	"github.com/tmc/it2/internal/sessionid"
)

// isWindowUUID checks if a string looks like a window UUID (starts with "pty-")
func isWindowUUID(s string) bool {
	return strings.HasPrefix(s, "pty-") && len(s) > 4
}

// NewCommand creates the window command with support for both flat and hierarchical modes.
func NewCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "window [<window-id>] [command]",
		Short:   "Manage iTerm2 windows or navigate to window context",
		GroupID: "core",
		Long: `Manage iTerm2 windows with support for both flat and hierarchical command styles.

Flat Command Examples:
  it2 window list                   # List all windows
  it2 window create                 # Create new window
  it2 window close 1                # Close window 1

Hierarchical Command Examples:
  it2 window 1 tab list            # List tabs in window 1
  it2 window 1 session list        # List sessions in window 1
  it2 window 1 tab 2 session list  # List sessions in tab 2 of window 1

The hierarchical style provides context-aware navigation while maintaining
full compatibility with existing flat commands.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			// If no args, show help
			if len(args) == 0 {
				return cmd.Help()
			}

			// Check if first arg is a subcommand
			// This allows subcommands like "current" to be handled by cobra
			for _, subCmd := range cmd.Commands() {
				if subCmd.Name() == args[0] || subCmd.HasAlias(args[0]) {
					// Let cobra handle the subcommand
					return nil
				}
			}

			// Check if first arg is a number (window index for hierarchical mode)
			if windowID, err := strconv.Atoi(args[0]); err == nil {
				// Hierarchical mode with numeric index: it2 window <index> [subcommand]
				return handleHierarchicalWindow(cmd, fmt.Sprintf("%d", windowID), args[1:])
			}

			// Check if first arg looks like a window UUID (for hierarchical mode)
			if isWindowUUID(args[0]) {
				// Hierarchical mode with UUID: it2 window <uuid> [subcommand]
				return handleHierarchicalWindow(cmd, args[0], args[1:])
			}

			// If we get here, it's likely a flat command that wasn't matched
			// Let cobra's normal subcommand resolution handle it
			return fmt.Errorf("unknown subcommand or invalid window ID: %s\nRun 'it2 window --help' for usage", args[0])
		},
	}

	// Define command groups
	cmd.AddGroup(&cobra.Group{
		ID:    "management",
		Title: "Window Management Commands:",
	})
	cmd.AddGroup(&cobra.Group{
		ID:    "display",
		Title: "Display & Appearance Commands:",
	})
	cmd.AddGroup(&cobra.Group{
		ID:    "properties",
		Title: "Property Commands:",
	})
	cmd.AddGroup(&cobra.Group{
		ID:    "layout",
		Title: "Layout Commands:",
	})

	// Window Management Commands
	listCmd := newListCommand()
	listCmd.GroupID = "management"
	cmd.AddCommand(listCmd)

	createCmd := newCreateCommand()
	createCmd.GroupID = "management"
	cmd.AddCommand(createCmd)

	closeCmd := newCloseCommand()
	closeCmd.GroupID = "management"
	cmd.AddCommand(closeCmd)

	focusCmd := newFocusCommand()
	focusCmd.GroupID = "management"
	cmd.AddCommand(focusCmd)

	currentCmd := newCurrentCommand()
	currentCmd.GroupID = "management"
	cmd.AddCommand(currentCmd)

	// Display & Appearance Commands
	getTitleCmd := newGetTitleCommand()
	getTitleCmd.GroupID = "display"
	cmd.AddCommand(getTitleCmd)

	setTitleCmd := newSetTitleCommand()
	setTitleCmd.GroupID = "display"
	cmd.AddCommand(setTitleCmd)

	clearTitleCmd := newClearTitleCommand()
	clearTitleCmd.GroupID = "display"
	cmd.AddCommand(clearTitleCmd)

	// Property Commands
	getPropCmd := newGetPropertyCommand()
	getPropCmd.GroupID = "properties"
	cmd.AddCommand(getPropCmd)

	setPropCmd := newSetPropertyCommand()
	setPropCmd.GroupID = "properties"
	cmd.AddCommand(setPropCmd)

	// Layout Commands
	treeCmd := newTreeCommand()
	treeCmd.GroupID = "layout"
	cmd.AddCommand(treeCmd)

	// Add hierarchical completion
	cmd.ValidArgsFunction = completion.HierarchicalWindowCompletion

	return cmd
}

// handleHierarchicalWindow handles hierarchical window commands
func handleHierarchicalWindow(cmd *cobra.Command, windowID string, subArgs []string) error {
	if len(subArgs) == 0 {
		return showWindowContextHelp(windowID)
	}

	subcommand := subArgs[0]
	switch subcommand {
	case "tab":
		return handleWindowTabSubcommand(cmd, windowID, subArgs[1:])
	case "session":
		return handleWindowSessionSubcommand(cmd, windowID, subArgs[1:])
	case "create":
		return handleWindowCreateTab(cmd, windowID, subArgs[1:])
	default:
		return fmt.Errorf("unknown subcommand: %s\nRun 'it2 window %s' for available commands", subcommand, windowID)
	}
}

// showWindowContextHelp shows help for a specific window context
func showWindowContextHelp(windowID string) error {
	fmt.Printf("Window %s Context\n", windowID)
	fmt.Println(fmt.Sprintf("Available commands for window %s:", windowID))
	fmt.Println("")
	fmt.Println("  tab <id> [command]    Navigate to tab context")
	fmt.Println("  tab list              List tabs in this window")
	fmt.Println("  session list          List sessions in this window")
	fmt.Println("  create                Create new tab in this window")
	fmt.Println("")
	fmt.Println("Examples:")
	fmt.Printf("  it2 window %s tab list\n", windowID)
	fmt.Printf("  it2 window %s tab 1 session list\n", windowID)
	fmt.Printf("  it2 window %s session list\n", windowID)
	return nil
}

// handleWindowTabSubcommand handles tab operations in window context
func handleWindowTabSubcommand(cmd *cobra.Command, windowID string, tabArgs []string) error {
	if len(tabArgs) == 0 {
		return fmt.Errorf("tab command requires arguments\nRun 'it2 window %s' for help", windowID)
	}

	if tabArgs[0] == "list" {
		// List tabs in this window
		return executeListTabs(cmd, windowID)
	}

	// Parse tab ID
	tabID := tabArgs[0]
	if _, err := strconv.Atoi(tabID); err != nil {
		return fmt.Errorf("invalid tab ID: %s (must be a number)", tabID)
	}

	if len(tabArgs) == 1 {
		return showTabContextHelp(windowID, tabID)
	}

	// Handle tab subcommands
	return handleTabSubcommand(cmd, windowID, tabID, tabArgs[1:])
}

// handleWindowSessionSubcommand handles session operations in window context
func handleWindowSessionSubcommand(cmd *cobra.Command, windowID string, sessionArgs []string) error {
	if len(sessionArgs) == 0 {
		return fmt.Errorf("session command requires arguments\nRun 'it2 window %s' for help", windowID)
	}

	if sessionArgs[0] == "list" {
		// List sessions in this window
		return executeListSessions(cmd, windowID, "")
	}

	return fmt.Errorf("session subcommand '%s' not yet implemented", sessionArgs[0])
}

// handleWindowCreateTab handles creating a new tab in window context
func handleWindowCreateTab(cmd *cobra.Command, windowID string, createArgs []string) error {
	// TODO: Implement tab creation in specific window
	return fmt.Errorf("create tab in window %s not yet implemented", windowID)
}

// showTabContextHelp shows help for a specific tab context
func showTabContextHelp(windowID, tabID string) error {
	fmt.Printf("Tab %s Context (Window %s)\n", tabID, windowID)
	fmt.Printf("Available commands for tab %s in window %s:\n", tabID, windowID)
	fmt.Println("")
	fmt.Println("  session list          List sessions in this tab")
	fmt.Println("  session <id> [cmd]    Navigate to session context")
	fmt.Println("  split                 Split current session in this tab")
	fmt.Println("")
	fmt.Println("Examples:")
	fmt.Printf("  it2 window %s tab %s session list\n", windowID, tabID)
	fmt.Printf("  it2 window %s tab %s split --vertical\n", windowID, tabID)
	return nil
}

// handleTabSubcommand handles subcommands in tab context
func handleTabSubcommand(cmd *cobra.Command, windowID, tabID string, subArgs []string) error {
	if len(subArgs) == 0 {
		return showTabContextHelp(windowID, tabID)
	}

	subcommand := subArgs[0]
	switch subcommand {
	case "session":
		return handleTabSessionSubcommand(cmd, windowID, tabID, subArgs[1:])
	case "split":
		return handleTabSplit(cmd, windowID, tabID, subArgs[1:])
	default:
		return fmt.Errorf("unknown subcommand: %s\nRun 'it2 window %s tab %s' for available commands", subcommand, windowID, tabID)
	}
}

// handleTabSessionSubcommand handles session operations in tab context
func handleTabSessionSubcommand(cmd *cobra.Command, windowID, tabID string, sessionArgs []string) error {
	if len(sessionArgs) == 0 {
		return fmt.Errorf("session command requires arguments\nRun 'it2 window %s tab %s' for help", windowID, tabID)
	}

	operation := sessionArgs[0]
	switch operation {
	case "list":
		// List sessions in this tab
		return executeListSessions(cmd, windowID, tabID)
	case "focus":
		if len(sessionArgs) < 2 {
			return fmt.Errorf("focus requires a session ID\nUsage: it2 window %s tab %s session focus <session-id>", windowID, tabID)
		}
		return executeSessionFocus(cmd, sessionArgs[1])
	case "close":
		if len(sessionArgs) < 2 {
			return fmt.Errorf("close requires a session ID\nUsage: it2 window %s tab %s session close <session-id>", windowID, tabID)
		}
		return executeSessionClose(cmd, sessionArgs[1])
	case "send-text":
		if len(sessionArgs) < 3 {
			return fmt.Errorf("send-text requires a session ID and text\nUsage: it2 window %s tab %s session send-text <session-id> <text>", windowID, tabID)
		}
		return executeSessionSendText(cmd, sessionArgs[1], strings.Join(sessionArgs[2:], " "))
	case "split":
		if len(sessionArgs) < 2 {
			return fmt.Errorf("split requires a session ID\nUsage: it2 window %s tab %s session split <session-id> [--vertical|--horizontal]", windowID, tabID)
		}
		return executeSessionSplit(cmd, sessionArgs[1], sessionArgs[2:])
	default:
		return fmt.Errorf("unknown session operation: %s\nAvailable: list, focus, close, send-text, split", operation)
	}
}

// handleTabSplit handles splitting sessions in tab context
func handleTabSplit(cmd *cobra.Command, windowID, tabID string, splitArgs []string) error {
	// TODO: Implement session splitting in specific tab
	return fmt.Errorf("split in window %s tab %s not yet implemented", windowID, tabID)
}

// executeListTabs executes tab listing for a specific window
func executeListTabs(cmd *cobra.Command, windowID string) error {
	_, timeout, format := cmdcore.GetFlags(cmd)
	ctx, cancel := cmdcore.CreateContext(timeout)
	defer cancel()

	c, err := cmdcore.ConnectClient(ctx)
	if err != nil {
		return fmt.Errorf("failed to connect: %w", err)
	}
	defer c.Close()

	sharedOps := cmdutil.NewSharedListOperations(c, ctx)

	// Resolve window index to UUID if needed
	resolvedWindowID, err := sharedOps.ResolveWindowID(windowID)
	if err != nil {
		return fmt.Errorf("failed to resolve window ID: %w", err)
	}

	return sharedOps.ListTabs(cmdutil.SharedListOptions{
		WindowID: resolvedWindowID,
		Format:   format,
	})
}

// executeListSessions executes session listing with optional window/tab filtering
func executeListSessions(cmd *cobra.Command, windowID, tabID string) error {
	_, timeout, format := cmdcore.GetFlags(cmd)
	ctx, cancel := cmdcore.CreateContext(timeout)
	defer cancel()

	c, err := cmdcore.ConnectClient(ctx)
	if err != nil {
		return fmt.Errorf("failed to connect: %w", err)
	}
	defer c.Close()

	sharedOps := cmdutil.NewSharedListOperations(c, ctx)

	// Resolve window index to UUID if needed
	resolvedWindowID, err := sharedOps.ResolveWindowID(windowID)
	if err != nil {
		return fmt.Errorf("failed to resolve window ID: %w", err)
	}

	return sharedOps.ListSessions(cmdutil.SharedListOptions{
		WindowID: resolvedWindowID,
		TabID:    tabID,
		Format:   format,
	})
}

// executeSessionFocus focuses a specific session
func executeSessionFocus(cmd *cobra.Command, sessionID string) error {
	_, timeout, _ := cmdcore.GetFlags(cmd)
	ctx, cancel := cmdcore.CreateContext(timeout)
	defer cancel()

	c, err := cmdcore.ConnectClient(ctx)
	if err != nil {
		return fmt.Errorf("failed to connect: %w", err)
	}
	defer c.Close()

	sessionID = sessionid.Normalize(sessionID)
	if _, err := c.ActivateSession(ctx, sessionID, true); err != nil {
		return fmt.Errorf("failed to focus session %s: %w", sessionID, err)
	}

	fmt.Printf("Focused session: %s\n", sessionID)
	return nil
}

// executeSessionClose closes a specific session
func executeSessionClose(cmd *cobra.Command, sessionID string) error {
	_, timeout, _ := cmdcore.GetFlags(cmd)
	ctx, cancel := cmdcore.CreateContext(timeout)
	defer cancel()

	c, err := cmdcore.ConnectClient(ctx)
	if err != nil {
		return fmt.Errorf("failed to connect: %w", err)
	}
	defer c.Close()

	sessionID = sessionid.Normalize(sessionID)
	if _, err := c.CloseSessions(ctx, []string{sessionID}, false); err != nil {
		return fmt.Errorf("failed to close session %s: %w", sessionID, err)
	}

	fmt.Printf("Closed session: %s\n", sessionID)
	return nil
}

// executeSessionSendText sends text to a specific session
func executeSessionSendText(cmd *cobra.Command, sessionID, text string) error {
	_, timeout, _ := cmdcore.GetFlags(cmd)
	ctx, cancel := cmdcore.CreateContext(timeout)
	defer cancel()

	c, err := cmdcore.ConnectClient(ctx)
	if err != nil {
		return fmt.Errorf("failed to connect: %w", err)
	}
	defer c.Close()

	sessionID = sessionid.Normalize(sessionID)
	// Add newline for better UX
	if !strings.HasSuffix(text, "\n") {
		text += "\n"
	}

	if err := c.SendText(ctx, sessionID, text); err != nil {
		return fmt.Errorf("failed to send text to session %s: %w", sessionID, err)
	}

	fmt.Printf("Sent text to session: %s\n", sessionID)
	return nil
}

// executeSessionSplit splits a specific session
func executeSessionSplit(cmd *cobra.Command, sessionID string, args []string) error {
	_, timeout, _ := cmdcore.GetFlags(cmd)
	ctx, cancel := cmdcore.CreateContext(timeout)
	defer cancel()

	c, err := cmdcore.ConnectClient(ctx)
	if err != nil {
		return fmt.Errorf("failed to connect: %w", err)
	}
	defer c.Close()

	sessionID = sessionid.Normalize(sessionID)

	// Parse split direction (default horizontal)
	vertical := false
	for _, arg := range args {
		if arg == "--vertical" || arg == "-v" {
			vertical = true
		}
	}

	if _, err := c.SplitPane(ctx, sessionID, vertical, false, ""); err != nil {
		return fmt.Errorf("failed to split session %s: %w", sessionID, err)
	}

	direction := "horizontally"
	if vertical {
		direction = "vertically"
	}
	fmt.Printf("Split session %s %s\n", sessionID, direction)
	return nil
}
