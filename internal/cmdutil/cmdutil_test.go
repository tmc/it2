package cmdutil

import (
	"os"
	"reflect"
	"testing"
	"time"

	"github.com/spf13/cobra"
)

func TestGetFlags_Defaults(t *testing.T) {
	cmd := &cobra.Command{}
	cmd.Flags().String("url", "", "")
	cmd.Flags().Duration("timeout", 0, "")
	cmd.Flags().String("format", "", "")

	wsURL, timeout, format := GetFlags(cmd)

	if wsURL != "ws://localhost:1912" {
		t.Errorf("Expected default wsURL to be 'ws://localhost:1912', got '%s'", wsURL)
	}
	if timeout != 5*time.Second {
		t.Errorf("Expected default timeout to be 5s, got %v", timeout)
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

	wsURL, timeout, format := GetFlags(cmd)

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

	wsURL, timeout, format, columns, sortBy, sortReverse := GetExtendedFlags(cmd)

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
	if timeout != 5*time.Second {
		t.Errorf("Expected default timeout to be 5s, got %v", timeout)
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

	_, _, _, columns, _, _ := GetExtendedFlags(cmd)

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

func TestResolveSessionID(t *testing.T) {
	// Test with provided session ID
	result := ResolveSessionID("w0t1p12:ABC123")
	if result != "ABC123" {
		t.Errorf("Expected 'ABC123', got '%s'", result)
	}

	// Test with empty session ID and environment variable set
	os.Setenv("ITERM_SESSION_ID", "w0t2p34:DEF456")
	defer os.Unsetenv("ITERM_SESSION_ID")

	result = ResolveSessionID("")
	if result != "DEF456" {
		t.Errorf("Expected 'DEF456' from environment, got '%s'", result)
	}

	// Provided session ID should override environment
	result = ResolveSessionID("w0t3p56:GHI789")
	if result != "GHI789" {
		t.Errorf("Expected 'GHI789' from parameter, got '%s'", result)
	}
}

func TestResolveSessionID_NoEnvironment(t *testing.T) {
	os.Unsetenv("ITERM_SESSION_ID")
	result := ResolveSessionID("")
	if result != "" {
		t.Errorf("Expected empty string when no session ID available, got '%s'", result)
	}
}

func TestResolveSessionIDWithError(t *testing.T) {
	// Test with valid session ID
	result, err := ResolveSessionIDWithError("w0t1p12:ABC123")
	if err != nil {
		t.Errorf("Expected no error with valid session ID, got: %v", err)
	}
	if result != "ABC123" {
		t.Errorf("Expected 'ABC123', got '%s'", result)
	}

	// Test with no session ID and no environment
	os.Unsetenv("ITERM_SESSION_ID")
	result, err = ResolveSessionIDWithError("")
	if err == nil {
		t.Error("Expected error when no session ID available")
	}
	if result != "" {
		t.Errorf("Expected empty string with error, got '%s'", result)
	}

	// Check error type
	_, ok := err.(*NoSessionIDError)
	if !ok {
		t.Errorf("Expected NoSessionIDError, got %T", err)
	}
}

func TestNoSessionIDError(t *testing.T) {
	err := &NoSessionIDError{}
	expected := "no session ID provided and $ITERM_SESSION_ID environment variable not set"
	if err.Error() != expected {
		t.Errorf("Expected error message '%s', got '%s'", expected, err.Error())
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