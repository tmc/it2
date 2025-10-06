package attention

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"
	"github.com/tmc/it2/internal/client"
	"github.com/tmc/it2/internal/cmdutil"
	"github.com/tmc/it2/internal/completion"
)

// AttentionType represents the type of attention request
type AttentionType string

const (
	AttentionFireworks AttentionType = "fireworks"
	AttentionOnce      AttentionType = "once"
	AttentionFlash     AttentionType = "flash"
	AttentionStart     AttentionType = "start"
	AttentionStop      AttentionType = "stop"
)

// NewCommand creates the attention command with all subcommands.
func NewCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "attention",
		GroupID: "content",
		Short:   "Request attention with visual effects",
		Long: `Request attention with various visual effects like fireworks, bouncing dock icon, or screen flash.

Attention effects can be used to:
- Draw attention to important events or notifications
- Signal completion of long-running tasks
- Provide visual feedback for alerts`,
	}

	cmd.AddCommand(newFireworksCommand())
	cmd.AddCommand(newOnceCommand())
	cmd.AddCommand(newFlashCommand())
	cmd.AddCommand(newStartCommand())
	cmd.AddCommand(newStopCommand())

	return cmd
}

func newFireworksCommand() *cobra.Command {
	template := cmdutil.CommandTemplate{
		Use:   "fireworks [session-id]",
		Short: "Show fireworks animation",
		Long: `Display a fireworks animation effect in the terminal.

Examples:
  it2 attention fireworks                  # Show fireworks in current session
  it2 attention fireworks session123       # Show fireworks in specific session`,
		Args:           cobra.RangeArgs(0, 1),
		RequiresClient: true,
		SupportsFormat: true,
		ValidArgsFunc:  completion.SessionIDCompletion,
		RunE: func(sc *cmdutil.StandardCommand, args []string) error {
			return requestAttention(sc, args, AttentionFireworks)
		},
	}

	return cmdutil.NewCommandFromTemplate(template)
}

func newOnceCommand() *cobra.Command {
	template := cmdutil.CommandTemplate{
		Use:   "once [session-id]",
		Short: "Bounce dock icon once",
		Long: `Bounce the iTerm2 dock icon once to draw attention.

Examples:
  it2 attention once                       # Bounce dock icon for current session
  it2 attention once session123            # Bounce dock icon for specific session`,
		Args:           cobra.RangeArgs(0, 1),
		RequiresClient: true,
		SupportsFormat: true,
		ValidArgsFunc:  completion.SessionIDCompletion,
		RunE: func(sc *cmdutil.StandardCommand, args []string) error {
			return requestAttention(sc, args, AttentionOnce)
		},
	}

	return cmdutil.NewCommandFromTemplate(template)
}

func newFlashCommand() *cobra.Command {
	template := cmdutil.CommandTemplate{
		Use:   "flash [session-id]",
		Short: "Flash the screen",
		Long: `Flash the screen to draw attention.

Examples:
  it2 attention flash                      # Flash screen for current session
  it2 attention flash session123           # Flash screen for specific session`,
		Args:           cobra.RangeArgs(0, 1),
		RequiresClient: true,
		SupportsFormat: true,
		ValidArgsFunc:  completion.SessionIDCompletion,
		RunE: func(sc *cmdutil.StandardCommand, args []string) error {
			return requestAttention(sc, args, AttentionFlash)
		},
	}

	return cmdutil.NewCommandFromTemplate(template)
}

func newStartCommand() *cobra.Command {
	template := cmdutil.CommandTemplate{
		Use:   "start [session-id]",
		Short: "Start bouncing dock icon continuously",
		Long: `Start bouncing the iTerm2 dock icon continuously until stopped.

Examples:
  it2 attention start                      # Start bouncing dock icon
  it2 attention start session123           # Start bouncing for specific session`,
		Args:           cobra.RangeArgs(0, 1),
		RequiresClient: true,
		SupportsFormat: true,
		ValidArgsFunc:  completion.SessionIDCompletion,
		RunE: func(sc *cmdutil.StandardCommand, args []string) error {
			return requestAttention(sc, args, AttentionStart)
		},
	}

	return cmdutil.NewCommandFromTemplate(template)
}

func newStopCommand() *cobra.Command {
	template := cmdutil.CommandTemplate{
		Use:   "stop [session-id]",
		Short: "Stop bouncing dock icon",
		Long: `Stop the continuous dock icon bouncing started by 'it2 attention start'.

Examples:
  it2 attention stop                       # Stop bouncing dock icon
  it2 attention stop session123            # Stop bouncing for specific session`,
		Args:           cobra.RangeArgs(0, 1),
		RequiresClient: true,
		SupportsFormat: true,
		ValidArgsFunc:  completion.SessionIDCompletion,
		RunE: func(sc *cmdutil.StandardCommand, args []string) error {
			return requestAttention(sc, args, AttentionStop)
		},
	}

	return cmdutil.NewCommandFromTemplate(template)
}

// requestAttention sends an attention request escape sequence to the session
func requestAttention(sc *cmdutil.StandardCommand, args []string, attentionType AttentionType) error {
	var sessionID string
	if len(args) > 0 {
		sessionID = args[0]
	}

	// Resolve session ID with environment fallback and prefix matching
	ctx := sc.GetContext()
	sessionID, err := sc.GetClient().ResolveSessionID(ctx, sessionID)
	if err != nil {
		return sc.ReportError("resolve session ID", err)
	}

	// Send the attention request
	err = sendAttentionRequest(sc.GetClient(), ctx, sessionID, attentionType)
	if err != nil {
		return sc.ReportError("send attention request", err)
	}

	// Report success with JSON output support
	if sc.GetFlags().Format == "json" {
		result := map[string]interface{}{
			"session_id": sessionID,
			"type":       string(attentionType),
			"action":     "requested",
		}
		return sc.FormatOutput(result)
	}

	sc.ReportSuccess("Requested attention (%s) for session %s", attentionType, sessionID)
	return nil
}

// sendAttentionRequest sends the iTerm2 escape sequence for requesting attention
func sendAttentionRequest(c *client.Client, ctx context.Context, sessionID string, attentionType AttentionType) error {
	// iTerm2 escape sequence format: ESC ] 1337 ; RequestAttention=<type> BEL
	// where <type> is one of: fireworks, once, flash, start, stop
	escapeSequence := fmt.Sprintf("\x1b]1337;RequestAttention=%s\x07", attentionType)

	// Inject the escape sequence into the session
	err := c.InjectData(ctx, sessionID, []byte(escapeSequence))
	if err != nil {
		return fmt.Errorf("failed to inject attention escape sequence: %w", err)
	}

	return nil
}
