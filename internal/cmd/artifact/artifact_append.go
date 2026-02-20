package artifact

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/spf13/cobra"
	"github.com/tmc/it2/internal/cmdutil"
	"github.com/tmc/it2/internal/completion"
)

func newAppendCommand() *cobra.Command {
	template := cmdutil.CommandTemplate{
		Use:   "append <session-id> <artifact-name> <data>",
		Short: "Append data to a JSONL artifact",
		Long: `Append a JSON line to a JSONL (JSON Lines) artifact file.

JSONL files are newline-delimited JSON, where each line is a valid JSON object.
This is ideal for:
  - Event logs (command executions, notifications, state changes)
  - Time-series data
  - Append-only logs that can be streamed and processed line-by-line

If the artifact doesn't exist, it will be created.
Each line is automatically timestamped.

Examples:
  # Append a send-text event
  $ it2 artifact append sess_abc123 it2_send-text \\
      '{"source":"sess_abc","dest":"sess_def","text":"echo hello"}'

  # Append a state change
  $ it2 artifact append sess_abc123 state-changes \\
      '{"component":"session","old":"idle","new":"exec"}'

  # Append from stdin
  $ echo '{"event":"completed"}' | it2 artifact append sess_abc123 events -

  # Append with auto-timestamp
  $ it2 artifact append sess_abc123 commands \\
      '{"command":"session split","args":["--horizontal"]}'`,
		Args:           cobra.RangeArgs(2, 3),
		RequiresClient: false,
		ValidArgsFunc:  completion.SessionIDCompletion,
		RunE: func(sc *cmdutil.StandardCommand, args []string) error {
			sessionID := args[0]
			artifactName := args[1]

			var data string
			if len(args) == 3 {
				data = args[2]
			} else {
				// Read from stdin
				stdinData, err := os.ReadFile("/dev/stdin")
				if err != nil {
					return sc.ReportError("read stdin", err)
				}
				data = string(stdinData)
			}

			// Get flags
			timestamp, _ := sc.GetCommand().Flags().GetBool("timestamp")

			// Append to artifact
			if err := appendToArtifact(sessionID, artifactName, data, timestamp); err != nil {
				return sc.ReportError("append to artifact", err)
			}

			fmt.Fprintf(sc.GetCommand().OutOrStdout(), "Appended to %s\n", artifactName)

			return nil
		},
	}

	cmd := cmdutil.NewCommandFromTemplate(template)
	cmd.Flags().Bool("timestamp", true, "Automatically add timestamp field")

	return cmd
}

func appendToArtifact(sessionID, artifactName, data string, addTimestamp bool) error {
	// Validate JSON
	var jsonObj map[string]interface{}
	if err := json.Unmarshal([]byte(data), &jsonObj); err != nil {
		return fmt.Errorf("invalid JSON: %w", err)
	}

	// Add standard metadata (timestamp, session_id, pid, tty)
	if addTimestamp {
		jsonObj["timestamp"] = time.Now().Format(time.RFC3339Nano)
		if sessionID != "" {
			jsonObj["session_id"] = sessionID
		}
		jsonObj["pid"] = os.Getpid()
		if tty := os.Getenv("TTY"); tty != "" {
			jsonObj["tty"] = tty
		}
	}

	// Re-marshal with enriched metadata
	jsonLine, err := json.Marshal(jsonObj)
	if err != nil {
		return fmt.Errorf("failed to marshal JSON: %w", err)
	}

	// Get artifact path
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return err
	}

	// Store JSONL files in custom category
	artifactDir := filepath.Join(homeDir, ".it2", "sessions", sessionID, "artifacts", "jsonl")
	if err := os.MkdirAll(artifactDir, 0700); err != nil {
		return fmt.Errorf("failed to create artifact directory: %w", err)
	}

	// Ensure .jsonl extension
	if filepath.Ext(artifactName) != ".jsonl" {
		artifactName += ".jsonl"
	}

	artifactPath := filepath.Join(artifactDir, artifactName)

	// Append to file (create if doesn't exist)
	f, err := os.OpenFile(artifactPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("failed to open artifact file: %w", err)
	}
	defer f.Close()

	// Write JSON line with newline
	if _, err := f.Write(append(jsonLine, '\n')); err != nil {
		return fmt.Errorf("failed to write to artifact: %w", err)
	}

	return nil
}
