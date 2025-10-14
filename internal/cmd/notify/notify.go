package notify

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"
	"github.com/tmc/it2/internal/cmdcore"
)

// NewCommand creates the notify command.
func NewCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "notify <message> [session-id]",
		GroupID: "content",
		Short:   "Post macOS notification",
		Long: `Post a notification to macOS Notification Center via iTerm2.

The notification is posted using iTerm2's escape sequence mechanism.
If no session ID is provided, uses the current session.

Examples:
  it2 notify "Build completed"
  it2 notify "Task finished" session123
  it2 notify "Job done" $(it2 session current)`,
		Args: cobra.RangeArgs(1, 2),
		RunE: runNotify,
	}

	return cmd
}

func runNotify(cmd *cobra.Command, args []string) error {
	message := args[0]
	var sessionID string
	if len(args) > 1 {
		sessionID = args[1]
	}

	ctx := context.Background()
	c, err := cmdcore.ConnectClient(ctx)
	if err != nil {
		return fmt.Errorf("failed to connect: %w", err)
	}
	defer c.Close()

	// Resolve session ID with environment fallback
	sessionID, err = c.ResolveSessionID(ctx, sessionID)
	if err != nil {
		return fmt.Errorf("failed to resolve session ID: %w", err)
	}

	// iTerm2 notification escape sequence: ESC ] 9 ; <message> BEL
	// This posts a notification to macOS Notification Center
	escapeSequence := fmt.Sprintf("\x1b]9;%s\x07", message)

	// Inject the escape sequence into the session
	err = c.InjectData(ctx, sessionID, []byte(escapeSequence))
	if err != nil {
		return fmt.Errorf("failed to post notification: %w", err)
	}

	fmt.Printf("Posted notification to session %s\n", sessionID)
	return nil
}
