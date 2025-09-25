package badge

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"
	"github.com/tmc/it2/internal/client"
	"github.com/tmc/it2/internal/cmdutil"
	"github.com/tmc/it2/internal/completion"
	pb "github.com/tmc/it2/proto"
)

// NewCommand creates the badge command with all subcommands.
func NewCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "badge",
		Short: "Manage session badges",
		Long: `Commands for setting and clearing session badges.

Badges are small text labels that appear on terminal sessions, useful for:
- Identifying different environments (prod, dev, test)
- Showing current git branch or project
- Marking sessions with status information`,
	}

	cmd.AddCommand(newSetCommand())
	cmd.AddCommand(newClearCommand())
	cmd.AddCommand(newGetCommand())

	return cmd
}

func newSetCommand() *cobra.Command {
	template := cmdutil.CommandTemplate{
		Use:   "set <session-id> <text>",
		Short: "Set a badge for a session",
		Long: `Set a badge for a session to display status or identification text.

Examples:
  it2 badge set session123 "PRODUCTION"
  it2 badge set session123 "feature-branch"
  it2 badge set session123 "🔴 LIVE"`,
		Args:            cobra.ExactArgs(2),
		RequiresClient:  true,
		RequiresSession: true,
		SupportsFormat:  true,
		ValidArgsFunc:   completion.SessionIDCompletion,
		RunE: func(sc *cmdutil.StandardCommand, args []string) error {
			// Session ID is already normalized by template when RequiresSession: true
			sessionID := args[0]
			text := args[1]

			err := setBadge(sc.GetClient(), sc.GetContext(), sessionID, text)
			if err != nil {
				return sc.ReportError("set badge", err)
			}

			// Report success with JSON output support
			if sc.GetFlags().Format == "json" {
				result := map[string]interface{}{
					"session_id": sessionID,
					"badge":      text,
					"action":     "set",
				}
				return sc.FormatOutput(result)
			}

			sc.ReportSuccess("Set badge for session %s: %s", sessionID, text)
			return nil
		},
	}

	return cmdutil.NewCommandFromTemplate(template)
}

func newClearCommand() *cobra.Command {
	template := cmdutil.CommandTemplate{
		Use:   "clear <session-id>",
		Short: "Clear the badge for a session",
		Long: `Clear the badge for a session, removing any previously set text.

Examples:
  it2 badge clear session123
  it2 badge clear $(it2 session list --format json | jq -r '.[0].id')`,
		Args:            cobra.ExactArgs(1),
		RequiresClient:  true,
		RequiresSession: true,
		SupportsFormat:  true,
		ValidArgsFunc:   completion.SessionIDCompletion,
		RunE: func(sc *cmdutil.StandardCommand, args []string) error {
			// Session ID is already normalized by template when RequiresSession: true
			sessionID := args[0]

			err := setBadge(sc.GetClient(), sc.GetContext(), sessionID, "")
			if err != nil {
				return sc.ReportError("clear badge", err)
			}

			// Report success with JSON output support
			if sc.GetFlags().Format == "json" {
				result := map[string]interface{}{
					"session_id": sessionID,
					"badge":      nil,
					"action":     "cleared",
				}
				return sc.FormatOutput(result)
			}

			sc.ReportSuccess("Cleared badge for session %s", sessionID)
			return nil
		},
	}

	return cmdutil.NewCommandFromTemplate(template)
}

func newGetCommand() *cobra.Command {
	template := cmdutil.CommandTemplate{
		Use:   "get [session-id]",
		Short: "Get the current badge for a session",
		Long: `Get the current badge for a session. If no session-id is provided,
uses $ITERM_SESSION_ID environment variable.

Examples:
  it2 badge get                    # Get badge for current session
  it2 badge get session123         # Get specific session's badge
  it2 badge get --format json      # Output as JSON`,
		Args:            cobra.RangeArgs(0, 1),
		RequiresClient:  true,
		SupportsFormat:  true,
		ValidArgsFunc:   completion.SessionIDCompletion,
		RunE: func(sc *cmdutil.StandardCommand, args []string) error {
			var sessionID string
			if len(args) > 0 {
				sessionID = args[0]
			}
			sessionID = cmdutil.ResolveSessionID(sessionID)
			if sessionID == "" {
				return cmdutil.NewRequiredArgumentError("session ID (or $ITERM_SESSION_ID)")
			}

			badge, err := getBadge(sc.GetClient(), sc.GetContext(), sessionID)
			if err != nil {
				return sc.ReportError("get badge", err)
			}

			// Format output based on format flag
			if sc.GetFlags().Format == "json" {
				result := map[string]interface{}{
					"session_id": sessionID,
					"badge":      badge,
				}
				if badge == "" {
					result["badge"] = nil
				}
				return sc.FormatOutput(result)
			}

			if badge == "" {
				fmt.Printf("No badge set for session %s\n", sessionID)
			} else {
				fmt.Printf("Badge for session %s: %s\n", sessionID, badge)
			}
			return nil
		},
	}

	return cmdutil.NewCommandFromTemplate(template)
}

// setBadge sets the badge text for a session using the badge property
func setBadge(c *client.Client, ctx context.Context, sessionID, text string) error {
	msg := &pb.ClientOriginatedMessage{
		Submessage: &pb.ClientOriginatedMessage_SetPropertyRequest{
			SetPropertyRequest: &pb.SetPropertyRequest{
				Identifier: &pb.SetPropertyRequest_SessionId{
					SessionId: sessionID,
				},
				Name:      cmdutil.StringPtr("badge"),
				JsonValue: cmdutil.StringPtr(fmt.Sprintf(`"%s"`, text)),
			},
		},
	}

	response, err := c.SendRequest(ctx, msg)
	if err != nil {
		return err
	}

	resp := response.GetSetPropertyResponse()
	if resp == nil {
		return fmt.Errorf("no set property response received")
	}

	if resp.GetStatus() != pb.SetPropertyResponse_OK {
		return fmt.Errorf("failed to set badge property: %v", resp.GetStatus())
	}

	return nil
}

// getBadge gets the badge text for a session using the badge property
func getBadge(c *client.Client, ctx context.Context, sessionID string) (string, error) {
	msg := &pb.ClientOriginatedMessage{
		Submessage: &pb.ClientOriginatedMessage_GetPropertyRequest{
			GetPropertyRequest: &pb.GetPropertyRequest{
				Identifier: &pb.GetPropertyRequest_SessionId{
					SessionId: sessionID,
				},
				Name: cmdutil.StringPtr("badge"),
			},
		},
	}

	response, err := c.SendRequest(ctx, msg)
	if err != nil {
		return "", err
	}

	resp := response.GetGetPropertyResponse()
	if resp == nil {
		return "", fmt.Errorf("no get property response received")
	}

	if resp.GetStatus() != pb.GetPropertyResponse_OK {
		if resp.GetStatus() == pb.GetPropertyResponse_UNRECOGNIZED_NAME {
			// Property doesn't exist yet, return empty string
			return "", nil
		}
		return "", fmt.Errorf("failed to get badge property: %v", resp.GetStatus())
	}

	if resp.JsonValue == nil {
		return "", nil
	}

	// The value is stored as a JSON string, so we need to remove the quotes
	value := *resp.JsonValue
	if len(value) >= 2 && value[0] == '"' && value[len(value)-1] == '"' {
		value = value[1 : len(value)-1]
	}

	return value, nil
}