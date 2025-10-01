package session

import (
	"fmt"
	"os/exec"
	"strconv"
	"strings"

	"github.com/spf13/cobra"
	"github.com/tmc/it2/internal/cmdutil"
)

func newHasProcessCommand() *cobra.Command {
	var maxDepth int
	var exactMatch bool

	cmd := &cobra.Command{
		Use:   "has-process <process-name> [<session-id>]",
		Short: "Check if a process name exists in session's process tree",
		Long: `Check if a process name exists in the session's process tree.

Searches the job PID first, then walks up the parent process chain from the shell PID.
Returns exit code 0 if found, 1 if not found.

Examples:
  # Check if current session is running node
  it2 session has-process node

  # Check if session ABC123 has vim in its process tree
  it2 session has-process vim ABC123

  # Exact match only (no substring matching)
  it2 session has-process --exact claude

  # Search up to 10 levels in parent chain
  it2 session has-process --depth 10 python`,
		Args: cobra.RangeArgs(1, 2),
		RunE: func(cmd *cobra.Command, args []string) error {
			processName := args[0]
			sessionID := ""
			if len(args) > 1 {
				sessionID = args[1]
			}

			quiet, _ := cmd.Flags().GetBool("quiet")
			verbose, _ := cmd.Flags().GetBool("verbose")

			timeout, _ := cmd.Flags().GetDuration("timeout")
			ctx, cancel := cmdutil.CreateContext(timeout)
			defer cancel()

			c, err := cmdutil.ConnectClient(ctx)
			if err != nil {
				return fmt.Errorf("failed to connect: %w", err)
			}
			defer c.Close()

			// Resolve session ID
			sessionID, err = c.ResolveSessionID(ctx, sessionID)
			if err != nil {
				return fmt.Errorf("failed to resolve session ID: %w", err)
			}

			// Get both PIDs
			jobPIDStr, _ := c.GetVariableWithScope(ctx, "session", sessionID, "jobPid")
			shellPIDStr, _ := c.GetVariableWithScope(ctx, "session", sessionID, "pid")

			// Check job PID first (most common case - active foreground process)
			if jobPIDStr != "" {
				if pid, err := strconv.Atoi(jobPIDStr); err == nil {
					if found, foundPID := checkProcess(pid, processName, exactMatch, verbose); found {
						if !quiet {
							fmt.Printf("Process '%s' found at PID %d (job)\n", processName, foundPID)
						}
						return nil
					}
				}
			}

			// Walk up parent chain from shell PID
			if shellPIDStr != "" {
				if pid, err := strconv.Atoi(shellPIDStr); err == nil {
					if found, foundPID := walkProcessTree(pid, processName, maxDepth, exactMatch, verbose); found {
						if !quiet {
							fmt.Printf("Process '%s' found at PID %d (ancestor)\n", processName, foundPID)
						}
						return nil
					}
				}
			}

			// Not found
			if verbose {
				fmt.Printf("Process '%s' not found in session %s\n", processName, sessionID)
			}
			return fmt.Errorf("process not found")
		},
	}

	cmd.Flags().IntVar(&maxDepth, "depth", 10, "Maximum depth to search in parent chain")
	cmd.Flags().BoolVar(&exactMatch, "exact", false, "Require exact match (no substring matching)")
	cmd.Flags().Bool("quiet", false, "Suppress output (only use exit code)")
	cmd.Flags().Bool("verbose", false, "Show detailed search information")

	return cmd
}

// checkProcess checks if a PID matches the process name
func checkProcess(pid int, processName string, exactMatch, verbose bool) (bool, int) {
	cmd := exec.Command("ps", "-p", strconv.Itoa(pid), "-o", "comm=")
	output, err := cmd.Output()
	if err != nil {
		return false, 0
	}

	procName := strings.TrimSpace(string(output))
	if verbose {
		fmt.Printf("Checking PID %d: %s\n", pid, procName)
	}

	matches := false
	if exactMatch {
		matches = procName == processName
	} else {
		matches = strings.Contains(procName, processName)
	}

	if matches {
		return true, pid
	}
	return false, 0
}

// walkProcessTree walks up the parent process chain
func walkProcessTree(startPID int, processName string, maxDepth int, exactMatch, verbose bool) (bool, int) {
	currentPID := startPID

	for i := 0; i < maxDepth; i++ {
		if currentPID <= 1 {
			break
		}

		// Check current process
		if found, foundPID := checkProcess(currentPID, processName, exactMatch, verbose); found {
			return true, foundPID
		}

		// Get parent PID
		cmd := exec.Command("ps", "-p", strconv.Itoa(currentPID), "-o", "ppid=")
		output, err := cmd.Output()
		if err != nil {
			break
		}

		ppid := strings.TrimSpace(string(output))
		if ppid == "" {
			break
		}

		parentPID, err := strconv.Atoi(ppid)
		if err != nil || parentPID <= 1 {
			break
		}

		currentPID = parentPID
	}

	return false, 0
}
