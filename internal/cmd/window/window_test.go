package window

import (
	"testing"

	"github.com/spf13/cobra"
)

func TestNewCommand(t *testing.T) {
	cmd := NewCommand()

	if cmd == nil {
		t.Fatal("Expected non-nil command")
	}

	if cmd.Use != "window [window-id] [command]" {
		t.Errorf("Expected Use to be 'window [window-id] [command]', got '%s'", cmd.Use)
	}

	if cmd.Short != "Manage iTerm2 windows or navigate to window context" {
		t.Errorf("Expected Short to be 'Manage iTerm2 windows or navigate to window context', got '%s'", cmd.Short)
	}

	expectedLong := `Manage iTerm2 windows with support for both flat and hierarchical command styles.

Flat Command Examples:
  it2 window list                   # List all windows
  it2 window create                 # Create new window
  it2 window close 1                # Close window 1

Hierarchical Command Examples:
  it2 window 1 tab list            # List tabs in window 1
  it2 window 1 session list        # List sessions in window 1
  it2 window 1 tab 2 session list  # List sessions in tab 2 of window 1

The hierarchical style provides context-aware navigation while maintaining
full compatibility with existing flat commands.`
	if cmd.Long != expectedLong {
		t.Errorf("Expected Long to be '%s', got '%s'", expectedLong, cmd.Long)
	}

	// Check that subcommands are added
	expectedSubcommands := []string{
		"list",
		"create",
		"close",
		"focus",
		"get-property",
		"set-property",
	}

	subcommands := cmd.Commands()
	if len(subcommands) != len(expectedSubcommands) {
		t.Errorf("Expected %d subcommands, got %d", len(expectedSubcommands), len(subcommands))
	}

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

	// The main window command should have a RunE function for hierarchical support
	if cmd.RunE == nil {
		t.Error("Expected main window command to have RunE function for hierarchical support")
	}
}

func TestWindowSubcommands_Structure(t *testing.T) {
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

		// Commands that should have RunE functions
		commandsWithRunE := map[string]bool{
			"list":         true,
			"create":       true,
			"close":        true,
			"focus":        true,
			"get-property": true,
			"set-property": true,
		}

		if commandsWithRunE[subcmd.Name()] && subcmd.RunE == nil {
			t.Errorf("Subcommand %s should have RunE function", subcmd.Name())
		}
	}
}

// Test helper functions exist
func TestWindowCommandCreationFunctions(t *testing.T) {
	testCases := []struct {
		name string
		fn   func() *cobra.Command
	}{
		{"list", newListCommand},
		{"create", newCreateCommand},
		{"close", newCloseCommand},
		{"focus", newFocusCommand},
		{"get-property", newGetPropertyCommand},
		{"set-property", newSetPropertyCommand},
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

func TestListCommand_Structure(t *testing.T) {
	cmd := newListCommand()

	if cmd.Use != "list" {
		t.Errorf("Expected Use to be 'list', got '%s'", cmd.Use)
	}

	if cmd.RunE == nil {
		t.Error("Expected list command to have RunE function")
	}
}

func TestCreateCommand_Structure(t *testing.T) {
	cmd := newCreateCommand()

	if cmd.Use != "create [profile]" {
		t.Errorf("Expected Use to be 'create [profile]', got '%s'", cmd.Use)
	}

	if cmd.RunE == nil {
		t.Error("Expected create command to have RunE function")
	}

	// Check argument requirements
	if cmd.Args == nil {
		t.Error("Expected create command to have argument validation")
	}
}

func TestCloseCommand_Structure(t *testing.T) {
	cmd := newCloseCommand()

	if cmd.RunE == nil {
		t.Error("Expected close command to have RunE function")
	}

	// Check for force flag
	forceFlag := cmd.Flags().Lookup("force")
	if forceFlag == nil {
		t.Error("Expected close command to have 'force' flag")
	}
}

func TestFocusCommand_Structure(t *testing.T) {
	cmd := newFocusCommand()

	if cmd.RunE == nil {
		t.Error("Expected focus command to have RunE function")
	}

	// Check for order-front flag
	orderFrontFlag := cmd.Flags().Lookup("order-front")
	if orderFrontFlag == nil {
		t.Error("Expected focus command to have 'order-front' flag")
	}
}

func TestGetPropertyCommand_Structure(t *testing.T) {
	cmd := newGetPropertyCommand()

	if cmd.RunE == nil {
		t.Error("Expected get-property command to have RunE function")
	}

	// Check argument requirements
	if cmd.Args == nil {
		t.Error("Expected get-property command to have argument validation")
	}
}

func TestSetPropertyCommand_Structure(t *testing.T) {
	cmd := newSetPropertyCommand()

	if cmd.RunE == nil {
		t.Error("Expected set-property command to have RunE function")
	}

	// Check argument requirements
	if cmd.Args == nil {
		t.Error("Expected set-property command to have argument validation")
	}
}
