package annotation

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

// Annotation represents an annotation stored in session properties
type Annotation struct {
	ID        string    `json:"id"`
	Text      string    `json:"text"`
	Line      int       `json:"line,omitempty"`
	Column    int       `json:"column,omitempty"`
	Timestamp time.Time `json:"timestamp"`
	Type      string    `json:"type,omitempty"`
}

// AnnotationStore represents the collection of annotations
type AnnotationStore struct {
	Annotations []*Annotation `json:"annotations"`
}

// NewCommand creates the annotation command with all subcommands.
func NewCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "annotation",
		Short: "Manage session annotations",
		Long:  "Commands for adding, removing, and listing annotations in sessions",
	}

	cmd.AddCommand(newAddCommand())
	cmd.AddCommand(newRemoveCommand())
	cmd.AddCommand(newListCommand())
	cmd.AddCommand(newClearCommand())

	return cmd
}

func newAddCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "add <session-id> <text>",
		Short: "Add an annotation to a session",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			sessionID := args[0]
			text := args[1]
			line, _ := cmd.Flags().GetInt("line")
			column, _ := cmd.Flags().GetInt("column")
			annotationType, _ := cmd.Flags().GetString("type")

			wsURL, timeout, _ := cmdutil.GetFlags(cmd)

			ctx, cancel := cmdutil.CreateContext(timeout)
			defer cancel()

			c, err := cmdutil.ConnectClient(ctx, wsURL)
			if err != nil {
				return fmt.Errorf("failed to connect: %w", err)
			}
			defer c.Close()

			// Get existing annotations
			store, err := getAnnotationStore(c, ctx, sessionID)
			if err != nil {
				return fmt.Errorf("failed to get existing annotations: %w", err)
			}

			// Create new annotation
			annotation := &Annotation{
				ID:        fmt.Sprintf("anno_%d", time.Now().UnixNano()),
				Text:      text,
				Line:      line,
				Column:    column,
				Timestamp: time.Now(),
				Type:      annotationType,
			}

			// Add to store
			store.Annotations = append(store.Annotations, annotation)

			// Save back to session
			err = setAnnotationStore(c, ctx, sessionID, store)
			if err != nil {
				return fmt.Errorf("failed to save annotation: %w", err)
			}

			fmt.Printf("Added annotation: %s (ID: %s)\n", text, annotation.ID)
			return nil
		},
	}

	cmd.Flags().Int("line", 0, "Line number for the annotation")
	cmd.Flags().Int("column", 0, "Column number for the annotation")
	cmd.Flags().String("type", "note", "Annotation type (note, warning, error, etc.)")

	return cmd
}

func newRemoveCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "remove <session-id> <annotation-id>",
		Short: "Remove an annotation from a session",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			sessionID := args[0]
			annotationID := args[1]

			wsURL, timeout, _ := cmdutil.GetFlags(cmd)

			ctx, cancel := cmdutil.CreateContext(timeout)
			defer cancel()

			c, err := cmdutil.ConnectClient(ctx, wsURL)
			if err != nil {
				return fmt.Errorf("failed to connect: %w", err)
			}
			defer c.Close()

			// Get existing annotations
			store, err := getAnnotationStore(c, ctx, sessionID)
			if err != nil {
				return fmt.Errorf("failed to get existing annotations: %w", err)
			}

			// Find and remove annotation
			var found bool
			for i, annotation := range store.Annotations {
				if annotation.ID == annotationID {
					store.Annotations = append(store.Annotations[:i], store.Annotations[i+1:]...)
					found = true
					break
				}
			}

			if !found {
				return fmt.Errorf("annotation with ID '%s' not found", annotationID)
			}

			// Save back to session
			err = setAnnotationStore(c, ctx, sessionID, store)
			if err != nil {
				return fmt.Errorf("failed to save annotations: %w", err)
			}

			fmt.Printf("Removed annotation: %s\n", annotationID)
			return nil
		},
	}

	return cmd
}

func newListCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list <session-id>",
		Short: "List all annotations in a session",
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

			// Get annotations
			store, err := getAnnotationStore(c, ctx, sessionID)
			if err != nil {
				return fmt.Errorf("failed to get annotations: %w", err)
			}

			// Convert to formatting annotations
			formattingAnnotations := make([]*formatting.Annotation, len(store.Annotations))
			for i, anno := range store.Annotations {
				formattingAnnotations[i] = &formatting.Annotation{
					ID:        anno.ID,
					Text:      anno.Text,
					Line:      anno.Line,
					Column:    anno.Column,
					Timestamp: anno.Timestamp,
					Type:      anno.Type,
				}
			}

			formatter := formatting.New(format)
			return formatter.FormatAnnotations(formattingAnnotations)
		},
	}

	return cmd
}

func newClearCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "clear <session-id>",
		Short: "Clear all annotations from a session",
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

			// Clear annotations
			store := &AnnotationStore{Annotations: []*Annotation{}}

			err = setAnnotationStore(c, ctx, sessionID, store)
			if err != nil {
				return fmt.Errorf("failed to clear annotations: %w", err)
			}

			fmt.Printf("Cleared all annotations from session %s\n", sessionID)
			return nil
		},
	}

	return cmd
}

// getAnnotationStore retrieves the annotation store from session properties
func getAnnotationStore(c *client.Client, ctx context.Context, sessionID string) (*AnnotationStore, error) {
	msg := &pb.ClientOriginatedMessage{
		Submessage: &pb.ClientOriginatedMessage_GetPropertyRequest{
			GetPropertyRequest: &pb.GetPropertyRequest{
				Identifier: &pb.GetPropertyRequest_SessionId{
					SessionId: sessionID,
				},
				Name: cmdutil.StringPtr("annotations"),
			},
		},
	}

	response, err := c.SendRequest(ctx, msg)
	if err != nil {
		return nil, err
	}

	resp := response.GetGetPropertyResponse()
	if resp == nil {
		return &AnnotationStore{Annotations: []*Annotation{}}, nil
	}

	if resp.GetStatus() != pb.GetPropertyResponse_OK {
		if resp.GetStatus() == pb.GetPropertyResponse_UNRECOGNIZED_NAME {
			// Property doesn't exist yet, return empty store
			return &AnnotationStore{Annotations: []*Annotation{}}, nil
		}
		return nil, fmt.Errorf("failed to get annotations property: %v", resp.GetStatus())
	}

	store := &AnnotationStore{Annotations: []*Annotation{}}
	if resp.JsonValue != nil && *resp.JsonValue != "" {
		err = json.Unmarshal([]byte(*resp.JsonValue), store)
		if err != nil {
			// If unmarshal fails, return empty store rather than error
			return &AnnotationStore{Annotations: []*Annotation{}}, nil
		}
	}

	return store, nil
}

// setAnnotationStore saves the annotation store to session properties
func setAnnotationStore(c *client.Client, ctx context.Context, sessionID string, store *AnnotationStore) error {
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
				Name:      cmdutil.StringPtr("annotations"),
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
		return fmt.Errorf("failed to set annotations property: %v", resp.GetStatus())
	}

	return nil
}
