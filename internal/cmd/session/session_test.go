package session

import (
	"testing"

	"github.com/spf13/cobra"
)

func TestNewCommand(t *testing.T) {
	cmd := NewCommand()

	if cmd == nil {
		t.Fatal("Expected non-nil command")
	}

	if cmd.Use != "session" {
		t.Errorf("Expected Use to be 'session', got '%s'", cmd.Use)
	}

	if cmd.Short != "Manage iTerm2 sessions" {
		t.Errorf("Expected Short to be 'Manage iTerm2 sessions', got '%s'", cmd.Short)
	}

	if cmd.Long != "Commands for creating, listing, and managing iTerm2 sessions" {
		t.Errorf("Expected Long to be 'Commands for creating, listing, and managing iTerm2 sessions', got '%s'", cmd.Long)
	}

	// Check that subcommands are added
	expectedSubcommands := []string{
		"list",
		"create",
		"close",
		"activate",
		"send-text",
		"send-key",
		"restart",
		"split",
		"get-pid",
		"get-property",
		"set-property",
		"monitor",
		"get-info",
		"autorespond",
		"watch",
	}

	subcommands := cmd.Commands()
	if len(subcommands) < len(expectedSubcommands) {
		t.Errorf("Expected at least %d subcommands, got %d", len(expectedSubcommands), len(subcommands))
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

	// The main session command should not have a RunE function
	// since it's a parent command that requires subcommands
	if cmd.RunE != nil {
		t.Error("Expected main session command to not have RunE function")
	}
}

func TestSessionSubcommands_Structure(t *testing.T) {
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
			"activate":     true,
			"send-text":    true,
			"send-key":     true,
			"restart":      true,
			"split":        true,
			"get-pid":      true,
			"get-property": true,
			"set-property": true,
			"monitor":      true,
			"get-info":     true,
			"autorespond":  true,
			"watch":        true,
		}

		if commandsWithRunE[subcmd.Name()] && subcmd.RunE == nil {
			t.Errorf("Subcommand %s should have RunE function", subcmd.Name())
		}
	}
}

func TestListCommand_Flags(t *testing.T) {
	listCmd := newListCommand()

	// Check that the list command has expected flags
	flags := listCmd.Flags()

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

func TestListCommand_Structure(t *testing.T) {
	cmd := newListCommand()

	if cmd.Use != "list" {
		t.Errorf("Expected Use to be 'list', got '%s'", cmd.Use)
	}

	if cmd.Short != "List all iTerm2 sessions" {
		t.Errorf("Expected Short to be 'List all iTerm2 sessions', got '%s'", cmd.Short)
	}

	if cmd.RunE == nil {
		t.Error("Expected list command to have RunE function")
	}
}

// Test helper functions
func TestCommandCreationFunctions(t *testing.T) {
	testCases := []struct {
		name string
		fn   func() *cobra.Command
	}{
		{"list", newListCommand},
		{"create", newCreateCommand},
		{"close", newCloseCommand},
		{"activate", newActivateCommand},
		{"send-text", newSendTextCommand},
		{"send-key", newSendKeyCommand},
		{"restart", newRestartCommand},
		{"split", newSplitCommand},
		{"get-pid", newGetPIDCommand},
		{"get-property", newGetPropertyCommand},
		{"set-property", newSetPropertyCommand},
		{"monitor", newMonitorCommand},
		{"get-info", newGetInfoCommand},
		{"autorespond", newAutoRespondCommand},
		{"watch", newWatchCommand},
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