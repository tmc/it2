package trigger

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/spf13/cobra"
	"github.com/tmc/it2/internal/client"
	"github.com/tmc/it2/internal/cmdutil"
	"github.com/tmc/it2/internal/formatting"
	pb "github.com/tmc/it2/proto"
)

// Trigger represents a text trigger
type Trigger struct {
	ID      string    `json:"id"`
	Pattern string    `json:"pattern"`
	Action  string    `json:"action"`
	Enabled bool      `json:"enabled"`
	Created time.Time `json:"created"`
}

// TriggerStore represents the collection of triggers
type TriggerStore struct {
	Triggers []*Trigger `json:"triggers"`
}

// NewCommand creates the trigger command with all subcommands.
func NewCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "trigger",
		Short: "Manage text triggers",
		Long:  "Commands for adding, removing, and listing text triggers that execute actions when patterns are matched",
	}

	cmd.AddCommand(newAddCommand())
	cmd.AddCommand(newRemoveCommand())
	cmd.AddCommand(newListCommand())
	cmd.AddCommand(newEnableCommand())
	cmd.AddCommand(newDisableCommand())
	cmd.AddCommand(newClearCommand())

	return cmd
}

func newAddCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "add <session-id> <pattern> <action>",
		Short: "Add a new trigger",
		Long:  "Add a new trigger that executes an action when the pattern is matched in the session output",
		Args:  cobra.ExactArgs(3),
		RunE: func(cmd *cobra.Command, args []string) error {
			sessionID := args[0]
			pattern := args[1]
			action := args[2]

			wsURL, timeout, _ := cmdutil.GetFlags(cmd)

			ctx, cancel := cmdutil.CreateContext(timeout)
			defer cancel()

			c, err := cmdutil.ConnectClient(ctx, wsURL)
			if err != nil {
				return fmt.Errorf("failed to connect: %w", err)
			}
			defer c.Close()

			// Get existing triggers
			store, err := getTriggerStore(c, ctx, sessionID)
			if err != nil {
				return fmt.Errorf("failed to get existing triggers: %w", err)
			}

			// Create new trigger
			trigger := &Trigger{
				ID:      fmt.Sprintf("trigger_%d", time.Now().UnixNano()),
				Pattern: pattern,
				Action:  action,
				Enabled: true,
				Created: time.Now(),
			}

			// Add to store
			store.Triggers = append(store.Triggers, trigger)

			// Save back to session
			err = setTriggerStore(c, ctx, sessionID, store)
			if err != nil {
				return fmt.Errorf("failed to save trigger: %w", err)
			}

			fmt.Printf("Added trigger: %s -> %s (ID: %s)\n", pattern, action, trigger.ID)
			return nil
		},
	}

	return cmd
}

func newRemoveCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "remove <session-id> <trigger-id>",
		Short: "Remove a trigger",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			sessionID := args[0]
			triggerID := args[1]

			wsURL, timeout, _ := cmdutil.GetFlags(cmd)

			ctx, cancel := cmdutil.CreateContext(timeout)
			defer cancel()

			c, err := cmdutil.ConnectClient(ctx, wsURL)
			if err != nil {
				return fmt.Errorf("failed to connect: %w", err)
			}
			defer c.Close()

			// Get existing triggers
			store, err := getTriggerStore(c, ctx, sessionID)
			if err != nil {
				return fmt.Errorf("failed to get existing triggers: %w", err)
			}

			// Find and remove trigger
			var found bool
			for i, trigger := range store.Triggers {
				if trigger.ID == triggerID {
					store.Triggers = append(store.Triggers[:i], store.Triggers[i+1:]...)
					found = true
					break
				}
			}

			if !found {
				return fmt.Errorf("trigger with ID '%s' not found", triggerID)
			}

			// Save back to session
			err = setTriggerStore(c, ctx, sessionID, store)
			if err != nil {
				return fmt.Errorf("failed to save triggers: %w", err)
			}

			fmt.Printf("Removed trigger: %s\n", triggerID)
			return nil
		},
	}

	return cmd
}

func newListCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list <session-id>",
		Short: "List all triggers in a session",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			sessionID := args[0]

			wsURL, timeout, format := cmdutil.GetFlags(cmd)

			ctx, cancel := cmdutil.CreateContext(timeout)
			defer cancel()

			c, err := cmdutil.ConnectClient(ctx, wsURL)
			if err != nil {
				return fmt.Errorf("failed to connect: %w", err)
			}
			defer c.Close()

			// Get triggers
			store, err := getTriggerStore(c, ctx, sessionID)
			if err != nil {
				return fmt.Errorf("failed to get triggers: %w", err)
			}

			// Convert to formatting triggers
			formattingTriggers := make([]*formatting.Trigger, len(store.Triggers))
			for i, trigger := range store.Triggers {
				formattingTriggers[i] = &formatting.Trigger{
					ID:      trigger.ID,
					Pattern: trigger.Pattern,
					Action:  trigger.Action,
					Enabled: trigger.Enabled,
					Created: trigger.Created,
				}
			}

			formatter := formatting.New(format)
			return formatter.FormatTriggers(formattingTriggers)
		},
	}

	return cmd
}

func newEnableCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "enable <session-id> <trigger-id>",
		Short: "Enable a trigger",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			return setTriggerEnabled(args[0], args[1], true, cmd)
		},
	}

	return cmd
}

func newDisableCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "disable <session-id> <trigger-id>",
		Short: "Disable a trigger",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			return setTriggerEnabled(args[0], args[1], false, cmd)
		},
	}

	return cmd
}

func newClearCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "clear <session-id>",
		Short: "Clear all triggers from a session",
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

			// Clear triggers
			store := &TriggerStore{Triggers: []*Trigger{}}

			err = setTriggerStore(c, ctx, sessionID, store)
			if err != nil {
				return fmt.Errorf("failed to clear triggers: %w", err)
			}

			fmt.Printf("Cleared all triggers from session %s\n", sessionID)
			return nil
		},
	}

	return cmd
}

// setTriggerEnabled enables or disables a trigger
func setTriggerEnabled(sessionID, triggerID string, enabled bool, cmd *cobra.Command) error {
	wsURL, timeout, _ := cmdutil.GetFlags(cmd)

	ctx, cancel := cmdutil.CreateContext(timeout)
	defer cancel()

	c, err := cmdutil.ConnectClient(ctx, wsURL)
	if err != nil {
		return fmt.Errorf("failed to connect: %w", err)
	}
	defer c.Close()

	// Get existing triggers
	store, err := getTriggerStore(c, ctx, sessionID)
	if err != nil {
		return fmt.Errorf("failed to get existing triggers: %w", err)
	}

	// Find and update trigger
	var found bool
	for _, trigger := range store.Triggers {
		if trigger.ID == triggerID {
			trigger.Enabled = enabled
			found = true
			break
		}
	}

	if !found {
		return fmt.Errorf("trigger with ID '%s' not found", triggerID)
	}

	// Save back to session
	err = setTriggerStore(c, ctx, sessionID, store)
	if err != nil {
		return fmt.Errorf("failed to save triggers: %w", err)
	}

	status := "disabled"
	if enabled {
		status = "enabled"
	}
	fmt.Printf("Trigger %s %s\n", triggerID, status)
	return nil
}

// getTriggerStore retrieves the trigger store from session properties
func getTriggerStore(c *client.Client, ctx context.Context, sessionID string) (*TriggerStore, error) {
	msg := &pb.ClientOriginatedMessage{
		Submessage: &pb.ClientOriginatedMessage_GetPropertyRequest{
			GetPropertyRequest: &pb.GetPropertyRequest{
				Identifier: &pb.GetPropertyRequest_SessionId{
					SessionId: sessionID,
				},
				Name: cmdutil.StringPtr("triggers"),
			},
		},
	}

	response, err := c.SendRequest(ctx, msg)
	if err != nil {
		return nil, err
	}

	resp := response.GetGetPropertyResponse()
	if resp == nil {
		return &TriggerStore{Triggers: []*Trigger{}}, nil
	}

	if resp.GetStatus() != pb.GetPropertyResponse_OK {
		if resp.GetStatus() == pb.GetPropertyResponse_UNRECOGNIZED_NAME {
			// Property doesn't exist yet, return empty store
			return &TriggerStore{Triggers: []*Trigger{}}, nil
		}
		return nil, fmt.Errorf("failed to get triggers property: %v", resp.GetStatus())
	}

	store := &TriggerStore{Triggers: []*Trigger{}}
	if resp.JsonValue != nil && *resp.JsonValue != "" {
		err = json.Unmarshal([]byte(*resp.JsonValue), store)
		if err != nil {
			// If unmarshal fails, return empty store rather than error
			return &TriggerStore{Triggers: []*Trigger{}}, nil
		}
	}

	return store, nil
}

// setTriggerStore saves the trigger store to session properties
func setTriggerStore(c *client.Client, ctx context.Context, sessionID string, store *TriggerStore) error {
	data, err := json.Marshal(store)
	if err != nil {
		return err
	}

	msg := &pb.ClientOriginatedMessage{
		Submessage: &pb.ClientOriginatedMessage_SetPropertyRequest{
			SetPropertyRequest: &pb.SetPropertyRequest{
				Identifier: &pb.SetPropertyRequest_SessionId{
					SessionId: sessionID,
				},
				Name:      cmdutil.StringPtr("triggers"),
				JsonValue: cmdutil.StringPtr(string(data)),
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
		return fmt.Errorf("failed to set triggers property: %v", resp.GetStatus())
	}

	return nil
}
