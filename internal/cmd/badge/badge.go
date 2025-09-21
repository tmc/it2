package badge

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"
	"github.com/tmc/it2/internal/client"
	"github.com/tmc/it2/internal/cmdutil"
	pb "github.com/tmc/it2/proto"
)

// NewCommand creates the badge command with all subcommands.
func NewCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "badge",
		Short: "Manage session badges",
		Long:  "Commands for setting and clearing session badges",
	}

	cmd.AddCommand(newSetCommand())
	cmd.AddCommand(newClearCommand())
	cmd.AddCommand(newGetCommand())

	return cmd
}

func newSetCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "set <session-id> <text>",
		Short: "Set a badge for a session",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			sessionID := args[0]
			text := args[1]

			wsURL, timeout, _ := cmdutil.GetFlags(cmd)

			ctx, cancel := cmdutil.CreateContext(timeout)
			defer cancel()

			c, err := cmdutil.ConnectClient(ctx, wsURL)
			if err != nil {
				return fmt.Errorf("failed to connect: %w", err)
			}
			defer c.Close()

			err = setBadge(c, ctx, sessionID, text)
			if err != nil {
				return fmt.Errorf("failed to set badge: %w", err)
			}

			fmt.Printf("Set badge for session %s: %s\n", sessionID, text)
			return nil
		},
	}

	return cmd
}

func newClearCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "clear <session-id>",
		Short: "Clear the badge for a session",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			sessionID := args[0]

			wsURL, timeout, _ := cmdutil.GetFlags(cmd)

			ctx, cancel := cmdutil.CreateContext(timeout)
			defer cancel()

			c, err := cmdutil.ConnectClient(ctx, wsURL)
			if err != nil {
				return fmt.Errorf("failed to connect: %w", err)
			}
			defer c.Close()

			err = setBadge(c, ctx, sessionID, "")
			if err != nil {
				return fmt.Errorf("failed to clear badge: %w", err)
			}

			fmt.Printf("Cleared badge for session %s\n", sessionID)
			return nil
		},
	}

	return cmd
}

func newGetCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "get <session-id>",
		Short: "Get the current badge for a session",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			sessionID := args[0]

			wsURL, timeout, _ := cmdutil.GetFlags(cmd)

			ctx, cancel := cmdutil.CreateContext(timeout)
			defer cancel()

			c, err := cmdutil.ConnectClient(ctx, wsURL)
			if err != nil {
				return fmt.Errorf("failed to connect: %w", err)
			}
			defer c.Close()

			badge, err := getBadge(c, ctx, sessionID)
			if err != nil {
				return fmt.Errorf("failed to get badge: %w", err)
			}

			if badge == "" {
				fmt.Printf("No badge set for session %s\n", sessionID)
			} else {
				fmt.Printf("Badge for session %s: %s\n", sessionID, badge)
			}
			return nil
		},
	}

	return cmd
}

// setBadge sets the badge text for a session using the badge property
func setBadge(c *client.Client, ctx context.Context, sessionID, text string) error {
	msg := &pb.ClientOriginatedMessage{
		Submessage: &pb.ClientOriginatedMessage_SetPropertyRequest{
			SetPropertyRequest: &pb.SetPropertyRequest{
				Identifier: &pb.SetPropertyRequest_SessionId{
					SessionId: sessionID,
				},
				Name:      cmdutil.StringPtr("badge"),
				JsonValue: cmdutil.StringPtr(fmt.Sprintf(`"%s"`, text)),
			},
		},
	}

	response, err := c.SendRequest(ctx, msg)
	if err != nil {
		return err
	}

	resp := response.GetSetPropertyResponse()
	if resp == nil {
		return fmt.Errorf("no set property response received")
	}

	if resp.GetStatus() != pb.SetPropertyResponse_OK {
		return fmt.Errorf("failed to set badge property: %v", resp.GetStatus())
	}

	return nil
}

// getBadge gets the badge text for a session using the badge property
func getBadge(c *client.Client, ctx context.Context, sessionID string) (string, error) {
	msg := &pb.ClientOriginatedMessage{
		Submessage: &pb.ClientOriginatedMessage_GetPropertyRequest{
			GetPropertyRequest: &pb.GetPropertyRequest{
				Identifier: &pb.GetPropertyRequest_SessionId{
					SessionId: sessionID,
				},
				Name: cmdutil.StringPtr("badge"),
			},
		},
	}

	response, err := c.SendRequest(ctx, msg)
	if err != nil {
		return "", err
	}

	resp := response.GetGetPropertyResponse()
	if resp == nil {
		return "", fmt.Errorf("no get property response received")
	}

	if resp.GetStatus() != pb.GetPropertyResponse_OK {
		if resp.GetStatus() == pb.GetPropertyResponse_UNRECOGNIZED_NAME {
			// Property doesn't exist yet, return empty string
			return "", nil
		}
		return "", fmt.Errorf("failed to get badge property: %v", resp.GetStatus())
	}

	if resp.JsonValue == nil {
		return "", nil
	}

	// The value is stored as a JSON string, so we need to remove the quotes
	value := *resp.JsonValue
	if len(value) >= 2 && value[0] == '"' && value[len(value)-1] == '"' {
		value = value[1 : len(value)-1]
	}

	return value, nil
}
