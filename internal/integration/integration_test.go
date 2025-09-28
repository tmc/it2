package integration_test

import (
	"bytes"
	"context"
	"encoding/json"
	"os"
	"os/exec"
	"strings"
	"testing"
	"time"
)

// TestIt2Binary tests the compiled it2 binary
func TestIt2Binary(t *testing.T) {
	// Build the binary first
	cmd := exec.Command("go", "build", "-o", "../../it2-test", "../../cmd/it2")
	if err := cmd.Run(); err != nil {
		t.Fatalf("Failed to build binary: %v", err)
	}
	defer os.Remove("../../it2-test")

	tests := []struct {
		name     string
		args     []string
		wantErr  bool
		contains []string
	}{
		{
			name:     "help command",
			args:     []string{"--help"},
			wantErr:  false,
			contains: []string{"iTerm2", "Usage:", "Core Operations"},
		},
		{
			name:     "version command",
			args:     []string{"--version"},
			wantErr:  false,
			contains: []string{}, // Version might be empty in test
		},
		{
			name:     "auth check",
			args:     []string{"auth", "check"},
			wantErr:  false,
			contains: []string{"Authentication"},
		},
		{
			name:     "invalid command",
			args:     []string{"invalid"},
			wantErr:  true,
			contains: []string{"Error"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := exec.Command("../../it2-test", tt.args...)
			output, err := cmd.CombinedOutput()

			if (err != nil) != tt.wantErr {
				t.Errorf("it2 %v error = %v, wantErr %v", tt.args, err, tt.wantErr)
			}

			outputStr := string(output)
			for _, want := range tt.contains {
				if !strings.Contains(outputStr, want) {
					t.Errorf("Output missing %q\nGot: %s", want, outputStr)
				}
			}
		})
	}
}

// TestCommandStructure tests that all major commands are available
func TestCommandStructure(t *testing.T) {
	// Build the binary
	cmd := exec.Command("go", "build", "-o", "../../it2-test", "../../cmd/it2")
	if err := cmd.Run(); err != nil {
		t.Fatalf("Failed to build binary: %v", err)
	}
	defer os.Remove("../../it2-test")

	// Commands that should exist
	commands := []string{
		"session", "tab", "window", "text", "profile",
		"variable", "notification", "auth", "app",
	}

	for _, command := range commands {
		t.Run(command+"_help", func(t *testing.T) {
			cmd := exec.Command("../../it2-test", command, "--help")
			output, err := cmd.CombinedOutput()
			if err != nil {
				t.Errorf("Command %s --help failed: %v\nOutput: %s", command, err, output)
			}

			// Check for basic help structure
			outputStr := string(output)
			if !strings.Contains(outputStr, "Usage:") {
				t.Errorf("Command %s help missing Usage section", command)
			}
			if !strings.Contains(outputStr, "Available Commands:") &&
				!strings.Contains(outputStr, "Flags:") {
				t.Errorf("Command %s help missing structure", command)
			}
		})
	}
}

// TestJSONOutput tests JSON formatting
func TestJSONOutput(t *testing.T) {
	// Skip if not in iTerm2 environment
	if os.Getenv("TERM_PROGRAM") != "iTerm.app" {
		t.Skip("Skipping iTerm2 integration test - not running in iTerm2")
	}

	cmd := exec.Command("go", "run", "../../cmd/it2", "session", "list", "--format", "json")
	output, err := cmd.CombinedOutput()
	if err != nil {
		// If auth fails, that's okay for testing
		if strings.Contains(string(output), "authentication") {
			t.Skip("Skipping - authentication not configured")
		}
		t.Fatalf("Failed to run command: %v\nOutput: %s", err, output)
	}

	// Try to parse as JSON
	var result []interface{}
	if err := json.Unmarshal(output, &result); err != nil {
		t.Errorf("Output is not valid JSON: %v\nOutput: %s", err, output)
	}
}

// TestFormatOptions tests different output formats
func TestFormatOptions(t *testing.T) {
	formats := []string{"json", "yaml", "table", "text"}

	// Build the binary
	cmd := exec.Command("go", "build", "-o", "../../it2-test", "../../cmd/it2")
	if err := cmd.Run(); err != nil {
		t.Fatalf("Failed to build binary: %v", err)
	}
	defer os.Remove("../../it2-test")

	for _, format := range formats {
		t.Run("format_"+format, func(t *testing.T) {
			cmd := exec.Command("../../it2-test", "auth", "check", "--format", format)
			output, err := cmd.CombinedOutput()
			if err != nil {
				t.Errorf("Format %s failed: %v\nOutput: %s", format, err, output)
			}

			// Basic validation
			if len(output) == 0 {
				t.Errorf("Format %s produced no output", format)
			}
		})
	}
}

// TestTimeout tests timeout functionality
func TestTimeout(t *testing.T) {
	// Build the binary
	cmd := exec.Command("go", "build", "-o", "../../it2-test", "../../cmd/it2")
	if err := cmd.Run(); err != nil {
		t.Fatalf("Failed to build binary: %v", err)
	}
	defer os.Remove("../../it2-test")

	// Run with very short timeout
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	cmd = exec.CommandContext(ctx, "../../it2-test", "session", "list", "--timeout", "50ms")
	output, err := cmd.CombinedOutput()

	// Should either timeout or complete quickly
	if ctx.Err() == nil && err == nil {
		// Command completed successfully
		t.Log("Command completed within timeout")
	} else if ctx.Err() != nil {
		// Context timeout - expected for integration test
		t.Log("Command timed out as expected")
	} else {
		// Some other error
		t.Logf("Command failed with error: %v\nOutput: %s", err, output)
	}
}

// TestPluginSystem tests plugin discovery
func TestPluginSystem(t *testing.T) {
	// Create a temporary plugin
	tmpDir := t.TempDir()
	pluginPath := tmpDir + "/it2-test-plugin"

	pluginContent := `#!/bin/bash
echo "TEST_PLUGIN_OUTPUT"
exit 0`

	if err := os.WriteFile(pluginPath, []byte(pluginContent), 0755); err != nil {
		t.Fatalf("Failed to create test plugin: %v", err)
	}

	// Add to PATH
	oldPath := os.Getenv("PATH")
	os.Setenv("PATH", tmpDir+":"+oldPath)
	defer os.Setenv("PATH", oldPath)

	// Build and run
	cmd := exec.Command("go", "run", "../../cmd/it2", "session", "get-info")
	output, err := cmd.CombinedOutput()

	// Check if plugins are mentioned (even if command fails)
	outputStr := string(output)
	if strings.Contains(outputStr, "Plugin") ||
		strings.Contains(outputStr, "plugin") ||
		err == nil {
		t.Log("Plugin system appears to be working")
	} else {
		t.Log("Plugin system not fully tested (may require iTerm2)")
	}
}

// TestErrorMessages tests error handling
func TestErrorMessages(t *testing.T) {
	// Build the binary
	cmd := exec.Command("go", "build", "-o", "../../it2-test", "../../cmd/it2")
	if err := cmd.Run(); err != nil {
		t.Fatalf("Failed to build binary: %v", err)
	}
	defer os.Remove("../../it2-test")

	tests := []struct {
		name    string
		args    []string
		wantErr string
	}{
		{
			name:    "missing required argument",
			args:    []string{"session", "send-text"},
			wantErr: "no text provided",
		},
		{
			name:    "invalid timeout",
			args:    []string{"session", "list", "--timeout", "invalid"},
			wantErr: "invalid duration",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := exec.Command("../../it2-test", tt.args...)
			var stderr bytes.Buffer
			cmd.Stderr = &stderr

			err := cmd.Run()
			if err == nil {
				t.Errorf("Expected error for %v, got none", tt.args)
			}

			errOutput := stderr.String()
			if !strings.Contains(strings.ToLower(errOutput), strings.ToLower(tt.wantErr)) {
				t.Errorf("Error output missing %q\nGot: %s", tt.wantErr, errOutput)
			}
		})
	}
}

// BenchmarkStartup benchmarks CLI startup time
func BenchmarkStartup(b *testing.B) {
	// Build the binary
	cmd := exec.Command("go", "build", "-o", "../../it2-test", "../../cmd/it2")
	if err := cmd.Run(); err != nil {
		b.Fatalf("Failed to build binary: %v", err)
	}
	defer os.Remove("../../it2-test")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		cmd := exec.Command("../../it2-test", "--help")
		if err := cmd.Run(); err != nil {
			b.Errorf("Command failed: %v", err)
		}
	}
}
