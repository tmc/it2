package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/spf13/cobra"
	"github.com/spf13/cobra/doc"
	"github.com/spf13/pflag"
	"github.com/tmc/it2/internal/cmd/app"
	"github.com/tmc/it2/internal/cmd/arrangement"
	"github.com/tmc/it2/internal/cmd/auth"
	"github.com/tmc/it2/internal/cmd/badge"
	"github.com/tmc/it2/internal/cmd/broadcast"
	"github.com/tmc/it2/internal/cmd/color"
	"github.com/tmc/it2/internal/cmd/job"
	"github.com/tmc/it2/internal/cmd/notification"
	"github.com/tmc/it2/internal/cmd/profile"
	"github.com/tmc/it2/internal/cmd/prompt"
	"github.com/tmc/it2/internal/cmd/selection"
	"github.com/tmc/it2/internal/cmd/session"
	"github.com/tmc/it2/internal/cmd/statusbar"
	"github.com/tmc/it2/internal/cmd/tab"
	"github.com/tmc/it2/internal/cmd/text"
	"github.com/tmc/it2/internal/cmd/tmux"
	"github.com/tmc/it2/internal/cmd/variable"
	"github.com/tmc/it2/internal/cmd/window"
	"github.com/tmc/it2/internal/docs"
)

// CommandInfo represents a command in the hierarchy
type CommandInfo struct {
	Name        string        `json:"name"`
	Short       string        `json:"short"`
	Path        string        `json:"path"`
	Subcommands []CommandInfo `json:"subcommands,omitempty"`
}

// CommandGroup represents a group of commands
type CommandGroup struct {
	ID          string        `json:"id"`
	Title       string        `json:"title"`
	Description string        `json:"description,omitempty"`
	Commands    []CommandInfo `json:"commands"`
}

// CommandHierarchy represents the full command hierarchy
type CommandHierarchy struct {
	CLI           CommandInfo    `json:"cli"`
	Groups        []CommandGroup `json:"groups"`
	TotalCommands int            `json:"total_commands"`
	GeneratedAt   string         `json:"generated_at"`
}

func main() {
	if err := run(os.Args); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run(args []string) error {
	flags := pflag.NewFlagSet("", pflag.ContinueOnError)
	manPage := flags.BoolP("man-page", "", false, "Generate manual pages")
	website := flags.BoolP("website", "", false, "Generate website pages")
	hierarchy := flags.BoolP("hierarchy", "", false, "Generate command hierarchy file")
	hierarchyFormat := flags.StringP("hierarchy-format", "", "json", "Hierarchy file format (json, yaml)")
	dir := flags.StringP("doc-path", "", "", "Path directory where you want generate doc files")
	help := flags.BoolP("help", "h", false, "Help about any command")

	if err := flags.Parse(args); err != nil {
		return err
	}

	if *help {
		fmt.Fprintf(os.Stderr, "Usage of %s:\n\n%s", filepath.Base(args[0]), flags.FlagUsages())
		return nil
	}

	if *dir == "" {
		return fmt.Errorf("error: --doc-path not set")
	}

	// Create the root command with all subcommands
	rootCmd := createRootCommand()

	if err := os.MkdirAll(*dir, 0755); err != nil {
		return err
	}

	if *website {
		if err := docs.GenEnhancedMarkdownTree(rootCmd, *dir, filePrepender, linkHandler); err != nil {
			return err
		}
	}

	if *manPage {
		header := &doc.GenManHeader{
			Title:   "IT2",
			Section: "1",
		}
		if err := doc.GenManTree(rootCmd, header, *dir); err != nil {
			return err
		}
	}

	if *hierarchy {
		if err := generateHierarchyFile(rootCmd, *dir, *hierarchyFormat); err != nil {
			return err
		}
	}

	return nil
}

func createRootCommand() *cobra.Command {
	rootCmd := &cobra.Command{
		Use:   "it2",
		Short: "iTerm2 API Command-Line Interface",
		Long: `A powerful command-line tool for controlling iTerm2 programmatically through its WebSocket API.

it2 provides comprehensive control over iTerm2 sessions, tabs, windows, and content.
It supports both traditional flat command syntax and intuitive hierarchical navigation.`,
	}

	// Add command groups
	rootCmd.AddGroup(&cobra.Group{ID: "core", Title: "Core Operations"})
	rootCmd.AddGroup(&cobra.Group{ID: "content", Title: "Content & Text"})
	rootCmd.AddGroup(&cobra.Group{ID: "config", Title: "Configuration"})
	rootCmd.AddGroup(&cobra.Group{ID: "monitoring", Title: "Monitoring"})
	rootCmd.AddGroup(&cobra.Group{ID: "advanced", Title: "Advanced"})

	// Add persistent flags
	rootCmd.PersistentFlags().String("url", "", "WebSocket URL for iTerm2 API")
	rootCmd.PersistentFlags().Duration("timeout", 0, "Timeout for API operations")
	rootCmd.PersistentFlags().String("format", "text", "Output format (text, json, yaml)")

	// Core Operations - primary user-facing commands
	sessionCmd := session.NewCommand()
	sessionCmd.GroupID = "core"
	rootCmd.AddCommand(sessionCmd)

	tabCmd := tab.NewCommand()
	tabCmd.GroupID = "core"
	rootCmd.AddCommand(tabCmd)

	windowCmd := window.NewCommand()
	windowCmd.GroupID = "core"
	rootCmd.AddCommand(windowCmd)

	appCmd := app.NewCommand()
	appCmd.GroupID = "core"
	rootCmd.AddCommand(appCmd)

	// Content & Text - text operations and content management
	textCmd := text.NewCommand()
	textCmd.GroupID = "content"
	rootCmd.AddCommand(textCmd)

	selectionCmd := selection.NewCommand()
	selectionCmd.GroupID = "content"
	rootCmd.AddCommand(selectionCmd)

	badgeCmd := badge.NewCommand()
	badgeCmd.GroupID = "content"
	rootCmd.AddCommand(badgeCmd)

	// Configuration - settings and authentication
	authCmd := auth.NewCommand()
	authCmd.GroupID = "config"
	rootCmd.AddCommand(authCmd)

	profileCmd := profile.NewCommand()
	profileCmd.GroupID = "config"
	rootCmd.AddCommand(profileCmd)

	variableCmd := variable.NewCommand()
	variableCmd.GroupID = "config"
	rootCmd.AddCommand(variableCmd)

	colorCmd := color.NewCommand()
	colorCmd.GroupID = "config"
	rootCmd.AddCommand(colorCmd)

	statusbarCmd := statusbar.NewCommand()
	statusbarCmd.GroupID = "config"
	rootCmd.AddCommand(statusbarCmd)

	// Monitoring - observation and tracking
	jobCmd := job.NewCommand()
	jobCmd.GroupID = "monitoring"
	rootCmd.AddCommand(jobCmd)

	promptCmd := prompt.NewCommand()
	promptCmd.GroupID = "monitoring"
	rootCmd.AddCommand(promptCmd)

	notificationCmd := notification.NewCommand()
	notificationCmd.GroupID = "monitoring"
	rootCmd.AddCommand(notificationCmd)

	// Advanced - specialized functionality
	arrangementCmd := arrangement.NewCommand()
	arrangementCmd.GroupID = "advanced"
	rootCmd.AddCommand(arrangementCmd)

	broadcastCmd := broadcast.NewCommand()
	broadcastCmd.GroupID = "advanced"
	rootCmd.AddCommand(broadcastCmd)

	tmuxCmd := tmux.NewCommand()
	tmuxCmd.GroupID = "advanced"
	rootCmd.AddCommand(tmuxCmd)

	return rootCmd
}

func filePrepender(filename string) string {
	return ""
}

func linkHandler(name string) string {
	return fmt.Sprintf("./%s", name)
}

// generateHierarchyFile creates a JSON file with the command hierarchy
func generateHierarchyFile(rootCmd *cobra.Command, dir, format string) error {
	hierarchy := buildCommandHierarchy(rootCmd)

	var filename string
	var data []byte
	var err error

	switch format {
	case "json":
		filename = filepath.Join(dir, "command-hierarchy.json")
		data, err = json.MarshalIndent(hierarchy, "", "  ")
	default:
		return fmt.Errorf("unsupported hierarchy format: %s", format)
	}

	if err != nil {
		return fmt.Errorf("failed to marshal hierarchy: %w", err)
	}

	return os.WriteFile(filename, data, 0644)
}

// buildCommandHierarchy builds the command hierarchy structure
func buildCommandHierarchy(rootCmd *cobra.Command) CommandHierarchy {
	// Build CLI root command info
	cliInfo := buildCommandInfo(rootCmd)

	// Build command groups
	groups := buildCommandGroups(rootCmd)

	// Count total commands
	totalCommands := countTotalCommands(rootCmd)

	return CommandHierarchy{
		CLI:           cliInfo,
		Groups:        groups,
		TotalCommands: totalCommands,
		GeneratedAt:   time.Now().UTC().Format(time.RFC3339),
	}
}

// buildCommandInfo recursively builds command information
func buildCommandInfo(cmd *cobra.Command) CommandInfo {
	info := CommandInfo{
		Name:  cmd.Name(),
		Short: cmd.Short,
		Path:  cmd.CommandPath(),
	}

	// Add subcommands
	for _, subcmd := range cmd.Commands() {
		if !subcmd.Hidden {
			info.Subcommands = append(info.Subcommands, buildCommandInfo(subcmd))
		}
	}

	return info
}

// buildCommandGroups builds command groups with their commands
func buildCommandGroups(rootCmd *cobra.Command) []CommandGroup {
	var groups []CommandGroup

	// Get command groups from Cobra
	for _, group := range rootCmd.Groups() {
		cmdGroup := CommandGroup{
			ID:    group.ID,
			Title: group.Title,
		}

		// Find commands in this group
		for _, cmd := range rootCmd.Commands() {
			if cmd.GroupID == group.ID && !cmd.Hidden {
				cmdGroup.Commands = append(cmdGroup.Commands, buildCommandInfo(cmd))
			}
		}

		if len(cmdGroup.Commands) > 0 {
			groups = append(groups, cmdGroup)
		}
	}

	// Add ungrouped commands
	var ungroupedCommands []CommandInfo
	for _, cmd := range rootCmd.Commands() {
		if cmd.GroupID == "" && !cmd.Hidden {
			ungroupedCommands = append(ungroupedCommands, buildCommandInfo(cmd))
		}
	}

	if len(ungroupedCommands) > 0 {
		groups = append(groups, CommandGroup{
			ID:       "ungrouped",
			Title:    "Other Commands",
			Commands: ungroupedCommands,
		})
	}

	return groups
}

// countTotalCommands recursively counts all commands
func countTotalCommands(cmd *cobra.Command) int {
	count := 1 // Count the current command
	for _, subcmd := range cmd.Commands() {
		if !subcmd.Hidden {
			count += countTotalCommands(subcmd)
		}
	}
	return count
}
