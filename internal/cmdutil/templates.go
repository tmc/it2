package cmdutil

import (
	"context"

	"github.com/spf13/cobra"
	"github.com/tmc/it2/internal/client"
)

// CommandTemplate provides a standard command structure
type CommandTemplate struct {
	Use             string
	Short           string
	Long            string
	Example         string
	Args            cobra.PositionalArgs
	RequiresClient  bool
	RequiresSession bool
	SupportsSorting bool
	SupportsColumns bool
	SupportsFormat  bool
	ValidArgsFunc   func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective)
	PreRunE         func(cmd *cobra.Command, args []string) error
	RunE            func(*StandardCommand, []string) error
}

// NewCommandFromTemplate creates a standardized cobra command from a template
func NewCommandFromTemplate(template CommandTemplate) *cobra.Command {
	// Create the standard command wrapper
	sc := NewStandardCommand(template.Use, template.Short)
	cmd := sc.GetCommand()

	// Set description fields
	if template.Long != "" {
		cmd.Long = template.Long
	}
	if template.Example != "" {
		cmd.Example = template.Example
	}

	// Set args validation
	if template.Args != nil {
		cmd.Args = template.Args
	}

	// Add appropriate flags based on template settings
	if template.SupportsSorting || template.SupportsColumns {
		sc.AddExtendedFlags()
	} else if template.RequiresClient || template.SupportsFormat {
		sc.AddStandardFlags()
	}

	// Set completion function if provided
	if template.ValidArgsFunc != nil {
		cmd.ValidArgsFunction = template.ValidArgsFunc
	}

	// Set pre-run function if provided
	if template.PreRunE != nil {
		cmd.PreRunE = template.PreRunE
	}

	// Set the main run function
	cmd.RunE = func(cmd *cobra.Command, args []string) error {
		// Parse flags into the StandardCommand
		if err := sc.ParseFlags(); err != nil {
			return err
		}

		// Validate format if supported
		if template.SupportsFormat {
			if err := ValidateFormat(sc.flags.Format); err != nil {
				return err
			}
		}

		// Execute with client if required
		if template.RequiresClient {
			return sc.ExecuteWithClient(func(c *client.Client, ctx context.Context) error {
				// Update StandardCommand's client reference
				sc.client = c
				sc.ctx = ctx

				// Handle session requirement
				if template.RequiresSession {
					// If no args and session required, try to resolve from environment
					if len(args) == 0 {
						sessionID, err := c.ResolveSessionID(ctx, "")
						if err != nil {
							return err
						}
						args = []string{sessionID}
					} else {
						// Resolve the session ID in the first argument with prefix matching
						resolvedID, err := c.ResolveSessionID(ctx, args[0])
						if err != nil {
							return err
						}
						args[0] = resolvedID
					}
				}

				return template.RunE(sc, args)
			})
		}

		// Execute without client
		return template.RunE(sc, args)
	}

	return cmd
}

// ListCommandTemplate creates a standard list command template
func ListCommandTemplate(resourceType string, listFunc func(*StandardCommand) error) CommandTemplate {
	return CommandTemplate{
		Use:             "list",
		Short:           "List " + resourceType + "s",
		Long:            "Display a list of all " + resourceType + "s with optional filtering and sorting",
		RequiresClient:  true,
		SupportsSorting: true,
		SupportsColumns: true,
		SupportsFormat:  true,
		RunE: func(sc *StandardCommand, args []string) error {
			return listFunc(sc)
		},
	}
}

// GetCommandTemplate creates a standard get command template
func GetCommandTemplate(resourceType string, getFunc func(*StandardCommand, string) error) CommandTemplate {
	return CommandTemplate{
		Use:            resourceType + " <id>",
		Short:          "Get " + resourceType + " details",
		Long:           "Display detailed information about a specific " + resourceType,
		Args:           cobra.ExactArgs(1),
		RequiresClient: true,
		SupportsFormat: true,
		PreRunE: func(cmd *cobra.Command, args []string) error {
			return ValidateResourceArgs(args, resourceType, false)
		},
		RunE: func(sc *StandardCommand, args []string) error {
			return getFunc(sc, args[0])
		},
	}
}

// CreateCommandTemplate creates a standard create command template
func CreateCommandTemplate(resourceType string, createFunc func(*StandardCommand, []string) error) CommandTemplate {
	return CommandTemplate{
		Use:            "create",
		Short:          "Create a new " + resourceType,
		Long:           "Create a new " + resourceType + " with the specified configuration",
		RequiresClient: true,
		SupportsFormat: true,
		RunE: func(sc *StandardCommand, args []string) error {
			return createFunc(sc, args)
		},
	}
}

// DeleteCommandTemplate creates a standard delete command template
func DeleteCommandTemplate(resourceType string, deleteFunc func(*StandardCommand, string) error) CommandTemplate {
	return CommandTemplate{
		Use:            "delete <id>",
		Short:          "Delete a " + resourceType,
		Long:           "Delete the specified " + resourceType,
		Args:           cobra.ExactArgs(1),
		RequiresClient: true,
		PreRunE: func(cmd *cobra.Command, args []string) error {
			return ValidateResourceArgs(args, resourceType, false)
		},
		RunE: func(sc *StandardCommand, args []string) error {
			return deleteFunc(sc, args[0])
		},
	}
}

// UpdateCommandTemplate creates a standard update command template
func UpdateCommandTemplate(resourceType string, updateFunc func(*StandardCommand, string, []string) error) CommandTemplate {
	return CommandTemplate{
		Use:            "update <id>",
		Short:          "Update a " + resourceType,
		Long:           "Update the specified " + resourceType + " with new values",
		Args:           cobra.MinimumNArgs(1),
		RequiresClient: true,
		SupportsFormat: true,
		PreRunE: func(cmd *cobra.Command, args []string) error {
			return ValidateResourceArgs([]string{args[0]}, resourceType, false)
		},
		RunE: func(sc *StandardCommand, args []string) error {
			return updateFunc(sc, args[0], args[1:])
		},
	}
}

// ActionCommandTemplate creates a standard action command template
func ActionCommandTemplate(action, resourceType string, actionFunc func(*StandardCommand, string) error) CommandTemplate {
	return CommandTemplate{
		Use:            action + " <id>",
		Short:          action + " a " + resourceType,
		Long:           "Perform " + action + " action on the specified " + resourceType,
		Args:           cobra.ExactArgs(1),
		RequiresClient: true,
		PreRunE: func(cmd *cobra.Command, args []string) error {
			return ValidateResourceArgs(args, resourceType, false)
		},
		RunE: func(sc *StandardCommand, args []string) error {
			if err := actionFunc(sc, args[0]); err != nil {
				return err
			}
			sc.ReportSuccess("Successfully performed %s on %s: %s", action, resourceType, args[0])
			return nil
		},
	}
}

// SessionCommandTemplate creates a template for session-specific commands
func SessionCommandTemplate(use, short string, sessionFunc func(*StandardCommand, string, []string) error) CommandTemplate {
	return CommandTemplate{
		Use:             use,
		Short:           short,
		RequiresClient:  true,
		RequiresSession: true,
		SupportsFormat:  true,
		RunE: func(sc *StandardCommand, args []string) error {
			if len(args) == 0 {
				return NewRequiredArgumentError("session_id")
			}
			return sessionFunc(sc, args[0], args[1:])
		},
	}
}
