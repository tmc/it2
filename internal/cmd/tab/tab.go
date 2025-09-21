package tab

import (
	"fmt"
	"strconv"

	"github.com/spf13/cobra"
	"github.com/tmc/it2/internal/client"
	"github.com/tmc/it2/internal/cmdutil"
	"github.com/tmc/it2/internal/formatting"
	_ "github.com/tmc/it2/proto" // for future use
)

// NewCommand creates the tab command with all subcommands.
func NewCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "tab",
		Short: "Manage iTerm2 tabs",
		Long:  "Commands for creating, closing, and managing iTerm2 tabs",
	}

	cmd.AddCommand(newListCommand())
	cmd.AddCommand(newCreateCommand())
	cmd.AddCommand(newCloseCommand())
	cmd.AddCommand(newActivateCommand())
	cmd.AddCommand(newReorderCommand())
	cmd.AddCommand(newSetTitleCommand())
	cmd.AddCommand(newSetLayoutCommand())
	cmd.AddCommand(newGetInfoCommand())

	return cmd
}

func newListCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list [window-id]",
		Short: "List tabs in windows",
		Long:  "List tabs in a specific window or all windows if no window-id provided",
		Args:  cobra.MaximumNArgs(1),
		RunE:  runListCommand,
	}

	return cmd
}

func runListCommand(cmd *cobra.Command, args []string) error {
	wsURL, timeout, format := cmdutil.GetFlags(cmd)
	ctx, cancel := cmdutil.CreateContext(timeout)
	defer cancel()

	c, err := cmdutil.ConnectClient(ctx, wsURL)
	if err != nil {
		return fmt.Errorf("failed to connect: %w", err)
	}
	defer c.Close()

	// Check if window ID is specified
	var targetWindowID string
	if len(args) > 0 {
		targetWindowID = args[0]
	}

	// Get all sessions which include tab information
	sessions, err := c.ListSessions(ctx)
	if err != nil {
		return fmt.Errorf("failed to list sessions: %w", err)
	}

	// Group sessions by tab and window
	tabMap := make(map[string]*formatting.TabInfo)
	for _, session := range sessions {
		if session.TabID == "" {
			// Skip sessions not associated with a tab
			continue
		}

		// Filter by window ID if specified
		if targetWindowID != "" && session.WindowID != targetWindowID {
			continue
		}

		tab, exists := tabMap[session.TabID]
		if !exists {
			tab = &formatting.TabInfo{
				TabID:    session.TabID,
				WindowID: session.WindowID,
				Sessions: make([]*client.SessionInfo, 0),
			}
			tabMap[session.TabID] = tab
		}
		tab.Sessions = append(tab.Sessions, session)
	}

	// Convert map to slice for consistent ordering
	tabs := make([]*formatting.TabInfo, 0, len(tabMap))
	for _, tab := range tabMap {
		tabs = append(tabs, tab)
	}

	// If targeting specific window but no tabs found, check if window exists
	if targetWindowID != "" && len(tabs) == 0 {
		windows, err := c.ListWindows(ctx)
		if err == nil {
			found := false
			for _, window := range windows {
				if window.WindowID == targetWindowID {
					found = true
					break
				}
			}
			if !found {
				return fmt.Errorf("window not found: %s", targetWindowID)
			}
		}
		return fmt.Errorf("no tabs found in window: %s", targetWindowID)
	}

	// Format and display results
	formatter := formatting.New(format)
	return formatter.FormatTabs(tabs)
}

func newCreateCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create <profile> [window-id]",
		Short: "Create a new tab",
		Long:  "Create a new tab with specified profile in a window (uses default window if not specified)",
		Args:  cobra.RangeArgs(1, 2),
		RunE:  runCreateCommand,
	}

	cmd.Flags().Uint32("index", 0, "Tab index position (0-based, appends to end if not specified)")
	cmd.Flags().String("command", "", "Initial command to run in the new tab")

	return cmd
}

func runCreateCommand(cmd *cobra.Command, args []string) error {
	wsURL, timeout, format := cmdutil.GetFlags(cmd)
	ctx, cancel := cmdutil.CreateContext(timeout)
	defer cancel()

	c, err := cmdutil.ConnectClient(ctx, wsURL)
	if err != nil {
		return fmt.Errorf("failed to connect: %w", err)
	}
	defer c.Close()

	profile := args[0]
	var windowID string
	if len(args) > 1 {
		windowID = args[1]
		// Validate window exists if specified
		windows, err := c.ListWindows(ctx)
		if err != nil {
			return fmt.Errorf("failed to validate window: %w", err)
		}
		found := false
		for _, window := range windows {
			if window.WindowID == windowID {
				found = true
				break
			}
		}
		if !found {
			return fmt.Errorf("window not found: %s", windowID)
		}
	}

	index, _ := cmd.Flags().GetUint32("index")
	initialCommand, _ := cmd.Flags().GetString("command")

	response, err := c.CreateTabWithOptions(ctx, profile, windowID, index, initialCommand)
	if err != nil {
		return fmt.Errorf("failed to create tab: %w", err)
	}

	// Format the response
	formatter := formatting.New(format)
	return formatter.FormatTabResponse(response)
}

func newCloseCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "close <tab-id> [tab-id...]",
		Short: "Close one or more tabs",
		Long:  "Close one or more tabs by their IDs",
		Args:  cobra.MinimumNArgs(1),
		RunE:  runCloseCommand,
	}

	cmd.Flags().Bool("force", false, "Force close the tabs")
	return cmd
}

func runCloseCommand(cmd *cobra.Command, args []string) error {
	wsURL, timeout, _ := cmdutil.GetFlags(cmd)
	ctx, cancel := cmdutil.CreateContext(timeout)
	defer cancel()

	c, err := cmdutil.ConnectClient(ctx, wsURL)
	if err != nil {
		return fmt.Errorf("failed to connect: %w", err)
	}
	defer c.Close()

	force, _ := cmd.Flags().GetBool("force")

	_, err = c.CloseTabs(ctx, args, force)
	if err != nil {
		return fmt.Errorf("failed to close tabs: %w", err)
	}

	if len(args) == 1 {
		fmt.Printf("Successfully closed tab: %s\n", args[0])
	} else {
		fmt.Printf("Successfully closed %d tabs\n", len(args))
	}

	return nil
}

func newActivateCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "activate <tab-id>",
		Short: "Activate a tab",
		Long:  "Activate and optionally select a tab",
		Args:  cobra.ExactArgs(1),
		RunE:  runActivateCommand,
	}

	cmd.Flags().Bool("select-tab", true, "Select the tab after activation")
	return cmd
}

func runActivateCommand(cmd *cobra.Command, args []string) error {
	wsURL, timeout, _ := cmdutil.GetFlags(cmd)
	ctx, cancel := cmdutil.CreateContext(timeout)
	defer cancel()

	c, err := cmdutil.ConnectClient(ctx, wsURL)
	if err != nil {
		return fmt.Errorf("failed to connect: %w", err)
	}
	defer c.Close()

	tabID := args[0]
	selectTab, _ := cmd.Flags().GetBool("select-tab")

	_, err = c.ActivateTab(ctx, tabID, selectTab)
	if err != nil {
		return fmt.Errorf("failed to activate tab: %w", err)
	}

	fmt.Printf("Successfully activated tab: %s\n", tabID)
	return nil
}

func newReorderCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "reorder <tab-id> <new-index>",
		Short: "Reorder tabs within a window",
		Long:  "Reorder tabs by specifying new position for a tab within its window",
		Args:  cobra.ExactArgs(2),
		RunE:  runReorderCommand,
	}

	return cmd
}

func runReorderCommand(cmd *cobra.Command, args []string) error {
	wsURL, timeout, _ := cmdutil.GetFlags(cmd)
	ctx, cancel := cmdutil.CreateContext(timeout)
	defer cancel()

	c, err := cmdutil.ConnectClient(ctx, wsURL)
	if err != nil {
		return fmt.Errorf("failed to connect: %w", err)
	}
	defer c.Close()

	tabID := args[0]
	newIndex, err := strconv.Atoi(args[1])
	if err != nil {
		return fmt.Errorf("invalid index: %w", err)
	}

	if newIndex < 0 {
		return fmt.Errorf("index must be non-negative")
	}

	// Find the window ID for this tab
	sessions, err := c.ListSessions(ctx)
	if err != nil {
		return fmt.Errorf("failed to list sessions: %w", err)
	}

	var windowID string
	for _, session := range sessions {
		if session.TabID == tabID {
			windowID = session.WindowID
			break
		}
	}

	if windowID == "" {
		return fmt.Errorf("tab not found: %s", tabID)
	}

	// Get all tabs in the window to build reorder list
	tabMap := make(map[string]*formatting.TabInfo)
	for _, session := range sessions {
		if session.TabID != "" && session.WindowID == windowID {
			tab, exists := tabMap[session.TabID]
			if !exists {
				tab = &formatting.TabInfo{
					TabID:    session.TabID,
					WindowID: session.WindowID,
				}
				tabMap[session.TabID] = tab
			}
		}
	}

	// Build ordered list of tab IDs
	var tabIDs []string
	for _, tab := range tabMap {
		if tab.TabID != tabID {
			tabIDs = append(tabIDs, tab.TabID)
		}
	}

	// Insert the target tab at the new position
	if newIndex >= len(tabIDs) {
		tabIDs = append(tabIDs, tabID)
	} else {
		// Insert at specific position
		newTabIDs := make([]string, 0, len(tabIDs)+1)
		newTabIDs = append(newTabIDs, tabIDs[:newIndex]...)
		newTabIDs = append(newTabIDs, tabID)
		newTabIDs = append(newTabIDs, tabIDs[newIndex:]...)
		tabIDs = newTabIDs
	}

	// Perform the reorder
	assignments := map[string][]string{
		windowID: tabIDs,
	}

	_, err = c.ReorderTabs(ctx, assignments)
	if err != nil {
		return fmt.Errorf("failed to reorder tabs: %w", err)
	}

	fmt.Printf("Successfully moved tab %s to position %d in window %s\n", tabID, newIndex, windowID)
	return nil
}

func newSetLayoutCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "set-layout <tab-id> <layout>",
		Short: "Set the layout of a tab",
		Long:  "Set the split layout of a tab using predefined layouts (single, split, grid)",
		Args:  cobra.ExactArgs(2),
		RunE:  runSetLayoutCommand,
	}

	cmd.Flags().Bool("horizontal", false, "Split horizontally (default is vertical)")

	return cmd
}

func runSetLayoutCommand(cmd *cobra.Command, args []string) error {
	wsURL, timeout, _ := cmdutil.GetFlags(cmd)
	ctx, cancel := cmdutil.CreateContext(timeout)
	defer cancel()

	c, err := cmdutil.ConnectClient(ctx, wsURL)
	if err != nil {
		return fmt.Errorf("failed to connect: %w", err)
	}
	defer c.Close()

	tabID := args[0]
	layout := args[1]

	// Validate tab exists
	sessions, err := c.ListSessions(ctx)
	if err != nil {
		return fmt.Errorf("failed to list sessions: %w", err)
	}

	tabFound := false
	var tabSessions []*client.SessionInfo
	for _, session := range sessions {
		if session.TabID == tabID {
			tabFound = true
			tabSessions = append(tabSessions, session)
		}
	}

	if !tabFound {
		return fmt.Errorf("tab not found: %s", tabID)
	}

	// For now, return an error as the SplitTreeNode API structure is complex
	// and would require session management integration for proper implementation
	_ = layout // Mark as used to avoid compile error
	return fmt.Errorf("set-layout command not yet fully implemented - requires complex SplitTreeNode construction")
}

func newSetTitleCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "set-title <tab-id> <title>",
		Short: "Set the title of a tab",
		Long:  "Set the title of a tab using iTerm2's variable system",
		Args:  cobra.ExactArgs(2),
		RunE:  runSetTitleCommand,
	}

	return cmd
}

func runSetTitleCommand(cmd *cobra.Command, args []string) error {
	wsURL, timeout, _ := cmdutil.GetFlags(cmd)
	ctx, cancel := cmdutil.CreateContext(timeout)
	defer cancel()

	c, err := cmdutil.ConnectClient(ctx, wsURL)
	if err != nil {
		return fmt.Errorf("failed to connect: %w", err)
	}
	defer c.Close()

	tabID := args[0]
	title := args[1]

	// Set the tab title using the standard iTerm2 variable
	err = c.SetTabVariable(ctx, tabID, "title", title)
	if err != nil {
		return fmt.Errorf("failed to set tab title: %w", err)
	}

	fmt.Printf("Successfully set title of tab %s to: %s\n", tabID, title)
	return nil
}

func newGetInfoCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "get-info <tab-id>",
		Short: "Get detailed information about a tab",
		Long:  "Get detailed information about a tab including its sessions and properties",
		Args:  cobra.ExactArgs(1),
		RunE:  runGetInfoCommand,
	}

	return cmd
}

func runGetInfoCommand(cmd *cobra.Command, args []string) error {
	wsURL, timeout, format := cmdutil.GetFlags(cmd)
	ctx, cancel := cmdutil.CreateContext(timeout)
	defer cancel()

	c, err := cmdutil.ConnectClient(ctx, wsURL)
	if err != nil {
		return fmt.Errorf("failed to connect: %w", err)
	}
	defer c.Close()

	tabID := args[0]

	// Get all sessions to find sessions in this tab
	sessions, err := c.ListSessions(ctx)
	if err != nil {
		return fmt.Errorf("failed to list sessions: %w", err)
	}

	// Find sessions belonging to this tab
	var tabSessions []*client.SessionInfo
	var windowID string
	for _, session := range sessions {
		if session.TabID == tabID {
			tabSessions = append(tabSessions, session)
			if windowID == "" {
				windowID = session.WindowID
			}
		}
	}

	if len(tabSessions) == 0 {
		return fmt.Errorf("tab not found: %s", tabID)
	}

	// Try to get tab title from variables
	title, _ := c.GetTabVariable(ctx, tabID, "title")

	// Create TabInfo structure
	tabInfo := &formatting.TabInfo{
		TabID:    tabID,
		WindowID: windowID,
		Title:    title,
		Sessions: tabSessions,
	}

	// Format and display the information
	formatter := formatting.New(format)
	return formatter.FormatTabInfo(tabInfo)
}
