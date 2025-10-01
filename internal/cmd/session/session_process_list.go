package session

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"text/tabwriter"

	"github.com/spf13/cobra"
	"github.com/tmc/it2/internal/cmdutil"
	"github.com/tmc/it2/internal/formatting"
	"github.com/tmc/it2/internal/plugins"
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
  - table: Tabular output with process details and plugin data (default)
  - text: Simple text output showing shell and job processes
  - json: JSON output with process details
  - yaml: YAML output with process details
  - tree: ASCII tree showing process hierarchy

Examples:
  # List processes for current session (table view)
  it2 session process list

  # Show process tree
  it2 session process list --format=tree

  # Show tree with PIDs
  it2 session process list --format=tree --pids

  # JSON output
  it2 session process list --format=json

  # Deeper tree traversal
  it2 session process list --format=tree --depth 15`,
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
			case "text":
				return printProcessText(jobPID, shellPID)
			default: // "table"
				return printProcessTable(ctx, sessionID, jobPID, shellPID)
			}
		},
	}

	cmd.Flags().StringVar(&format, "format", "table", "Output format: table, text, json, yaml, tree")
	cmd.Flags().IntVar(&maxDepth, "depth", 10, "Maximum tree depth (tree format only)")
	cmd.Flags().BoolVar(&showPIDs, "pids", false, "Show PIDs in tree format")

	return cmd
}

func printProcessText(jobPID, shellPID int) error {
	fmt.Println("Session Processes")
	fmt.Println("=================")
	fmt.Println()

	// Collect process chain from job to shell
	var chain []struct {
		pid  int
		name string
	}

	// Start from job PID and walk up to shell PID
	currentPID := jobPID
	for currentPID > 0 {
		procName, _ := getProcessName(currentPID)
		chain = append(chain, struct {
			pid  int
			name string
		}{currentPID, procName})

		// Stop if we reached the shell PID
		if currentPID == shellPID {
			break
		}

		// Get parent PID
		cmd := exec.Command("ps", "-p", strconv.Itoa(currentPID), "-o", "ppid=")
		output, err := cmd.Output()
		if err != nil {
			break
		}

		ppidStr := strings.TrimSpace(string(output))
		ppid, err := strconv.Atoi(ppidStr)
		if err != nil || ppid <= 1 {
			break
		}

		currentPID = ppid

		// Safety: limit depth
		if len(chain) > 20 {
			break
		}
	}

	// Print the chain
	if len(chain) > 0 {
		fmt.Println("Process Chain (job → shell):")
		for i, proc := range chain {
			prefix := "├─"
			if i == len(chain)-1 {
				prefix = "└─"
			}
			fmt.Printf("%s PID %d: %s\n", prefix, proc.pid, proc.name)
		}
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
	fmt.Println("Process Tree:")

	// Start from job PID if available, otherwise shell PID
	startPID := jobPID
	if startPID == 0 {
		startPID = shellPID
	}

	if startPID > 0 {
		printProcessTree(startPID, maxDepth, showPIDs, 0)
	}

	return nil
}

func printProcessTable(ctx context.Context, sessionID string, jobPID, shellPID int) error {
	// Collect process chain from job to shell
	var chain []struct {
		pid  int
		name string
	}

	// Start from job PID and walk up to shell PID
	currentPID := jobPID
	for currentPID > 0 {
		procName, _ := getProcessName(currentPID)
		chain = append(chain, struct {
			pid  int
			name string
		}{currentPID, procName})

		// Stop if we reached the shell PID
		if currentPID == shellPID {
			break
		}

		// Get parent PID
		cmd := exec.Command("ps", "-p", strconv.Itoa(currentPID), "-o", "ppid=")
		output, err := cmd.Output()
		if err != nil {
			break
		}

		ppidStr := strings.TrimSpace(string(output))
		ppid, err := strconv.Atoi(ppidStr)
		if err != nil || ppid <= 1 {
			break
		}

		currentPID = ppid

		// Safety: limit depth
		if len(chain) > 20 {
			break
		}
	}

	// Discover process enricher plugins
	registry := plugins.NewRegistry()
	if err := registry.DiscoverAndRegister(); err == nil {
		processPlugins := registry.GetProcessEnrichers()

		// Create tabwriter for aligned columns
		w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
		defer w.Flush()

		// Print header
		if len(processPlugins) > 0 {
			fmt.Fprintf(w, "PID\tProcess\tRole\tPlugins\n")
			fmt.Fprintf(w, "---\t-------\t----\t-------\n")
		} else {
			fmt.Fprintf(w, "PID\tProcess\tRole\n")
			fmt.Fprintf(w, "---\t-------\t----\n")
		}

		// Print each process in the chain
		for i, proc := range chain {
			role := "intermediate"
			if proc.pid == jobPID {
				role = "job"
			} else if proc.pid == shellPID {
				role = "shell"
			}

			// Collect plugin data
			var pluginData []string
			for _, plugin := range processPlugins {
				data, err := plugin.EnrichProcess(ctx, sessionID, proc.pid)
				if err == nil && len(data) > 0 {
					// Format plugin output
					for key, value := range data {
						pluginData = append(pluginData, fmt.Sprintf("%s=%v", key, value))
					}
				}
			}

			if len(pluginData) > 0 {
				fmt.Fprintf(w, "%d\t%s\t%s\t%s\n", proc.pid, proc.name, role, strings.Join(pluginData, ", "))
			} else if len(processPlugins) > 0 {
				fmt.Fprintf(w, "%d\t%s\t%s\t-\n", proc.pid, proc.name, role)
			} else {
				fmt.Fprintf(w, "%d\t%s\t%s\n", proc.pid, proc.name, role)
			}

			_ = i // suppress unused variable warning
		}
	} else {
		// Fallback to simple table without plugins
		w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
		defer w.Flush()

		fmt.Fprintf(w, "PID\tProcess\tRole\n")
		fmt.Fprintf(w, "---\t-------\t----\n")

		for _, proc := range chain {
			role := "intermediate"
			if proc.pid == jobPID {
				role = "job"
			} else if proc.pid == shellPID {
				role = "shell"
			}
			fmt.Fprintf(w, "%d\t%s\t%s\n", proc.pid, proc.name, role)
		}
	}

	return nil
}
