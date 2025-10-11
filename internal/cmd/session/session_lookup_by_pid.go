package session

import (
	"context"
	"fmt"
	"strconv"

	"github.com/spf13/cobra"
	"github.com/tmc/it2/internal/client"
)

func newLookupByPIDCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "by-pid <pid>",
		Short: "Look up session ID by process ID",
		Long: `Look up which iTerm2 session contains a specific process ID.

Searches through all sessions to find which one has the given PID as either:
- ShellPID: The shell process ID
- JobPID: The currently running foreground job PID

This is useful for finding which session is running a specific process.`,
		Example: `  # Find session for PID 12345
  $ it2 session lookup by-pid 12345

  # Use in scripts to get artifacts for a PID
  $ it2 session artifacts directory $(it2 session lookup by-pid 12345)

  # Monitor the session running a specific process
  $ it2 session monitor $(it2 session lookup by-pid 98077) --tail --log`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			wsURL, _ := cmd.Flags().GetString("url")
			timeout, _ := cmd.Flags().GetDuration("timeout")
			verbose, _ := cmd.Flags().GetBool("verbose")

			// Use parent command flags if not set
			if wsURL == "" {
				wsURL = cmd.Root().PersistentFlags().Lookup("url").Value.String()
			}
			if timeout == 0 {
				timeout, _ = cmd.Root().PersistentFlags().GetDuration("timeout")
			}

			pidStr := args[0]
			pid, err := strconv.Atoi(pidStr)
			if err != nil {
				return fmt.Errorf("invalid PID: %s", pidStr)
			}

			ctx, cancel := context.WithTimeout(context.Background(), timeout)
			defer cancel()

			c := client.New(wsURL)
			if err := c.Connect(ctx); err != nil {
				return fmt.Errorf("failed to connect: %w", err)
			}
			defer c.Close()

			// List all sessions
			sessions, err := c.ListSessions(ctx)
			if err != nil {
				return fmt.Errorf("failed to list sessions: %w", err)
			}

			// Find session with matching PID
			var foundSession *client.SessionInfo
			var pidType string

			for _, session := range sessions {
				if int(session.ShellPID) == pid {
					foundSession = session
					pidType = "shell"
					break
				} else if int(session.JobPID) == pid {
					foundSession = session
					pidType = "job"
					break
				}
			}

			if foundSession == nil {
				return fmt.Errorf("no session found for PID %d", pid)
			}

			// Output
			if verbose {
				fmt.Printf("Session ID: %s\n", foundSession.SessionID)
				fmt.Printf("Short ID:   %s\n", foundSession.ShortID)
				fmt.Printf("PID Type:   %s\n", pidType)
				fmt.Printf("Shell PID:  %d\n", foundSession.ShellPID)
				fmt.Printf("Job PID:    %d\n", foundSession.JobPID)
				if foundSession.SessionName != "" {
					fmt.Printf("Name:       %s\n", foundSession.SessionName)
				}
				if foundSession.WorkingDirectory != "" {
					fmt.Printf("Directory:  %s\n", foundSession.WorkingDirectory)
				}
			} else {
				// Just output session ID for easy scripting
				fmt.Println(foundSession.SessionID)
			}

			return nil
		},
	}

	cmd.Flags().BoolP("verbose", "v", false, "Show detailed information")

	return cmd
}
