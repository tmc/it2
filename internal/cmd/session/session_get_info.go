package session

import (
	"context"
	"encoding/json"
	"fmt"
	"os/exec"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"github.com/tmc/it2/internal/client"
	"github.com/tmc/it2/internal/cmdcore"
	"github.com/tmc/it2/internal/formatting"
	"github.com/tmc/it2/internal/plugins"
)

func newGetInfoCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "get-info [<session-id>]",
		Short: "Get comprehensive session information",
		Long: `Get comprehensive information about a session including:
  - Basic session details (ID, name, title)
  - Window and tab associations
  - Session properties
  - Current prompt information (if Shell Integration enabled)
  - Process information

Session ID can be:
  - Omitted (uses current session from $ITERM_SESSION_ID)
  - Full UUID
  - Partial ID (minimum 4 characters)

This command combines multiple API calls to provide a complete view of the session.`,
		Example: `  # Basic Usage (defaults to current session)

  it2 session get-info

  # Using full session ID
  it2 session get-info 7AA97682-C080-4D65-8C19-FDEF4669AA84

  # Using partial session ID (4+ characters)
  it2 session get-info 7aa9
  it2 session get-info 6b1a4c

  # Output Formats (default: text)

  it2 session get-info --format json
  it2 session get-info --format yaml

  # Include session properties
  it2 session get-info --properties

  # Include prompt information
  it2 session get-info --prompt

  # Extract specific fields
  it2 session get-info --extract name
  it2 session get-info --extract grid_size.width`,
		Args: cobra.RangeArgs(0, 1),
		RunE: func(cmd *cobra.Command, args []string) error {
			var sessionID string
			if len(args) > 0 {
				sessionID = args[0]
			}

			format, _ := cmd.Flags().GetString("format")
			jsonOutput, _ := cmd.Flags().GetBool("json")
			includeProperties, _ := cmd.Flags().GetBool("properties")
			extractPath, _ := cmd.Flags().GetString("extract")

			// Handle --json flag for backwards compatibility
			if jsonOutput {
				format = "json"
			}
			if format == "" {
				format = "text"
			}

			timeout, _ := cmd.Flags().GetDuration("timeout")
			if timeout == 0 {
				timeout = 60 * time.Second
			}
			ctx, cancel := cmdcore.CreateContext(timeout)
			defer cancel()

			c, err := cmdcore.ConnectClient(ctx)
			if err != nil {
				return fmt.Errorf("failed to connect: %w", err)
			}
			defer c.Close()

			// Resolve session ID with proper client method
			sessionID, err = c.ResolveSessionID(ctx, sessionID)
			if err != nil {
				return fmt.Errorf("failed to resolve session ID: %w", err)
			}

			// Gather comprehensive session information (always include prompt by default)
			info, err := gatherSessionInfo(ctx, c, sessionID, includeProperties, true)
			if err != nil {
				return fmt.Errorf("failed to gather session info: %w", err)
			}

			// Handle property extraction if requested
			if extractPath != "" {
				value, err := extractProperty(info, extractPath)
				if err != nil {
					return fmt.Errorf("failed to extract property '%s': %w", extractPath, err)
				}
				fmt.Println(value)
				return nil
			}

			switch format {
			case "json":
				return formatting.PrintJSON(info)
			case "yaml":
				return formatting.PrintYAML(info)
			case "text":
				return printSessionInfo(info)
			default:
				return fmt.Errorf("unsupported format: %s (supported: text, json, yaml)", format)
			}
		},
	}

	cmd.Flags().String("format", "", "Output format: text, json, yaml (default: text)")
	cmd.Flags().Bool("json", false, "Output result as JSON (deprecated, use --format=json)")
	cmd.Flags().Bool("properties", false, "Include session properties")
	cmd.Flags().Bool("prompt", false, "Include current prompt information")
	cmd.Flags().String("extract", "", "Extract specific property value (e.g., 'frame', 'frame.coords', 'name')")
	return cmd
}

func gatherSessionInfo(ctx context.Context, c *client.Client, sessionID string, includeProperties, includePrompt bool) (map[string]interface{}, error) {
	info := make(map[string]interface{})
	info["session_id"] = sessionID

	// Get basic session information
	sessions, err := c.ListSessions(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to list sessions: %w", err)
	}

	var targetSession *client.SessionInfo
	for _, session := range sessions {
		if session.SessionID == sessionID {
			targetSession = session
			break
		}
	}

	if targetSession == nil {
		return nil, fmt.Errorf("session not found: %s", sessionID)
	}

	// Basic session details
	info["window_id"] = targetSession.WindowID
	info["tab_id"] = targetSession.TabID
	info["window_title"] = targetSession.WindowTitle
	info["tab_title"] = targetSession.TabTitle
	if targetSession.ParentSessionID != "" {
		info["parent_session_id"] = targetSession.ParentSessionID
	}

	// Add frame information if available
	if targetSession.Frame != nil {
		frame := make(map[string]interface{})
		if origin := targetSession.Frame.GetOrigin(); origin != nil {
			frame["origin"] = map[string]interface{}{
				"x": origin.GetX(),
				"y": origin.GetY(),
			}
		}
		if size := targetSession.Frame.GetSize(); size != nil {
			frame["size"] = map[string]interface{}{
				"width":  size.GetWidth(),
				"height": size.GetHeight(),
			}
		}
		if len(frame) > 0 {
			info["frame"] = frame
		}
	}

	// Add grid size if available
	if targetSession.GridSize != nil {
		info["grid_size"] = map[string]interface{}{
			"width":  targetSession.GridSize.GetWidth(),
			"height": targetSession.GridSize.GetHeight(),
		}
	}

	// Session title is the SessionName
	if targetSession.SessionName != "" {
		info["session_title"] = targetSession.SessionName
	}

	// Try to get session badge from profile property
	if badge, err := c.GetSessionBadge(ctx, sessionID); err == nil && badge != "" {
		// Parse JSON if needed
		var badgeStr string
		if err := json.Unmarshal([]byte(badge), &badgeStr); err == nil {
			info["session_badge"] = badgeStr
		} else {
			info["session_badge"] = badge
		}
	}

	// Include properties if requested
	if includeProperties {
		properties := make(map[string]interface{})

		// Only properties that actually work with the real API
		propertiesToFetch := []string{
			"grid_size", "buried", "number_of_lines",
		}

		for _, prop := range propertiesToFetch {
			if value, err := c.GetSessionProperty(ctx, sessionID, prop); err == nil {
				properties[prop] = value
			}
		}

		if len(properties) > 0 {
			info["properties"] = properties
		}
	}

	// Include prompt information if requested
	if includePrompt {
		if promptResp, err := c.GetPrompt(ctx, sessionID); err == nil {
			prompt := make(map[string]interface{})

			prompt["status"] = promptResp.GetStatus().String()

			if workingDir := promptResp.GetWorkingDirectory(); workingDir != "" {
				prompt["working_directory"] = workingDir
			}

			if command := promptResp.GetCommand(); command != "" {
				prompt["command"] = command
			}

			prompt["state"] = promptResp.GetPromptState().String()

			if exitStatus := promptResp.GetExitStatus(); exitStatus != 0 {
				prompt["exit_status"] = exitStatus
			}

			if uniqueID := promptResp.GetUniquePromptId(); uniqueID != "" {
				prompt["unique_id"] = uniqueID
			}

			if len(prompt) > 0 {
				info["prompt"] = prompt
			}
		}
	}

	// Get shell PID if available
	if targetSession.ShellPID > 0 {
		info["shell_pid"] = targetSession.ShellPID
	}

	// Get job PID and process name if available
	if targetSession.JobPID > 0 {
		info["job_pid"] = targetSession.JobPID
		// Try to get the process name from the PID
		if processName, err := getProcessName(int(targetSession.JobPID)); err == nil && processName != "" {
			info["process"] = processName
		}
	}

	// Load session tags
	if tags, err := LoadSessionTags(sessionID); err == nil && len(tags) > 0 {
		info["tags"] = tags
	}

	// Apply plugin enrichment
	registry := plugins.NewRegistry()
	if err := registry.DiscoverAndRegister(); err == nil {
		enrichers := registry.GetEnrichers()
		pluginData := make(map[string]interface{})

		// Initialize plugin data in the target session if needed
		if targetSession.PluginData == nil {
			targetSession.PluginData = make(map[string]interface{})
		}

		// Apply each enricher
		for _, enricher := range enrichers {
			data, err := enricher.EnrichSession(ctx, targetSession)
			if err == nil && len(data) > 0 {
				for k, v := range data {
					pluginData[k] = v
				}
			}
		}

		// Add plugin data if any plugins provided data
		if len(pluginData) > 0 {
			info["plugin_data"] = pluginData
		}
	}

	return info, nil
}

func printSessionInfo(info map[string]interface{}) error {
	fmt.Printf("Session Information\n")
	fmt.Printf("==================\n\n")

	// Order: session, parent, tab, window, badge, title, window title, tab title, command, pid
	fmt.Printf("Session ID:    %v\n", info["session_id"])

	if parentID, ok := info["parent_session_id"]; ok && parentID != "" {
		fmt.Printf("Parent ID:     %v\n", parentID)
	}

	if tabID, ok := info["tab_id"]; ok && tabID != "" {
		fmt.Printf("Tab ID:        %v\n", tabID)
	}

	if windowID, ok := info["window_id"]; ok && windowID != "" {
		fmt.Printf("Window ID:     %v\n", windowID)
	}

	// Show badge
	if badge, ok := info["session_badge"]; ok && badge != "" {
		fmt.Printf("Badge:         %v\n", badge)
	}

	// Show session title
	if title, ok := info["session_title"]; ok && title != "" {
		fmt.Printf("Title:         %v\n", title)
	}

	// Show window title
	if windowTitle, ok := info["window_title"]; ok && windowTitle != "" {
		fmt.Printf("Window Title:  %v\n", windowTitle)
	}

	// Show tab title
	if tabTitle, ok := info["tab_title"]; ok && tabTitle != "" {
		fmt.Printf("Tab Title:     %v\n", tabTitle)
	}

	// Show process name if available (the main running program, not the shell)
	if process, ok := info["process"]; ok && process != "" {
		fmt.Printf("Process:       %v\n", process)
	}

	// Show shell PID if available
	if shellPID, ok := info["shell_pid"]; ok {
		fmt.Printf("Shell PID:     %v\n", shellPID)
	}

	// Show job PID if available
	if jobPID, ok := info["job_pid"]; ok {
		fmt.Printf("Job PID:       %v\n", jobPID)
	}

	// Print frame information if available
	if frame, ok := info["frame"].(map[string]interface{}); ok && len(frame) > 0 {
		fmt.Printf("Frame:         ")
		if origin, ok := frame["origin"].(map[string]interface{}); ok {
			fmt.Printf("origin(%v,%v) ", origin["x"], origin["y"])
		}
		if size, ok := frame["size"].(map[string]interface{}); ok {
			fmt.Printf("size(%v×%v)", size["width"], size["height"])
		}
		fmt.Printf("\n")
	}

	// Print grid size if available
	if gridSize, ok := info["grid_size"].(map[string]interface{}); ok {
		fmt.Printf("Grid Size:     %v×%v\n", gridSize["width"], gridSize["height"])
	}

	// Print properties if included
	if properties, ok := info["properties"].(map[string]interface{}); ok && len(properties) > 0 {
		fmt.Printf("\nProperties:\n")
		for prop, value := range properties {
			fmt.Printf("  %-20s %v\n", prop+":", value)
		}
	}

	// Print prompt information if included (in stable order)
	if prompt, ok := info["prompt"].(map[string]interface{}); ok && len(prompt) > 0 {
		fmt.Printf("\nCurrent Prompt:\n")
		// Print in stable order
		if status, ok := prompt["status"]; ok {
			fmt.Printf("  Status:            %v\n", status)
		}
		if state, ok := prompt["state"]; ok {
			fmt.Printf("  State:             %v\n", state)
		}
		if workingDir, ok := prompt["working_directory"]; ok {
			fmt.Printf("  Working Dir:       %v\n", workingDir)
		}
		if command, ok := prompt["command"]; ok {
			fmt.Printf("  Command:           %v\n", command)
		}
		if exitStatus, ok := prompt["exit_status"]; ok {
			fmt.Printf("  Exit Status:       %v\n", exitStatus)
		}
		if uniqueID, ok := prompt["unique_id"]; ok {
			fmt.Printf("  Unique ID:         %v\n", uniqueID)
		}
	}

	// Print tags if included
	if tags, ok := info["tags"].(map[string]string); ok && len(tags) > 0 {
		fmt.Printf("\nTags:\n")
		for key, value := range tags {
			fmt.Printf("  %s=%s\n", key, value)
		}
	}

	// Print plugin data if included
	if pluginData, ok := info["plugin_data"].(map[string]interface{}); ok && len(pluginData) > 0 {
		fmt.Printf("\nPlugin Data:\n")
		for key, value := range pluginData {
			switch key {
			case "is-claude-session":
				if value == "✅" {
					fmt.Printf("  Claude Session:    Yes ✅\n")
				} else {
					fmt.Printf("  Claude Session:    %v\n", value)
				}
			case "is-claude-in-progress":
				if value == "🚧" {
					fmt.Printf("  Claude In Progress: Yes 🚧\n")
				} else {
					fmt.Printf("  Claude In Progress: %v\n", value)
				}
			case "is-claude-in-vim-mode":
				if value == "✏️" {
					fmt.Printf("  Claude Vim Mode:   Yes ✏️\n")
				} else {
					fmt.Printf("  Claude Vim Mode:   %v\n", value)
				}
			default:
				// Format plugin key nicely
				displayKey := strings.ReplaceAll(key, "-", " ")
				displayKey = strings.Title(displayKey)
				fmt.Printf("  %-18s %v\n", displayKey+":", value)
			}
		}
	}

	return nil
}

// extractProperty extracts a specific property value from session info
func extractProperty(info map[string]interface{}, path string) (string, error) {
	// Handle special case for frame.coords (commonly needed for screencapture)
	if path == "frame.coords" {
		if frame, ok := info["frame"].(map[string]interface{}); ok {
			origin, hasOrigin := frame["origin"].(map[string]interface{})
			size, hasSize := frame["size"].(map[string]interface{})
			if hasOrigin && hasSize {
				return fmt.Sprintf("%v,%v,%v,%v",
					origin["x"], origin["y"],
					size["width"], size["height"]), nil
			}
		}
		return "", fmt.Errorf("frame coordinates not available")
	}

	// Split path for nested properties
	parts := strings.Split(path, ".")
	current := info

	for i, part := range parts {
		if i == len(parts)-1 {
			// Last part - return the value
			if val, ok := current[part]; ok {
				return formatValue(val), nil
			}
			return "", fmt.Errorf("property not found")
		}

		// Navigate deeper
		if next, ok := current[part].(map[string]interface{}); ok {
			current = next
		} else {
			return "", fmt.Errorf("invalid path at '%s'", part)
		}
	}

	return "", fmt.Errorf("property not found")
}

// formatValue formats a value for direct output
func formatValue(val interface{}) string {
	switch v := val.(type) {
	case string:
		return v
	case map[string]interface{}:
		// Format maps as JSON for complex structures
		data, _ := json.Marshal(v)
		return string(data)
	default:
		return fmt.Sprintf("%v", v)
	}
}

// getProcessName retrieves the process name for a given PID using ps
func getProcessName(pid int) (string, error) {
	cmd := exec.Command("ps", "-p", fmt.Sprintf("%d", pid), "-o", "comm=")
	output, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(output)), nil
}
