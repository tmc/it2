package session

import (
	"fmt"
	"regexp"

	"github.com/spf13/cobra"
	"github.com/tmc/it2/internal/cmdcore"
	"github.com/tmc/it2/internal/cmdutil"
	"github.com/tmc/it2/internal/completion"
)

func newListCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List all iTerm2 sessions",
		Long:  `List iTerm2 sessions with optional filtering and output customization.`,
		Example: cmdutil.Doc(`
			# List all sessions in table format (default: all sessions)
			$ it2 session list

			# List sessions in JSON format for scripting
			$ it2 session list --format json

			# Output only session IDs (quiet mode)
			$ it2 session list -q

			# Select specific JSON fields
			$ it2 session list --json id,name,cwd

			# Scope filtering - only sessions in current window
			$ it2 session list --scope window

			# Scope filtering - only sessions in current tab
			$ it2 session list --scope tab

			# Filter sessions with jq (requires --json)
			$ it2 session list --json id,name --jq '.[] | select(.name | test("vim"))'

			# Complex jq filter: wide vim sessions in /src
			$ it2 session list --json name,cwd,grid_size --jq '.[] | select(.name | test("vim") and .grid_size.width > 120 and .cwd | test("/src"))'

			# Format output with jq
			$ it2 session list --json id,name --jq '.[] | "\(.id): \(.name)"'

			# List sessions in a specific window (by window ID)
			$ it2 session list --window 1

			# List sessions with custom columns
			$ it2 session list --columns id,name,title

			# Sort sessions by name
			$ it2 session list --sort name

			# Use with xargs to send text to all sessions
			$ it2 session list -q | xargs -n1 -I {} it2 session send-text {} "echo hello"

			# Include buried sessions when needed
			$ it2 session list --include-buried

			# Count total sessions
			$ it2 session list -q | wc -l
		`),
		RunE: func(cmd *cobra.Command, args []string) error {
			_, timeout, format, columns, sortBy, sortReverse := cmdcore.GetExtendedFlags(cmd)
			ctx, cancel := cmdcore.CreateContext(timeout)
			defer cancel()

			c, err := cmdcore.ConnectClient(ctx)
			if err != nil {
				return fmt.Errorf("failed to connect: %w", err)
			}
			defer c.Close()

			// Get quiet flag
			quiet, _ := cmd.Flags().GetBool("quiet")

			// Include buried flag
			includeBuried, _ := cmd.Flags().GetBool("include-buried")

			// Print scope notice if IT2_SCOPE is set or --scope flag is used
			scopeFlag, _ := cmd.Flags().GetString("scope")
			cmdutil.PrintScopeNoticeWithFlag(format, scopeFlag)

			// Get filter flags (check both short and long forms)
			windowID, _ := cmd.Flags().GetString("window")
			if windowID == "" {
				windowID, _ = cmd.Flags().GetString("window-id")
			}
			tabID, _ := cmd.Flags().GetString("tab")
			if tabID == "" {
				tabID, _ = cmd.Flags().GetString("tab-id")
			}

			// Get --json and --jq flags
			_, _ = cmd.Flags().GetString("json") // Keep flag registered but value unused
			jqExpr, _ := cmd.Flags().GetString("jq")

			// Validate: --jq requires --json
			if jqExpr != "" && !cmd.Flags().Changed("json") {
				return fmt.Errorf("cannot use --jq without specifying --json")
			}

			// If --json is specified, override format to json
			if cmd.Flags().Changed("json") {
				format = "json"
			}

			// Get no-hyperlinks flag
			noHyperlinks, _ := cmd.Flags().GetBool("no-hyperlinks")

			// Plugin control flags
			pluginPatternStr, _ := cmd.Flags().GetString("plugins")
			noPlugins, _ := cmd.Flags().GetBool("no-plugins")
			pluginsFlagSet := cmd.Flags().Changed("plugins")
			noPluginsFlagSet := cmd.Flags().Changed("no-plugins")
			var pluginPattern *regexp.Regexp
			if pluginPatternStr != "" {
				var err error
				pluginPattern, err = regexp.Compile(pluginPatternStr)
				if err != nil {
					return fmt.Errorf("invalid --plugins pattern: %w", err)
				}
			}
			skipPlugins := false
			if format == "tree" && !pluginsFlagSet && !noPluginsFlagSet {
				skipPlugins = true
			}
			if pluginsFlagSet {
				skipPlugins = false
			}
			if noPlugins {
				skipPlugins = true
			}

			// Use shared operations
			sharedOps := cmdutil.NewSharedListOperations(c, ctx)
			return sharedOps.ListSessions(cmdutil.SharedListOptions{
				WindowID:      windowID,
				TabID:         tabID,
				ScopeFlag:     scopeFlag,
				Format:        format,
				Columns:       columns,
				SortBy:        sortBy,
				SortReverse:   sortReverse,
				Quiet:         quiet,
				NoHyperlinks:  noHyperlinks,
				PluginPattern: pluginPattern,
				SkipPlugins:   skipPlugins,
				IncludeBuried: includeBuried,
			})
		},
	}

	// Add flags for column selection and sorting
	cmd.Flags().String("columns", "", "Comma-separated list of columns to display (e.g., 'session id,window id,name')")
	cmd.Flags().String("sort", "", "Column to sort by (e.g., 'Session ID', 'Window ID', 'Name')")
	cmd.Flags().Bool("reverse", false, "Reverse sort order (descending)")

	// Add filtering flags
	cmd.Flags().String("window", "", "Filter sessions by window ID")
	cmd.Flags().String("tab", "", "Filter sessions by tab ID")
	// Keep old flag names for backward compatibility (hidden)
	cmd.Flags().String("window-id", "", "Filter sessions by window ID (deprecated, use --window)")
	cmd.Flags().String("tab-id", "", "Filter sessions by tab ID (deprecated, use --tab)")
	cmd.Flags().MarkHidden("window-id")
	cmd.Flags().MarkHidden("tab-id")

	// Add JSON and jq flags (GitHub CLI pattern)
	cmd.Flags().String("json", "", "Output JSON with specified fields (comma-separated, or empty for all fields)")
	cmd.Flags().String("jq", "", "Filter JSON output using a jq expression (requires --json)")

	// Plugin control flags
	cmd.Flags().String("plugins", "", "Regular expression of plugin names to run (case-sensitive)")
	cmd.Flags().Bool("no-plugins", false, "Disable plugin enrichment")

	// Add scope support for session filtering
	cmd.Flags().String("scope", "", "Filter sessions by scope (default: all sessions). Options: none, window, tab, parents, siblings, peers, lineage. Overrides IT2_SCOPE env var.")

	// Add quiet flag
	cmd.Flags().BoolP("quiet", "q", false, "Output only session IDs")

	// Add include-buried flag
	cmd.Flags().Bool("include-buried", false, "Include buried sessions")

	// Add no-hyperlinks flag (hidden)
	cmd.Flags().Bool("no-hyperlinks", false, "Disable OSC 8 terminal hyperlinks in output")
	cmd.Flags().MarkHidden("no-hyperlinks")

	// Add completion functions
	cmd.RegisterFlagCompletionFunc("window", completion.WindowIDCompletion)
	cmd.RegisterFlagCompletionFunc("tab", completion.TabIDCompletion)
	cmd.RegisterFlagCompletionFunc("window-id", completion.WindowIDCompletion)
	cmd.RegisterFlagCompletionFunc("tab-id", completion.TabIDCompletion)

	return cmd
}
