package main

import (
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"github.com/tmc/it2/internal/cmd/alert"
	"github.com/tmc/it2/internal/cmd/annotation"
	"github.com/tmc/it2/internal/cmd/app"
	"github.com/tmc/it2/internal/cmd/arrangement"
	"github.com/tmc/it2/internal/cmd/auth"
	"github.com/tmc/it2/internal/cmd/badge"
	"github.com/tmc/it2/internal/cmd/broadcast"
	"github.com/tmc/it2/internal/cmd/color"
	"github.com/tmc/it2/internal/cmd/control"
	"github.com/tmc/it2/internal/cmd/event"
	"github.com/tmc/it2/internal/cmd/filepanel"
	"github.com/tmc/it2/internal/cmd/focus"
	"github.com/tmc/it2/internal/cmd/job"
	"github.com/tmc/it2/internal/cmd/keyboard"
	"github.com/tmc/it2/internal/cmd/lifecycle"
	"github.com/tmc/it2/internal/cmd/notification"
	"github.com/tmc/it2/internal/cmd/plugins"
	"github.com/tmc/it2/internal/cmd/preference"
	"github.com/tmc/it2/internal/cmd/profile"
	"github.com/tmc/it2/internal/cmd/prompt"
	"github.com/tmc/it2/internal/cmd/selection"
	"github.com/tmc/it2/internal/cmd/session"
	"github.com/tmc/it2/internal/cmd/shell"
	"github.com/tmc/it2/internal/cmd/snippet"
	"github.com/tmc/it2/internal/cmd/statusbar"
	"github.com/tmc/it2/internal/cmd/tab"
	"github.com/tmc/it2/internal/cmd/text"
	"github.com/tmc/it2/internal/cmd/tmux"
	"github.com/tmc/it2/internal/cmd/transaction"
	"github.com/tmc/it2/internal/cmd/trigger"
	"github.com/tmc/it2/internal/cmd/utility"
	"github.com/tmc/it2/internal/cmd/variable"
	"github.com/tmc/it2/internal/cmd/window"
	"github.com/tmc/it2/internal/completion"
	"github.com/tmc/it2/internal/config"
)

var (
	wsURL          string
	timeout        time.Duration
	format         string
	pluginPaths    []string
	pluginDeadline time.Duration
	configCmd      *cobra.Command
)

var rootCmd = &cobra.Command{
	Use:   "it2",
	Short: "Comprehensive command-line interface for iTerm2 automation",
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		// Set IT2_PLUGIN_PATHS environment variable from flag
		if len(pluginPaths) > 0 {
			os.Setenv("IT2_PLUGIN_PATHS", strings.Join(pluginPaths, string(os.PathListSeparator)))
		}
		// Set IT2_PLUGIN_DEADLINE environment variable from flag
		if pluginDeadline > 0 {
			os.Setenv("IT2_PLUGIN_DEADLINE", pluginDeadline.String())
		}
	},
	Long: `A powerful command-line tool for controlling iTerm2. Provides comprehensive access to terminal automation,
session management, and advanced iTerm2 features.

COMMAND GROUPS:

Core Operations
  app, session, tab, window - Control application and core UI elements

Content & Text
  badge, selection, text - Manage content, buffers, and text operations

Configuration
  auth, color, plugins, profile, statusbar, variable - Manage settings and appearance

Monitoring
  job, notification, prompt, shell - Monitor state and events

Advanced
  arrangement, broadcast, tmux - Advanced features and integrations

AUTHENTICATION:
  The tool automatically requests authentication from iTerm2 on first use.
  iTerm2 will prompt to allow API access. No manual setup required.

REQUIREMENTS:
  • iTerm2 version 3.3.0 or later
  • Python API enabled in iTerm2 preferences
  • macOS (iTerm2 is macOS-only)`,
	Example: `  # Core Operations

  it2 session list
  it2 session list --scope=tab
  it2 session split --horizontal
  it2 tab create "Default"
  it2 window list

  # Content & Text

  it2 text get-buffer sess_abc123
  it2 badge set "PROD"
  it2 selection copy

  # Configuration

  it2 profile list
  it2 variable get session sess_abc123 user.environment
  it2 color list

  # Monitoring

  it2 prompt list
  it2 prompt search "git commit"
  it2 notification monitor --type keystroke`,
}

// newCompletionCommand creates the completion command
func newCompletionCommand() *cobra.Command {
	completionCmd := &cobra.Command{
		Use:   "completion [bash|zsh|fish|powershell]",
		Short: "Generate completion scripts for various shells",
		Long: `Generate shell completion scripts for it2.

To load completions:

Bash:
  # Load completion into current session
  source <(it2 completion bash)

  # Install permanently (Linux)
  it2 completion bash > /etc/bash_completion.d/it2

  # Install permanently (macOS with Homebrew)
  it2 completion bash > $(brew --prefix)/etc/bash_completion.d/it2

Zsh:
  # Load completion into current session
  source <(it2 completion zsh)

  # Install permanently
  it2 completion zsh > "${fpath[1]}/_it2"

Fish:
  it2 completion fish | source

  # Install permanently
  it2 completion fish > ~/.config/fish/completions/it2.fish

PowerShell:
  it2 completion powershell | Out-String | Invoke-Expression

  # Install permanently, run:
  it2 completion powershell > it2.ps1
  # and source this file from your PowerShell profile.`,
		DisableFlagsInUseLine: true,
		ValidArgs:             []string{"bash", "zsh", "fish", "powershell"},
		Args:                  cobra.MatchAll(cobra.ExactArgs(1), cobra.OnlyValidArgs),
		Run: func(cmd *cobra.Command, args []string) {
			switch args[0] {
			case "bash":
				cmd.Root().GenBashCompletion(cmd.OutOrStdout())
			case "zsh":
				cmd.Root().GenZshCompletion(cmd.OutOrStdout())
			case "fish":
				cmd.Root().GenFishCompletion(cmd.OutOrStdout(), true)
			case "powershell":
				cmd.Root().GenPowerShellCompletionWithDesc(cmd.OutOrStdout())
			}
		},
	}
	return completionCmd
}

// newConfigCommand creates the config management command
func newConfigCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "config",
		Short: "Manage it2 configuration",
		Long:  "Commands for viewing, editing, and managing it2 configuration files",
	}

	// Add config subcommands
	cmd.AddCommand(&cobra.Command{
		Use:   "show",
		Short: "Show current configuration",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := config.Load()
			if err != nil {
				return err
			}

			// Display configuration in YAML format
			fmt.Printf("Current configuration:\n")
			fmt.Printf("URL: %s\n", cfg.URL)
			fmt.Printf("Timeout: %s\n", cfg.Timeout)
			fmt.Printf("Format: %s\n", cfg.Format)
			fmt.Printf("Color: %t\n", cfg.Color)
			fmt.Printf("Verbose: %t\n", cfg.Verbose)
			fmt.Printf("Debug: %t\n", cfg.Debug)
			fmt.Printf("Show Progress: %t\n", cfg.ShowProgress)
			fmt.Printf("Log Level: %s\n", cfg.LogLevel)
			fmt.Printf("Retry Attempts: %d\n", cfg.RetryAttempts)
			fmt.Printf("Retry Delay: %s\n", cfg.RetryDelay)

			return nil
		},
	})

	cmd.AddCommand(&cobra.Command{
		Use:   "path",
		Short: "Show configuration file path",
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Printf("Configuration file path: %s\n", config.GetConfigPath())
			return nil
		},
	})

	cmd.AddCommand(&cobra.Command{
		Use:   "init",
		Short: "Create default configuration file",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg := config.DefaultConfig()
			err := cfg.Save()
			if err != nil {
				return err
			}
			fmt.Printf("Created default configuration at: %s\n", config.GetConfigPath())
			return nil
		},
	})

	return cmd
}

func init() {
	// Load configuration from file/env and use as defaults
	cfg, err := config.Load()
	defaultURL := "ws://localhost:1912"
	defaultTimeout := 60 * time.Second
	defaultFormat := "table"

	if err == nil {
		defaultURL = cfg.URL
		defaultTimeout = cfg.Timeout
		defaultFormat = cfg.Format
	}

	rootCmd.PersistentFlags().StringVar(&wsURL, "url", defaultURL, "WebSocket URL to connect to iTerm2")
	rootCmd.PersistentFlags().DurationVar(&timeout, "timeout", defaultTimeout, "Connection timeout")
	rootCmd.PersistentFlags().StringVar(&format, "format", defaultFormat, "Output format (table, text, json, yaml)")
	rootCmd.PersistentFlags().StringSliceVar(&pluginPaths, "plugin-path", nil, "Additional plugin search paths (higher priority than embedded plugins)")
	rootCmd.PersistentFlags().DurationVar(&pluginDeadline, "plugin-deadline", 0, "Plugin execution deadline (hidden flag, default 5s)")

	// Hide advanced flags
	rootCmd.PersistentFlags().MarkHidden("plugin-deadline")
	rootCmd.PersistentFlags().MarkHidden("url")
	rootCmd.PersistentFlags().MarkHidden("plugin-path")

	// Add flag completion
	rootCmd.RegisterFlagCompletionFunc("format", completion.FormatCompletion)

	// Add command groups like GitHub CLI
	rootCmd.AddGroup(&cobra.Group{ID: "core", Title: "Core Operations"})
	rootCmd.AddGroup(&cobra.Group{ID: "content", Title: "Content & Text"})
	rootCmd.AddGroup(&cobra.Group{ID: "config", Title: "Configuration"})
	rootCmd.AddGroup(&cobra.Group{ID: "monitoring", Title: "Monitoring"})
	rootCmd.AddGroup(&cobra.Group{ID: "advanced", Title: "Advanced"})

	// Add shell completion commands
	rootCmd.AddCommand(newCompletionCommand())
	rootCmd.AddCommand(newConfigCommand())

	// Add organized command groups
	rootCmd.AddCommand(app.NewCommand())
	rootCmd.AddCommand(arrangement.NewCommand())
	rootCmd.AddCommand(auth.NewCommand())
	rootCmd.AddCommand(badge.NewCommand())
	rootCmd.AddCommand(broadcast.NewCommand())
	rootCmd.AddCommand(color.NewCommand())
	rootCmd.AddCommand(job.NewCommand())
	rootCmd.AddCommand(notification.NewCommand())
	rootCmd.AddCommand(plugins.NewCommand())
	rootCmd.AddCommand(profile.NewCommand())
	rootCmd.AddCommand(prompt.NewCommand())
	rootCmd.AddCommand(selection.NewCommand())
	rootCmd.AddCommand(session.NewCommand())
	rootCmd.AddCommand(shell.NewCommand())
	rootCmd.AddCommand(statusbar.NewCommand())
	rootCmd.AddCommand(tab.NewCommand())
	rootCmd.AddCommand(text.NewCommand())
	rootCmd.AddCommand(tmux.NewCommand())
	rootCmd.AddCommand(variable.NewCommand())
	rootCmd.AddCommand(window.NewCommand())

	// Add hidden/experimental commands
	alertCmd := alert.NewCommand()
	alertCmd.Hidden = true
	rootCmd.AddCommand(alertCmd)

	annotationCmd := annotation.NewCommand()
	annotationCmd.Hidden = true
	rootCmd.AddCommand(annotationCmd)

	controlCmd := control.NewCommand()
	controlCmd.Hidden = true
	rootCmd.AddCommand(controlCmd)

	eventCmd := event.NewCommand()
	eventCmd.Hidden = true
	rootCmd.AddCommand(eventCmd)

	filepanelCmd := filepanel.NewCommand()
	filepanelCmd.Hidden = true
	rootCmd.AddCommand(filepanelCmd)

	focusCmd := focus.NewCommand()
	focusCmd.Hidden = true
	rootCmd.AddCommand(focusCmd)

	keyboardCmd := keyboard.NewCommand()
	keyboardCmd.Hidden = true
	rootCmd.AddCommand(keyboardCmd)

	lifecycleCmd := lifecycle.NewCommand()
	lifecycleCmd.Hidden = true
	rootCmd.AddCommand(lifecycleCmd)

	preferenceCmd := preference.NewCommand()
	preferenceCmd.Hidden = true
	rootCmd.AddCommand(preferenceCmd)

	snippetCmd := snippet.NewCommand()
	snippetCmd.Hidden = true
	rootCmd.AddCommand(snippetCmd)

	transactionCmd := transaction.NewCommand()
	transactionCmd.Hidden = true
	rootCmd.AddCommand(transactionCmd)

	triggerCmd := trigger.NewCommand()
	triggerCmd.Hidden = true
	rootCmd.AddCommand(triggerCmd)

	utilityCmd := utility.NewCommand()
	utilityCmd.Hidden = true
	rootCmd.AddCommand(utilityCmd)
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		log.Fatal(err)
	}
}
