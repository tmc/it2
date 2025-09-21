package statusbar

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/tmc/it2/internal/cmdutil"
	pb "github.com/tmc/it2/proto"
)

// NewCommand creates the status-bar command with all subcommands.
func NewCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "statusbar",
		Short: "Manage iTerm2 status bar components",
		Long:  "Commands for managing iTerm2 status bar components",
	}

	cmd.AddCommand(newOpenPopoverCommand())

	return cmd
}

func newOpenPopoverCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "open-popover <identifier> <session-id> <html>",
		Short: "Open a popover from a status bar component",
		Long:  "Open a popover from a status bar component with HTML content",
		Args:  cobra.ExactArgs(3),
		RunE: func(cmd *cobra.Command, args []string) error {
			identifier := args[0]
			sessionID := cmdutil.NormalizeSessionID(args[1])
			html := args[2]

			wsURL, timeout, _ := cmdutil.GetFlags(cmd)
			widthFloat, _ := cmd.Flags().GetFloat32("width")
			heightFloat, _ := cmd.Flags().GetFloat32("height")

			width := int32(widthFloat)
			height := int32(heightFloat)

			ctx, cancel := cmdutil.CreateContext(timeout)
			defer cancel()

			c, err := cmdutil.ConnectClient(ctx, wsURL)
			if err != nil {
				return fmt.Errorf("failed to connect: %w", err)
			}
			defer c.Close()

			size := &pb.Size{
				Width:  &width,
				Height: &height,
			}

			msg := &pb.ClientOriginatedMessage{
				Submessage: &pb.ClientOriginatedMessage_StatusBarComponentRequest{
					StatusBarComponentRequest: &pb.StatusBarComponentRequest{
						Request: &pb.StatusBarComponentRequest_OpenPopover_{
							OpenPopover: &pb.StatusBarComponentRequest_OpenPopover{
								SessionId: &sessionID,
								Html:      &html,
								Size:      size,
							},
						},
						Identifier: &identifier,
					},
				},
			}

			response, err := c.SendRequest(ctx, msg)
			if err != nil {
				return fmt.Errorf("failed to open popover: %w", err)
			}

			if resp := response.GetStatusBarComponentResponse(); resp != nil {
				switch resp.GetStatus() {
				case pb.StatusBarComponentResponse_OK:
					fmt.Printf("Popover opened successfully for component %s\n", identifier)
					return nil
				case pb.StatusBarComponentResponse_SESSION_NOT_FOUND:
					return fmt.Errorf("session not found: %s", sessionID)
				case pb.StatusBarComponentResponse_INVALID_IDENTIFIER:
					return fmt.Errorf("invalid component identifier: %s", identifier)
				default:
					return fmt.Errorf("failed to open popover: %v", resp.GetStatus())
				}
			}

			return fmt.Errorf("unexpected response type")
		},
	}

	cmd.Flags().Float32("width", 400, "Width of the popover in points")
	cmd.Flags().Float32("height", 300, "Height of the popover in points")

	return cmd
}
