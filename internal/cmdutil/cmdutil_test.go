package cmdutil

import (
	"reflect"
	"testing"
	"time"

	"github.com/spf13/cobra"
	"github.com/tmc/it2/internal/cmdcore"
)

func TestGetFlags_Defaults(t *testing.T) {
	cmd := &cobra.Command{}
	cmd.Flags().String("url", "", "")
	cmd.Flags().Duration("timeout", 0, "")
	cmd.Flags().String("format", "", "")

	wsURL, timeout, format := cmdcore.GetFlags(cmd)

	if wsURL != "ws://localhost:1912" {
		t.Errorf("Expected default wsURL to be 'ws://localhost:1912', got '%s'", wsURL)
	}
	if timeout != 60*time.Second {
		t.Errorf("Expected default timeout to be 60s, got %v", timeout)
	}
	if format != "table" {
		t.Errorf("Expected default format to be 'table', got '%s'", format)
	}
}

func TestGetFlags_CustomValues(t *testing.T) {
	cmd := &cobra.Command{}
	cmd.Flags().String("url", "ws://example.com:1234", "")
	cmd.Flags().Duration("timeout", 10*time.Second, "")
	cmd.Flags().String("format", "json", "")

	wsURL, timeout, format := cmdcore.GetFlags(cmd)

	if wsURL != "ws://example.com:1234" {
		t.Errorf("Expected wsURL to be 'ws://example.com:1234', got '%s'", wsURL)
	}
	if timeout != 10*time.Second {
		t.Errorf("Expected timeout to be 10s, got %v", timeout)
	}
	if format != "json" {
		t.Errorf("Expected format to be 'json', got '%s'", format)
	}
}

func TestGetExtendedFlags_WithColumns(t *testing.T) {
	cmd := &cobra.Command{}
	cmd.Flags().String("url", "", "")
	cmd.Flags().Duration("timeout", 0, "")
	cmd.Flags().String("format", "", "")
	cmd.Flags().String("columns", "session id,window id,name", "")
	cmd.Flags().String("sort", "Session ID", "")
	cmd.Flags().Bool("reverse", true, "")

	wsURL, timeout, format, columns, sortBy, sortReverse := cmdcore.GetExtendedFlags(cmd)

	expectedColumns := []string{"session id", "window id", "name"}
	if !reflect.DeepEqual(columns, expectedColumns) {
		t.Errorf("Expected columns to be %v, got %v", expectedColumns, columns)
	}
	if sortBy != "Session ID" {
		t.Errorf("Expected sortBy to be 'Session ID', got '%s'", sortBy)
	}
	if !sortReverse {
		t.Errorf("Expected sortReverse to be true, got %v", sortReverse)
	}
	// Also test base flags
	if wsURL != "ws://localhost:1912" {
		t.Errorf("Expected default wsURL to be 'ws://localhost:1912', got '%s'", wsURL)
	}
	if timeout != 60*time.Second {
		t.Errorf("Expected default timeout to be 60s, got %v", timeout)
	}
	if format != "table" {
		t.Errorf("Expected default format to be 'table', got '%s'", format)
	}
}

func TestGetExtendedFlags_WithSpacedColumns(t *testing.T) {
	cmd := &cobra.Command{}
	cmd.Flags().String("url", "", "")
	cmd.Flags().Duration("timeout", 0, "")
	cmd.Flags().String("format", "", "")
	cmd.Flags().String("columns", " session id , window id , name ", "")

	_, _, _, columns, _, _ := cmdcore.GetExtendedFlags(cmd)

	expectedColumns := []string{"session id", "window id", "name"}
	if !reflect.DeepEqual(columns, expectedColumns) {
		t.Errorf("Expected trimmed columns to be %v, got %v", expectedColumns, columns)
	}
}

func TestBoolPtr(t *testing.T) {
	result := BoolPtr(true)
	if result == nil {
		t.Fatal("Expected non-nil pointer")
	}
	if *result != true {
		t.Errorf("Expected dereferenced value to be true, got %v", *result)
	}

	result = BoolPtr(false)
	if result == nil {
		t.Fatal("Expected non-nil pointer")
	}
	if *result != false {
		t.Errorf("Expected dereferenced value to be false, got %v", *result)
	}
}

func TestStringPtr(t *testing.T) {
	result := StringPtr("test")
	if result == nil {
		t.Fatal("Expected non-nil pointer")
	}
	if *result != "test" {
		t.Errorf("Expected dereferenced value to be 'test', got '%s'", *result)
	}

	result = StringPtr("")
	if result == nil {
		t.Fatal("Expected non-nil pointer")
	}
	if *result != "" {
		t.Errorf("Expected dereferenced value to be '', got '%s'", *result)
	}
}

func TestInt32Ptr(t *testing.T) {
	result := Int32Ptr(42)
	if result == nil {
		t.Fatal("Expected non-nil pointer")
	}
	if *result != 42 {
		t.Errorf("Expected dereferenced value to be 42, got %d", *result)
	}

	result = Int32Ptr(-1)
	if result == nil {
		t.Fatal("Expected non-nil pointer")
	}
	if *result != -1 {
		t.Errorf("Expected dereferenced value to be -1, got %d", *result)
	}
}

func TestNormalizeSessionID(t *testing.T) {
	testCases := []struct {
		input    string
		expected string
	}{
		{"w0t1p12:ABC123-DEF456-GHI789", "ABC123-DEF456-GHI789"},
		{"ABC123-DEF456-GHI789", "ABC123-DEF456-GHI789"},
		{"simple", "simple"},
		{"", ""},
		{"w0t1p12:", ""},
		{"a:b:c:d", "d"},
		{"prefix:suffix", "suffix"},
	}

	for _, tc := range testCases {
		t.Run(tc.input, func(t *testing.T) {
			result := NormalizeSessionID(tc.input)
			if result != tc.expected {
				t.Errorf("Expected '%s', got '%s'", tc.expected, result)
			}
		})
	}
}

func TestIsSessionCommand(t *testing.T) {
	testCases := []struct {
		name     string
		cmdName  string
		expected bool
	}{
		{"session command", "session", true},
		{"session-list command", "session-list", true},
		{"list command", "list", false},
		{"window command", "window", false},
		{"tab command", "tab", false},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			cmd := &cobra.Command{Use: tc.cmdName}
			result := IsSessionCommand(cmd)
			if result != tc.expected {
				t.Errorf("Expected %v for command '%s', got %v", tc.expected, tc.cmdName, result)
			}
		})
	}
}

func TestIsSessionCommand_WithParent(t *testing.T) {
	rootCmd := &cobra.Command{Use: "it2"}
	sessionCmd := &cobra.Command{Use: "session"}
	listCmd := &cobra.Command{Use: "list"}

	rootCmd.AddCommand(sessionCmd)
	sessionCmd.AddCommand(listCmd)

	// The list command under session should be considered a session command
	if !IsSessionCommand(listCmd) {
		t.Error("Expected list command under session to be considered a session command")
	}

	// Create a different hierarchy
	windowCmd := &cobra.Command{Use: "window"}
	windowListCmd := &cobra.Command{Use: "list"}
	rootCmd.AddCommand(windowCmd)
	windowCmd.AddCommand(windowListCmd)

	// The list command under window should not be considered a session command
	if IsSessionCommand(windowListCmd) {
		t.Error("Expected list command under window to not be considered a session command")
	}
}
