package session

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/tmc/it2/internal/cmdutil"
	"github.com/tmc/it2/internal/completion"
	"github.com/tmc/it2/internal/formatting"
	"github.com/tmc/it2/internal/sessionid"
	"github.com/tmc/it2/internal/validate"
	pb "github.com/tmc/it2/proto"
)

func newCreateCommand() *cobra.Command {
	template := cmdutil.CommandTemplate{
		Use:   "create [<session-id>]",
		Short: "[DEPRECATED] Create a new session",
		Long: `[DEPRECATED] This command is deprecated. Please use the clearer alternatives:
  - For splits: 'it2 session split [<session-id>] --vertical|--horizontal'
  - For new tabs: 'it2 tab create'
  - For new windows: 'it2 window create'

This command creates a new iTerm2 session by either:
  1. Creating a new tab (default)
  2. Splitting an existing session (if session-id provided or --split flag used)

When no session-id is provided, creates a new tab with the specified profile.
When a session-id is provided, splits that session to create a new session.
When --split flag is used without session-id, splits the current session.`,
		Example: cmdutil.Doc(`
			# DEPRECATED: Create new session in a new tab with Default profile
			$ it2 session create
			# USE INSTEAD: it2 tab create

			# DEPRECATED: Split current session vertically
			$ it2 session create --split vertical
			# USE INSTEAD: it2 session split --vertical

			# DEPRECATED: Split existing session by ID
			$ it2 session create session-id --split vertical
			# USE INSTEAD: it2 session split <session-id> --vertical
		`),
		Args:           cobra.RangeArgs(0, 1),
		RequiresClient: true,
		SupportsFormat: true,
		ValidArgsFunc:  completion.SessionIDCompletion,
		RunE: func(sc *cmdutil.StandardCommand, args []string) error {
			// Print deprecation warning to stderr
			fmt.Fprintf(os.Stderr, "WARNING: 'it2 session create' is deprecated.\n")
			fmt.Fprintf(os.Stderr, "  For splits, use: it2 session split [<session-id>] --vertical|--horizontal\n")
			fmt.Fprintf(os.Stderr, "  For new tabs, use: it2 tab create\n")
			fmt.Fprintf(os.Stderr, "  For new windows, use: it2 window create\n\n")
			profile, _ := sc.GetCommand().Flags().GetString("profile")
			windowID, _ := sc.GetCommand().Flags().GetString("window")
			splitDirection, _ := sc.GetCommand().Flags().GetString("split")
			command, _ := sc.GetCommand().Flags().GetString("command")

			// Check if split flag was explicitly set
			splitFlagChanged := sc.GetCommand().Flags().Changed("split")

			if len(args) == 0 {
				// If --split flag was explicitly provided, use current session
				if splitFlagChanged {
					// Get current session ID from environment
					currentSessionID := os.Getenv("ITERM_SESSION_ID")
					if currentSessionID == "" {
						return fmt.Errorf("--split requires either a session-id argument or ITERM_SESSION_ID environment variable (are you running inside iTerm2?)")
					}
					currentSessionID = sessionid.Normalize(currentSessionID)
					return createSessionBySplit(sc, currentSessionID, profile, splitDirection, command)
				}
				// Create new tab (which creates a new session)
				return createSessionByNewTab(sc, profile, windowID, command)
			} else {
				// Split existing session
				sessionID, err := sc.GetClient().ResolveSessionID(sc.GetContext(), args[0])
				if err != nil {
					return sc.ReportError("resolve session ID", err)
				}
				return createSessionBySplit(sc, sessionID, profile, splitDirection, command)
			}
		},
	}

	cmd := cmdutil.NewCommandFromTemplate(template)
	cmd.Flags().String("profile", "Default", "Profile to use for the new session")
	cmd.Flags().String("window", "", "Window ID to create session in (for new tab)")
	cmd.Flags().String("split", "horizontal", "Split direction: horizontal or vertical (when splitting existing session)")
	cmd.Flags().String("command", "", "Initial command to run in the new session")

	// Override the default format to 'text' for better scripting ergonomics
	// This allows: SESSION_ID=$(it2 session create --split vertical)
	cmd.Flags().Lookup("format").DefValue = "text"
	cmd.Flags().Set("format", "text")

	// Hide deprecated command from help output
	cmd.Hidden = true

	return cmd
}

func createSessionByNewTab(sc *cmdutil.StandardCommand, profile, windowID, command string) error {
	// Validate window exists if provided
	if windowID != "" {
		if err := validate.WindowExists(sc.GetContext(), sc.GetClient(), windowID); err != nil {
			return err
		}
	}

	// Create new tab (which automatically creates a session)
	response, err := sc.GetClient().CreateTabWithOptions(
		sc.GetContext(),
		profile,
		windowID,
		0, // index (0 = append to end)
		command,
	)
	if err != nil {
		return sc.ReportError("create session (new tab)", err)
	}

	// The response contains the new session ID
	formatter := formatting.New(sc.GetFlags().Format)
	return formatter.FormatTabResponse(response)
}

func createSessionBySplit(sc *cmdutil.StandardCommand, sessionID, profile, splitDirection, command string) error {
	// Validate session exists
	if err := validate.SessionExists(sc.GetContext(), sc.GetClient(), sessionID); err != nil {
		return err
	}

	// Determine split direction
	var isVertical bool
	switch splitDirection {
	case "vertical", "v":
		isVertical = true
	case "horizontal", "h":
		isVertical = false
	default:
		return fmt.Errorf("invalid split direction: %s (must be 'horizontal' or 'vertical')", splitDirection)
	}

	// Split the session
	response, err := sc.GetClient().SplitPane(
		sc.GetContext(),
		sessionID,
		isVertical,
		false, // before (false = after)
		profile,
		nil, // no custom profile properties
	)
	if err != nil {
		return sc.ReportError("create session (split)", err)
	}

	// Format the response based on status
	switch response.GetStatus() {
	case pb.SplitPaneResponse_OK:
		newSessionIDs := response.GetSessionId()
		if len(newSessionIDs) > 0 {
			// If a command was specified, send it to the new session
			if command != "" {
				if err := sc.GetClient().SendText(sc.GetContext(), newSessionIDs[0], command+"\n"); err != nil {
					fmt.Fprintf(os.Stderr, "Warning: Failed to send command to new session: %v\n", err)
				}
			}

			switch sc.GetFlags().Format {
			case "json":
				result := map[string]interface{}{
					"success":         true,
					"new_session_id":  newSessionIDs[0],
					"all_session_ids": newSessionIDs,
				}
				if command != "" {
					result["command_sent"] = command
				}
				return formatting.PrintJSON(result)
			case "text":
				// For text format, output just the session ID for easy scripting
				fmt.Println(newSessionIDs[0])
			default:
				fmt.Printf("Session created successfully. New session ID: %s\n", newSessionIDs[0])
				if command != "" {
					fmt.Printf("Command sent: %s\n", command)
				}
			}
		}
		return nil
	case pb.SplitPaneResponse_SESSION_NOT_FOUND:
		return fmt.Errorf("session not found: %s", sessionID)
	case pb.SplitPaneResponse_INVALID_PROFILE_NAME:
		return fmt.Errorf("invalid profile name: %s", profile)
	case pb.SplitPaneResponse_CANNOT_SPLIT:
		return fmt.Errorf("cannot split session %s (may be at maximum split level)", sessionID)
	default:
		return fmt.Errorf("split failed with status: %v", response.GetStatus())
	}
}
