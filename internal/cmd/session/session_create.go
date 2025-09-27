package session

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/tmc/it2/internal/cmdutil"
	"github.com/tmc/it2/internal/completion"
	"github.com/tmc/it2/internal/formatting"
	pb "github.com/tmc/it2/proto"
)

func newCreateCommand() *cobra.Command {
	template := cmdutil.CommandTemplate{
		Use:   "create [<session-id>]",
		Short: "Create a new session",
		Long: `Create a new iTerm2 session by either:
  1. Creating a new tab (default)
  2. Splitting an existing session (if session-id provided)

When no session-id is provided, creates a new tab with the specified profile.
When a session-id is provided, splits that session to create a new session.`,
		Example: cmdutil.Doc(`
			# Create new session in a new tab with Default profile
			$ it2 session create

			# Create new session with specific profile in new tab
			$ it2 session create --profile "Development"

			# Create new session in specific window
			$ it2 session create --window window-id

			# Split existing session vertically to create new session
			$ it2 session create session-id --split vertical

			# Split existing session horizontally with specific profile
			$ it2 session create session-id --split horizontal --profile "SSH"

			# Create session and get details
			$ it2 session create --format json

			# Create session with initial command
			$ it2 session create --command "cd ~/project && npm run dev"
		`),
		Args:           cobra.RangeArgs(0, 1),
		RequiresClient: true,
		SupportsFormat: true,
		ValidArgsFunc:  completion.SessionIDCompletion,
		RunE: func(sc *cmdutil.StandardCommand, args []string) error {
			profile, _ := sc.GetCommand().Flags().GetString("profile")
			windowID, _ := sc.GetCommand().Flags().GetString("window")
			splitDirection, _ := sc.GetCommand().Flags().GetString("split")
			command, _ := sc.GetCommand().Flags().GetString("command")

			if len(args) == 0 {
				// Create new tab (which creates a new session)
				return createSessionByNewTab(sc, profile, windowID, command)
			} else {
				// Split existing session
				sessionID := cmdutil.ResolveSessionID(args[0])
				if sessionID == "" {
					return fmt.Errorf("no session ID provided and $ITERM_SESSION_ID not set")
				}
				return createSessionBySplit(sc, sessionID, profile, splitDirection)
			}
		},
	}

	cmd := cmdutil.NewCommandFromTemplate(template)
	cmd.Flags().String("profile", "Default", "Profile to use for the new session")
	cmd.Flags().String("window", "", "Window ID to create session in (for new tab)")
	cmd.Flags().String("split", "horizontal", "Split direction: horizontal or vertical (when splitting existing session)")
	cmd.Flags().String("command", "", "Initial command to run in the new session")

	return cmd
}

func createSessionByNewTab(sc *cmdutil.StandardCommand, profile, windowID, command string) error {
	// Validate window exists if provided
	if windowID != "" {
		if err := cmdutil.ValidateWindowExists(sc.GetContext(), sc.GetClient(), windowID); err != nil {
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

func createSessionBySplit(sc *cmdutil.StandardCommand, sessionID, profile, splitDirection string) error {
	// Validate session exists
	if err := cmdutil.ValidateSessionExists(sc.GetContext(), sc.GetClient(), sessionID); err != nil {
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
	)
	if err != nil {
		return sc.ReportError("create session (split)", err)
	}

	// Format the response based on status
	switch response.GetStatus() {
	case pb.SplitPaneResponse_OK:
		newSessionIDs := response.GetSessionId()
		if len(newSessionIDs) > 0 {
			if sc.GetFlags().Format == "json" {
				result := map[string]interface{}{
					"success":         true,
					"new_session_id":  newSessionIDs[0],
					"all_session_ids": newSessionIDs,
				}
				return formatting.PrintJSON(result)
			} else {
				fmt.Printf("Session created successfully. New session ID: %s\n", newSessionIDs[0])
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
