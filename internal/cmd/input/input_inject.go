package input

import (
	"github.com/spf13/cobra"
	"github.com/tmc/it2/internal/cmdutil"
	"github.com/tmc/it2/internal/completion"
)

func newInjectCommand() *cobra.Command {
	template := cmdutil.CommandTemplate{
		Use:             "inject <session-id> <data>",
		Short:           "Inject data into a session as if it came from the program",
		Long:            "Inject data directly into a session's output stream, as if the running program had produced it. Supports hex-encoded data with --hex flag.",
		Args:            cobra.ExactArgs(2),
		RequiresClient:  true,
		RequiresSession: true,
		SupportsFormat:  true,
		ValidArgsFunc:   completion.SessionIDCompletion,
		RunE: func(sc *cmdutil.StandardCommand, args []string) error {
			sessionID := cmdutil.NormalizeSessionID(args[0])
			dataStr := args[1]

			isHex, _ := sc.GetCommand().Flags().GetBool("hex")

			var data []byte
			var err error

			if isHex {
				// Decode hex string to bytes
				data, err = cmdutil.ParseHexData(dataStr)
				if err != nil {
					return err
				}
			} else {
				// Use string as bytes
				data = []byte(dataStr)
			}

			// Inject the data
			if err := sc.GetClient().InjectData(sc.GetContext(), sessionID, data); err != nil {
				return sc.ReportError("inject data", err)
			}

			// Report success with JSON output support
			if sc.GetFlags().Format == "json" {
				result := map[string]interface{}{
					"session_id": sessionID,
					"data_size":  len(data),
					"hex":        isHex,
					"injected":   true,
				}
				return sc.FormatOutput(result)
			}

			sc.ReportSuccess("Data injected successfully (%d bytes) to session %s", len(data), sessionID)
			return nil
		},
	}

	cmd := cmdutil.NewCommandFromTemplate(template)
	cmd.Flags().Bool("hex", false, "Treat data as hex-encoded bytes")

	return cmd
}
