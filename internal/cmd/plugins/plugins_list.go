package plugins

import (
	"fmt"
	"text/tabwriter"
	"os"

	"github.com/spf13/cobra"
	"github.com/tmc/it2/internal/plugins"
)

func newListCommand() *cobra.Command {
	var showPaths bool

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List discovered plugins",
		Long: `List all discovered session enrichment plugins.

Plugins are discovered from PATH, custom paths (--plugin-path), and embedded sources.
The first matching plugin found (by priority) is used, others are shadowed.`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			metadata, err := plugins.DiscoverPluginMetadata()
			if err != nil {
				return fmt.Errorf("failed to discover plugins: %w", err)
			}

			if len(metadata) == 0 {
				fmt.Println("No plugins found")
				return nil
			}

			fmt.Printf("Discovered %d plugin(s):\n\n", len(metadata))

			w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
			if showPaths {
				fmt.Fprintln(w, "NAME\tSHA256\tSOURCE\tDUPLICATES\tPATH")
				fmt.Fprintln(w, "----\t------\t------\t----------\t----")
			} else {
				fmt.Fprintln(w, "NAME\tSHA256\tSOURCE\tDUPLICATES")
				fmt.Fprintln(w, "----\t------\t------\t----------")
			}

			for _, meta := range metadata {
				dupStr := ""
				if meta.Duplicates > 0 {
					dupStr = fmt.Sprintf("+%d", meta.Duplicates)
				}
				if showPaths {
					fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\n",
						meta.Name,
						meta.SHA256,
						meta.Source,
						dupStr,
						meta.Path,
					)
				} else {
					fmt.Fprintf(w, "%s\t%s\t%s\t%s\n",
						meta.Name,
						meta.SHA256[:8],
						meta.Source,
						dupStr,
					)
				}
			}
			w.Flush()

			return nil
		},
	}

	cmd.Flags().BoolVarP(&showPaths, "paths", "p", false, "Show full paths to plugin executables")

	return cmd
}
