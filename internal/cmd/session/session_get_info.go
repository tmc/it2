package session

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	"github.com/tmc/it2/internal/client"
	"github.com/tmc/it2/internal/cmdutil"
	"github.com/tmc/it2/internal/formatting"
)

func newGetInfoCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "get-info [session-id]",
		Short: "Get comprehensive session information",
		Long: `Get comprehensive information about a session including:
  - Basic session details (ID, name, title)
  - Window and tab associations
  - Session properties
  - Current prompt information (if Shell Integration enabled)
  - Process information

This command combines multiple API calls to provide a complete view of the session.`,
		Args: cobra.RangeArgs(0, 1),
		RunE: func(cmd *cobra.Command, args []string) error {
			var sessionID string
			if len(args) == 0 {
				var err error
				sessionID, err = cmdutil.ResolveSessionIDWithError("")
				if err != nil {
					return err
				}
			} else {
				sessionID = cmdutil.ResolveSessionID(args[0])
			}

			wsURL, _ := cmd.Flags().GetString("url")
			timeout, _ := cmd.Flags().GetDuration("timeout")
			jsonOutput, _ := cmd.Flags().GetBool("json")
			includeProperties, _ := cmd.Flags().GetBool("properties")
			includePrompt, _ := cmd.Flags().GetBool("prompt")
			extractPath, _ := cmd.Flags().GetString("extract")

			// Use parent command flags if not set
			if wsURL == "" {
				wsURL = cmd.Parent().PersistentFlags().Lookup("url").Value.String()
			}
			if timeout == 0 {
				timeout, _ = cmd.Parent().PersistentFlags().GetDuration("timeout")
			}

			ctx, cancel := context.WithTimeout(context.Background(), timeout)
			defer cancel()

			c := client.New(wsURL)
			if err := c.Connect(ctx); err != nil {
				return fmt.Errorf("failed to connect: %w", err)
			}
			defer c.Close()

			// Gather comprehensive session information
			info, err := gatherSessionInfo(ctx, c, sessionID, includeProperties, includePrompt)
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

			if jsonOutput {
				return formatting.PrintJSON(info)
			} else {
				return printSessionInfo(info)
			}
		},
	}

	cmd.Flags().Bool("json", false, "Output result as JSON")
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
	info["name"] = targetSession.SessionName
	info["window_id"] = targetSession.WindowID
	info["tab_id"] = targetSession.TabID
	info["window_title"] = targetSession.WindowTitle
	info["tab_title"] = targetSession.TabTitle

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

	// Try to get session profile
	if profile, err := c.GetSessionProfile(ctx, sessionID); err == nil {
		info["profile"] = strings.Trim(profile, "\"")
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

	// Try to get PID information
	if pid, err := getShellPIDQuiet(ctx, c, sessionID); err == nil && pid > 0 {
		info["shell_pid"] = pid
	}

	return info, nil
}

func printSessionInfo(info map[string]interface{}) error {
	fmt.Printf("Session Information\n")
	fmt.Printf("==================\n\n")

	fmt.Printf("Session ID:    %v\n", info["session_id"])
	if name, ok := info["name"]; ok && name != "" {
		fmt.Printf("Name:          %v\n", name)
	}
	if windowID, ok := info["window_id"]; ok && windowID != "" {
		fmt.Printf("Window ID:     %v\n", windowID)
	}
	if tabID, ok := info["tab_id"]; ok && tabID != "" {
		fmt.Printf("Tab ID:        %v\n", tabID)
	}
	if windowTitle, ok := info["window_title"]; ok && windowTitle != "" {
		fmt.Printf("Window Title:  %v\n", windowTitle)
	}
	if tabTitle, ok := info["tab_title"]; ok && tabTitle != "" {
		fmt.Printf("Tab Title:     %v\n", tabTitle)
	}
	if profile, ok := info["profile"]; ok && profile != "" {
		fmt.Printf("Profile:       %v\n", profile)
	}
	if pid, ok := info["shell_pid"]; ok {
		fmt.Printf("Shell PID:     %v\n", pid)
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

	// Print prompt information if included
	if prompt, ok := info["prompt"].(map[string]interface{}); ok && len(prompt) > 0 {
		fmt.Printf("\nCurrent Prompt:\n")
		for key, value := range prompt {
			switch key {
			case "working_directory":
				fmt.Printf("  Working Dir:       %v\n", value)
			case "command":
				fmt.Printf("  Command:           %v\n", value)
			case "state":
				fmt.Printf("  State:             %v\n", value)
			case "exit_status":
				fmt.Printf("  Exit Status:       %v\n", value)
			case "unique_id":
				fmt.Printf("  Unique ID:         %v\n", value)
			default:
				fmt.Printf("  %-18s %v\n", strings.Title(key)+":", value)
			}
		}
	}

	return nil
}

// getShellPIDQuiet attempts to get PID without output or errors
func getShellPIDQuiet(ctx context.Context, c *client.Client, sessionID string) (int, error) {
	// This is a simplified version that doesn't send commands to the terminal
	// to avoid interfering with the user's session
	sessions, err := c.ListSessions(ctx)
	if err != nil {
		return 0, err
	}

	for _, session := range sessions {
		if session.SessionID == sessionID {
			// Try to extract PID from session name if available
			if session.SessionName != "" {
				// This is a basic implementation - in practice you might want
				// to use more sophisticated PID extraction methods
				return 0, fmt.Errorf("PID not available from session title")
			}
			break
		}
	}

	return 0, fmt.Errorf("unable to determine PID")
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
