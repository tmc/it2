package cmdutil

import (
	"context"

	"github.com/spf13/cobra"
	"github.com/tmc/it2/internal/client"
	"github.com/tmc/it2/internal/validate"
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
	NoTimeout       bool // Override default timeout for long-running commands
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

		// Override timeout if NoTimeout is set
		if template.NoTimeout {
			sc.flags.Timeout = 0
		}

		// Validate format if supported
		if template.SupportsFormat {
			if err := validate.Format(sc.flags.Format); err != nil {
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
