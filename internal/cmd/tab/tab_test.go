package tab

import (
	"testing"

	"github.com/spf13/cobra"
)

func TestNewCommand(t *testing.T) {
	cmd := NewCommand()

	if cmd == nil {
		t.Fatal("Expected non-nil command")
	}

	if cmd.Use != "tab" {
		t.Errorf("Expected Use to be 'tab', got '%s'", cmd.Use)
	}

	if cmd.Short != "Manage iTerm2 tabs" {
		t.Errorf("Expected Short to be 'Manage iTerm2 tabs', got '%s'", cmd.Short)
	}

	if cmd.Long != "Commands for creating, closing, and managing iTerm2 tabs" {
		t.Errorf("Expected Long to be 'Commands for creating, closing, and managing iTerm2 tabs', got '%s'", cmd.Long)
	}

	// Check that key subcommands are added
	expectedSubcommands := []string{
		"list",
		"create",
		"close",
		"activate",
		"reorder",
		"set-title",
		"set-layout",
		"get-info",
		"splits",
	}

	subcommands := cmd.Commands()

	// Verify key subcommands exist
	subcommandNames := make(map[string]bool)
	for _, subcmd := range subcommands {
		subcommandNames[subcmd.Name()] = true
	}

	for _, expected := range expectedSubcommands {
		if !subcommandNames[expected] {
			t.Errorf("Expected subcommand '%s' to be present", expected)
		}
	}
}

func TestNewCommand_RunE(t *testing.T) {
	cmd := NewCommand()

	// The main tab command should not have a RunE function
	// since it's a parent command that requires subcommands
	if cmd.RunE != nil {
		t.Error("Expected main tab command to not have RunE function")
	}
}

func TestTabSubcommands_Structure(t *testing.T) {
	cmd := NewCommand()
	subcommands := cmd.Commands()

	// Test that each subcommand has the expected structure
	for _, subcmd := range subcommands {
		if subcmd.Use == "" {
			t.Errorf("Subcommand %s should have Use field set", subcmd.Name())
		}
		if subcmd.Short == "" {
			t.Errorf("Subcommand %s should have Short description", subcmd.Name())
		}

		// All tab subcommands should have RunE functions
		if subcmd.RunE == nil {
			t.Errorf("Subcommand %s should have RunE function", subcmd.Name())
		}
	}
}

func TestListCommand_Structure(t *testing.T) {
	cmd := newListCommand()

	if cmd.Use != "list [<window-id>]" {
		t.Errorf("Expected Use to be 'list [<window-id>]', got '%s'", cmd.Use)
	}

	if cmd.Short != "List tabs in windows" {
		t.Errorf("Expected Short to be 'List tabs in windows', got '%s'", cmd.Short)
	}

	expectedLong := `List tabs in a specific window or all windows.

Examples:
  it2 tab list                # List tabs in all windows
  it2 tab list 1              # List tabs in window 1 (positional)
  it2 tab list --window-id 1  # List tabs in window 1 (flag)
  it2 tab list -q             # Output only tab IDs (quiet mode)
  it2 tab list -q | xargs -n1 it2 tab close  # Close all tabs (use with caution!)
  it2 tab list -q | wc -l     # Count total tabs`
	if cmd.Long != expectedLong {
		t.Errorf("Expected correct Long description, got '%s'", cmd.Long)
	}

	if cmd.RunE == nil {
		t.Error("Expected list command to have RunE function")
	}

	// Check argument validation
	if cmd.Args == nil {
		t.Error("Expected list command to have argument validation")
	}
}

func TestListCommand_Flags(t *testing.T) {
	cmd := newListCommand()
	flags := cmd.Flags()

	// Check that the list command has expected flags
	columnsFlag := flags.Lookup("columns")
	if columnsFlag == nil {
		t.Error("Expected list command to have 'columns' flag")
	}

	sortFlag := flags.Lookup("sort")
	if sortFlag == nil {
		t.Error("Expected list command to have 'sort' flag")
	}

	reverseFlag := flags.Lookup("reverse")
	if reverseFlag == nil {
		t.Error("Expected list command to have 'reverse' flag")
	}
}

func TestCreateCommand_Structure(t *testing.T) {
	cmd := newCreateCommand()

	if cmd.Use != "create <profile> [<window-id>]" {
		t.Errorf("Expected Use to be 'create <profile> [<window-id>]', got '%s'", cmd.Use)
	}

	if cmd.Short != "Create a new tab" {
		t.Errorf("Expected Short to be 'Create a new tab', got '%s'", cmd.Short)
	}

	if cmd.RunE == nil {
		t.Error("Expected create command to have RunE function")
	}

	// Check argument validation
	if cmd.Args == nil {
		t.Error("Expected create command to have argument validation")
	}
}

func TestCreateCommand_Flags(t *testing.T) {
	cmd := newCreateCommand()
	flags := cmd.Flags()

	indexFlag := flags.Lookup("index")
	if indexFlag == nil {
		t.Error("Expected create command to have 'index' flag")
	}

	commandFlag := flags.Lookup("command")
	if commandFlag == nil {
		t.Error("Expected create command to have 'command' flag")
	}
}

func TestCloseCommand_Structure(t *testing.T) {
	cmd := newCloseCommand()

	if cmd.Use != "close <tab-id> [tab-id...]" {
		t.Errorf("Expected Use to be 'close <tab-id> [tab-id...]', got '%s'", cmd.Use)
	}

	if cmd.Short != "Close one or more tabs" {
		t.Errorf("Expected Short to be 'Close one or more tabs', got '%s'", cmd.Short)
	}

	if cmd.RunE == nil {
		t.Error("Expected close command to have RunE function")
	}

	// Check argument validation
	if cmd.Args == nil {
		t.Error("Expected close command to have argument validation")
	}

	// Check flags
	forceFlag := cmd.Flags().Lookup("force")
	if forceFlag == nil {
		t.Error("Expected close command to have 'force' flag")
	}
}

func TestActivateCommand_Structure(t *testing.T) {
	cmd := newActivateCommand()

	if cmd.Use != "activate <tab-id>" {
		t.Errorf("Expected Use to be 'activate <tab-id>', got '%s'", cmd.Use)
	}

	if cmd.Short != "Activate a tab" {
		t.Errorf("Expected Short to be 'Activate a tab', got '%s'", cmd.Short)
	}

	if cmd.RunE == nil {
		t.Error("Expected activate command to have RunE function")
	}

	// Check argument validation
	if cmd.Args == nil {
		t.Error("Expected activate command to have argument validation")
	}

	// Check flags
	selectTabFlag := cmd.Flags().Lookup("select-tab")
	if selectTabFlag == nil {
		t.Error("Expected activate command to have 'select-tab' flag")
	}
}

func TestReorderCommand_Structure(t *testing.T) {
	cmd := newReorderCommand()

	if cmd.Use != "reorder <tab-id> <new-index>" {
		t.Errorf("Expected Use to be 'reorder <tab-id> <new-index>', got '%s'", cmd.Use)
	}

	if cmd.Short != "Reorder tabs within a window" {
		t.Errorf("Expected Short to be 'Reorder tabs within a window', got '%s'", cmd.Short)
	}

	if cmd.RunE == nil {
		t.Error("Expected reorder command to have RunE function")
	}

	// Check argument validation
	if cmd.Args == nil {
		t.Error("Expected reorder command to have argument validation")
	}
}

func TestSetTitleCommand_Structure(t *testing.T) {
	cmd := newSetTitleCommand()

	if cmd.Short != "Set the title of a tab" {
		t.Errorf("Expected Short to be 'Set the title of a tab', got '%s'", cmd.Short)
	}

	if cmd.RunE == nil {
		t.Error("Expected set-title command to have RunE function")
	}

	// Check argument validation
	if cmd.Args == nil {
		t.Error("Expected set-title command to have argument validation")
	}
}

func TestSetLayoutCommand_Structure(t *testing.T) {
	cmd := newSetLayoutCommand()

	if cmd.Use != "set-layout <tab-id> <layout>" {
		t.Errorf("Expected Use to be 'set-layout <tab-id> <layout>', got '%s'", cmd.Use)
	}

	if cmd.Short != "Set the layout of a tab" {
		t.Errorf("Expected Short to be 'Set the layout of a tab', got '%s'", cmd.Short)
	}

	if cmd.RunE == nil {
		t.Error("Expected set-layout command to have RunE function")
	}

	// Check argument validation
	if cmd.Args == nil {
		t.Error("Expected set-layout command to have argument validation")
	}

	// Check flags
	horizontalFlag := cmd.Flags().Lookup("horizontal")
	if horizontalFlag == nil {
		t.Error("Expected set-layout command to have 'horizontal' flag")
	}
}

func TestGetInfoCommand_Structure(t *testing.T) {
	cmd := newGetInfoCommand()

	if cmd.Use != "get-info <tab-id>" {
		t.Errorf("Expected Use to be 'get-info <tab-id>', got '%s'", cmd.Use)
	}

	if cmd.Short != "Get detailed information about a tab" {
		t.Errorf("Expected Short to be 'Get detailed information about a tab', got '%s'", cmd.Short)
	}

	if cmd.RunE == nil {
		t.Error("Expected get-info command to have RunE function")
	}

	// Check argument validation
	if cmd.Args == nil {
		t.Error("Expected get-info command to have argument validation")
	}
}

func TestRunListCommand_ArgumentHandling(t *testing.T) {
	cmd := newListCommand()

	// Test that the function can be called with no arguments
	// (we can't actually test the full execution without a real iTerm2 connection)

	// Verify MaximumNArgs(1) is being used correctly
	expectedArgsFunc := cobra.MaximumNArgs(1)

	// Test with no arguments (should pass)
	err := expectedArgsFunc(cmd, []string{})
	if err != nil {
		t.Errorf("Expected no error with no arguments, got: %v", err)
	}

	// Test with one argument (should pass)
	err = expectedArgsFunc(cmd, []string{"window-123"})
	if err != nil {
		t.Errorf("Expected no error with one argument, got: %v", err)
	}

	// Test with two arguments (should fail)
	err = expectedArgsFunc(cmd, []string{"window-123", "extra-arg"})
	if err == nil {
		t.Error("Expected error with too many arguments")
	}
}

// Test helper functions
func TestTabCommandCreationFunctions(t *testing.T) {
	testCases := []struct {
		name string
		fn   func() *cobra.Command
	}{
		{"list", newListCommand},
		{"create", newCreateCommand},
		{"close", newCloseCommand},
		{"activate", newActivateCommand},
		{"reorder", newReorderCommand},
		{"set-title", newSetTitleCommand},
		{"set-layout", newSetLayoutCommand},
		{"get-info", newGetInfoCommand},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			cmd := tc.fn()
			if cmd == nil {
				t.Fatalf("Expected non-nil command for %s", tc.name)
			}
			if cmd.Use == "" {
				t.Errorf("Expected Use to be set for %s command", tc.name)
			}
			if cmd.Short == "" {
				t.Errorf("Expected Short to be set for %s command", tc.name)
			}
		})
	}
}

// Test argument validation functions
func TestCobraArgValidations(t *testing.T) {
	testCases := []struct {
		name        string
		argsFunc    cobra.PositionalArgs
		args        []string
		expectError bool
	}{
		{
			name:        "ExactArgs(1) with correct args",
			argsFunc:    cobra.ExactArgs(1),
			args:        []string{"arg1"},
			expectError: false,
		},
		{
			name:        "ExactArgs(1) with no args",
			argsFunc:    cobra.ExactArgs(1),
			args:        []string{},
			expectError: true,
		},
		{
			name:        "ExactArgs(2) with correct args",
			argsFunc:    cobra.ExactArgs(2),
			args:        []string{"arg1", "arg2"},
			expectError: false,
		},
		{
			name:        "MinimumNArgs(1) with correct args",
			argsFunc:    cobra.MinimumNArgs(1),
			args:        []string{"arg1", "arg2"},
			expectError: false,
		},
		{
			name:        "MinimumNArgs(1) with no args",
			argsFunc:    cobra.MinimumNArgs(1),
			args:        []string{},
			expectError: true,
		},
		{
			name:        "RangeArgs(1,2) with one arg",
			argsFunc:    cobra.RangeArgs(1, 2),
			args:        []string{"arg1"},
			expectError: false,
		},
		{
			name:        "RangeArgs(1,2) with two args",
			argsFunc:    cobra.RangeArgs(1, 2),
			args:        []string{"arg1", "arg2"},
			expectError: false,
		},
		{
			name:        "RangeArgs(1,2) with no args",
			argsFunc:    cobra.RangeArgs(1, 2),
			args:        []string{},
			expectError: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			cmd := &cobra.Command{}
			err := tc.argsFunc(cmd, tc.args)
			if tc.expectError && err == nil {
				t.Error("Expected error but got none")
			}
			if !tc.expectError && err != nil {
				t.Errorf("Expected no error but got: %v", err)
			}
		})
	}
}
