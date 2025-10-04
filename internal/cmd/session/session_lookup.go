package session

import (
	"github.com/spf13/cobra"
)

// NewLookupCommand creates the session lookup command with subcommands.
func NewLookupCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "lookup",
		Short: "Look up session relationships and hierarchy",
		Long: `Commands for looking up session relationships and hierarchy information.

These commands help you navigate the session tree structure by finding
parent-child relationships, split trees, and container lookups.`,
		Example: `  # Get the parent session
  it2 session lookup parent

  # Get the root of a split tree
  it2 session lookup split-root sess_abc123

  # Get which tab contains a session
  it2 session lookup tab

  # Get which window contains a session
  it2 session lookup window`,
	}

	// Add hierarchy/lineage subcommands
	cmd.AddCommand(newLookupParentCommand())
	cmd.AddCommand(newLookupChildrenCommand())
	cmd.AddCommand(newLookupSiblingsCommand())
	cmd.AddCommand(newLookupAncestorsCommand())
	cmd.AddCommand(newLookupDescendantsCommand())
	cmd.AddCommand(newLookupLineageCommand())
	cmd.AddCommand(newLookupSplitRootCommand())

	// Add container lookup subcommands
	cmd.AddCommand(newLookupTabCommand())
	cmd.AddCommand(newLookupWindowCommand())

	// Add spatial lookup subcommands
	cmd.AddCommand(newLookupAboveCommand())

	return cmd
}
