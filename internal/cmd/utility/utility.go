package utility

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/spf13/cobra"
	"github.com/tmc/it2/internal/cmdutil"
	"github.com/tmc/it2/internal/formatting"
	pb "github.com/tmc/it2/proto"
)

// NewCommand creates the utility command with all subcommands.
func NewCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "utility",
		Short: "iTerm2 utility and helper commands",
		Long:  "Commands for cursor guides, timestamps, shell integration, and other utilities",
	}

	cmd.AddCommand(newCursorGuideCommand())
	cmd.AddCommand(newTimestampCommand())
	cmd.AddCommand(newShellIntegrationCommand())
	cmd.AddCommand(newPasswordManagerCommand())
	cmd.AddCommand(newSSHCommand())
	cmd.AddCommand(newBadgeCommand())
	cmd.AddCommand(newTouchBarCommand())

	return cmd
}

func newCursorGuideCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "cursor-guide",
		Short: "Control cursor guide visibility",
		Long:  "Show or hide cursor guides that help track cursor position",
	}

	cmd.AddCommand(newCursorGuideShowCommand())
	cmd.AddCommand(newCursorGuideHideCommand())
	cmd.AddCommand(newCursorGuideToggleCommand())

	return cmd
}

func newCursorGuideShowCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "show [session-id]",
		Short: "Show cursor guides",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			sessionID := "active"
			if len(args) > 0 {
				sessionID = args[0]
			}

			wsURL, timeout, _ := cmdutil.GetFlags(cmd)

			ctx, cancel := cmdutil.CreateContext(timeout)
			defer cancel()

			c, err := cmdutil.ConnectClient(ctx, wsURL)
			if err != nil {
				return fmt.Errorf("failed to connect: %w", err)
			}
			defer c.Close()

			// Use InvokeFunction to show cursor guides
			invocation := "iterm2.set_cursor_guide(True)"

			request := &pb.InvokeFunctionRequest{
				Context: &pb.InvokeFunctionRequest_Session{
					Session: &pb.InvokeFunctionRequest_Session{
						SessionId: sessionID,
					},
				},
				Invocation: invocation,
				Timeout:    float64(timeout.Seconds()),
			}

			response, err := c.InvokeFunction(ctx, request)
			if err != nil {
				return fmt.Errorf("failed to show cursor guides: %w", err)
			}

			if response.GetError() != nil {
				return fmt.Errorf("cursor guide command failed: %s", response.GetError().GetErrorReason())
			}

			fmt.Println("Cursor guides shown")
			return nil
		},
	}

	return cmd
}

func newCursorGuideHideCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "hide [session-id]",
		Short: "Hide cursor guides",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			sessionID := "active"
			if len(args) > 0 {
				sessionID = args[0]
			}

			wsURL, timeout, _ := cmdutil.GetFlags(cmd)

			ctx, cancel := cmdutil.CreateContext(timeout)
			defer cancel()

			c, err := cmdutil.ConnectClient(ctx, wsURL)
			if err != nil {
				return fmt.Errorf("failed to connect: %w", err)
			}
			defer c.Close()

			// Use InvokeFunction to hide cursor guides
			invocation := "iterm2.set_cursor_guide(False)"

			request := &pb.InvokeFunctionRequest{
				Context: &pb.InvokeFunctionRequest_Session{
					Session: &pb.InvokeFunctionRequest_Session{
						SessionId: sessionID,
					},
				},
				Invocation: invocation,
				Timeout:    float64(timeout.Seconds()),
			}

			response, err := c.InvokeFunction(ctx, request)
			if err != nil {
				return fmt.Errorf("failed to hide cursor guides: %w", err)
			}

			if response.GetError() != nil {
				return fmt.Errorf("cursor guide command failed: %s", response.GetError().GetErrorReason())
			}

			fmt.Println("Cursor guides hidden")
			return nil
		},
	}

	return cmd
}

func newCursorGuideToggleCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "toggle [session-id]",
		Short: "Toggle cursor guides",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			sessionID := "active"
			if len(args) > 0 {
				sessionID = args[0]
			}

			wsURL, timeout, _ := cmdutil.GetFlags(cmd)

			ctx, cancel := cmdutil.CreateContext(timeout)
			defer cancel()

			c, err := cmdutil.ConnectClient(ctx, wsURL)
			if err != nil {
				return fmt.Errorf("failed to connect: %w", err)
			}
			defer c.Close()

			// Use InvokeFunction to toggle cursor guides
			invocation := "iterm2.toggle_cursor_guide()"

			request := &pb.InvokeFunctionRequest{
				Context: &pb.InvokeFunctionRequest_Session{
					Session: &pb.InvokeFunctionRequest_Session{
						SessionId: sessionID,
					},
				},
				Invocation: invocation,
				Timeout:    float64(timeout.Seconds()),
			}

			response, err := c.InvokeFunction(ctx, request)
			if err != nil {
				return fmt.Errorf("failed to toggle cursor guides: %w", err)
			}

			if response.GetError() != nil {
				return fmt.Errorf("cursor guide command failed: %s", response.GetError().GetErrorReason())
			}

			fmt.Println("Cursor guides toggled")
			return nil
		},
	}

	return cmd
}

func newTimestampCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "timestamp",
		Short: "Control timestamp visibility",
		Long:  "Show or hide timestamps for commands in the session",
	}

	cmd.AddCommand(newTimestampShowCommand())
	cmd.AddCommand(newTimestampHideCommand())
	cmd.AddCommand(newTimestampToggleCommand())

	return cmd
}

func newTimestampShowCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "show [session-id]",
		Short: "Show timestamps",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			sessionID := "active"
			if len(args) > 0 {
				sessionID = args[0]
			}

			wsURL, timeout, _ := cmdutil.GetFlags(cmd)

			ctx, cancel := cmdutil.CreateContext(timeout)
			defer cancel()

			c, err := cmdutil.ConnectClient(ctx, wsURL)
			if err != nil {
				return fmt.Errorf("failed to connect: %w", err)
			}
			defer c.Close()

			// Set timestamp property
			msg := &pb.ClientOriginatedMessage{
				Submessage: &pb.ClientOriginatedMessage_SetPropertyRequest{
					SetPropertyRequest: &pb.SetPropertyRequest{
						Identifier: &pb.SetPropertyRequest_SessionId{
							SessionId: sessionID,
						},
						Name:      cmdutil.StringPtr("show_timestamps"),
						JsonValue: cmdutil.StringPtr("true"),
					},
				},
			}

			response, err := c.SendRequest(ctx, msg)
			if err != nil {
				return fmt.Errorf("failed to show timestamps: %w", err)
			}

			if resp := response.GetSetPropertyResponse(); resp != nil {
				if resp.GetStatus() != pb.SetPropertyResponse_OK {
					return fmt.Errorf("failed to show timestamps: %v", resp.GetStatus())
				}
			}

			fmt.Println("Timestamps shown")
			return nil
		},
	}

	return cmd
}

func newTimestampHideCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "hide [session-id]",
		Short: "Hide timestamps",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			sessionID := "active"
			if len(args) > 0 {
				sessionID = args[0]
			}

			wsURL, timeout, _ := cmdutil.GetFlags(cmd)

			ctx, cancel := cmdutil.CreateContext(timeout)
			defer cancel()

			c, err := cmdutil.ConnectClient(ctx, wsURL)
			if err != nil {
				return fmt.Errorf("failed to connect: %w", err)
			}
			defer c.Close()

			// Set timestamp property
			msg := &pb.ClientOriginatedMessage{
				Submessage: &pb.ClientOriginatedMessage_SetPropertyRequest{
					SetPropertyRequest: &pb.SetPropertyRequest{
						Identifier: &pb.SetPropertyRequest_SessionId{
							SessionId: sessionID,
						},
						Name:      cmdutil.StringPtr("show_timestamps"),
						JsonValue: cmdutil.StringPtr("false"),
					},
				},
			}

			response, err := c.SendRequest(ctx, msg)
			if err != nil {
				return fmt.Errorf("failed to hide timestamps: %w", err)
			}

			if resp := response.GetSetPropertyResponse(); resp != nil {
				if resp.GetStatus() != pb.SetPropertyResponse_OK {
					return fmt.Errorf("failed to hide timestamps: %v", resp.GetStatus())
				}
			}

			fmt.Println("Timestamps hidden")
			return nil
		},
	}

	return cmd
}

func newTimestampToggleCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "toggle [session-id]",
		Short: "Toggle timestamps",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			sessionID := "active"
			if len(args) > 0 {
				sessionID = args[0]
			}

			wsURL, timeout, _ := cmdutil.GetFlags(cmd)

			ctx, cancel := cmdutil.CreateContext(timeout)
			defer cancel()

			c, err := cmdutil.ConnectClient(ctx, wsURL)
			if err != nil {
				return fmt.Errorf("failed to connect: %w", err)
			}
			defer c.Close()

			// First get current state
			getMsg := &pb.ClientOriginatedMessage{
				Submessage: &pb.ClientOriginatedMessage_GetPropertyRequest{
					GetPropertyRequest: &pb.GetPropertyRequest{
						Identifier: &pb.GetPropertyRequest_SessionId{
							SessionId: sessionID,
						},
						Name: cmdutil.StringPtr("show_timestamps"),
					},
				},
			}

			response, err := c.SendRequest(ctx, getMsg)
			if err != nil {
				return fmt.Errorf("failed to get timestamp status: %w", err)
			}

			currentValue := "false"
			if resp := response.GetPropertyResponse(); resp != nil {
				if resp.GetStatus() == pb.GetPropertyResponse_OK && resp.JsonValue != nil {
					currentValue = *resp.JsonValue
				}
			}

			// Toggle the value
			newValue := "true"
			if currentValue == "true" {
				newValue = "false"
			}

			setMsg := &pb.ClientOriginatedMessage{
				Submessage: &pb.ClientOriginatedMessage_SetPropertyRequest{
					SetPropertyRequest: &pb.SetPropertyRequest{
						Identifier: &pb.SetPropertyRequest_SessionId{
							SessionId: sessionID,
						},
						Name:      cmdutil.StringPtr("show_timestamps"),
						JsonValue: cmdutil.StringPtr(newValue),
					},
				},
			}

			response, err = c.SendRequest(ctx, setMsg)
			if err != nil {
				return fmt.Errorf("failed to toggle timestamps: %w", err)
			}

			if resp := response.GetSetPropertyResponse(); resp != nil {
				if resp.GetStatus() != pb.SetPropertyResponse_OK {
					return fmt.Errorf("failed to toggle timestamps: %v", resp.GetStatus())
				}
			}

			action := "shown"
			if newValue == "false" {
				action = "hidden"
			}
			fmt.Printf("Timestamps %s\n", action)
			return nil
		},
	}

	return cmd
}

func newShellIntegrationCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "shell-integration",
		Short: "Manage shell integration installation",
		Long:  "Install or check status of iTerm2 shell integration",
	}

	cmd.AddCommand(newShellIntegrationInstallCommand())
	cmd.AddCommand(newShellIntegrationStatusCommand())
	cmd.AddCommand(newShellIntegrationUninstallCommand())

	return cmd
}

func newShellIntegrationInstallCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "install [shell]",
		Short: "Install shell integration scripts",
		Long:  "Install iTerm2 shell integration for the specified shell (bash, zsh, fish, tcsh)",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			shell := ""
			if len(args) > 0 {
				shell = args[0]
			}

			force, _ := cmd.Flags().GetBool("force")

			// Detect shell if not provided
			if shell == "" {
				if envShell := os.Getenv("SHELL"); envShell != "" {
					shell = filepath.Base(envShell)
				} else {
					shell = "bash" // default
				}
			}

			// Validate shell
			validShells := []string{"bash", "zsh", "fish", "tcsh"}
			valid := false
			for _, vs := range validShells {
				if shell == vs {
					valid = true
					break
				}
			}
			if !valid {
				return fmt.Errorf("unsupported shell: %s. Valid shells: %v", shell, validShells)
			}

			homeDir, err := os.UserHomeDir()
			if err != nil {
				return fmt.Errorf("failed to get home directory: %w", err)
			}

			// Determine config file based on shell
			var configFile string
			switch shell {
			case "bash":
				configFile = filepath.Join(homeDir, ".bashrc")
				if runtime.GOOS == "darwin" {
					// On macOS, use .bash_profile
					configFile = filepath.Join(homeDir, ".bash_profile")
				}
			case "zsh":
				configFile = filepath.Join(homeDir, ".zshrc")
			case "fish":
				configFile = filepath.Join(homeDir, ".config/fish/config.fish")
			case "tcsh":
				configFile = filepath.Join(homeDir, ".tcshrc")
			}

			// Check if integration is already installed
			if !force {
				if data, err := os.ReadFile(configFile); err == nil {
					if strings.Contains(string(data), "iterm2_shell_integration") {
						fmt.Printf("Shell integration already installed in %s\n", configFile)
						fmt.Println("Use --force to reinstall")
						return nil
					}
				}
			}

			// Integration script content
			integrationScript := fmt.Sprintf(`
# iTerm2 Shell Integration
test -e "${HOME}/.iterm2_shell_integration.%s" && source "${HOME}/.iterm2_shell_integration.%s"
`, shell, shell)

			// Download and install integration script
			scriptFile := filepath.Join(homeDir, fmt.Sprintf(".iterm2_shell_integration.%s", shell))

			// Create a simple integration script (in a real implementation, this would download from iTerm2)
			basicScript := `#!/bin/bash
# iTerm2 Shell Integration
# This is a simplified version - the real script would be downloaded from iTerm2

# Basic functionality
export ITERM_SESSION_ID="$TERM_SESSION_ID"

# Hook function to report directory changes
if [[ "$SHELL" == *"bash"* ]]; then
    export PROMPT_COMMAND='echo -e "\033]1337;CurrentDir=$(pwd)\a";'$PROMPT_COMMAND
elif [[ "$SHELL" == *"zsh"* ]]; then
    precmd() { echo -e "\033]1337;CurrentDir=$(pwd)\a"; }
fi

# Function to set iTerm2 badge
iterm2_set_badge() {
    echo -e "\033]1337;SetBadgeFormat=$1\a"
}

# Function to set iTerm2 tab title
iterm2_set_title() {
    echo -e "\033]0;$1\a"
}

echo "iTerm2 Shell Integration loaded for $SHELL"
`

			err = os.WriteFile(scriptFile, []byte(basicScript), 0755)
			if err != nil {
				return fmt.Errorf("failed to create integration script: %w", err)
			}

			// Add to shell config
			configData := ""
			if data, err := os.ReadFile(configFile); err == nil {
				configData = string(data)
			}

			// Ensure directory exists for fish config
			if shell == "fish" {
				configDir := filepath.Dir(configFile)
				os.MkdirAll(configDir, 0755)
			}

			// Append integration to config file
			configData += integrationScript

			err = os.WriteFile(configFile, []byte(configData), 0644)
			if err != nil {
				return fmt.Errorf("failed to update config file %s: %w", configFile, err)
			}

			fmt.Printf("Shell integration installed for %s\n", shell)
			fmt.Printf("Script: %s\n", scriptFile)
			fmt.Printf("Config: %s\n", configFile)
			fmt.Println("Restart your shell or run 'source ~/.iterm2_shell_integration." + shell + "' to activate")

			return nil
		},
	}

	cmd.Flags().BoolP("force", "f", false, "Force reinstallation even if already installed")

	return cmd
}

func newShellIntegrationStatusCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "status",
		Short: "Check shell integration status",
		RunE: func(cmd *cobra.Command, args []string) error {
			format, _ := cmd.Flags().GetString("format")

			homeDir, err := os.UserHomeDir()
			if err != nil {
				return fmt.Errorf("failed to get home directory: %w", err)
			}

			shells := []string{"bash", "zsh", "fish", "tcsh"}
			status := make(map[string]interface{})

			for _, shell := range shells {
				shellStatus := map[string]interface{}{
					"script_exists":    false,
					"config_modified":  false,
					"config_file":      "",
					"script_file":      "",
				}

				// Check if script file exists
				scriptFile := filepath.Join(homeDir, fmt.Sprintf(".iterm2_shell_integration.%s", shell))
				if _, err := os.Stat(scriptFile); err == nil {
					shellStatus["script_exists"] = true
					shellStatus["script_file"] = scriptFile
				}

				// Check config file
				var configFile string
				switch shell {
				case "bash":
					configFile = filepath.Join(homeDir, ".bashrc")
					if runtime.GOOS == "darwin" {
						configFile = filepath.Join(homeDir, ".bash_profile")
					}
				case "zsh":
					configFile = filepath.Join(homeDir, ".zshrc")
				case "fish":
					configFile = filepath.Join(homeDir, ".config/fish/config.fish")
				case "tcsh":
					configFile = filepath.Join(homeDir, ".tcshrc")
				}

				shellStatus["config_file"] = configFile
				if data, err := os.ReadFile(configFile); err == nil {
					if strings.Contains(string(data), "iterm2_shell_integration") {
						shellStatus["config_modified"] = true
					}
				}

				status[shell] = shellStatus
			}

			formatter := formatting.New(format)
			if format == "json" {
				return formatter.FormatGeneric(status)
			}

			// Text format
			fmt.Println("iTerm2 Shell Integration Status:")
			fmt.Println()

			for _, shell := range shells {
				shellStatus := status[shell].(map[string]interface{})
				fmt.Printf("%s:\n", strings.ToUpper(shell))
				fmt.Printf("  Script exists: %v\n", shellStatus["script_exists"])
				fmt.Printf("  Config modified: %v\n", shellStatus["config_modified"])
				fmt.Printf("  Script file: %s\n", shellStatus["script_file"])
				fmt.Printf("  Config file: %s\n", shellStatus["config_file"])
				fmt.Println()
			}

			return nil
		},
	}

	cmd.Flags().StringP("format", "f", "text", "Output format (text, json)")

	return cmd
}

func newShellIntegrationUninstallCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "uninstall [shell]",
		Short: "Uninstall shell integration",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			shell := ""
			if len(args) > 0 {
				shell = args[0]
			}

			homeDir, err := os.UserHomeDir()
			if err != nil {
				return fmt.Errorf("failed to get home directory: %w", err)
			}

			shells := []string{}
			if shell != "" {
				shells = []string{shell}
			} else {
				shells = []string{"bash", "zsh", "fish", "tcsh"}
			}

			for _, sh := range shells {
				// Remove script file
				scriptFile := filepath.Join(homeDir, fmt.Sprintf(".iterm2_shell_integration.%s", sh))
				if err := os.Remove(scriptFile); err == nil {
					fmt.Printf("Removed script: %s\n", scriptFile)
				}

				// Remove from config file
				var configFile string
				switch sh {
				case "bash":
					configFile = filepath.Join(homeDir, ".bashrc")
					if runtime.GOOS == "darwin" {
						configFile = filepath.Join(homeDir, ".bash_profile")
					}
				case "zsh":
					configFile = filepath.Join(homeDir, ".zshrc")
				case "fish":
					configFile = filepath.Join(homeDir, ".config/fish/config.fish")
				case "tcsh":
					configFile = filepath.Join(homeDir, ".tcshrc")
				}

				if data, err := os.ReadFile(configFile); err == nil {
					lines := strings.Split(string(data), "\n")
					var newLines []string
					skipNext := false
					for _, line := range lines {
						if strings.Contains(line, "iTerm2 Shell Integration") {
							skipNext = true
							continue
						}
						if skipNext && strings.Contains(line, "iterm2_shell_integration") {
							skipNext = false
							continue
						}
						if !strings.Contains(line, "iterm2_shell_integration") {
							newLines = append(newLines, line)
						}
					}

					newContent := strings.Join(newLines, "\n")
					if err := os.WriteFile(configFile, []byte(newContent), 0644); err == nil {
						fmt.Printf("Updated config: %s\n", configFile)
					}
				}
			}

			fmt.Println("Shell integration uninstalled")
			fmt.Println("Restart your shell to complete removal")

			return nil
		},
	}

	return cmd
}

func newPasswordManagerCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "password-manager",
		Short: "Password manager integration",
		Long:  "Commands for integrating with password managers",
	}

	cmd.AddCommand(newPasswordManagerGetCommand())
	cmd.AddCommand(newPasswordManagerSetCommand())

	return cmd
}

func newPasswordManagerGetCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "get <account>",
		Short: "Get password from manager",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			account := args[0]
			service, _ := cmd.Flags().GetString("service")

			wsURL, timeout, _ := cmdutil.GetFlags(cmd)

			ctx, cancel := cmdutil.CreateContext(timeout)
			defer cancel()

			c, err := cmdutil.ConnectClient(ctx, wsURL)
			if err != nil {
				return fmt.Errorf("failed to connect: %w", err)
			}
			defer c.Close()

			// Use InvokeFunction to get password via iTerm2's password manager integration
			invocation := fmt.Sprintf(`
				set theAccount to "%s"
				set theService to "%s"

				tell application "iTerm2"
					try
						set thePassword to get password for account theAccount service theService
						return thePassword
					on error
						return "Password not found"
					end try
				end tell
			`, account, service)

			request := &pb.InvokeFunctionRequest{
				Context: &pb.InvokeFunctionRequest_App{
					App: &pb.InvokeFunctionRequest_App{},
				},
				Invocation: invocation,
				Timeout:    float64(timeout.Seconds()),
			}

			response, err := c.InvokeFunction(ctx, request)
			if err != nil {
				return fmt.Errorf("failed to get password: %w", err)
			}

			if response.GetError() != nil {
				return fmt.Errorf("password retrieval failed: %s", response.GetError().GetErrorReason())
			}

			if success := response.GetSuccess(); success != nil {
				fmt.Println(success.GetJsonResult())
			}

			return nil
		},
	}

	cmd.Flags().StringP("service", "s", "default", "Service name for password storage")

	return cmd
}

func newPasswordManagerSetCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "set <account> <password>",
		Short: "Set password in manager",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			account := args[0]
			password := args[1]
			service, _ := cmd.Flags().GetString("service")

			wsURL, timeout, _ := cmdutil.GetFlags(cmd)

			ctx, cancel := cmdutil.CreateContext(timeout)
			defer cancel()

			c, err := cmdutil.ConnectClient(ctx, wsURL)
			if err != nil {
				return fmt.Errorf("failed to connect: %w", err)
			}
			defer c.Close()

			// Use InvokeFunction to set password via iTerm2's password manager integration
			invocation := fmt.Sprintf(`
				set theAccount to "%s"
				set thePassword to "%s"
				set theService to "%s"

				tell application "iTerm2"
					set password for account theAccount service theService to thePassword
					return "Password stored"
				end tell
			`, account, password, service)

			request := &pb.InvokeFunctionRequest{
				Context: &pb.InvokeFunctionRequest_App{
					App: &pb.InvokeFunctionRequest_App{},
				},
				Invocation: invocation,
				Timeout:    float64(timeout.Seconds()),
			}

			response, err := c.InvokeFunction(ctx, request)
			if err != nil {
				return fmt.Errorf("failed to set password: %w", err)
			}

			if response.GetError() != nil {
				return fmt.Errorf("password storage failed: %s", response.GetError().GetErrorReason())
			}

			fmt.Printf("Password stored for account: %s\n", account)
			return nil
		},
	}

	cmd.Flags().StringP("service", "s", "default", "Service name for password storage")

	return cmd
}

func newSSHCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "ssh",
		Short: "SSH integration utilities",
		Long:  "Commands for SSH integration and key management",
	}

	cmd.AddCommand(newSSHConnectCommand())
	cmd.AddCommand(newSSHKeyCommand())

	return cmd
}

func newSSHConnectCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "connect <host>",
		Short: "Connect to SSH host with integration",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			host := args[0]
			user, _ := cmd.Flags().GetString("user")
			port, _ := cmd.Flags().GetInt("port")
			key, _ := cmd.Flags().GetString("key")
			sessionID, _ := cmd.Flags().GetString("session")

			if sessionID == "" {
				sessionID = "active"
			}

			wsURL, timeout, _ := cmdutil.GetFlags(cmd)

			ctx, cancel := cmdutil.CreateContext(timeout)
			defer cancel()

			c, err := cmdutil.ConnectClient(ctx, wsURL)
			if err != nil {
				return fmt.Errorf("failed to connect: %w", err)
			}
			defer c.Close()

			// Build SSH command
			sshCmd := fmt.Sprintf("ssh")
			if user != "" {
				sshCmd += fmt.Sprintf(" -l %s", user)
			}
			if port != 0 {
				sshCmd += fmt.Sprintf(" -p %d", port)
			}
			if key != "" {
				sshCmd += fmt.Sprintf(" -i %s", key)
			}
			sshCmd += fmt.Sprintf(" %s", host)

			// Send SSH command to session
			err = c.SendText(ctx, sessionID, sshCmd+"\n")
			if err != nil {
				return fmt.Errorf("failed to send SSH command: %w", err)
			}

			fmt.Printf("SSH command sent to session %s: %s\n", sessionID, sshCmd)
			return nil
		},
	}

	cmd.Flags().StringP("user", "u", "", "SSH username")
	cmd.Flags().IntP("port", "p", 22, "SSH port")
	cmd.Flags().StringP("key", "k", "", "SSH private key file")
	cmd.Flags().StringP("session", "s", "", "Target session ID (default: active)")

	return cmd
}

func newSSHKeyCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "key",
		Short: "SSH key management",
		Long:  "Manage SSH keys and agent integration",
	}

	cmd.AddCommand(newSSHKeyListCommand())
	cmd.AddCommand(newSSHKeyAddCommand())

	return cmd
}

func newSSHKeyListCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List SSH keys in agent",
		RunE: func(cmd *cobra.Command, args []string) error {
			sessionID, _ := cmd.Flags().GetString("session")
			if sessionID == "" {
				sessionID = "active"
			}

			wsURL, timeout, _ := cmdutil.GetFlags(cmd)

			ctx, cancel := cmdutil.CreateContext(timeout)
			defer cancel()

			c, err := cmdutil.ConnectClient(ctx, wsURL)
			if err != nil {
				return fmt.Errorf("failed to connect: %w", err)
			}
			defer c.Close()

			// Send ssh-add -l command
			err = c.SendText(ctx, sessionID, "ssh-add -l\n")
			if err != nil {
				return fmt.Errorf("failed to list SSH keys: %w", err)
			}

			fmt.Println("SSH key list command sent to session")
			return nil
		},
	}

	cmd.Flags().StringP("session", "s", "", "Target session ID (default: active)")

	return cmd
}

func newSSHKeyAddCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "add <key-file>",
		Short: "Add SSH key to agent",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			keyFile := args[0]
			sessionID, _ := cmd.Flags().GetString("session")
			if sessionID == "" {
				sessionID = "active"
			}

			wsURL, timeout, _ := cmdutil.GetFlags(cmd)

			ctx, cancel := cmdutil.CreateContext(timeout)
			defer cancel()

			c, err := cmdutil.ConnectClient(ctx, wsURL)
			if err != nil {
				return fmt.Errorf("failed to connect: %w", err)
			}
			defer c.Close()

			// Send ssh-add command
			sshAddCmd := fmt.Sprintf("ssh-add %s\n", keyFile)
			err = c.SendText(ctx, sessionID, sshAddCmd)
			if err != nil {
				return fmt.Errorf("failed to add SSH key: %w", err)
			}

			fmt.Printf("SSH key add command sent: %s", sshAddCmd)
			return nil
		},
	}

	cmd.Flags().StringP("session", "s", "", "Target session ID (default: active)")

	return cmd
}

func newBadgeCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "badge",
		Short: "Manage session badges",
		Long:  "Set or clear session badges for visual identification",
	}

	cmd.AddCommand(newBadgeSetCommand())
	cmd.AddCommand(newBadgeClearCommand())

	return cmd
}

func newBadgeSetCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "set <text> [session-id]",
		Short: "Set session badge text",
		Args:  cobra.RangeArgs(1, 2),
		RunE: func(cmd *cobra.Command, args []string) error {
			badgeText := args[0]
			sessionID := "active"
			if len(args) > 1 {
				sessionID = args[1]
			}

			wsURL, timeout, _ := cmdutil.GetFlags(cmd)

			ctx, cancel := cmdutil.CreateContext(timeout)
			defer cancel()

			c, err := cmdutil.ConnectClient(ctx, wsURL)
			if err != nil {
				return fmt.Errorf("failed to connect: %w", err)
			}
			defer c.Close()

			// Set badge property
			msg := &pb.ClientOriginatedMessage{
				Submessage: &pb.ClientOriginatedMessage_SetPropertyRequest{
					SetPropertyRequest: &pb.SetPropertyRequest{
						Identifier: &pb.SetPropertyRequest_SessionId{
							SessionId: sessionID,
						},
						Name:      cmdutil.StringPtr("badge"),
						JsonValue: cmdutil.StringPtr(fmt.Sprintf(`"%s"`, badgeText)),
					},
				},
			}

			response, err := c.SendRequest(ctx, msg)
			if err != nil {
				return fmt.Errorf("failed to set badge: %w", err)
			}

			if resp := response.GetSetPropertyResponse(); resp != nil {
				if resp.GetStatus() != pb.SetPropertyResponse_OK {
					return fmt.Errorf("failed to set badge: %v", resp.GetStatus())
				}
			}

			fmt.Printf("Badge set to: %s\n", badgeText)
			return nil
		},
	}

	return cmd
}

func newBadgeClearCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "clear [session-id]",
		Short: "Clear session badge",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			sessionID := "active"
			if len(args) > 0 {
				sessionID = args[0]
			}

			wsURL, timeout, _ := cmdutil.GetFlags(cmd)

			ctx, cancel := cmdutil.CreateContext(timeout)
			defer cancel()

			c, err := cmdutil.ConnectClient(ctx, wsURL)
			if err != nil {
				return fmt.Errorf("failed to connect: %w", err)
			}
			defer c.Close()

			// Clear badge property
			msg := &pb.ClientOriginatedMessage{
				Submessage: &pb.ClientOriginatedMessage_SetPropertyRequest{
					SetPropertyRequest: &pb.SetPropertyRequest{
						Identifier: &pb.SetPropertyRequest_SessionId{
							SessionId: sessionID,
						},
						Name:      cmdutil.StringPtr("badge"),
						JsonValue: cmdutil.StringPtr(`""`),
					},
				},
			}

			response, err := c.SendRequest(ctx, msg)
			if err != nil {
				return fmt.Errorf("failed to clear badge: %w", err)
			}

			if resp := response.GetSetPropertyResponse(); resp != nil {
				if resp.GetStatus() != pb.SetPropertyResponse_OK {
					return fmt.Errorf("failed to clear badge: %v", resp.GetStatus())
				}
			}

			fmt.Println("Badge cleared")
			return nil
		},
	}

	return cmd
}

func newTouchBarCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "touchbar",
		Short: "TouchBar customization (macOS)",
		Long:  "Customize TouchBar buttons and actions (macOS only)",
	}

	cmd.AddCommand(newTouchBarSetCommand())
	cmd.AddCommand(newTouchBarClearCommand())

	return cmd
}

func newTouchBarSetCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "set <label> <command>",
		Short: "Set TouchBar button",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			label := args[0]
			command := args[1]

			wsURL, timeout, _ := cmdutil.GetFlags(cmd)

			ctx, cancel := cmdutil.CreateContext(timeout)
			defer cancel()

			c, err := cmdutil.ConnectClient(ctx, wsURL)
			if err != nil {
				return fmt.Errorf("failed to connect: %w", err)
			}
			defer c.Close()

			// Use InvokeFunction to set TouchBar button
			invocation := fmt.Sprintf(`
				tell application "iTerm2"
					set touchBarButton to make new touch bar button
					set label of touchBarButton to "%s"
					set action of touchBarButton to "%s"
				end tell
				return "TouchBar button set"
			`, label, command)

			request := &pb.InvokeFunctionRequest{
				Context: &pb.InvokeFunctionRequest_App{
					App: &pb.InvokeFunctionRequest_App{},
				},
				Invocation: invocation,
				Timeout:    float64(timeout.Seconds()),
			}

			response, err := c.InvokeFunction(ctx, request)
			if err != nil {
				return fmt.Errorf("failed to set TouchBar button: %w", err)
			}

			if response.GetError() != nil {
				return fmt.Errorf("TouchBar command failed: %s", response.GetError().GetErrorReason())
			}

			fmt.Printf("TouchBar button set: %s -> %s\n", label, command)
			return nil
		},
	}

	return cmd
}

func newTouchBarClearCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "clear",
		Short: "Clear TouchBar customizations",
		RunE: func(cmd *cobra.Command, args []string) error {
			wsURL, timeout, _ := cmdutil.GetFlags(cmd)

			ctx, cancel := cmdutil.CreateContext(timeout)
			defer cancel()

			c, err := cmdutil.ConnectClient(ctx, wsURL)
			if err != nil {
				return fmt.Errorf("failed to connect: %w", err)
			}
			defer c.Close()

			// Use InvokeFunction to clear TouchBar buttons
			invocation := `
				tell application "iTerm2"
					delete every touch bar button
				end tell
				return "TouchBar cleared"
			`

			request := &pb.InvokeFunctionRequest{
				Context: &pb.InvokeFunctionRequest_App{
					App: &pb.InvokeFunctionRequest_App{},
				},
				Invocation: invocation,
				Timeout:    float64(timeout.Seconds()),
			}

			response, err := c.InvokeFunction(ctx, request)
			if err != nil {
				return fmt.Errorf("failed to clear TouchBar: %w", err)
			}

			if response.GetError() != nil {
				return fmt.Errorf("TouchBar command failed: %s", response.GetError().GetErrorReason())
			}

			fmt.Println("TouchBar customizations cleared")
			return nil
		},
	}

	return cmd
}