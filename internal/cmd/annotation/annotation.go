package annotation

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/spf13/cobra"
	"github.com/tmc/it2/internal/client"
	"github.com/tmc/it2/internal/cmderr"
	"github.com/tmc/it2/internal/cmdutil"
	"github.com/tmc/it2/internal/completion"
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
		Long: `Commands for adding, removing, and listing annotations in sessions.

Annotations are metadata attached to terminal sessions that can be used to:
- Mark important locations in command output
- Add notes and reminders about session state
- Create navigable markers for debugging sessions`,
	}

	cmd.AddCommand(newAddCommand())
	cmd.AddCommand(newRemoveCommand())
	cmd.AddCommand(newListCommand())
	cmd.AddCommand(newClearCommand())

	return cmd
}

func newAddCommand() *cobra.Command {
	template := cmdutil.CommandTemplate{
		Use:   "add <session-id> <text>",
		Short: "Add an annotation to a session",
		Long: `Add an annotation to a session with optional line and column position.

Examples:
  it2 annotation add session123 "Error occurred here"
  it2 annotation add session123 "Stack trace" --line 42
  it2 annotation add session123 "Breakpoint" --line 10 --column 5 --type error`,
		Args:            cobra.ExactArgs(2),
		RequiresClient:  true,
		RequiresSession: true,
		SupportsFormat:  true,
		ValidArgsFunc:   completion.SessionIDCompletion,
		RunE: func(sc *cmdutil.StandardCommand, args []string) error {
			// Session ID is already normalized by template when RequiresSession: true
			sessionID := args[0]
			text := args[1]
			line, _ := sc.GetCommand().Flags().GetInt("line")
			column, _ := sc.GetCommand().Flags().GetInt("column")
			annotationType, _ := sc.GetCommand().Flags().GetString("type")

			// Get existing annotations
			store, err := getAnnotationStore(sc.GetClient(), sc.GetContext(), sessionID)
			if err != nil {
				return sc.ReportError("get existing annotations", err)
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
			err = setAnnotationStore(sc.GetClient(), sc.GetContext(), sessionID, store)
			if err != nil {
				return sc.ReportError("save annotation", err)
			}

			// Report success with JSON output support
			if sc.GetFlags().Format == "json" {
				result := map[string]interface{}{
					"session_id": sessionID,
					"annotation": annotation,
					"action":     "added",
				}
				return sc.FormatOutput(result)
			}

			sc.ReportSuccess("Added annotation: %s (ID: %s)", text, annotation.ID)
			return nil
		},
	}

	cmd := cmdutil.NewCommandFromTemplate(template)
	cmd.Flags().Int("line", 0, "Line number for the annotation")
	cmd.Flags().Int("column", 0, "Column number for the annotation")
	cmd.Flags().String("type", "note", "Annotation type (note, warning, error, etc.)")

	return cmd
}

func newRemoveCommand() *cobra.Command {
	template := cmdutil.CommandTemplate{
		Use:   "remove <session-id> <annotation-id>",
		Short: "Remove an annotation from a session",
		Long: `Remove a specific annotation from a session by its ID.

Examples:
  it2 annotation remove session123 anno_1234567890
  it2 annotation list session123 --format json | jq -r '.[-1].id' | xargs it2 annotation remove session123`,
		Args:            cobra.ExactArgs(2),
		RequiresClient:  true,
		RequiresSession: true,
		SupportsFormat:  true,
		ValidArgsFunc:   completion.SessionIDCompletion,
		RunE: func(sc *cmdutil.StandardCommand, args []string) error {
			// Session ID is already normalized by template when RequiresSession: true
			sessionID := args[0]
			annotationID := args[1]

			// Get existing annotations
			store, err := getAnnotationStore(sc.GetClient(), sc.GetContext(), sessionID)
			if err != nil {
				return sc.ReportError("get existing annotations", err)
			}

			// Find and remove annotation
			var found bool
			var removedAnnotation *Annotation
			for i, annotation := range store.Annotations {
				if annotation.ID == annotationID {
					removedAnnotation = annotation
					store.Annotations = append(store.Annotations[:i], store.Annotations[i+1:]...)
					found = true
					break
				}
			}

			if !found {
				return cmderr.NewNotFoundError("annotation", annotationID)
			}

			// Save back to session
			err = setAnnotationStore(sc.GetClient(), sc.GetContext(), sessionID, store)
			if err != nil {
				return sc.ReportError("save annotations", err)
			}

			// Report success with JSON output support
			if sc.GetFlags().Format == "json" {
				result := map[string]interface{}{
					"session_id": sessionID,
					"annotation": removedAnnotation,
					"action":     "removed",
				}
				return sc.FormatOutput(result)
			}

			sc.ReportSuccess("Removed annotation: %s", annotationID)
			return nil
		},
	}

	return cmdutil.NewCommandFromTemplate(template)
}

func newListCommand() *cobra.Command {
	template := cmdutil.CommandTemplate{
		Use:   "list [<session-id>]",
		Short: "List all annotations in a session",
		Long: `List all annotations in a session. If no session-id is provided,
uses $ITERM_SESSION_ID environment variable.

Examples:
  it2 annotation list                  # List for current session
  it2 annotation list session123       # List for specific session
  it2 annotation list --format json    # Output as JSON for processing`,
		Args:           cobra.RangeArgs(0, 1),
		RequiresClient: true,
		SupportsFormat: true,
		ValidArgsFunc:  completion.SessionIDCompletion,
		RunE: func(sc *cmdutil.StandardCommand, args []string) error {
			var sessionID string
			if len(args) > 0 {
				sessionID = args[0]
			}
			// Resolve session ID with environment fallback and prefix matching
			ctx := sc.GetContext()
			sessionID, err := sc.GetClient().ResolveSessionID(ctx, sessionID)
			if err != nil {
				return sc.ReportError("resolve session ID", err)
			}

			// Get annotations
			store, err := getAnnotationStore(sc.GetClient(), sc.GetContext(), sessionID)
			if err != nil {
				return sc.ReportError("get annotations", err)
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

			formatter := formatting.New(sc.GetFlags().Format)
			return formatter.FormatAnnotations(formattingAnnotations)
		},
	}

	return cmdutil.NewCommandFromTemplate(template)
}

func newClearCommand() *cobra.Command {
	template := cmdutil.CommandTemplate{
		Use:   "clear <session-id>",
		Short: "Clear all annotations from a session",
		Long: `Clear all annotations from a session, removing all stored metadata.

Examples:
  it2 annotation clear session123
  it2 session list --format json | jq -r '.[].id' | xargs -I{} it2 annotation clear {}`,
		Args:            cobra.ExactArgs(1),
		RequiresClient:  true,
		RequiresSession: true,
		SupportsFormat:  true,
		ValidArgsFunc:   completion.SessionIDCompletion,
		RunE: func(sc *cmdutil.StandardCommand, args []string) error {
			// Session ID is already normalized by template when RequiresSession: true
			sessionID := args[0]

			// Get current count before clearing
			store, _ := getAnnotationStore(sc.GetClient(), sc.GetContext(), sessionID)
			previousCount := len(store.Annotations)

			// Clear annotations
			store = &AnnotationStore{Annotations: []*Annotation{}}

			err := setAnnotationStore(sc.GetClient(), sc.GetContext(), sessionID, store)
			if err != nil {
				return sc.ReportError("clear annotations", err)
			}

			// Report success with JSON output support
			if sc.GetFlags().Format == "json" {
				result := map[string]interface{}{
					"session_id":    sessionID,
					"cleared_count": previousCount,
					"action":        "cleared",
				}
				return sc.FormatOutput(result)
			}

			if previousCount > 0 {
				sc.ReportSuccess("Cleared %d annotations from session %s", previousCount, sessionID)
			} else {
				sc.ReportSuccess("No annotations to clear in session %s", sessionID)
			}
			return nil
		},
	}

	return cmdutil.NewCommandFromTemplate(template)
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
