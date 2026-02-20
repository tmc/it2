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

const (
	indentUnit               = "  "
	splitVerticalIndicator   = "[V]"
	splitHorizontalIndicator = "[H]"
)

var asciiCleaner = strings.NewReplacer(
	"\u00a0", " ",
	"\u2014", "--",
	"\u2013", "-",
)

func indent(level int) string {
	if level <= 0 {
		return ""
	}
	return strings.Repeat(indentUnit, level)
}

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

	for _, window := range filteredWindows {
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

		line := fmt.Sprintf("Window %d", windowNum)
		if currentSessionID != "" && windowContainsCurrent {
			line += " [current]"
		}
		fmt.Println()
		fmt.Println(line)

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

		for _, tab := range filteredTabs {
			tabID := tab.GetTabId()

			// Check if this tab contains the current session or ancestors
			root := tab.GetRoot()
			inHighlightedPath := false
			if root != nil {
				inHighlightedPath = tabContainsSession(root, currentSessionID, ancestorSet)
			}

			line := fmt.Sprintf("%sTab %s", indent(1), tabID)
			if currentSessionID != "" && inHighlightedPath {
				line += " [path]"
			}
			fmt.Println(line)

			if root != nil {
				printSplitTreeNode(root, 2, sessionInfoMap, pluginCols, currentSessionID, ancestorSet, inHighlightedPath)
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

	// Sort windows by number
	var windowNumbers []int32
	for windowNum := range windowGroups {
		windowNumbers = append(windowNumbers, windowNum)
	}
	sort.Slice(windowNumbers, func(i, j int) bool {
		return windowNumbers[i] < windowNumbers[j]
	})

	// Print tree structure: Windows -> Tabs -> Sessions (with splits)
	for _, windowNum := range windowNumbers {
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

		// Print window header
		fmt.Printf("Window %d\n", windowNum)

		// Print tabs within this window
		for _, tabKey := range tabKeys {
			sessionList := tabGroups[tabKey]

			tabTitle := fmt.Sprintf("Tab %s", tabKey)
			if tabKey == "no-tab" {
				tabTitle = "Tab (no ID)"
			}

			fmt.Printf("%s%s\n", indent(1), tabTitle)

			// Build session hierarchy within this tab (with panel structure)
			sessionHierarchy := buildSessionHierarchy(sessionList)

			printSessionHierarchy(sessionHierarchy, 2, pluginCols)
		}
	}

	return nil
}

// printFormattedTreeLine prints a tree line with session details.
func printFormattedTreeLine(prefix, title, sessionID, pidDisplay, state, command string, pluginCols []string, pluginData map[string]interface{}, suffix string) {
	var output strings.Builder

	title = normalizeSpaces(title)
	command = normalizeSpaces(command)
	state = normalizeSpaces(state)

	output.WriteString(prefix)
	if sessionID != "" {
		output.WriteString("[")
		output.WriteString(sessionID)
		output.WriteString("]")
	}

	// Column 2: PID (right-aligned, fixed width)
	if pidDisplay != "" {
		if output.Len() > 0 {
			output.WriteByte(' ')
		}
		output.WriteString(pidDisplay)
	}

	// Column 3: Title + state + command
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
		titleSection.WriteString(fmt.Sprintf("-- %s", command))
	}

	titleStr := normalizeSpaces(titleSection.String())
	if titleStr != "" {
		if output.Len() > 0 {
			output.WriteByte(' ')
		}
		output.WriteString(titleStr)
	}

	// Add plugin data
	var pluginParts []string
	if pluginData != nil {
		for _, col := range pluginCols {
			value, ok := pluginData[col]
			if !ok {
				continue
			}
			valueStr := strings.TrimSpace(stripControlChars(fmt.Sprintf("%v", value)))
			valueStr = normalizeSpaces(valueStr)
			if valueStr == "" || valueStr == "-" {
				continue
			}
			pluginParts = append(pluginParts, fmt.Sprintf("%s:%s", col, valueStr))
		}
	}

	if len(pluginParts) > 0 {
		if output.Len() > 0 {
			output.WriteByte(' ')
		}
		output.WriteString(strings.Join(pluginParts, " "))
	}

	if suffix != "" {
		if output.Len() > 0 {
			output.WriteByte(' ')
		}
		output.WriteString(suffix)
	}

	fmt.Println(output.String())
}

func normalizeSpaces(s string) string {
	if s == "" {
		return s
	}
	return asciiCleaner.Replace(s)
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
func printSessionHierarchy(nodeMap map[string]*TreeNode, level int, pluginCols []string) {
	var roots []*TreeNode
	for _, node := range nodeMap {
		if node.ParentID == "" {
			roots = append(roots, node)
		} else if _, exists := nodeMap[node.ParentID]; !exists {
			roots = append(roots, node)
		}
	}

	sortTreeNodes(roots)

	for _, root := range roots {
		printSessionNode(root, level, pluginCols)
	}
}

// printSessionNode prints a session node with its children.
func printSessionNode(node *TreeNode, level int, pluginCols []string) {
	if node == nil {
		return
	}

	indicator := ""
	if node.SplitVertical != nil {
		if *node.SplitVertical {
			indicator = splitVerticalIndicator + " "
		} else {
			indicator = splitHorizontalIndicator + " "
		}
	}

	sessionID := node.ShortID
	pidDisplay := ""
	if node.ShellPID != 0 && node.JobPID != 0 && node.ShellPID != node.JobPID {
		pidDisplay = fmt.Sprintf("%d/%d", node.ShellPID, node.JobPID)
	} else if node.JobPID != 0 {
		pidDisplay = fmt.Sprintf("%d", node.JobPID)
	} else if node.ShellPID != 0 {
		pidDisplay = fmt.Sprintf("%d", node.ShellPID)
	}

	state := node.PromptState
	switch state {
	case "AT_COMMAND_LINE", "PROMPT_STATE_AT_COMMAND_LINE":
		state = "ready"
	case "IN_COMMAND", "PROMPT_STATE_IN_COMMAND":
		state = "busy"
	case "AT_PASSWORD_PROMPT", "PROMPT_STATE_AT_PASSWORD_PROMPT":
		state = "locked"
	case "FILE_TRANSFER", "PROMPT_STATE_FILE_TRANSFER":
		state = "file-xfer"
	case "UNKNOWN", "PROMPT_STATE_UNKNOWN", "":
		state = ""
	default:
		state = ""
	}

	linePrefix := indent(level) + indicator
	printFormattedTreeLine(linePrefix, node.Name, sessionID, pidDisplay, state, node.CurrentCommand, pluginCols, node.PluginData, "")

	for _, child := range node.Children {
		printSessionNode(child, level+1, pluginCols)
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
func printSplitTreeNode(node *pb.SplitTreeNode, level int, sessionInfoMap map[string]*client.SessionInfo, pluginCols []string, currentSessionID string, ancestorSet map[string]bool, inHighlightedPath bool) {
	if node == nil {
		return
	}

	links := node.GetLinks()
	if len(links) == 0 {
		return
	}

	if len(links) > 1 {
		dirIndicator := splitHorizontalIndicator
		label := "horizontal split"
		if node.GetVertical() {
			dirIndicator = splitVerticalIndicator
			label = "vertical split"
		}

		line := fmt.Sprintf("%s%s %s", indent(level), dirIndicator, label)
		if currentSessionID != "" && inHighlightedPath {
			line += " [path]"
		}
		fmt.Println(line)

		for _, link := range links {
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
			printSplitTreeLink(link, level+1, sessionInfoMap, pluginCols, currentSessionID, ancestorSet, childInPath)
		}
		return
	}

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
	printSplitTreeLink(links[0], level, sessionInfoMap, pluginCols, currentSessionID, ancestorSet, linkInPath)
}

// printSplitTreeLink prints a link in the split tree (can be a session or another split node).
func printSplitTreeLink(link *pb.SplitTreeNode_SplitTreeLink, level int, sessionInfoMap map[string]*client.SessionInfo, pluginCols []string, currentSessionID string, ancestorSet map[string]bool, inHighlightedPath bool) {
	if link == nil {
		return
	}

	switch child := link.GetChild().(type) {
	case *pb.SplitTreeNode_SplitTreeLink_Session:
		session := child.Session
		if session == nil {
			return
		}

		sessionID := session.GetUniqueIdentifier()
		shortID := sessionID
		if len(sessionID) > 8 {
			shortID = sessionID[:8]
		}

		sessionInfo := sessionInfoMap[sessionID]

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

			switch sessionInfo.PromptState {
			case "AT_COMMAND_LINE", "PROMPT_STATE_AT_COMMAND_LINE":
				state = "ready"
			case "IN_COMMAND", "PROMPT_STATE_IN_COMMAND":
				state = "busy"
			case "AT_PASSWORD_PROMPT", "PROMPT_STATE_AT_PASSWORD_PROMPT":
				state = "locked"
			case "FILE_TRANSFER", "PROMPT_STATE_FILE_TRANSFER":
				state = "file-xfer"
			}

			command = sessionInfo.CurrentCommand
			pluginData = sessionInfo.PluginData
		}

		suffix := ""
		if currentSessionID != "" {
			if sessionID == currentSessionID {
				suffix = "[current]"
			} else if ancestorSet[sessionID] {
				suffix = "[path]"
			}
		}

		printFormattedTreeLine(indent(level), title, shortID, pidDisplay, state, command, pluginCols, pluginData, suffix)

	case *pb.SplitTreeNode_SplitTreeLink_Node:
		childInPath := tabContainsSession(child.Node, currentSessionID, ancestorSet)
		printSplitTreeNode(child.Node, level, sessionInfoMap, pluginCols, currentSessionID, ancestorSet, childInPath)
	}
}
