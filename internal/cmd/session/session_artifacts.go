package session

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/spf13/cobra"
	"github.com/tmc/it2/internal/client"
)

func newArtifactsCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "artifacts",
		Short: "Manage session artifact directories",
		Long: `Manage and inspect session artifact directories.

Session artifacts are stored in ~/.it2/sessions/<session-id>/ and include:
  - it2-monitor.jsonl: Event logs
  - correlated-network.harl: Network activity
  - plugins/: Plugin-specific data

This command helps you find and work with these directories.`,
		Example: `  # Get artifacts directory for current session
  $ it2 session artifacts directory

  # Get artifacts directory by session ID
  $ it2 session artifacts directory 5AC0

  # Get artifacts directory by PID
  $ it2 session artifacts directory 12345

  # Open artifacts directory in Finder
  $ open $(it2 session artifacts directory)

  # List all files in artifacts directory
  $ ls -la $(it2 session artifacts directory 5AC0)

  # Tail monitor log for current session
  $ tail -f $(it2 session artifacts directory)/it2-monitor.jsonl`,
	}

	cmd.AddCommand(newArtifactsDirectoryCommand())
	cmd.AddCommand(newArtifactsListCommand())
	cmd.AddCommand(newArtifactsOpenCommand())

	return cmd
}

func newArtifactsDirectoryCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "directory [session-id-or-pid]",
		Short: "Get the artifacts directory path for a session",
		Long: `Get the artifacts directory path for a session.

If no argument is provided, uses the current session ($ITERM_SESSION_ID).
The argument can be:
  - A session ID (e.g., 5AC0 or full UUID)
  - A process ID (e.g., 12345) - will lookup the session for that PID

The directory path is written to stdout, making it easy to use in scripts:
  $ cd $(it2 session artifacts directory)
  $ ls $(it2 session artifacts directory 5AC0)
  $ open $(it2 session artifacts directory)`,
		Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			wsURL, _ := cmd.Flags().GetString("url")
			timeout, _ := cmd.Flags().GetDuration("timeout")

			// Use parent command flags if not set
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
				// Use current session
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

			// Determine if input is a PID or session ID
			var sessionID string
			if isPID(sessionIDOrPID) {
				pid, _ := strconv.Atoi(sessionIDOrPID)
				foundSessionID, err := findSessionByPID(ctx, c, pid)
				if err != nil {
					return fmt.Errorf("failed to find session for PID %d: %w", pid, err)
				}
				sessionID = foundSessionID
			} else {
				// Resolve session ID
				resolvedID, err := c.ResolveSessionID(ctx, sessionIDOrPID)
				if err != nil {
					return fmt.Errorf("failed to resolve session ID: %w", err)
				}
				sessionID = resolvedID
			}

			// Build artifacts directory path
			homeDir, err := os.UserHomeDir()
			if err != nil {
				return fmt.Errorf("failed to get home directory: %w", err)
			}

			artifactsDir := filepath.Join(homeDir, ".it2", "sessions", sessionID)

			// Create directory if it doesn't exist
			if err := os.MkdirAll(artifactsDir, 0700); err != nil {
				return fmt.Errorf("failed to create artifacts directory: %w", err)
			}

			// Output just the path (for easy scripting)
			fmt.Println(artifactsDir)
			return nil
		},
	}

	return cmd
}

func newArtifactsListCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list [session-id-or-pid]",
		Short: "List all files in a session's artifacts directory",
		Long:  `List all files in a session's artifacts directory with sizes and modification times.`,
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			wsURL, _ := cmd.Flags().GetString("url")
			timeout, _ := cmd.Flags().GetDuration("timeout")
			showAll, _ := cmd.Flags().GetBool("all")

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

			var sessionID string
			if isPID(sessionIDOrPID) {
				pid, _ := strconv.Atoi(sessionIDOrPID)
				foundSessionID, err := findSessionByPID(ctx, c, pid)
				if err != nil {
					return fmt.Errorf("failed to find session for PID %d: %w", pid, err)
				}
				sessionID = foundSessionID
			} else {
				resolvedID, err := c.ResolveSessionID(ctx, sessionIDOrPID)
				if err != nil {
					return fmt.Errorf("failed to resolve session ID: %w", err)
				}
				sessionID = resolvedID
			}

			homeDir, err := os.UserHomeDir()
			if err != nil {
				return fmt.Errorf("failed to get home directory: %w", err)
			}

			artifactsDir := filepath.Join(homeDir, ".it2", "sessions", sessionID)

			if _, err := os.Stat(artifactsDir); os.IsNotExist(err) {
				return fmt.Errorf("artifacts directory does not exist: %s", artifactsDir)
			}

			fmt.Printf("Artifacts for session %s:\n", sessionID)
			fmt.Printf("Location: %s\n\n", artifactsDir)

			// Walk directory tree
			err = filepath.Walk(artifactsDir, func(path string, info os.FileInfo, err error) error {
				if err != nil {
					return err
				}

				// Skip hidden files unless --all
				if !showAll && strings.HasPrefix(info.Name(), ".") && path != artifactsDir {
					if info.IsDir() {
						return filepath.SkipDir
					}
					return nil
				}

				// Get relative path
				relPath, _ := filepath.Rel(artifactsDir, path)
				if relPath == "." {
					return nil
				}

				// Format output
				if info.IsDir() {
					fmt.Printf("📁 %s/\n", relPath)
				} else {
					size := humanizeSize(info.Size())
					modTime := info.ModTime().Format("2006-01-02 15:04:05")
					fmt.Printf("📄 %s (%s, %s)\n", relPath, size, modTime)
				}

				return nil
			})

			return err
		},
	}

	cmd.Flags().Bool("all", false, "Show hidden files")

	return cmd
}

func newArtifactsOpenCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "open [session-id-or-pid]",
		Short: "Open artifacts directory in Finder/file manager",
		Long:  `Open the session's artifacts directory in the system file manager (Finder on macOS).`,
		Args:  cobra.MaximumNArgs(1),
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

			var sessionID string
			if isPID(sessionIDOrPID) {
				pid, _ := strconv.Atoi(sessionIDOrPID)
				foundSessionID, err := findSessionByPID(ctx, c, pid)
				if err != nil {
					return fmt.Errorf("failed to find session for PID %d: %w", pid, err)
				}
				sessionID = foundSessionID
			} else {
				resolvedID, err := c.ResolveSessionID(ctx, sessionIDOrPID)
				if err != nil {
					return fmt.Errorf("failed to resolve session ID: %w", err)
				}
				sessionID = resolvedID
			}

			homeDir, err := os.UserHomeDir()
			if err != nil {
				return fmt.Errorf("failed to get home directory: %w", err)
			}

			artifactsDir := filepath.Join(homeDir, ".it2", "sessions", sessionID)

			if _, err := os.Stat(artifactsDir); os.IsNotExist(err) {
				return fmt.Errorf("artifacts directory does not exist: %s", artifactsDir)
			}

			// Open in file manager (macOS-specific, but could be extended)
			openCmd := exec.Command("open", artifactsDir)
			if err := openCmd.Run(); err != nil {
				return fmt.Errorf("failed to open directory: %w", err)
			}

			fmt.Printf("Opened: %s\n", artifactsDir)
			return nil
		},
	}

	return cmd
}

// Helper functions

func isPID(s string) bool {
	// Check if string is all digits and reasonable PID range
	if _, err := strconv.Atoi(s); err == nil {
		return len(s) < 10 // PIDs are typically < 10 digits
	}
	return false
}

func findSessionByPID(ctx context.Context, c *client.Client, pid int) (string, error) {
	// List all sessions
	sessions, err := c.ListSessions(ctx)
	if err != nil {
		return "", err
	}

	// Find session with matching PID (check both ShellPID and JobPID)
	for _, session := range sessions {
		if int(session.ShellPID) == pid || int(session.JobPID) == pid {
			return session.SessionID, nil
		}
	}

	return "", fmt.Errorf("no session found for PID %d", pid)
}

func humanizeSize(size int64) string {
	const unit = 1024
	if size < unit {
		return fmt.Sprintf("%d B", size)
	}
	div, exp := int64(unit), 0
	for n := size / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(size)/float64(div), "KMGTPE"[exp])
}
