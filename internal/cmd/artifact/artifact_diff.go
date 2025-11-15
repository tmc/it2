package artifact

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/spf13/cobra"
	"github.com/tmc/it2/internal/cmdutil"
	"github.com/tmc/it2/internal/completion"
)

func newDiffCommand() *cobra.Command {
	template := cmdutil.CommandTemplate{
		Use:   "diff <session-id> <artifact1> <artifact2>",
		Short: "Show diff between two artifacts",
		Long: `Compare two artifacts and show their differences.

Uses the system's diff command to show differences between artifacts.

Examples:
  # Diff two buffer snapshots
  $ it2 artifact diff sess_abc123 \\
      buffer-snapshot/2025-10-10T14-30-00Z.txt \\
      buffer-snapshot/2025-10-10T15-00-00Z.txt

  # Unified diff format
  $ it2 artifact diff sess_abc123 file1.txt file2.txt --unified`,
		Args:           cobra.ExactArgs(3),
		RequiresClient: false,
		ValidArgsFunc:  completion.SessionIDCompletion,
		RunE: func(sc *cmdutil.StandardCommand, args []string) error {
			sessionID := args[0]
			artifact1 := args[1]
			artifact2 := args[2]

			// Get flags
			unified, _ := sc.GetCommand().Flags().GetBool("unified")

			// Get artifact contents
			content1, err := getArtifact(sessionID, artifact1)
			if err != nil {
				return sc.ReportError("get first artifact", err)
			}

			content2, err := getArtifact(sessionID, artifact2)
			if err != nil {
				return sc.ReportError("get second artifact", err)
			}

			// Create temp files for diff
			tmp1, err := os.CreateTemp("", "artifact1-*.txt")
			if err != nil {
				return err
			}
			defer os.Remove(tmp1.Name())

			tmp2, err := os.CreateTemp("", "artifact2-*.txt")
			if err != nil {
				return err
			}
			defer os.Remove(tmp2.Name())

			if _, err := tmp1.Write(content1); err != nil {
				return err
			}
			if _, err := tmp2.Write(content2); err != nil {
				return err
			}
			tmp1.Close()
			tmp2.Close()

			// Run diff
			diffArgs := []string{tmp1.Name(), tmp2.Name()}
			if unified {
				diffArgs = append([]string{"-u"}, diffArgs...)
			}

			cmd := exec.Command("diff", diffArgs...)
			cmd.Stdout = sc.GetCommand().OutOrStdout()
			cmd.Stderr = sc.GetCommand().ErrOrStderr()

			// diff returns exit code 1 when files differ, which is expected
			if err := cmd.Run(); err != nil {
				if exitErr, ok := err.(*exec.ExitError); ok && exitErr.ExitCode() == 1 {
					return nil // Files differ, but that's okay
				}
				return fmt.Errorf("diff command failed: %w", err)
			}

			return nil
		},
	}

	cmd := cmdutil.NewCommandFromTemplate(template)
	cmd.Flags().BoolP("unified", "u", true, "Use unified diff format")

	return cmd
}
