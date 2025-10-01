package plugins

import (
	"fmt"
	"os"
	"text/tabwriter"
	"time"

	"github.com/spf13/cobra"
	"github.com/tmc/it2/internal/plugins"
)

func newListCommand() *cobra.Command {
	var showPaths bool
	var showMetrics bool

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

			// Get metrics if requested
			var allMetrics map[string]*plugins.PluginMetrics
			if showMetrics {
				allMetrics = plugins.GetMetricsStore().GetAllMetrics()
			}

			w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
			if showMetrics {
				if showPaths {
					fmt.Fprintln(w, "NAME\tSHA256\tSOURCE\tEXEC\tAVG\tP99\tPATH")
					fmt.Fprintln(w, "----\t------\t------\t----\t---\t---\t----")
				} else {
					fmt.Fprintln(w, "NAME\tSHA256\tSOURCE\tEXEC\tAVG\tP99")
					fmt.Fprintln(w, "----\t------\t------\t----\t---\t---")
				}
			} else if showPaths {
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

				if showMetrics {
					execCount := "0"
					avgStr := "-"
					p99Str := "-"

					if m := allMetrics[meta.Name]; m != nil {
						execCount = fmt.Sprintf("%d", m.ExecutionCount)
						avg, _, _, p99 := m.CalculateStats()
						if avg > 0 {
							avgStr = formatDuration(avg)
							p99Str = formatDuration(p99)
						}
					}

					if showPaths {
						fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\t%s\t%s\n",
							meta.Name,
							meta.SHA256,
							meta.Source,
							execCount,
							avgStr,
							p99Str,
							meta.Path,
						)
					} else {
						fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\t%s\n",
							meta.Name,
							meta.SHA256[:8],
							meta.Source,
							execCount,
							avgStr,
							p99Str,
						)
					}
				} else {
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
			}
			w.Flush()

			return nil
		},
	}

	cmd.Flags().BoolVarP(&showPaths, "paths", "p", false, "Show full paths to plugin executables")
	cmd.Flags().BoolVarP(&showMetrics, "metrics", "m", false, "Show execution metrics (count, avg, p99)")

	return cmd
}

// formatDuration formats a duration in a compact human-readable form
func formatDuration(d time.Duration) string {
	if d < time.Microsecond {
		return fmt.Sprintf("%dns", d.Nanoseconds())
	}
	if d < time.Millisecond {
		return fmt.Sprintf("%dµs", d.Microseconds())
	}
	if d < time.Second {
		return fmt.Sprintf("%dms", d.Milliseconds())
	}
	return fmt.Sprintf("%.2fs", d.Seconds())
}
