package session

import (
	"fmt"
	"os/exec"
	"strconv"
	"strings"

	"github.com/spf13/cobra"
	"github.com/tmc/it2/internal/cmdutil"
	"github.com/tmc/it2/internal/formatting"
)

// printProcessTree recursively prints the process tree
func printProcessTree(pid int, maxDepth int, showPIDs bool, level int) {
	if level >= maxDepth {
		return
	}

	// Get process name
	procName, err := getProcessName(pid)
	if err != nil {
		procName = "<unknown>"
	}

	// Print with indentation
	indent := strings.Repeat("  ", level)
	if showPIDs {
		fmt.Printf("%s└─ %s (PID %d)\n", indent, procName, pid)
	} else {
		fmt.Printf("%s└─ %s\n", indent, procName)
	}

	// Get parent PID
	cmd := exec.Command("ps", "-p", strconv.Itoa(pid), "-o", "ppid=")
	output, err := cmd.Output()
	if err != nil {
		return
	}

	ppidStr := strings.TrimSpace(string(output))
	ppid, err := strconv.Atoi(ppidStr)
	if err != nil || ppid <= 1 {
		return
	}

	// Recurse to parent
	printProcessTree(ppid, maxDepth, showPIDs, level+1)
}

func newProcessListCommand() *cobra.Command {
	var format string
	var maxDepth int
	var showPIDs bool

	cmd := &cobra.Command{
		Use:   "process-list [<session-id>]",
		Short: "List processes for a session",
		Long: `List and inspect processes associated with a session.

Supports multiple output formats:
  - text: Simple text output showing shell and job processes
  - json: JSON output with process details
  - yaml: YAML output with process details
  - tree: ASCII tree showing process hierarchy

Examples:
  # List processes for current session
  it2 session process-list

  # Show process tree
  it2 session process-list --format=tree

  # Show tree with PIDs
  it2 session process-list --format=tree --pids

  # JSON output
  it2 session process-list --format=json

  # Deeper tree traversal
  it2 session process-list --format=tree --depth 15`,
		Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			sessionID := ""
			if len(args) > 0 {
				sessionID = args[0]
			}

			timeout, _ := cmd.Flags().GetDuration("timeout")
			ctx, cancel := cmdutil.CreateContext(timeout)
			defer cancel()

			c, err := cmdutil.ConnectClient(ctx)
			if err != nil {
				return fmt.Errorf("failed to connect: %w", err)
			}
			defer c.Close()

			sessionID, err = c.ResolveSessionID(ctx, sessionID)
			if err != nil {
				return fmt.Errorf("failed to resolve session ID: %w", err)
			}

			// Get both PIDs
			jobPIDStr, _ := c.GetVariableWithScope(ctx, "session", sessionID, "jobPid")
			shellPIDStr, _ := c.GetVariableWithScope(ctx, "session", sessionID, "pid")

			var jobPID, shellPID int
			if jobPIDStr != "" {
				jobPID, _ = strconv.Atoi(jobPIDStr)
			}
			if shellPIDStr != "" {
				shellPID, _ = strconv.Atoi(shellPIDStr)
			}

			// Handle different formats
			switch format {
			case "tree":
				return printProcessTreeFormat(jobPID, shellPID, maxDepth, showPIDs)
			case "json", "yaml":
				return printProcessData(jobPID, shellPID, format)
			default:
				return printProcessText(jobPID, shellPID)
			}
		},
	}

	cmd.Flags().StringVar(&format, "format", "text", "Output format: text, json, yaml, tree")
	cmd.Flags().IntVar(&maxDepth, "depth", 10, "Maximum tree depth (tree format only)")
	cmd.Flags().BoolVar(&showPIDs, "pids", false, "Show PIDs in tree format")

	return cmd
}

func printProcessText(jobPID, shellPID int) error {
	fmt.Println("Session Processes")
	fmt.Println("=================")
	fmt.Println()

	if shellPID > 0 {
		shellProc, _ := getProcessName(shellPID)
		fmt.Printf("Shell PID:     %d\n", shellPID)
		fmt.Printf("Shell Process: %s\n", shellProc)
		fmt.Println()
	}

	if jobPID > 0 {
		jobProc, _ := getProcessName(jobPID)
		fmt.Printf("Job PID:       %d\n", jobPID)
		fmt.Printf("Job Process:   %s\n", jobProc)
	}

	return nil
}

func printProcessData(jobPID, shellPID int, format string) error {
	data := make(map[string]interface{})

	if shellPID > 0 {
		data["shell_pid"] = shellPID
		shellProc, _ := getProcessName(shellPID)
		data["shell_process"] = shellProc
	}

	if jobPID > 0 {
		data["job_pid"] = jobPID
		jobProc, _ := getProcessName(jobPID)
		data["job_process"] = jobProc
	}

	switch format {
	case "json":
		return formatting.PrintJSON(data)
	case "yaml":
		return formatting.PrintYAML(data)
	}

	return nil
}

func printProcessTreeFormat(jobPID, shellPID int, maxDepth int, showPIDs bool) error {
	// Print job process if available
	if jobPID > 0 {
		procName, _ := getProcessName(jobPID)
		if showPIDs {
			fmt.Printf("Job:   %s (PID %d)\n", procName, jobPID)
		} else {
			fmt.Printf("Job:   %s\n", procName)
		}
	}

	// Print shell and its ancestors
	if shellPID > 0 {
		fmt.Println("\nProcess Tree:")
		printProcessTree(shellPID, maxDepth, showPIDs, 0)
	}

	return nil
}
