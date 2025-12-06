// Package plugins provides the plugin manifest schema and verification system.
package plugins

import (
	"time"
)

// Manifest describes a plugin's metadata, source, permissions, and verification info.
// This is the source of truth for what a plugin needs and how to verify it.
type Manifest struct {
	// Plugin metadata
	Name        string `json:"name" yaml:"name"`
	Version     string `json:"version" yaml:"version"`
	Description string `json:"description" yaml:"description"`
	Author      string `json:"author" yaml:"author"`

	// Source information for reproducible builds
	Source Source `json:"source" yaml:"source"`

	// Binary verification
	Binary Binary `json:"binary" yaml:"binary"`

	// Input requirements (what data the plugin needs)
	Inputs Inputs `json:"inputs" yaml:"inputs"`

	// Permissions (what the plugin can do)
	Permissions Permissions `json:"permissions" yaml:"permissions"`

	// Sandbox configuration
	Sandbox SandboxConfig `json:"sandbox" yaml:"sandbox"`
}

// Source describes where to find the source code for reproducible builds.
type Source struct {
	// Repository URL (e.g., "github.com/tmc/it2-session-claude-has-modal")
	Repository string `json:"repository" yaml:"repository"`

	// Git commit hash (full SHA)
	Commit string `json:"commit" yaml:"commit"`

	// Git tag (optional, for human readability)
	Tag string `json:"tag,omitempty" yaml:"tag,omitempty"`

	// Build configuration for reproducible builds
	Build BuildConfig `json:"build" yaml:"build"`
}

// BuildConfig specifies exactly how to build the plugin reproducibly.
type BuildConfig struct {
	// Go version (e.g., "1.21.5")
	GoVersion string `json:"go_version" yaml:"go_version"`

	// Build flags (e.g., ["-trimpath", "-ldflags=-s -w"])
	Flags []string `json:"flags" yaml:"flags"`

	// Environment variables (e.g., {"CGO_ENABLED": "0", "SOURCE_DATE_EPOCH": "1"})
	Env map[string]string `json:"env" yaml:"env"`

	// Main package path (e.g., "./cmd/plugin")
	MainPackage string `json:"main_package,omitempty" yaml:"main_package,omitempty"`
}

// Binary describes the compiled binary and its verification info.
type Binary struct {
	// SHA256 hash of the binary
	SHA256 string `json:"sha256" yaml:"sha256"`

	// Download URL for pre-built binary (optional)
	URL string `json:"url,omitempty" yaml:"url,omitempty"`

	// Code signature info (macOS)
	CodeSignature *CodeSignature `json:"code_signature,omitempty" yaml:"code_signature,omitempty"`
}

// CodeSignature describes expected code signature (macOS only).
type CodeSignature struct {
	// Developer ID (e.g., "Developer ID: Travis Cline (XXXXXXXXXX)")
	DeveloperID string `json:"developer_id" yaml:"developer_id"`

	// Team ID (e.g., "XXXXXXXXXX")
	TeamID string `json:"team_id,omitempty" yaml:"team_id,omitempty"`
}

// Inputs declares what data the plugin needs to function.
// This implements input minimization - plugins only get what they declare.
type Inputs struct {
	// Session ID (always provided)
	SessionID bool `json:"session_id" yaml:"session_id"`

	// Screen lines (visible terminal content)
	ScreenLines *ScreenLinesInput `json:"screen_lines,omitempty" yaml:"screen_lines,omitempty"`

	// Buffer lines (scrollback history)
	BufferLines *BufferLinesInput `json:"buffer_lines,omitempty" yaml:"buffer_lines,omitempty"`

	// Session metadata (title, working directory, etc.)
	SessionMetadata []string `json:"session_metadata,omitempty" yaml:"session_metadata,omitempty"`

	// Environment variables (specific vars only, not full env)
	EnvironmentVars []string `json:"environment_vars,omitempty" yaml:"environment_vars,omitempty"`
}

// ScreenLinesInput specifies requirements for screen content.
type ScreenLinesInput struct {
	// Number of lines needed (e.g., 50)
	Lines int `json:"lines" yaml:"lines"`

	// Whether to include escape sequences
	IncludeEscapes bool `json:"include_escapes,omitempty" yaml:"include_escapes,omitempty"`
}

// BufferLinesInput specifies requirements for scrollback buffer.
type BufferLinesInput struct {
	// Number of lines needed (e.g., 1000)
	Lines int `json:"lines" yaml:"lines"`

	// Whether to include escape sequences
	IncludeEscapes bool `json:"include_escapes,omitempty" yaml:"include_escapes,omitempty"`
}

// Permissions declares what the plugin is allowed to do.
// This is used to configure the sandbox.
type Permissions struct {
	// Filesystem access
	Filesystem *FilesystemPermissions `json:"filesystem,omitempty" yaml:"filesystem,omitempty"`

	// Network access (default: denied)
	Network bool `json:"network,omitempty" yaml:"network,omitempty"`

	// Execute other programs (default: denied)
	Execute []string `json:"execute,omitempty" yaml:"execute,omitempty"`
}

// FilesystemPermissions specifies allowed filesystem access.
type FilesystemPermissions struct {
	// Read-only paths (e.g., ["/tmp"])
	ReadOnly []string `json:"read_only,omitempty" yaml:"read_only,omitempty"`

	// Read-write paths (e.g., ["~/.cache/plugin-name"])
	ReadWrite []string `json:"read_write,omitempty" yaml:"read_write,omitempty"`
}

// SandboxConfig specifies sandbox constraints.
type SandboxConfig struct {
	// Timeout for execution
	Timeout time.Duration `json:"timeout" yaml:"timeout"`

	// Maximum output size in bytes
	MaxOutputBytes int `json:"max_output_bytes" yaml:"max_output_bytes"`

	// Maximum memory in bytes (0 = no limit)
	MaxMemoryBytes int64 `json:"max_memory_bytes,omitempty" yaml:"max_memory_bytes,omitempty"`

	// Maximum CPU time in milliseconds (0 = no limit)
	MaxCPUMillis int64 `json:"max_cpu_millis,omitempty" yaml:"max_cpu_millis,omitempty"`
}

// DefaultSandboxConfig returns sensible defaults for sandbox configuration.
func DefaultSandboxConfig() SandboxConfig {
	return SandboxConfig{
		Timeout:        500 * time.Millisecond,
		MaxOutputBytes: 1024, // 1KB
		MaxMemoryBytes: 0,    // No limit (relies on OS)
		MaxCPUMillis:   0,    // No limit
	}
}

// RelaxedSandboxConfig returns looser constraints for verified plugins.
func RelaxedSandboxConfig() SandboxConfig {
	return SandboxConfig{
		Timeout:        2 * time.Second, // More time for verified plugins
		MaxOutputBytes: 4096,             // 4KB
		MaxMemoryBytes: 0,
		MaxCPUMillis:   0,
	}
}
