package session

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"github.com/tmc/it2/internal/cmdcore"
	"github.com/tmc/it2/internal/connect"
)

// ClaudeInfo contains Claude Code session information linked to an iTerm2 session.
type ClaudeInfo struct {
	ITerm2SessionID  string `json:"iterm2_session_id"`
	ClaudeSessionID  string `json:"claude_session_id,omitempty"`
	TranscriptPath   string `json:"transcript_path,omitempty"`
	LastEvent        string `json:"last_event,omitempty"`
	HasClaudeSession bool   `json:"has_claude_session"`
}

func newClaudeCommand() *cobra.Command {
	var formatFlag string

	cmd := &cobra.Command{
		Use:   "claude [session-id]",
		Short: "Get Claude Code session info for an iTerm2 session",
		Long: `Get Claude Code session information linked to an iTerm2 session.

This command retrieves user variables set by the it2-claude-session-link hook,
which links Claude Code session IDs to iTerm2 sessions.

To enable this feature, add the following hook to ~/.claude/settings.json:

  "hooks": {
    "SessionStart": [{
      "type": "command",
      "command": "it2-claude-session-link"
    }]
  }

The hook sets these iTerm2 user variables:
  - user.claude_session_id: The Claude Code session UUID
  - user.claude_transcript: Path to the session transcript file
  - user.claude_last_event: Most recent hook event type

Examples:
  # Get Claude session ID for current iTerm2 session
  it2 session claude

  # Get Claude session for a specific iTerm2 session
  it2 session claude DE94CE8E

  # Output as JSON
  it2 session claude --format json

  # Use with claude --resume
  claude --resume $(it2 session claude DE94CE8E)`,
		Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			sessionID := ""
			if len(args) > 0 {
				sessionID = args[0]
			}

			timeout, _ := cmd.Flags().GetDuration("timeout")
			if timeout == 0 {
				timeout = 5 * time.Second
			}
			ctx, cancel := cmdcore.CreateContext(timeout)
			defer cancel()

			c, err := connect.ConnectClient(ctx)
			if err != nil {
				return fmt.Errorf("failed to connect: %w", err)
			}
			defer c.Close()

			// Resolve session ID
			sessionID, err = c.ResolveSessionID(ctx, sessionID)
			if err != nil {
				return fmt.Errorf("failed to resolve session ID: %w", err)
			}

			info := ClaudeInfo{
				ITerm2SessionID: sessionID,
			}

			// Get Claude session ID
			claudeID, err := c.GetVariableWithScope(ctx, "session", sessionID, "user.claude_session_id")
			if err == nil && claudeID != "" {
				// Remove quotes if present (JSON string format)
				claudeID = strings.Trim(claudeID, "\"")
				info.ClaudeSessionID = claudeID
				info.HasClaudeSession = true
			}

			// Get transcript path
			transcript, err := c.GetVariableWithScope(ctx, "session", sessionID, "user.claude_transcript")
			if err == nil && transcript != "" {
				transcript = strings.Trim(transcript, "\"")
				info.TranscriptPath = transcript
			}

			// Get last event
			lastEvent, err := c.GetVariableWithScope(ctx, "session", sessionID, "user.claude_last_event")
			if err == nil && lastEvent != "" {
				lastEvent = strings.Trim(lastEvent, "\"")
				info.LastEvent = lastEvent
			}

			// Output based on format
			switch formatFlag {
			case "json":
				data, err := json.MarshalIndent(info, "", "  ")
				if err != nil {
					return fmt.Errorf("failed to marshal JSON: %w", err)
				}
				fmt.Println(string(data))
			default:
				if info.HasClaudeSession {
					fmt.Println(info.ClaudeSessionID)
				}
				// Print nothing if no Claude session (useful for scripting)
			}

			return nil
		},
	}

	cmd.Flags().StringVar(&formatFlag, "format", "text", "Output format: text, json")

	return cmd
}
