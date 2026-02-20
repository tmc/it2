package session

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"github.com/tmc/it2/internal/cmdcore"
	"github.com/tmc/it2/internal/cmdutil"
	"github.com/tmc/it2/internal/completion"
	"github.com/tmc/it2/internal/sessionid"
)

func newRunCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:          "run [<session-id>] <command>",
		Short:        "Send text to a session and execute it (appends newline)",
		SilenceUsage: true,
		Long: `Send text to a session and execute it by appending a newline.

This is a convenience wrapper around send-text that always appends a newline
character to execute the command. Equivalent to: send-text --skip-confirm <session-id> <text>\n

The session can be specified as a positional argument or with the -s flag.`,
		Example: cmdutil.Doc(`
			# Run a command in a specific session (positional)
			$ it2 session run 7AA9 "echo hello"

			# Run a command using -s flag
			$ it2 session run -s 7AA9 "echo hello"

			# Run a command in the current session
			$ it2 session run "ls -la"
		`),
		Args: cobra.RangeArgs(1, 2),
		RunE: func(cmd *cobra.Command, args []string) error {
			sessionFlag, _ := cmd.Flags().GetString("session")

			var rawSessionID, text string

			if sessionFlag != "" {
				// -s flag provided: all positional args are the command text
				rawSessionID = sessionFlag
				text = strings.Join(args, " ")
			} else if len(args) == 2 {
				// Two positional args: session-id and command
				rawSessionID = args[0]
				text = args[1]
			} else {
				// One positional arg: could be command (use env session) or session-id
				// Treat as command text, use env for session
				text = args[0]
			}

			if text == "" {
				return fmt.Errorf("command text is required")
			}

			// Append newline to execute the command
			text = text + "\n"

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

			sessionID, err := c.ResolveSessionID(ctx, rawSessionID)
			if err != nil {
				return fmt.Errorf("failed to resolve session ID: %w", err)
			}

			// Log session context to stderr
			srcSessionID := sessionid.Normalize(os.Getenv("ITERM_SESSION_ID"))
			srcShort := sessionid.Shorten(srcSessionID)
			dstShort := sessionid.Shorten(sessionID)
			fmt.Fprintf(os.Stderr, "[it2:run src=%s dst=%s]\n", srcShort, dstShort)

			// Check for self-send unless explicitly allowed
			allowSelf, _ := cmd.Flags().GetBool("allow-self")
			if !allowSelf && srcSessionID != "" && srcSessionID == sessionID {
				return fmt.Errorf("refusing to run command in the same session (src=%s dst=%s); use --allow-self to override", srcShort, dstShort)
			}

			if err := c.SendText(ctx, sessionID, text); err != nil {
				return fmt.Errorf("failed to send command: %w", err)
			}

			return nil
		},
	}

	cmd.Flags().StringP("session", "s", "", "Target session ID")
	cmd.Flags().Bool("allow-self", false, "Allow running command in the same session (disabled by default for safety)")

	cmd.ValidArgsFunction = completion.SessionIDCompletion

	return cmd
}
