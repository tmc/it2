package app

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"time"

	"github.com/spf13/cobra"
	"github.com/tmc/it2/internal/cmd/app/settings"
	"github.com/tmc/it2/internal/cmdcore"
	"github.com/tmc/it2/internal/launcher"
)

// NewCommand creates the app command with all subcommands.
func NewCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "app",
		GroupID: "core",
		Short:   "Control the iTerm2 application",
		Long:    "Commands for interacting with iTerm2 as a whole, including activation and window management",
	}

	cmd.AddCommand(newLaunchCommand())
	cmd.AddCommand(newQuitCommand())
	cmd.AddCommand(newVersionCommand())
	cmd.AddCommand(newGetVariableCommand())
	cmd.AddCommand(newSetVariableCommand())
	cmd.AddCommand(settings.NewCommand())

	return cmd
}

func newGetVariableCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "get-variable <variable-name>",
		Short: "Get the value of an application-level variable",
		Long: `Get the value of an application-level variable.
Common variables:
  id              - Current session ID
  tab.id          - Current tab ID
  window.id       - Current window ID
  effectiveTheme  - Current theme (light/dark)`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			varName := args[0]

			// First check if we can get it from environment
			// iTerm2 sets various environment variables
			envVars := map[string]string{
				"id":          "ITERM_SESSION_ID",
				"sessionId":   "ITERM_SESSION_ID",
				"profileName": "ITERM_PROFILE",
			}

			if envVar, ok := envVars[varName]; ok {
				if value := os.Getenv(envVar); value != "" {
					fmt.Println(value)
					return nil
				}
			}

			_, timeout, _ := cmdcore.GetFlags(cmd)
			ctx, cancel := cmdcore.CreateContext(timeout)
			defer cancel()

			c, err := cmdcore.ConnectClient(ctx)
			if err != nil {
				return fmt.Errorf("failed to connect: %w", err)
			}
			defer c.Close()

			// Get variable via API
			value, err := c.GetVariable(ctx, "", varName)
			if err != nil {
				return fmt.Errorf("failed to get variable: %w", err)
			}

			fmt.Println(value)
			return nil
		},
	}
}

func newSetVariableCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "set-variable <variable-name> <json-value>",
		Short: "Set a user-defined application-level variable",
		Long:  "Set a user-defined application-level variable. Variable name must begin with 'user.'",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			varName := args[0]
			varValue := args[1]

			// Validate variable name
			if len(varName) < 5 || varName[:5] != "user." {
				return fmt.Errorf("variable name must begin with 'user.'")
			}

			// Validate JSON
			var jsonValue interface{}
			if err := json.Unmarshal([]byte(varValue), &jsonValue); err != nil {
				return fmt.Errorf("invalid JSON value: %w", err)
			}

			_, timeout, _ := cmdcore.GetFlags(cmd)
			ctx, cancel := cmdcore.CreateContext(timeout)
			defer cancel()

			c, err := cmdcore.ConnectClient(ctx)
			if err != nil {
				return fmt.Errorf("failed to connect: %w", err)
			}
			defer c.Close()

			// Set variable via API
			err = c.SetVariable(ctx, "", varName, varValue)
			if err != nil {
				return fmt.Errorf("failed to set variable: %w", err)
			}

			fmt.Fprintf(os.Stderr, "Variable '%s' set successfully\n", varName)
			return nil
		},
	}
}

func newVersionCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Get iTerm2 version information",
		Long:  "Display iTerm2 version and build information",
		RunE: func(cmd *cobra.Command, args []string) error {
			_, timeout, format := cmdcore.GetFlags(cmd)

			ctx, cancel := cmdcore.CreateContext(timeout)
			defer cancel()

			c, err := cmdcore.ConnectClient(ctx)
			if err != nil {
				return fmt.Errorf("failed to connect: %w", err)
			}
			defer c.Close()

			// Try to get version information through available variables first
			versionInfo := map[string]interface{}{}

			// Try to get app variables that might contain version info
			if variables, err := c.GetVariable(ctx, "", ""); err == nil {
				var variablesMap map[string]interface{}
				if err := json.Unmarshal([]byte(variables), &variablesMap); err == nil {
					versionInfo = variablesMap
				}
			}

			// If no version info available from variables, try fallback to system info
			if len(versionInfo) == 0 {
				versionInfo = make(map[string]interface{})
				// Get basic app information that is available
				if effectiveTheme, err := c.GetVariable(ctx, "", "effectiveTheme"); err == nil {
					versionInfo["effectiveTheme"] = effectiveTheme
				}
				if pid, err := c.GetVariable(ctx, "", "pid"); err == nil {
					versionInfo["pid"] = pid
				}
				if hostname, err := c.GetVariable(ctx, "", "localhostName"); err == nil {
					versionInfo["hostname"] = hostname
				}
			}

			// Format output
			switch format {
			case "json":
				output, _ := json.MarshalIndent(versionInfo, "", "  ")
				fmt.Println(string(output))
			default:
				if len(versionInfo) == 0 {
					fmt.Println("iTerm2 application information not available")
					return nil
				}

				fmt.Println("iTerm2 Application Information:")
				fmt.Println("==============================")

				// Show available information
				if v, ok := versionInfo["version"]; ok {
					fmt.Printf("Version:        %v\n", v)
				}
				if b, ok := versionInfo["build"]; ok {
					fmt.Printf("Build:          %v\n", b)
				}
				if pid, ok := versionInfo["pid"]; ok {
					fmt.Printf("Process ID:     %v\n", pid)
				}
				if theme, ok := versionInfo["effectiveTheme"]; ok {
					fmt.Printf("Theme:          %v\n", theme)
				}
				if hostname, ok := versionInfo["hostname"]; ok {
					fmt.Printf("Hostname:       %v\n", hostname)
				}

				// Show any other available variables
				for key, value := range versionInfo {
					if key != "version" && key != "build" && key != "pid" && key != "effectiveTheme" && key != "hostname" {
						fmt.Printf("%-15s %v\n", key+":", value)
					}
				}
			}

			return nil
		},
	}
}

func newLaunchCommand() *cobra.Command {
	var wait bool
	var timeout time.Duration

	cmd := &cobra.Command{
		Use:   "launch",
		Short: "Launch iTerm2 if not running",
		Long: `Launch iTerm2 if it's not already running.
This is useful when running it2 from Terminal.app or other terminals.

Examples:
  # Launch iTerm2 and wait for it to be ready
  $ it2 app launch

  # Launch iTerm2 without waiting
  $ it2 app launch --no-wait

  # Launch with custom timeout
  $ it2 app launch --timeout 60s`,
		RunE: func(cmd *cobra.Command, args []string) error {
			// Check current terminal
			termInfo := launcher.GetTerminalInfo()
			if launcher.IsTerminalApp() {
				fmt.Println("Running from Terminal.app")
			} else {
				fmt.Printf("Running from: %s\n", termInfo)
			}

			// Check if already running
			if launcher.IsITerm2Running() {
				fmt.Println("iTerm2 is already running")
				if wait && !launcher.SocketExists() {
					fmt.Println("Waiting for API socket...")
					ctx, cancel := context.WithTimeout(context.Background(), timeout)
					defer cancel()
					if err := launcher.WaitForITerm2(ctx); err != nil {
						return fmt.Errorf("failed waiting for iTerm2 socket: %w", err)
					}
					fmt.Println("API socket is now available")
				} else if launcher.SocketExists() {
					fmt.Println("API socket is available")
				}
				return nil
			}

			fmt.Println("Launching iTerm2...")
			if wait {
				ctx, cancel := context.WithTimeout(context.Background(), timeout)
				defer cancel()
				if err := launcher.LaunchITerm2(ctx); err != nil {
					return fmt.Errorf("failed to launch iTerm2: %w", err)
				}
				fmt.Println("iTerm2 launched successfully and ready")
			} else {
				if err := launcher.LaunchITerm2(context.Background()); err != nil {
					return fmt.Errorf("failed to launch iTerm2: %w", err)
				}
				fmt.Println("iTerm2 launch initiated")
			}

			return nil
		},
	}

	cmd.Flags().BoolVar(&wait, "wait", true, "Wait for iTerm2 to be ready")
	cmd.Flags().BoolVar(&wait, "no-wait", false, "Don't wait for iTerm2 to be ready (opposite of --wait)")
	cmd.Flags().DurationVar(&timeout, "timeout", 30*time.Second, "Timeout for waiting")

	return cmd
}

func newQuitCommand() *cobra.Command {
	var force bool

	cmd := &cobra.Command{
		Use:   "quit",
		Short: "Quit iTerm2 application",
		Long: `Quit the iTerm2 application gracefully.

Examples:
  # Quit iTerm2
  $ it2 app quit

  # Force quit if unresponsive
  $ it2 app quit --force`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if force {
				// Force quit using kill
				if launcher.IsITerm2Running() {
					killCmd := exec.Command("pkill", "-x", "iTerm2")
					if err := killCmd.Run(); err != nil {
						return fmt.Errorf("failed to force quit iTerm2: %w", err)
					}
					fmt.Println("iTerm2 force quit")
				} else {
					fmt.Println("iTerm2 is not running")
				}
				return nil
			}

			// Graceful quit using AppleScript
			if launcher.IsITerm2Running() {
				quitCmd := exec.Command("osascript", "-e", `tell application "iTerm2" to quit`)
				if err := quitCmd.Run(); err != nil {
					// Try alternative app name
					quitCmd = exec.Command("osascript", "-e", `tell application "iTerm" to quit`)
					if err := quitCmd.Run(); err != nil {
						return fmt.Errorf("failed to quit iTerm2: %w", err)
					}
				}
				fmt.Println("iTerm2 quit successfully")
			} else {
				fmt.Println("iTerm2 is not running")
			}

			return nil
		},
	}

	cmd.Flags().BoolVar(&force, "force", false, "Force quit iTerm2 (using pkill)")

	return cmd
}
