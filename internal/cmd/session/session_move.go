package session

import (
	"fmt"
	"time"

	"github.com/spf13/cobra"
	"github.com/tmc/it2/internal/cmdcore"
	"github.com/tmc/it2/internal/cmdutil"
	"github.com/tmc/it2/internal/connect"
	"github.com/tmc/it2/internal/formatting"
)

func newMoveCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "move <source-session-id> <destination-session-id>",
		Short: "Move a session to be a split pane next to another session",
		Long: `Move a session to become a split pane by splitting another session.

The source session will be moved to split the destination session's pane.
This uses iTerm2's built-in move_session function.

The move will fail if:
  - Either session ID is invalid
  - Sessions are not compatible (e.g., tmux vs non-tmux)
  - Either session has no tab
  - Either session is locked
  - Panes are maximized (will auto-unmaximize)`,
		Example: cmdutil.Doc(`
			# Move current session to split below another session (horizontal)
			$ it2 session move $ITERM_SESSION_ID sess_abc123

			# Move session to the left of destination (vertical split before)
			$ it2 session move sess_abc123 sess_def456 --vertical --before

			# Move session to the right of destination (vertical split after)
			$ it2 session move sess_abc123 sess_def456 --vertical

			# Move with JSON output for scripting
			$ it2 session move sess_abc123 sess_def456 --json

			# Quiet mode - no output unless error
			$ it2 session move sess_abc123 sess_def456 --quiet
		`),
		Args: cobra.ExactArgs(2),
		RunE: runMove,
	}

	cmd.Flags().BoolP("vertical", "v", false, "Split destination vertically (default: horizontal)")
	cmd.Flags().BoolP("before", "b", false, "Place source before/above destination (default: after/below)")
	cmd.Flags().Bool("json", false, "Output result as JSON")
	cmd.Flags().BoolP("quiet", "q", false, "Quiet mode - minimal output")

	return cmd
}

func runMove(cmd *cobra.Command, args []string) error {
	sourceID := args[0]
	destID := args[1]

	vertical, _ := cmd.Flags().GetBool("vertical")
	before, _ := cmd.Flags().GetBool("before")
	jsonOutput, _ := cmd.Flags().GetBool("json")
	quiet, _ := cmd.Flags().GetBool("quiet")

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

	// Resolve both session IDs with environment fallback and prefix matching
	sourceID, err = c.ResolveSessionID(ctx, sourceID)
	if err != nil {
		return fmt.Errorf("failed to resolve source session ID: %w", err)
	}

	destID, err = c.ResolveSessionID(ctx, destID)
	if err != nil {
		return fmt.Errorf("failed to resolve destination session ID: %w", err)
	}

	// Move the session
	if err := c.MoveSession(ctx, sourceID, destID, vertical, before); err != nil {
		return fmt.Errorf("failed to move session: %w", err)
	}

	// Output results
	if quiet {
		return nil
	}

	if jsonOutput {
		result := map[string]interface{}{
			"success":     true,
			"source":      sourceID,
			"destination": destID,
			"vertical":    vertical,
			"before":      before,
		}
		return formatting.PrintJSON(result)
	}

	direction := "horizontal"
	if vertical {
		direction = "vertical"
	}
	position := "after"
	if before {
		position = "before"
	}

	fmt.Printf("Successfully moved session %s to %s split %s %s\n",
		sourceID, direction, position, destID)

	return nil
}
