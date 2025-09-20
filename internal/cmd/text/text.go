package text

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"
	"github.com/tmc/it2/internal/client"
	"github.com/tmc/it2/internal/formatting"
)

// NewCommand creates the text command with all subcommands.
func NewCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "text",
		Short: "Text operations in iTerm2",
		Long:  "Commands for sending text and retrieving buffer contents",
	}

	cmd.AddCommand(newSendCommand())
	cmd.AddCommand(newBufferCommand())
	cmd.AddCommand(newSelectCommand())

	return cmd
}

func newSendCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "send <session-id> <text>",
		Short: "Send text to a session",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			sessionID := args[0]
			text := args[1]

			wsURL, _ := cmd.Flags().GetString("url")
			timeout, _ := cmd.Flags().GetDuration("timeout")

			// Use parent command flags if not set
			if wsURL == "" {
				wsURL = cmd.Parent().PersistentFlags().Lookup("url").Value.String()
			}
			if timeout == 0 {
				timeout, _ = cmd.Parent().PersistentFlags().GetDuration("timeout")
			}

			ctx, cancel := context.WithTimeout(context.Background(), timeout)
			defer cancel()

			c := client.New(wsURL)
			if err := c.Connect(ctx); err != nil {
				return fmt.Errorf("failed to connect: %w", err)
			}
			defer c.Close()

			if err := c.SendText(ctx, sessionID, text); err != nil {
				return fmt.Errorf("failed to send text: %w", err)
			}

			fmt.Println("Text sent successfully")
			return nil
		},
	}
}

func newBufferCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "buffer <session-id>",
		Short: "Get buffer contents of a session",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			sessionID := args[0]

			wsURL, _ := cmd.Flags().GetString("url")
			timeout, _ := cmd.Flags().GetDuration("timeout")
			format, _ := cmd.Flags().GetString("format")
			lines, _ := cmd.Flags().GetInt32("lines")

			// Use parent command flags if not set
			if wsURL == "" {
				wsURL = cmd.Parent().PersistentFlags().Lookup("url").Value.String()
			}
			if timeout == 0 {
				timeout, _ = cmd.Parent().PersistentFlags().GetDuration("timeout")
			}
			if format == "" {
				format = cmd.Parent().PersistentFlags().Lookup("format").Value.String()
			}

			ctx, cancel := context.WithTimeout(context.Background(), timeout)
			defer cancel()

			c := client.New(wsURL)
			if err := c.Connect(ctx); err != nil {
				return fmt.Errorf("failed to connect: %w", err)
			}
			defer c.Close()

			resp, err := c.GetBuffer(ctx, sessionID, lines)
			if err != nil {
				return fmt.Errorf("failed to get buffer: %w", err)
			}

			formatter := formatting.New(format)
			return formatter.FormatBuffer(resp)
		},
	}

	cmd.Flags().Int32("lines", 100, "Number of lines to retrieve")
	cmd.Flags().Bool("scrollback", false, "Include scrollback history")

	return cmd
}

func newSelectCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "select <session-id>",
		Short: "Control text selection in a session",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			// TODO: Implement text selection
			return nil
		},
	}

	cmd.Flags().String("start", "", "Start position (row,col)")
	cmd.Flags().String("end", "", "End position (row,col)")
	cmd.Flags().Bool("clear", false, "Clear current selection")

	return cmd
}