package session

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"github.com/spf13/cobra"
	"github.com/tmc/it2/internal/client"
)

func newExportRecordingCommand() *cobra.Command {
	var outputPath string
	var autoName bool

	cmd := &cobra.Command{
		Use:   "export-recording [session-id]",
		Short: "Export session recording to .itr file",
		Long: `Export the iTerm2 instant replay (DVR) recording for a session to a .itr file.

iTerm2 continuously records terminal sessions for instant replay. This command
exports that recording to an .itr file that can be replayed later.

The recording includes all terminal output, timing information, and the session
profile used at the time of recording.

By default, recordings are saved to the session's artifacts directory:
  ~/.it2/sessions/<session-id>/recordings/

You can specify a custom output path with --output, or use --auto-name to
generate a timestamped filename in the artifacts directory.`,
		Example: `  # Export current session recording with auto-generated filename
  $ it2 session export-recording --auto-name

  # Export specific session by ID
  $ it2 session export-recording 5AC0 --auto-name

  # Export to custom path
  $ it2 session export-recording --output ~/my-recording.itr

  # Export and open in Finder
  $ it2 session export-recording --auto-name && open $(it2 session artifacts directory)/recordings/`,
		Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			wsURL, _ := cmd.Flags().GetString("url")
			timeout, _ := cmd.Flags().GetDuration("timeout")

			if wsURL == "" {
				wsURL = cmd.Root().PersistentFlags().Lookup("url").Value.String()
			}
			if timeout == 0 {
				timeout, _ = cmd.Root().PersistentFlags().GetDuration("timeout")
			}

			var sessionIDOrPID string
			if len(args) > 0 {
				sessionIDOrPID = args[0]
			} else {
				sessionIDOrPID = os.Getenv("ITERM_SESSION_ID")
				if sessionIDOrPID == "" {
					return fmt.Errorf("no session ID provided and ITERM_SESSION_ID not set")
				}
			}

			ctx, cancel := context.WithTimeout(context.Background(), timeout)
			defer cancel()

			c := client.New(wsURL)
			if err := c.Connect(ctx); err != nil {
				return fmt.Errorf("failed to connect: %w", err)
			}
			defer c.Close()

			// Resolve session ID
			sessionID, err := c.ResolveSessionID(ctx, sessionIDOrPID)
			if err != nil {
				return fmt.Errorf("failed to resolve session ID: %w", err)
			}

			// Determine output path
			var finalOutputPath string
			if outputPath != "" && autoName {
				return fmt.Errorf("cannot specify both --output and --auto-name")
			}

			if autoName || outputPath == "" {
				// Use artifacts directory
				homeDir, err := os.UserHomeDir()
				if err != nil {
					return fmt.Errorf("failed to get home directory: %w", err)
				}

				recordingsDir := filepath.Join(homeDir, ".it2", "sessions", sessionID, "recordings")
				if err := os.MkdirAll(recordingsDir, 0755); err != nil {
					return fmt.Errorf("failed to create recordings directory: %w", err)
				}

				timestamp := time.Now().Format("2006-01-02-150405")
				filename := fmt.Sprintf("recording-%s.itr", timestamp)
				finalOutputPath = filepath.Join(recordingsDir, filename)
			} else {
				finalOutputPath = outputPath
				// Expand ~ to home directory
				if finalOutputPath[0] == '~' {
					homeDir, err := os.UserHomeDir()
					if err != nil {
						return fmt.Errorf("failed to get home directory: %w", err)
					}
					finalOutputPath = filepath.Join(homeDir, finalOutputPath[1:])
				}
				// Ensure .itr extension
				if filepath.Ext(finalOutputPath) != ".itr" {
					finalOutputPath += ".itr"
				}
			}

			// Ensure parent directory exists
			if err := os.MkdirAll(filepath.Dir(finalOutputPath), 0755); err != nil {
				return fmt.Errorf("failed to create output directory: %w", err)
			}

			// Use AppleScript to trigger the export
			// This is necessary because iTerm2's API doesn't expose recording export
			script := fmt.Sprintf(`
tell application "iTerm2"
	tell current session of current window
		-- Get the session with matching ID
		set targetSession to my findSessionByID("%s")
		if targetSession is not missing value then
			-- Focus the session first
			select targetSession
			delay 0.2

			-- Try to invoke the export menu item
			tell application "System Events"
				tell process "iTerm2"
					-- Session > Export Recording...
					-- Note: This path may vary by iTerm2 version
					try
						click menu item "Export Recording…" of menu "Session" of menu bar 1
						delay 0.5

						-- Handle save dialog
						keystroke "g" using {command down, shift down} -- Go to folder
						delay 0.3
						keystroke "%s" -- Directory path
						keystroke return
						delay 0.3
						keystroke "%s" -- Filename
						delay 0.2
						keystroke return -- Save
					on error errMsg
						error "Failed to export recording: " & errMsg
					end try
				end tell
			end tell
		else
			error "Session not found: %s"
		end if
	end tell
end tell

on findSessionByID(targetID)
	tell application "iTerm2"
		repeat with w in windows
			repeat with t in tabs of w
				repeat with s in sessions of t
					if id of s is equal to targetID then
						return s
					end if
				end repeat
			end repeat
		end repeat
	end tell
	return missing value
end findSessionByID
`, sessionID, filepath.Dir(finalOutputPath), filepath.Base(finalOutputPath), sessionID)

			// Execute AppleScript
			appleScriptCmd := exec.Command("osascript", "-e", script)
			output, err := appleScriptCmd.CombinedOutput()
			if err != nil {
				return fmt.Errorf("failed to export recording via AppleScript: %w\nOutput: %s", err, string(output))
			}

			fmt.Printf("Recording exported successfully:\n")
			fmt.Printf("  Session: %s\n", sessionID)
			fmt.Printf("  File: %s\n", finalOutputPath)

			// Verify file was created
			if _, err := os.Stat(finalOutputPath); os.IsNotExist(err) {
				fmt.Println("\nWarning: Expected output file not found. The recording may have been saved to a different location.")
				fmt.Println("Please check iTerm2's save dialog and the session artifacts directory:")
				homeDir, _ := os.UserHomeDir()
				recordingsDir := filepath.Join(homeDir, ".it2", "sessions", sessionID, "recordings")
				fmt.Printf("  %s\n", recordingsDir)
			}

			return nil
		},
	}

	cmd.Flags().StringVarP(&outputPath, "output", "o", "", "Output file path (default: artifacts/recordings/recording-TIMESTAMP.itr)")
	cmd.Flags().BoolVar(&autoName, "auto-name", false, "Auto-generate timestamped filename in artifacts directory")

	return cmd
}
