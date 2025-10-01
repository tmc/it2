package settings

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/spf13/cobra"
	"github.com/tmc/it2/internal/auth"
	"github.com/tmc/it2/internal/iterm2settings"
)

// NewCommand creates the settings command with all subcommands.
func NewCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "settings",
		Short: "Manage iTerm2 application settings",
		Long:  "Configure various iTerm2 settings including API automation and key bindings",
	}

	cmd.AddCommand(newAPICommand())
	cmd.AddCommand(newKeybindingsCommand())

	return cmd
}

func newAPICommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "api",
		Short: "Manage iTerm2 API automation",
		Long:  "Check status, enable, or disable iTerm2 API automation",
	}

	cmd.AddCommand(newAPIEnabledCommand())
	cmd.AddCommand(newAPIEnableCommand())
	cmd.AddCommand(newAPIDisableCommand())

	return cmd
}

func newAPIEnabledCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "enabled",
		Short: "Check if iTerm2 API automation is enabled",
		Long:  "Check the status of iTerm2 API automation (EnableAPIServer preference)",
		Run: func(cmd *cobra.Command, args []string) {
			if iterm2settings.IsAPIEnabled() {
				fmt.Println("✓ iTerm2 API automation is ENABLED")
			} else {
				fmt.Println("❌ iTerm2 API automation is DISABLED")
				fmt.Println()
				fmt.Println("To enable it, run: it2 app settings api enable")
				os.Exit(1)
			}
		},
	}
}

func newAPIEnableCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "enable",
		Short: "Enable iTerm2 API automation",
		Long:  "Enable iTerm2 API automation by setting the EnableAPIServer preference to true",
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := auth.EnableAutomation(); err != nil {
				return fmt.Errorf("failed to enable API: %w", err)
			}

			fmt.Println("✓ iTerm2 API automation enabled")
			return nil
		},
	}
}

func newAPIDisableCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "disable",
		Short: "Disable iTerm2 API automation",
		Long:  "Disable iTerm2 API automation by setting the EnableAPIServer preference to false",
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := auth.DisableAutomation(); err != nil {
				return fmt.Errorf("failed to disable API: %w", err)
			}

			fmt.Println("✓ iTerm2 API automation disabled")
			return nil
		},
	}
}

func newKeybindingsCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "keybindings",
		Short: "Manage iTerm2 key bindings",
		Long:  "Configure iTerm2 global key bindings",
	}

	cmd.AddCommand(newShiftEnterCommand())

	return cmd
}

func newShiftEnterCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "shift-enter",
		Short: "Manage Shift+Enter key binding",
		Long:  "Enable or disable Shift+Enter sending newline (useful for Claude Code)",
	}

	cmd.AddCommand(newShiftEnterEnableCommand())
	cmd.AddCommand(newShiftEnterDisableCommand())
	cmd.AddCommand(newShiftEnterStatusCommand())

	return cmd
}

func newShiftEnterStatusCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "status",
		Short: "Check if Shift+Enter is configured",
		Run: func(cmd *cobra.Command, args []string) {
			if hasShiftEnterBinding() {
				fmt.Println("✓ Shift+Enter → newline is ENABLED")
			} else {
				fmt.Println("❌ Shift+Enter → newline is DISABLED")
				fmt.Println()
				fmt.Println("To enable it, run: it2 app settings keybindings shift-enter enable")
			}
		},
	}
}

func newShiftEnterEnableCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "enable",
		Short: "Enable Shift+Enter → newline binding",
		Long:  "Configure Shift+Enter to send a newline character (useful for multi-line input in Claude Code)",
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := setShiftEnterBinding(true); err != nil {
				return fmt.Errorf("failed to enable Shift+Enter binding: %w", err)
			}

			fmt.Println("✓ Shift+Enter → newline enabled")
			fmt.Println()
			fmt.Println("The binding is active immediately in new iTerm2 sessions.")
			return nil
		},
	}
}

func newShiftEnterDisableCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "disable",
		Short: "Disable Shift+Enter → newline binding",
		Long:  "Remove the Shift+Enter → newline key binding",
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := setShiftEnterBinding(false); err != nil {
				return fmt.Errorf("failed to disable Shift+Enter binding: %w", err)
			}

			fmt.Println("✓ Shift+Enter → newline disabled")
			return nil
		},
	}
}

// hasShiftEnterBinding checks if the Shift+Enter binding exists
func hasShiftEnterBinding() bool {
	cmd := exec.Command("defaults", "read", "com.googlecode.iterm2", "GlobalKeyMap")
	output, err := cmd.Output()
	if err != nil {
		return false
	}

	// Check if the key binding exists in the output
	// The key is "0xd-0x20000-0x24" for Shift+Enter
	return contains(string(output), "0xd-0x20000-0x24")
}

// setShiftEnterBinding enables or disables the Shift+Enter binding
func setShiftEnterBinding(enable bool) error {
	if enable {
		// Add the key binding using defaults write with dict-add
		// Key: 0xd-0x20000-0x24 (Shift+Enter)
		// Action: 12 (Send Text)
		// Text: \n (newline)
		script := `defaults write com.googlecode.iterm2 GlobalKeyMap -dict-add "0xd-0x20000-0x24" '{
			Action = 12;
			Keycode = 13;
			Modifiers = 131072;
			Text = "\\n";
			Version = 1;
		}'`
		cmd := exec.Command("sh", "-c", script)
		if output, err := cmd.CombinedOutput(); err != nil {
			return fmt.Errorf("failed to add key binding: %w\nOutput: %s", err, string(output))
		}
	} else {
		// Use plutil to remove the key from the GlobalKeyMap dict
		script := `/usr/libexec/PlistBuddy -c "Delete :GlobalKeyMap:0xd-0x20000-0x24" ~/Library/Preferences/com.googlecode.iterm2.plist 2>/dev/null || true`
		cmd := exec.Command("sh", "-c", script)
		_ = cmd.Run()
	}

	return nil
}

func contains(s, substr string) bool {
	return strings.Contains(s, substr)
}
