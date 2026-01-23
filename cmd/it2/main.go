package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"github.com/tmc/it2/internal/cmd/alert"
	"github.com/tmc/it2/internal/cmd/annotation"
	"github.com/tmc/it2/internal/cmd/app"
	"github.com/tmc/it2/internal/cmd/arrangement"
	"github.com/tmc/it2/internal/cmd/artifact"
	"github.com/tmc/it2/internal/cmd/attention"
	"github.com/tmc/it2/internal/cmd/auth"
	"github.com/tmc/it2/internal/cmd/badge"
	"github.com/tmc/it2/internal/cmd/broadcast"
	"github.com/tmc/it2/internal/cmd/color"
	"github.com/tmc/it2/internal/cmd/control"
	"github.com/tmc/it2/internal/cmd/event"
	"github.com/tmc/it2/internal/cmd/filepanel"
	"github.com/tmc/it2/internal/cmd/focus"
	"github.com/tmc/it2/internal/cmd/shortcuts"
	"github.com/tmc/it2/internal/cmd/job"
	"github.com/tmc/it2/internal/cmd/keyboard"
	"github.com/tmc/it2/internal/cmd/lifecycle"
	"github.com/tmc/it2/internal/cmd/notification"
	"github.com/tmc/it2/internal/cmd/notify"
	"github.com/tmc/it2/internal/cmd/plugins"
	"github.com/tmc/it2/internal/cmd/preference"
	"github.com/tmc/it2/internal/cmd/profile"
	"github.com/tmc/it2/internal/cmd/prompt"
	"github.com/tmc/it2/internal/cmd/selection"
	"github.com/tmc/it2/internal/cmd/session"
	"github.com/tmc/it2/internal/cmd/shell"
	"github.com/tmc/it2/internal/cmd/snippet"
	"github.com/tmc/it2/internal/cmd/statusbar"
	"github.com/tmc/it2/internal/cmd/subscribe"
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
	"github.com/tmc/it2/internal/suggestions"
)

var (
	wsURL          string
	timeout        time.Duration
	format         string
	pluginPaths    []string
	pluginDeadline time.Duration
)

var rootCmd = &cobra.Command{
	Use:   "it2",
	Short: "Comprehensive command-line interface for iTerm2 automation",
	// Disable Cobra's built-in suggestions to use our own
	DisableSuggestions: true,
	SilenceErrors:      true, // Handle errors ourselves
	SilenceUsage:       true, // Don't show usage on runtime errors
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

Quick Access Commands
  get-screen, get-buffer, split - Convenient shortcuts for common operations

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

// newDeprecatedToolCommand creates the deprecated 'tool' command that wraps 'plugin'
// It warns once per day and delegates to the plugin command.
func newDeprecatedToolCommand() *cobra.Command {
	// Get the plugin command to wrap
	pluginCmd := plugins.NewCommand()

	cmd := &cobra.Command{
		Use:        "tool",
		Short:      "[DEPRECATED] Use 'it2 plugin' instead",
		Long:       "DEPRECATED: The 'it2 tool' command has been renamed to 'it2 plugin'.\n\nPlease use 'it2 plugin' for all plugin-related operations.",
		Hidden:     true, // Hide from main help
		GroupID:    "config",
		Deprecated: "use 'it2 plugin' instead",
		PersistentPreRun: func(cmd *cobra.Command, args []string) {
			maybeWarnDeprecation()
		},
	}

	// Copy all subcommands from plugin to tool
	for _, subCmd := range pluginCmd.Commands() {
		cmd.AddCommand(subCmd)
	}

	// Copy command groups
	for _, group := range pluginCmd.Groups() {
		cmd.AddGroup(group)
	}

	return cmd
}

// maybeWarnDeprecation checks if we should warn about deprecation (once per day)
func maybeWarnDeprecation() {
	home, err := os.UserHomeDir()
	if err != nil {
		// Can't check, just warn
		fmt.Fprintln(os.Stderr, "Warning: 'it2 tool' is deprecated, use 'it2 plugin' instead")
		return
	}

	warnFile := filepath.Join(home, ".it2", "deprecation-warnings.json")

	// Try to read existing warnings
	var warnings map[string]string
	data, err := os.ReadFile(warnFile)
	if err == nil {
		json.Unmarshal(data, &warnings)
	}
	if warnings == nil {
		warnings = make(map[string]string)
	}

	// Check if we warned today
	today := time.Now().Format("2006-01-02")
	if lastWarn, ok := warnings["tool"]; ok && lastWarn == today {
		return // Already warned today
	}

	// Warn and record
	fmt.Fprintln(os.Stderr, "Warning: 'it2 tool' is deprecated, use 'it2 plugin' instead")
	warnings["tool"] = today

	// Save warnings
	if err := os.MkdirAll(filepath.Dir(warnFile), 0755); err == nil {
		if data, err := json.Marshal(warnings); err == nil {
			os.WriteFile(warnFile, data, 0644)
		}
	}
}

// newQuickstartCommand creates the quickstart help command
func newQuickstartCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "quickstart",
		Short: "Get started with it2 - comprehensive guide for core workflows",
		Long: `it2 - iTerm2 Command-Line Automation

Control iTerm2 sessions, automate workflows, and integrate with AI agents.

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
 QUICK START (Try These First!)
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

1. See your sessions:
   it2 session list

2. Create a new split and save its ID:
   SID=$(it2 session split --vertical)
   echo $SID

3. Label it (best practice: ID + name):
   it2 session set-badge "$SID" "$(echo ${SID:0:8})\nWork"

4. Send a command to it:
   it2 session send-text "$SID" 'echo "Hello from it2!"'

5. Check what happened:
   it2 session get-screen "$SID"

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
 CORE COMMANDS
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

Session Management:
  it2 session list                        List all sessions (IDs + names)
  it2 session split --horizontal          Split current session
  it2 session get-info $SID               Get session details
  it2 session lookup left|right|up|down   Find adjacent sessions
  it2 session tree                        Visualize session layout

Sending Commands:
  it2 session send-text $SID 'command'           Send text (auto-confirms delivery)
  it2 session send-key $SID enter                Send special keys
  it2 session send-key $SID ctrl-c               Send key combos

Reading Output:
  it2 session get-screen $SID             Get current visible screen
  it2 session get-buffer $SID             Get entire scrollback
  it2 get-screen --wait-stable            Wait for output to settle

Organization:
  it2 session set-badge $SID "Label"      Label sessions visually
  it2 window list                         List windows
  it2 tab list                            List tabs

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
 AUTOMATION & SAFETY
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

Pre-conditions prevent sending commands at the wrong time:

  # Only send if shell is ready
  it2 session send-text $SID --require is-at-prompt 'git status'

  # Ensure no partial input exists
  it2 session send-text $SID --require has-no-partial-input 'make build'

  # For Claude Code sessions (multiple conditions)
  it2 session send-text $SID \
    --require is-claude-session,is-at-prompt,has-no-queued-messages \
    'continue with next task'

  # Add retries for reliability
  it2 session send-text $SID --retry 3 --retry-delay 1s 'command'

Available plugins:
  it2 plugin list                        Show all plugins
  it2 plugin list --type condition       Show pre-conditions only

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
 PRACTICAL EXAMPLES
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

Example 1: Create labeled work session
  SID=$(it2 session split --vertical)
  it2 session set-badge "$SID" "$(echo ${SID:0:8})\nBuild"
  it2 session send-text "$SID" --require is-at-prompt 'make test'

Example 2: Monitor long-running command
  it2 session send-text $SID 'npm install'
  it2 session get-screen $SID --wait-stable  # Waits for completion

Example 3: Multi-session workflow
  BUILD=$(it2 session split --vertical)
  TEST=$(it2 session split --horizontal)
  it2 session set-badge "$BUILD" "$(echo ${BUILD:0:8})\nBuild"
  it2 session set-badge "$TEST" "$(echo ${TEST:0:8})\nTest"

Example 4: JSON output for scripts
  it2 session list --format json | jq '.[] | select(.name | contains("prod"))'
  it2 session get-info $SID --format json | jq .grid_size

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
 AGENT INTEGRATION
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

it2 is designed for AI-supervised workflows:
  • Session IDs are stable across restarts
  • Badges track session purpose
  • Pre-conditions prevent unsafe operations
  • JSON output for programmatic parsing
  • Delivery confirmation built into send-text

Agent workflow with bd (Beads issue tracker):
  # Create work session and claim task
  SID=$(it2 session split --vertical)
  TASK=$(bd ready --json | jq -r '.[0].id')
  bd update $TASK --assignee "$SID" --status in_progress
  it2 session set-badge "$SID" "$(echo ${SID:0:8})\n$TASK"

  # Execute work safely
  it2 session send-text "$SID" --require is-at-prompt 'make build'
  it2 session get-screen "$SID" --wait-stable | tail -20

  # Complete task
  bd close $TASK

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
 CONFIGURATION
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

  it2 config show                        Show current configuration
  it2 config init                        Create config file
  it2 auth enable                        Enable iTerm2 API automation

Global flags (work with any command):
  --format json|yaml|table               Output format
  --timeout 60s                          Operation timeout

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
 NEXT STEPS
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

Ready to get productive!

Try this now:
  1. it2 session list              (see your sessions)
  2. it2 session split --vertical  (create a new one)
  3. it2 session set-badge $SID "$(echo ${SID:0:8})\nTest"  (label it)

Learn more:
  it2 --help                       Full command reference
  it2 session --help               Session command details
  it2 plugin list                  Available plugins`,
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println(cmd.Long)
		},
	}
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
	rootCmd.AddGroup(&cobra.Group{ID: "shortcuts", Title: "Quick Access Commands"})
	rootCmd.AddGroup(&cobra.Group{ID: "core", Title: "Core Operations"})
	rootCmd.AddGroup(&cobra.Group{ID: "content", Title: "Content & Text"})
	rootCmd.AddGroup(&cobra.Group{ID: "config", Title: "Configuration"})
	rootCmd.AddGroup(&cobra.Group{ID: "monitoring", Title: "Monitoring"})
	rootCmd.AddGroup(&cobra.Group{ID: "advanced", Title: "Advanced"})

	// Add shell completion commands
	rootCmd.AddCommand(newCompletionCommand())
	rootCmd.AddCommand(newConfigCommand())
	rootCmd.AddCommand(newQuickstartCommand())

	// Add shortcut commands (top-level shortcuts)
	rootCmd.AddCommand(shortcuts.NewGetScreenCommand())
	rootCmd.AddCommand(shortcuts.NewGetBufferCommand())
	rootCmd.AddCommand(shortcuts.NewSplitCommand())

	// Add organized command groups
	rootCmd.AddCommand(app.NewCommand())
	rootCmd.AddCommand(arrangement.NewCommand())
	rootCmd.AddCommand(artifact.NewCommand())
	rootCmd.AddCommand(attention.NewCommand())
	rootCmd.AddCommand(auth.NewCommand())
	rootCmd.AddCommand(badge.NewCommand())
	rootCmd.AddCommand(broadcast.NewCommand())
	rootCmd.AddCommand(color.NewCommand())
	rootCmd.AddCommand(job.NewCommand())
	rootCmd.AddCommand(notification.NewCommand())
	rootCmd.AddCommand(notify.NewCommand())
	rootCmd.AddCommand(plugins.NewCommand())
	rootCmd.AddCommand(profile.NewCommand())
	rootCmd.AddCommand(prompt.NewCommand())
	rootCmd.AddCommand(selection.NewCommand())
	rootCmd.AddCommand(session.NewCommand())
	rootCmd.AddCommand(shell.NewCommand())
	rootCmd.AddCommand(statusbar.NewCommand())
	rootCmd.AddCommand(subscribe.NewCommand())
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

	// Add deprecated 'tool' command (hidden, warns once per day)
	rootCmd.AddCommand(newDeprecatedToolCommand())
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		errMsg := err.Error()

		// Print the base error message
		fmt.Fprintf(os.Stderr, "Error: %s\n", errMsg)

		// Check if this is an unknown command error and provide suggestions
		if strings.Contains(errMsg, "unknown command") {
			// Extract the invalid command from the error message
			// Error format: 'unknown command "COMMAND" for "it2"'
			if startIdx := strings.Index(errMsg, `"`); startIdx != -1 {
				endIdx := strings.Index(errMsg[startIdx+1:], `"`)
				if endIdx != -1 {
					invalidCmd := errMsg[startIdx+1 : startIdx+1+endIdx]
					similarCmds := suggestions.FindSimilarCommands(rootCmd, invalidCmd)
					if suggestMsg := suggestions.FormatSuggestions(invalidCmd, similarCmds); suggestMsg != "" {
						fmt.Fprint(os.Stderr, suggestMsg)
					}
				}
			}
		}

		fmt.Fprintln(os.Stderr, "Run 'it2 --help' for usage.")
		os.Exit(1)
	}
}
