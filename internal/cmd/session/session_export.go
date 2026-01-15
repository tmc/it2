package session

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"github.com/spf13/cobra"
	"github.com/tmc/it2/internal/cmdcore"
	"github.com/tmc/it2/internal/formatting"
)

// SessionExport represents exported session metadata for forensic correlation.
type SessionExport struct {
	SessionID   string            `json:"session_id"`
	ShortID     string            `json:"short_id"`
	StartTime   *time.Time        `json:"start_time,omitempty"`
	EndTime     *time.Time        `json:"end_time,omitempty"`
	WorkingDirs []string          `json:"working_directories,omitempty"`
	CurrentDir  string            `json:"current_directory,omitempty"`
	ShellPID    int32             `json:"shell_pid,omitempty"`
	JobPID      int32             `json:"job_pid,omitempty"`
	Process     string            `json:"process,omitempty"`
	Tags        map[string]string `json:"tags,omitempty"`
	ESLogs      *ESLogsInfo       `json:"eslogs,omitempty"`
}

// ESLogsInfo contains eslogs data for the session.
type ESLogsInfo struct {
	Available  bool       `json:"available"`
	EventCount int64      `json:"event_count,omitempty"`
	ExecCount  int64      `json:"exec_count,omitempty"`
	FileOps    int64      `json:"file_ops,omitempty"`
	FirstEvent *time.Time `json:"first_event,omitempty"`
	LastEvent  *time.Time `json:"last_event,omitempty"`
	Agent      string     `json:"agent,omitempty"`
}

func newExportCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "export [session-id]",
		Short: "Export session metadata for forensic correlation",
		Long: `Export session metadata in a structured format for forensic analysis.

This command collects:
- Session ID and timestamps
- Working directories visited (from prompt history)
- Process information
- Session tags
- ESLogs summary (if available)

The output is designed for correlation with eslogs terminal queries.`,
		Example: `  # Export current session metadata
  it2 session export

  # Export specific session as JSON
  it2 session export 5AC0 --format json

  # Export with eslogs data
  it2 session export --include-eslogs

  # Use with eslogs for correlation
  it2 session export --format json | jq -r .session_id | xargs eslogs terminal`,
		Args: cobra.MaximumNArgs(1),
		RunE: runExportCommand,
	}

	cmd.Flags().String("format", "json", "Output format: json, yaml, text")
	cmd.Flags().Bool("include-eslogs", false, "Include eslogs summary data")

	return cmd
}

func runExportCommand(cmd *cobra.Command, args []string) error {
	format, _ := cmd.Flags().GetString("format")
	includeEslogs, _ := cmd.Flags().GetBool("include-eslogs")
	timeout, _ := cmd.Flags().GetDuration("timeout")

	if timeout == 0 {
		timeout, _ = cmd.Root().PersistentFlags().GetDuration("timeout")
	}

	ctx, cancel := cmdcore.CreateContext(timeout)
	defer cancel()

	c, err := cmdcore.ConnectClient(ctx)
	if err != nil {
		return fmt.Errorf("connect: %w", err)
	}
	defer c.Close()

	// Get session ID
	var sessionID string
	if len(args) > 0 {
		sessionID = args[0]
	} else {
		sessionID = os.Getenv("ITERM_SESSION_ID")
	}

	if sessionID == "" {
		return fmt.Errorf("no session ID provided and ITERM_SESSION_ID not set")
	}

	// Resolve session ID
	resolved, err := c.ResolveSessionID(ctx, sessionID)
	if err != nil {
		return fmt.Errorf("resolve session: %w", err)
	}
	sessionID = resolved

	// Build export data
	export := SessionExport{
		SessionID: sessionID,
		ShortID:   sessionID[:8],
	}

	// Get session info from iTerm2
	sessions, err := c.ListSessions(ctx)
	if err != nil {
		return fmt.Errorf("list sessions: %w", err)
	}

	for _, sess := range sessions {
		if sess.SessionID == sessionID {
			export.ShellPID = sess.ShellPID
			export.JobPID = sess.JobPID
			if sess.JobPID > 0 {
				if name, err := getProcessName(int(sess.JobPID)); err == nil {
					export.Process = name
				}
			}
			break
		}
	}

	// Get prompt info for working directory
	if prompt, err := c.GetPrompt(ctx, sessionID); err == nil {
		if wd := prompt.GetWorkingDirectory(); wd != "" {
			export.CurrentDir = wd
			export.WorkingDirs = []string{wd}
		}
	}

	// Load tags
	if tags, err := LoadSessionTags(sessionID); err == nil && len(tags) > 0 {
		export.Tags = tags
	}

	// Load eslogs data if requested
	if includeEslogs {
		export.ESLogs = getESLogsInfo(sessionID)
	}

	// Output
	switch format {
	case "json":
		return formatting.PrintJSON(export)
	case "yaml":
		return formatting.PrintYAML(export)
	case "text":
		return printExportText(export)
	default:
		return fmt.Errorf("unsupported format: %s", format)
	}
}

func getESLogsInfo(sessionID string) *ESLogsInfo {
	info := &ESLogsInfo{Available: false}

	// Find eslogs binary
	eslogsBin := findEslogsBinary()
	if eslogsBin == "" {
		return info
	}

	// Check database exists
	dbPaths := []string{
		os.Getenv("ESLOGS_DB"),
		os.ExpandEnv("$HOME/.eslogs/eslogs.db"),
		os.ExpandEnv("$HOME/.eslogs/events.db"),
	}
	dbFound := false
	for _, p := range dbPaths {
		if p == "" {
			continue
		}
		if _, err := os.Stat(p); err == nil {
			dbFound = true
			break
		}
	}
	if !dbFound {
		return info
	}

	// Query eslogs for session info
	cmd := exec.Command(eslogsBin, "terminal", sessionID, "--format", "json", "--since", "720h")
	output, err := cmd.Output()
	if err != nil {
		return info
	}

	// Parse JSON output
	var result struct {
		EventCount int64  `json:"event_count"`
		ExecCount  int64  `json:"exec_count"`
		FileOps    int64  `json:"file_ops"`
		FirstEvent string `json:"first_event"`
		LastEvent  string `json:"last_event"`
		Agent      string `json:"agent"`
	}

	if err := json.Unmarshal(output, &result); err != nil {
		return info
	}

	info.Available = true
	info.EventCount = result.EventCount
	info.ExecCount = result.ExecCount
	info.FileOps = result.FileOps
	info.Agent = result.Agent

	if result.FirstEvent != "" {
		if t, err := time.Parse(time.RFC3339, result.FirstEvent); err == nil {
			info.FirstEvent = &t
		}
	}
	if result.LastEvent != "" {
		if t, err := time.Parse(time.RFC3339, result.LastEvent); err == nil {
			info.LastEvent = &t
		}
	}

	return info
}

func printExportText(export SessionExport) error {
	fmt.Printf("Session Export\n")
	fmt.Printf("==============\n\n")

	fmt.Printf("Session ID:    %s\n", export.SessionID)
	fmt.Printf("Short ID:      %s\n", export.ShortID)

	if export.ShellPID > 0 {
		fmt.Printf("Shell PID:     %d\n", export.ShellPID)
	}
	if export.JobPID > 0 {
		fmt.Printf("Job PID:       %d\n", export.JobPID)
	}
	if export.Process != "" {
		fmt.Printf("Process:       %s\n", export.Process)
	}
	if export.CurrentDir != "" {
		fmt.Printf("Working Dir:   %s\n", export.CurrentDir)
	}

	if len(export.Tags) > 0 {
		fmt.Printf("\nTags:\n")
		for k, v := range export.Tags {
			fmt.Printf("  %s=%s\n", k, v)
		}
	}

	if export.ESLogs != nil {
		fmt.Printf("\nESLogs Data:\n")
		if !export.ESLogs.Available {
			fmt.Printf("  (not available)\n")
		} else {
			fmt.Printf("  Events:      %d\n", export.ESLogs.EventCount)
			fmt.Printf("  Execs:       %d\n", export.ESLogs.ExecCount)
			fmt.Printf("  File Ops:    %d\n", export.ESLogs.FileOps)
			if export.ESLogs.Agent != "" {
				fmt.Printf("  Agent:       %s\n", export.ESLogs.Agent)
			}
			if export.ESLogs.FirstEvent != nil {
				fmt.Printf("  First Event: %s\n", export.ESLogs.FirstEvent.Format(time.RFC3339))
			}
			if export.ESLogs.LastEvent != nil {
				fmt.Printf("  Last Event:  %s\n", export.ESLogs.LastEvent.Format(time.RFC3339))
			}
		}
	}

	// Print eslogs command hint
	fmt.Printf("\n# Query eslogs events:\n")
	fmt.Printf("eslogs terminal %s\n", export.ShortID)

	return nil
}

// getArtifactsDir returns the artifacts directory for a session
func getArtifactsDir(sessionID string) string {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return ""
	}
	return filepath.Join(homeDir, ".it2", "sessions", sessionID)
}

// findWorkingDirsFromHistory tries to find working directories from session history
func findWorkingDirsFromHistory(sessionID string) []string {
	artifactsDir := getArtifactsDir(sessionID)
	if artifactsDir == "" {
		return nil
	}

	// Check if monitor log exists
	monitorLog := filepath.Join(artifactsDir, "it2-monitor.jsonl")
	if _, err := os.Stat(monitorLog); err != nil {
		return nil
	}

	// Read and parse monitor log for working directory changes
	// This would parse JSONL looking for prompt events with working_directory
	// For now, return nil as this is a future enhancement
	return nil
}
