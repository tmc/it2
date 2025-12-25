package session

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/tmc/it2/internal/cmdutil"
	"github.com/tmc/it2/internal/completion"
	"github.com/tmc/it2/internal/formatting"
	"github.com/tmc/it2/internal/sessionid"
	"github.com/tmc/it2/internal/sessionstate"

	// Register agent profiles
	_ "github.com/tmc/it2/internal/sessionstate/agents/claude"
)

func newHasModalCommand() *cobra.Command {
	template := cmdutil.CommandTemplate{
		Use:   "has-modal [<session-id>]",
		Short: "Check if a session is showing a modal dialog",
		Long: `Detect if a session is showing a modal dialog that requires user interaction.

Returns exit code 0 if a modal is present, 1 if not.
Prints the modal type to stdout:
  - "approval" - "Do you want to proceed?" style prompts
  - "choice" - Numbered option selections
  - "confirmation" - Simple yes/no prompts
  - "input" - Text input prompts
  - "none" - No modal detected

With --json, outputs detailed modal information including safety assessment.

Examples:
  # Check for any modal
  it2 session has-modal E0A8 && echo "needs approval"

  # Get modal type
  modal=$(it2 session has-modal E0A8)
  case $modal in
    approval) echo "Auto-approving..." ;;
    choice) echo "Selecting option 1..." ;;
    *) echo "No modal" ;;
  esac

  # Get detailed modal info as JSON
  it2 session has-modal E0A8 --format json`,
		Args:           cobra.RangeArgs(0, 1),
		RequiresClient: true,
		SupportsFormat: true,
		ValidArgsFunc:  completion.SessionIDCompletion,
		RunE: func(sc *cmdutil.StandardCommand, args []string) error {
			var sessionID string
			if len(args) > 0 {
				sessionID = args[0]
			}

			ctx := sc.GetContext()
			sessionID, err := sc.GetClient().ResolveSessionID(ctx, sessionID)
			if err != nil {
				return sc.ReportError("resolve session ID", err)
			}

			// Log session context
			srcSessionID := sessionid.Normalize(os.Getenv("ITERM_SESSION_ID"))
			srcShort := sessionid.Shorten(srcSessionID)
			dstShort := sessionid.Shorten(sessionID)
			fmt.Fprintf(os.Stderr, "[it2:has-modal src=%s dst=%s]\n", srcShort, dstShort)

			// Get agent flag
			agentName, _ := sc.GetCommand().Flags().GetString("agent")
			if agentName == "" {
				agentName = "claude" // Default to claude for modal detection
			}

			// Detect state with agent detection enabled for modal patterns
			opts := sessionstate.DetectOptions{
				AgentName:   agentName,
				RecentLines: 20,
			}

			state, err := sessionstate.DetectState(ctx, sc.GetClient(), sessionID, opts)
			if err != nil {
				return sc.ReportError("detect state", err)
			}

			// Check for modal
			hasModal := state.Agent != nil &&
				state.Agent.Modal != nil &&
				state.Agent.Modal.Type != sessionstate.ModalNone

			// Handle JSON output
			format := sc.GetFlags().Format
			if format == "json" || format == "yaml" {
				result := struct {
					HasModal      bool                   `json:"has_modal"`
					ModalType     sessionstate.ModalType `json:"modal_type"`
					SafeToApprove bool                   `json:"safe_to_approve"`
				}{
					HasModal:      hasModal,
					ModalType:     sessionstate.ModalNone,
					SafeToApprove: false,
				}

				if hasModal {
					result.ModalType = state.Agent.Modal.Type
					result.SafeToApprove = state.Agent.Modal.SafeToApprove
				}

				formatter := formatting.New(format)
				return formatter.FormatGeneric(result)
			}

			// Plain text output
			if hasModal {
				fmt.Println(state.Agent.Modal.Type)
				return nil
			}

			fmt.Println("none")
			os.Exit(1)
			return nil
		},
	}

	cmd := cmdutil.NewCommandFromTemplate(template)
	cmd.SilenceUsage = true

	// Add command-specific flags
	cmd.Flags().String("agent", "", "Agent to detect: claude (default), auto, or empty (generic)")

	return cmd
}
