package formatting

import (
	"fmt"
	"os"
	"sort"
	"strconv"
	"strings"

	"github.com/tmc/it2/internal/client"
	"github.com/tmc/it2/internal/sessionid"
	pb "github.com/tmc/it2/proto"
	"golang.org/x/term"
)

// shouldEnableColors determines if ANSI color codes should be used.
func shouldEnableColors() bool {
	if os.Getenv("NO_COLOR") != "" {
		return false
	}
	if os.Getenv("FORCE_COLOR") != "" || os.Getenv("CLICOLOR_FORCE") != "" {
		return true
	}
	if term.IsTerminal(int(os.Stdout.Fd())) {
		return true
	}
	if term.IsTerminal(int(os.Stderr.Fd())) {
		return true
	}
	return false
}

// buildAncestorSet builds a set of ancestor session IDs for the given session.
func buildAncestorSet(targetSessionID string, sessions []*client.SessionInfo) map[string]bool {
	ancestorSet := make(map[string]bool)
	if targetSessionID == "" {
		return ancestorSet
	}

	// Find ancestors by following parent links
	currentID := targetSessionID
	for {
		found := false
		for _, session := range sessions {
			if session.SessionID == currentID {
				if session.ParentSessionID != "" {
					ancestorSet[session.ParentSessionID] = true
					currentID = session.ParentSessionID
					found = true
				}
				break
			}
		}
		if !found {
			break
		}
	}

	return ancestorSet
}

// tabContainsSession checks if a split tree node contains the target session or any ancestors.
func tabContainsSession(node *pb.SplitTreeNode, targetSessionID string, ancestorSet map[string]bool) bool {
	if node == nil || targetSessionID == "" {
		return false
	}
	links := node.GetLinks()
	for _, link := range links {
		switch child := link.GetChild().(type) {
		case *pb.SplitTreeNode_SplitTreeLink_Session:
			if child.Session != nil {
				sessionID := child.Session.GetUniqueIdentifier()
				if sessionID == targetSessionID || ancestorSet[sessionID] {
					return true
				}
			}
		case *pb.SplitTreeNode_SplitTreeLink_Node:
			if tabContainsSession(child.Node, targetSessionID, ancestorSet) {
				return true
			}
		}
	}
	return false
}

// formatSessionsTreeFromRaw formats sessions as a tree using raw split tree structure.
func (f *Formatter) formatSessionsTreeFromRaw(listResp *pb.ListSessionsResponse, sessions []*client.SessionInfo) error {
	if listResp == nil || len(listResp.GetWindows()) == 0 {
		fmt.Println("✗ No sessions found")
		fmt.Println("  Run 'it2 session create' to create a new session")
		return nil
	}

	// Build session info lookup map
	sessionInfoMap := make(map[string]*client.SessionInfo)
	for _, session := range sessions {
		sessionInfoMap[session.SessionID] = session
	}

	// Check if any sessions have plugin data to determine additional columns
	pluginColumns := make(map[string]bool)
	for _, session := range sessions {
		for key := range session.PluginData {
			pluginColumns[key] = true
		}
	}

	// Add plugin columns to headers in sorted order for stability
	var pluginCols []string
	for col := range pluginColumns {
		pluginCols = append(pluginCols, col)
	}
	sort.Strings(pluginCols)

	// Determine if highlighting should be enabled and get current session
	shouldHighlight := shouldEnableColors()
	currentSessionID := ""
	var ancestorSet map[string]bool
	if shouldHighlight {
		if envSessionID := os.Getenv("ITERM_SESSION_ID"); envSessionID != "" {
			currentSessionID = sessionid.Normalize(envSessionID)
			ancestorSet = buildAncestorSet(currentSessionID, sessions)
		}
	}

	// Print header
	fmt.Println("Session Hierarchy:")
	fmt.Println()

	// Build sets of window IDs and tab IDs that contain our filtered sessions
	windowIDs := make(map[string]bool)
	tabKeys := make(map[string]bool) // key is "windowID:tabID"
	for _, session := range sessions {
		windowIDs[session.WindowID] = true
		tabKeys[session.WindowID+":"+session.TabID] = true
	}

	// Print each window (only those containing filtered sessions)
	windows := listResp.GetWindows()
	var filteredWindows []*pb.ListSessionsResponse_Window
	for _, window := range windows {
		windowID := window.GetWindowId()
		if windowIDs[windowID] {
			filteredWindows = append(filteredWindows, window)
		}
	}

	for windowIndex, window := range filteredWindows {
		isLastWindow := windowIndex == len(filteredWindows)-1
		windowConnector := "├─"
		if isLastWindow {
			windowConnector = "└─"
		}

		windowNum := window.GetNumber()

		// Check if this window contains the current session
		windowContainsCurrent := false
		if currentSessionID != "" {
			tabs := window.GetTabs()
			for _, tab := range tabs {
				if tabContainsSession(tab.GetRoot(), currentSessionID, ancestorSet) {
					windowContainsCurrent = true
					break
				}
			}
		}

		// Highlight window if it contains the current session
		if windowContainsCurrent {
			fmt.Printf("%s%s%s Window %d\n", colorBrightCyan, windowConnector, colorReset, windowNum)
		} else {
			fmt.Printf("%s Window %d\n", windowConnector, windowNum)
		}

		// Print tabs within this window (only those containing filtered sessions)
		windowID := window.GetWindowId()
		tabs := window.GetTabs()
		var filteredTabs []*pb.ListSessionsResponse_Tab
		for _, tab := range tabs {
			tabKey := windowID + ":" + tab.GetTabId()
			if tabKeys[tabKey] {
				filteredTabs = append(filteredTabs, tab)
			}
		}

		for tabIndex, tab := range filteredTabs {
			isLastTab := tabIndex == len(filteredTabs)-1

			tabPrefix := "│ "
			if isLastWindow {
				tabPrefix = "  "
			}

			tabConnector := "├─"
			if isLastTab {
				tabConnector = "└─"
			}

			tabID := tab.GetTabId()

			// Check if this tab contains the current session or ancestors
			root := tab.GetRoot()
			inHighlightedPath := false
			if root != nil {
				inHighlightedPath = tabContainsSession(root, currentSessionID, ancestorSet)
			}

			// Highlight tab if it contains the current session
			if inHighlightedPath {
				fmt.Printf("%s%s%s%s Tab %s\n", colorBrightCyan, tabPrefix, tabConnector, colorReset, tabID)
			} else {
				fmt.Printf("%s%s Tab %s\n", tabPrefix, tabConnector, tabID)
			}

			// Print the split tree for this tab
			sessionPrefix := tabPrefix
			if isLastTab {
				sessionPrefix += "  "
			} else {
				sessionPrefix += "│ "
			}

			if root != nil {
				printSplitTreeNode(root, sessionPrefix, true, sessionInfoMap, pluginCols, currentSessionID, ancestorSet, inHighlightedPath)
			}
		}
	}

	return nil
}

// formatSessionsTree formats sessions as a tree showing hierarchy ordered by window -> tab -> splits -> sessions.
func (f *Formatter) formatSessionsTree(sessions []*client.SessionInfo) error {
	// Note: To show proper panel hierarchy, we need access to the raw protobuf data
	// For now, fall back to the parent-child hierarchy from sessions
	// TODO: Refactor to pass raw ListSessionsResponse to show actual split tree
	if len(sessions) == 0 {
		fmt.Println("✗ No sessions found")
		fmt.Println("  Run 'it2 session create' to create a new session")
		return nil
	}

	// Check if any sessions have plugin data to determine additional columns
	pluginColumns := make(map[string]bool)
	for _, session := range sessions {
		for key := range session.PluginData {
			pluginColumns[key] = true
		}
	}

	// Add plugin columns to headers in sorted order for stability
	var pluginCols []string
	for col := range pluginColumns {
		pluginCols = append(pluginCols, col)
	}
	sort.Strings(pluginCols)

	// Group sessions by window -> tab structure
	windowGroups := make(map[int32]map[string][]*client.SessionInfo)

	for _, session := range sessions {
		if windowGroups[session.WindowNumber] == nil {
			windowGroups[session.WindowNumber] = make(map[string][]*client.SessionInfo)
		}
		tabKey := session.TabID
		if tabKey == "" {
			tabKey = "no-tab"
		}
		windowGroups[session.WindowNumber][tabKey] = append(windowGroups[session.WindowNumber][tabKey], session)
	}

	// Print header
	fmt.Println("Session Hierarchy:")
	fmt.Println()

	// Sort windows by number
	var windowNumbers []int32
	for windowNum := range windowGroups {
		windowNumbers = append(windowNumbers, windowNum)
	}
	sort.Slice(windowNumbers, func(i, j int) bool {
		return windowNumbers[i] < windowNumbers[j]
	})

	// Print tree structure: Windows -> Tabs -> Sessions (with splits)
	for windowIndex, windowNum := range windowNumbers {
		tabGroups := windowGroups[windowNum]

		// Sort tab keys numerically
		var tabKeys []string
		for tabKey := range tabGroups {
			tabKeys = append(tabKeys, tabKey)
		}
		sort.Slice(tabKeys, func(i, j int) bool {
			// Try to parse as numbers for proper numeric sorting
			numI, errI := strconv.Atoi(tabKeys[i])
			numJ, errJ := strconv.Atoi(tabKeys[j])

			// If both are numbers, sort numerically
			if errI == nil && errJ == nil {
				return numI < numJ
			}

			// Otherwise fall back to string sorting
			return tabKeys[i] < tabKeys[j]
		})

		isLastWindow := windowIndex == len(windowNumbers)-1
		windowConnector := "├─"
		if isLastWindow {
			windowConnector = "└─"
		}

		// Print window header
		fmt.Printf("%s Window %d\n", windowConnector, windowNum)

		// Print tabs within this window
		for tabIndex, tabKey := range tabKeys {
			sessionList := tabGroups[tabKey]
			isLastTab := tabIndex == len(tabKeys)-1

			tabPrefix := "│ "
			if isLastWindow {
				tabPrefix = "  "
			}

			tabConnector := "├─"
			if isLastTab {
				tabConnector = "└─"
			}

			tabTitle := fmt.Sprintf("Tab %s", tabKey)
			if tabKey == "no-tab" {
				tabTitle = "Tab (no ID)"
			}

			fmt.Printf("%s%s %s\n", tabPrefix, tabConnector, tabTitle)

			// Build session hierarchy within this tab (with panel structure)
			sessionHierarchy := buildSessionHierarchy(sessionList)

			// Print sessions within this tab
			sessionPrefix := tabPrefix
			if isLastTab {
				sessionPrefix += "  "
			} else {
				sessionPrefix += "│ "
			}

			printSessionHierarchy(sessionHierarchy, sessionPrefix, pluginCols)
		}
	}

	return nil
}

// printFormattedTreeLine prints a tree line with session details.
func printFormattedTreeLine(prefix, title, sessionID, pidDisplay, state, command string, pluginCols []string, pluginData map[string]interface{}) {
	// Build output with fixed column widths for proper alignment
	var output strings.Builder

	// Column 1: prefix + session ID (fixed width including tree chars)
	output.WriteString(prefix)
	if sessionID != "" {
		output.WriteString(fmt.Sprintf("[%s]", sessionID))
	} else {
		output.WriteString(strings.Repeat(" ", 10))
	}
	output.WriteString("    ") // 4 spaces separator

	// Column 2: PID (right-aligned, fixed width)
	if pidDisplay != "" {
		output.WriteString(fmt.Sprintf("%13s", pidDisplay)) // 13 chars for pid/pid format
	} else {
		output.WriteString(strings.Repeat(" ", 13))
	}
	output.WriteString("    ") // 4 spaces separator

	// Column 3: Title + state + command (variable width, but calculate for alignment)
	var titleSection strings.Builder
	if title != "" {
		// Truncate title to reasonable length
		if len(title) > 60 {
			title = title[:57] + "..."
		}
		titleSection.WriteString(title)
	}

	if state != "" {
		if titleSection.Len() > 0 {
			titleSection.WriteString(" ")
		}
		titleSection.WriteString(state)
	}

	if command != "" && command != "-" {
		// Truncate very long commands
		if len(command) > 80 {
			command = command[:77] + "..."
		}
		if titleSection.Len() > 0 {
			titleSection.WriteString(" ")
		}
		titleSection.WriteString(fmt.Sprintf("— %s", command))
	}

	titleStr := titleSection.String()
	output.WriteString(titleStr)

	// Add plugin data aligned to fixed column position
	var pluginParts []string
	for _, col := range pluginCols {
		if pluginData != nil {
			if value, exists := pluginData[col]; exists {
				if v := fmt.Sprintf("%v", value); v != "" && v != "-" {
					pluginParts = append(pluginParts, fmt.Sprintf("%s:%v", col, value))
				}
			}
		}
	}

	if len(pluginParts) > 0 {
		// Calculate total visible width so far (prefix + ID + PID + title)
		currentWidth := visibleLen(prefix) + 10 + 4 + 13 + 4 + visibleLen(titleStr)
		// Align plugin data to absolute column position
		targetCol := 120
		if currentWidth < targetCol {
			output.WriteString(strings.Repeat(" ", targetCol-currentWidth))
		} else {
			output.WriteString("    ") // minimum spacing if content is too long
		}
		output.WriteString(strings.Join(pluginParts, " "))
	}

	fmt.Println(output.String())
}

// buildSessionHierarchy builds parent-child relationships within a list of sessions.
func buildSessionHierarchy(sessions []*client.SessionInfo) map[string]*TreeNode {
	nodeMap := make(map[string]*TreeNode)

	// First pass: create all nodes
	for _, session := range sessions {
		node := &TreeNode{
			SessionID:      session.SessionID,
			ShortID:        session.ShortID,
			Name:           session.SessionName,
			ParentID:       session.ParentSessionID,
			SplitVertical:  session.SplitVertical,
			Children:       []*TreeNode{},
			CurrentCommand: session.CurrentCommand,
			ShellPID:       session.ShellPID,
			JobPID:         session.JobPID,
			PromptState:    session.PromptState,
			PluginData:     session.PluginData,
		}
		nodeMap[session.SessionID] = node
	}

	// Second pass: build parent-child relationships
	for _, node := range nodeMap {
		if node.ParentID != "" {
			if parent, exists := nodeMap[node.ParentID]; exists {
				parent.Children = append(parent.Children, node)
			}
		}
	}

	// Sort children by name for consistent output
	for _, node := range nodeMap {
		sortTreeNodes(node.Children)
	}

	return nodeMap
}

// printSessionHierarchy prints sessions with proper hierarchy within a tab.
func printSessionHierarchy(nodeMap map[string]*TreeNode, prefix string, pluginCols []string) {
	// Find root sessions (those without parents in this hierarchy)
	var roots []*TreeNode
	for _, node := range nodeMap {
		if node.ParentID == "" {
			roots = append(roots, node)
		} else {
			// Check if parent exists in this hierarchy
			if _, exists := nodeMap[node.ParentID]; !exists {
				// Parent doesn't exist in this hierarchy, treat as root
				roots = append(roots, node)
			}
		}
	}

	// Sort roots by name for consistent output
	sortTreeNodes(roots)

	// Print each root and its children
	for i, root := range roots {
		isLast := i == len(roots)-1
		printSessionNode(root, prefix, isLast, pluginCols)
	}
}

// printSessionNode prints a session node with its children.
func printSessionNode(node *TreeNode, prefix string, isLast bool, pluginCols []string) {
	// Determine the connector with split direction indicator
	connector := "├─"
	if isLast {
		connector = "└─"
	}

	// Add split direction indicator if this is a split
	if node.SplitVertical != nil {
		if *node.SplitVertical {
			connector += "⫴" // Vertical split indicator
		} else {
			connector += "⫻" // Horizontal split indicator
		}
	}

	// Format session information
	sessionID := node.ShortID
	pidDisplay := ""
	if node.ShellPID != 0 && node.JobPID != 0 && node.ShellPID != node.JobPID {
		// Show both PIDs when they differ (shell/job)
		pidDisplay = fmt.Sprintf("%d/%d", node.ShellPID, node.JobPID)
	} else if node.JobPID != 0 {
		// Show only job PID if available
		pidDisplay = fmt.Sprintf("%d", node.JobPID)
	} else if node.ShellPID != 0 {
		// Fall back to shell PID if no job PID
		pidDisplay = fmt.Sprintf("%d", node.ShellPID)
	}

	// Shorten state with indicators
	state := node.PromptState
	switch state {
	case "AT_COMMAND_LINE", "PROMPT_STATE_AT_COMMAND_LINE":
		state = "✓ READY"
	case "IN_COMMAND", "PROMPT_STATE_IN_COMMAND":
		state = "🚧"
	case "AT_PASSWORD_PROMPT", "PROMPT_STATE_AT_PASSWORD_PROMPT":
		state = "🔒"
	case "FILE_TRANSFER", "PROMPT_STATE_FILE_TRANSFER":
		state = "📁"
	case "UNKNOWN", "PROMPT_STATE_UNKNOWN", "":
		state = ""
	default:
		state = ""
	}

	// Format title and command - don't truncate, let terminal wrap naturally
	title := node.Name
	command := node.CurrentCommand

	// Print the session
	printFormattedTreeLine(prefix+connector+" ", title, sessionID, pidDisplay, state, command, pluginCols, node.PluginData)

	// Print children (splits) - use compressed indent
	newPrefix := prefix
	if isLast {
		newPrefix += "  "
	} else {
		newPrefix += "│ "
	}

	for i, child := range node.Children {
		childIsLast := i == len(node.Children)-1
		printSessionNode(child, newPrefix, childIsLast, pluginCols)
	}
}

// sortTreeNodes sorts tree nodes by name for consistent output.
func sortTreeNodes(nodes []*TreeNode) {
	// Simple bubble sort by name
	for i := 0; i < len(nodes); i++ {
		for j := i + 1; j < len(nodes); j++ {
			if nodes[i].Name > nodes[j].Name {
				nodes[i], nodes[j] = nodes[j], nodes[i]
			}
		}
	}
}

// printSplitTreeNode recursively prints a split tree node (either a split container or a session).
func printSplitTreeNode(node *pb.SplitTreeNode, prefix string, isLast bool, sessionInfoMap map[string]*client.SessionInfo, pluginCols []string, currentSessionID string, ancestorSet map[string]bool, inHighlightedPath bool) {
	if node == nil {
		return
	}

	links := node.GetLinks()
	if len(links) == 0 {
		return
	}

	// If this node has multiple links, it's a split container
	if len(links) > 1 {
		// Print split container
		connector := "├─"
		if isLast {
			connector = "└─"
		}

		// Add split direction indicator
		splitType := "Horizontal Split"
		dirIndicator := "⫻"
		if node.GetVertical() {
			splitType = "Vertical Split"
			dirIndicator = "⫴"
		}

		// Apply bright cyan highlighting if this split is in the path to current session
		if inHighlightedPath {
			fmt.Printf("%s%s%s%s %s [%s]\n", colorBrightCyan, prefix, connector, colorReset, dirIndicator, splitType)
		} else {
			fmt.Printf("%s%s %s [%s]\n", prefix, connector, dirIndicator, splitType)
		}

		// Print children with appropriate prefix
		newPrefix := prefix
		if isLast {
			newPrefix += "  "
		} else {
			newPrefix += "│ "
		}

		for i, link := range links {
			childIsLast := i == len(links)-1
			// Check if this child link contains the current session or ancestors
			var childInPath bool
			switch child := link.GetChild().(type) {
			case *pb.SplitTreeNode_SplitTreeLink_Session:
				if child.Session != nil {
					sessionID := child.Session.GetUniqueIdentifier()
					childInPath = sessionID == currentSessionID || ancestorSet[sessionID]
				}
			case *pb.SplitTreeNode_SplitTreeLink_Node:
				childInPath = tabContainsSession(child.Node, currentSessionID, ancestorSet)
			}
			printSplitTreeLink(link, newPrefix, childIsLast, sessionInfoMap, pluginCols, currentSessionID, ancestorSet, childInPath)
		}
	} else {
		// Single link - just print it directly
		// Check if this link is in the highlighted path
		var linkInPath bool
		switch child := links[0].GetChild().(type) {
		case *pb.SplitTreeNode_SplitTreeLink_Session:
			if child.Session != nil {
				sessionID := child.Session.GetUniqueIdentifier()
				linkInPath = sessionID == currentSessionID || ancestorSet[sessionID]
			}
		case *pb.SplitTreeNode_SplitTreeLink_Node:
			linkInPath = tabContainsSession(child.Node, currentSessionID, ancestorSet)
		}
		printSplitTreeLink(links[0], prefix, isLast, sessionInfoMap, pluginCols, currentSessionID, ancestorSet, linkInPath)
	}
}

// printSplitTreeLink prints a link in the split tree (can be a session or another split node).
func printSplitTreeLink(link *pb.SplitTreeNode_SplitTreeLink, prefix string, isLast bool, sessionInfoMap map[string]*client.SessionInfo, pluginCols []string, currentSessionID string, ancestorSet map[string]bool, inHighlightedPath bool) {
	if link == nil {
		return
	}

	switch child := link.GetChild().(type) {
	case *pb.SplitTreeNode_SplitTreeLink_Session:
		// Print session
		session := child.Session
		if session == nil {
			return
		}

		connector := "├─"
		if isLast {
			connector = "└─"
		}

		sessionID := session.GetUniqueIdentifier()
		shortID := sessionID
		if len(sessionID) > 8 {
			shortID = sessionID[:8]
		}

		// Get enriched session info if available
		sessionInfo := sessionInfoMap[sessionID]

		// Format session information
		pidDisplay := ""
		state := ""
		command := ""
		title := session.GetTitle()
		var pluginData map[string]interface{}

		if sessionInfo != nil {
			if sessionInfo.ShellPID != 0 && sessionInfo.JobPID != 0 && sessionInfo.ShellPID != sessionInfo.JobPID {
				pidDisplay = fmt.Sprintf("%d/%d", sessionInfo.ShellPID, sessionInfo.JobPID)
			} else if sessionInfo.JobPID != 0 {
				pidDisplay = fmt.Sprintf("%d", sessionInfo.JobPID)
			} else if sessionInfo.ShellPID != 0 {
				pidDisplay = fmt.Sprintf("%d", sessionInfo.ShellPID)
			}

			// Format state
			switch sessionInfo.PromptState {
			case "AT_COMMAND_LINE", "PROMPT_STATE_AT_COMMAND_LINE":
				state = "✳"
			case "IN_COMMAND", "PROMPT_STATE_IN_COMMAND":
				state = "🚧"
			case "AT_PASSWORD_PROMPT", "PROMPT_STATE_AT_PASSWORD_PROMPT":
				state = "🔒"
			case "FILE_TRANSFER", "PROMPT_STATE_FILE_TRANSFER":
				state = "📁"
			}

			command = sessionInfo.CurrentCommand
			pluginData = sessionInfo.PluginData
		}

		// Determine highlighting: current session vs ancestor vs normal
		isCurrent := currentSessionID != "" && sessionID == currentSessionID
		isAncestor := ancestorSet[sessionID]

		// Build the line prefix with appropriate highlighting
		var linePrefix string
		if isCurrent {
			// Current session: cyan connector with green indicator
			linePrefix = prefix + colorBrightCyan + connector + colorReset + " " + colorGreen + "● " + colorReset
		} else if isAncestor {
			// Ancestor in path: bright cyan connector
			linePrefix = prefix + colorBrightCyan + connector + colorReset + " "
		} else {
			// No highlighting - normal display
			linePrefix = prefix + connector + " "
		}

		// Print the session using the formatted tree line
		printFormattedTreeLine(linePrefix, title, shortID, pidDisplay, state, command, pluginCols, pluginData)

	case *pb.SplitTreeNode_SplitTreeLink_Node:
		// Check if this child node contains the current session or ancestors
		childInPath := tabContainsSession(child.Node, currentSessionID, ancestorSet)
		// Recursively print child split tree node
		printSplitTreeNode(child.Node, prefix, isLast, sessionInfoMap, pluginCols, currentSessionID, ancestorSet, childInPath)
	}
}
