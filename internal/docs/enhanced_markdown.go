package docs

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

// FlagView represents a command flag for HTML rendering
type FlagView struct {
	Name      string
	Varname   string
	Shorthand string
	DefValue  string
	Usage     string
}

// Enhanced flags template with better styling
var flagsTemplate = `
<div class="command-flags">
<dl class="flags">{{range .}}
	<dt>
		{{if .Shorthand}}<code class="flag-short">-{{.Shorthand}}</code>, {{end}}
		<code class="flag-long">--{{.Name}}{{if .Varname}} <{{.Varname}}>{{end}}{{.DefValue}}</code>
	</dt>
	<dd class="flag-desc">{{.Usage}}</dd>
{{end}}</dl>
</div>
`

var flagTemplate = template.Must(template.New("flags").Parse(flagsTemplate))

// GenEnhancedMarkdown generates enhanced markdown with HTML elements for better rendering
func GenEnhancedMarkdown(cmd *cobra.Command, w io.Writer, linkHandler func(string) string) error {
	// Add frontmatter with title
	fmt.Fprint(w, "---\n")
	fmt.Fprintf(w, "title: %s\n", cmd.CommandPath())
	fmt.Fprint(w, "---\n\n")

	// Command title with enhanced formatting
	fmt.Fprintf(w, "# %s\n\n", cmd.CommandPath())

	// Description section
	if cmd.Short != "" {
		fmt.Fprintf(w, "> %s\n\n", cmd.Short)
	}

	// Usage section with syntax highlighting
	if cmd.Runnable() {
		fmt.Fprint(w, "## Synopsis\n\n")
		fmt.Fprintf(w, "```bash\n%s\n```\n\n", cmd.UseLine())
	}

	// Long description
	if cmd.Long != "" {
		fmt.Fprint(w, "## Description\n\n")
		fmt.Fprintf(w, "%s\n\n", cmd.Long)
	}

	// Subcommands with grouping (similar to GitHub CLI)
	if len(cmd.Commands()) > 0 {
		fmt.Fprint(w, "## Commands\n\n")

		// Group commands by category if available
		groups := groupCommands(cmd)
		if len(groups) > 1 {
			for _, group := range groups {
				if group.Title != "" {
					fmt.Fprintf(w, "### %s\n\n", group.Title)
				}
				for _, subcmd := range group.Commands {
					fmt.Fprintf(w, "* [%s](%s) - %s\n",
						subcmd.CommandPath(),
						linkHandler(cmdManualPath(subcmd)),
						subcmd.Short)
				}
				fmt.Fprint(w, "\n")
			}
		} else {
			// Simple list if no grouping
			for _, subcmd := range cmd.Commands() {
				if !subcmd.Hidden {
					fmt.Fprintf(w, "* [%s](%s) - %s\n",
						subcmd.CommandPath(),
						linkHandler(cmdManualPath(subcmd)),
						subcmd.Short)
				}
			}
			fmt.Fprint(w, "\n")
		}
	}

	// Options with enhanced HTML rendering
	if err := printEnhancedOptions(w, cmd); err != nil {
		return err
	}

	// Aliases
	if len(cmd.Aliases) > 0 {
		fmt.Fprint(w, "## Aliases\n\n")
		for i, alias := range cmd.Aliases {
			if i > 0 {
				fmt.Fprint(w, ", ")
			}
			fmt.Fprintf(w, "`%s`", alias)
		}
		fmt.Fprint(w, "\n\n")
	}

	// See also section
	if cmd.HasParent() {
		fmt.Fprint(w, "## See Also\n\n")
		p := cmd.Parent()
		fmt.Fprintf(w, "* [%s](%s) - %s\n", p.CommandPath(), linkHandler(cmdManualPath(p)), p.Short)
		fmt.Fprint(w, "\n")
	}

	// Examples section - moved to the end
	if len(cmd.Example) > 0 {
		fmt.Fprint(w, "## Examples\n\n")
		fmt.Fprintf(w, "```bash\n%s\n```\n\n", cmd.Example)
	}

	return nil
}

// Enhanced options printing with better HTML structure
func printEnhancedOptions(w io.Writer, cmd *cobra.Command) error {
	flags := cmd.NonInheritedFlags()
	if flags.HasAvailableFlags() {
		fmt.Fprint(w, "## Options\n\n")
		if err := printFlagsHTML(w, flags); err != nil {
			return err
		}
		fmt.Fprint(w, "\n")
	}

	parentFlags := cmd.InheritedFlags()
	if hasNonHelpFlags(parentFlags) {
		fmt.Fprint(w, "## Options inherited from parent commands\n\n")
		if err := printInheritedFlagsHTML(w, parentFlags); err != nil {
			return err
		}
		fmt.Fprint(w, "\n")
	}
	return nil
}

func printFlagsHTML(w io.Writer, fs *pflag.FlagSet) error {
	var flags []FlagView
	fs.VisitAll(func(f *pflag.Flag) {
		if f.Hidden || f.Name == "help" {
			return
		}
		varname, usage := pflag.UnquoteUsage(f)

		flags = append(flags, FlagView{
			Name:      f.Name,
			Varname:   varname,
			Shorthand: f.Shorthand,
			DefValue:  getDefaultValueDisplayString(f),
			Usage:     usage,
		})
	})
	return flagTemplate.Execute(w, flags)
}

// Enhanced inherited flags template with more compact styling
var inheritedFlagsTemplate = `
<div class="inherited-flags">
<dl class="flags inherited">{{range .}}
	<dt>
		{{if .Shorthand}}<code class="flag-short">-{{.Shorthand}}</code>, {{end}}
		<code class="flag-long">--{{.Name}}{{if .Varname}} <{{.Varname}}>{{end}}{{.DefValue}}</code>
	</dt>
	<dd class="flag-desc inherited">{{.Usage}}</dd>
{{end}}</dl>
</div>
`

var inheritedFlagTemplate = template.Must(template.New("inherited-flags").Parse(inheritedFlagsTemplate))

// printInheritedFlagsHTML prints inherited flags with enhanced descriptions
func printInheritedFlagsHTML(w io.Writer, fs *pflag.FlagSet) error {
	var flags []FlagView
	fs.VisitAll(func(f *pflag.Flag) {
		if f.Hidden || f.Name == "help" {
			return
		}
		varname, usage := pflag.UnquoteUsage(f)

		// Enhanced descriptions for common global flags
		enhancedUsage := enhanceGlobalFlagDescription(f.Name, usage)

		flags = append(flags, FlagView{
			Name:      f.Name,
			Varname:   varname,
			Shorthand: f.Shorthand,
			DefValue:  getDefaultValueDisplayString(f),
			Usage:     enhancedUsage,
		})
	})
	return inheritedFlagTemplate.Execute(w, flags)
}

// enhanceGlobalFlagDescription provides better descriptions for global flags
func enhanceGlobalFlagDescription(flagName, originalUsage string) string {
	switch flagName {
	case "format":
		return "Output format (text, json, yaml) - affects how command results are displayed"
	case "timeout":
		return "Timeout for API operations - how long to wait for iTerm2 to respond"
	case "url":
		return "WebSocket URL for iTerm2 API - typically ws://localhost:1912 for local iTerm2"
	default:
		return originalUsage
	}
}

func hasNonHelpFlags(fs *pflag.FlagSet) bool {
	var found bool
	fs.VisitAll(func(f *pflag.Flag) {
		if !f.Hidden && f.Name != "help" {
			found = true
		}
	})
	return found
}

var hiddenFlagDefaults = map[string]bool{
	"false": true,
	"":      true,
	"[]":    true,
	"0s":    true,
}

var defaultValFormats = map[string]string{
	"string":   " (default \"%s\")",
	"duration": " (default \"%s\")",
}

func getDefaultValueDisplayString(f *pflag.Flag) string {
	if hiddenFlagDefaults[f.DefValue] || hiddenFlagDefaults[f.Value.Type()] {
		return ""
	}

	if dvf, found := defaultValFormats[f.Value.Type()]; found {
		return fmt.Sprintf(dvf, f.Value)
	}
	return fmt.Sprintf(" (default %s)", f.Value)
}

// Command grouping structure
type CommandGroup struct {
	Title    string
	Commands []*cobra.Command
}

// groupCommands groups commands by category (basic implementation)
func groupCommands(cmd *cobra.Command) []CommandGroup {
	var groups []CommandGroup
	var ungrouped []*cobra.Command

	for _, subcmd := range cmd.Commands() {
		if subcmd.Hidden {
			continue
		}
		ungrouped = append(ungrouped, subcmd)
	}

	// For now, just return all as one group
	// Could be enhanced to use command annotations for grouping
	groups = append(groups, CommandGroup{
		Title:    "",
		Commands: ungrouped,
	})

	return groups
}

func cmdManualPath(c *cobra.Command) string {
	if basenameOverride, found := c.Annotations["markdown:basename"]; found {
		return basenameOverride + ".md"
	}
	return strings.ReplaceAll(c.CommandPath(), " ", "_") + ".md"
}

// GenEnhancedMarkdownTree generates the complete documentation tree
func GenEnhancedMarkdownTree(cmd *cobra.Command, dir string, filePrepender, linkHandler func(string) string) error {
	for _, c := range cmd.Commands() {
		if c.Hidden {
			continue
		}

		if err := GenEnhancedMarkdownTree(c, dir, filePrepender, linkHandler); err != nil {
			return err
		}
	}

	filename := filepath.Join(dir, cmdManualPath(cmd))
	f, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer f.Close()

	if prepend := filePrepender(filename); prepend != "" {
		if _, err := io.WriteString(f, prepend); err != nil {
			return err
		}
	}

	if err := GenEnhancedMarkdown(cmd, f, linkHandler); err != nil {
		return err
	}
	return nil
}
