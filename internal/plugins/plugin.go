package plugins

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"bytes"

	"github.com/tmc/it2/internal/client"
	"github.com/tmc/it2/internal/formatting"
)

// PluginInput represents the data passed to a plugin via stdin
type PluginInput struct {
	SessionID   string   `json:"session_id"`
	SessionName string   `json:"session_name,omitempty"`
	Cwd         string   `json:"cwd,omitempty"`
	BufferLines []string `json:"buffer_lines,omitempty"`
}

// PluginType identifies the scope a plugin enriches.
type PluginType string

const (
	PluginTypeUnknown        PluginType = "unknown"
	PluginTypeSession        PluginType = "session"
	PluginTypeTab            PluginType = "tab"
	PluginTypeWindow         PluginType = "window"
	PluginTypeSessionProcess PluginType = "session-process"
)

// SessionEnricher is an interface for plugins that can add extra data to session listings
type SessionEnricher interface {
	// Name returns the name of the enricher
	Name() string

	// EnrichSession adds extra data to a session
	EnrichSession(ctx context.Context, session *client.SessionInfo) (map[string]interface{}, error)
}

// TabEnricher is an interface for plugins that can add extra data to tab listings
type TabEnricher interface {
	// Name returns the name of the enricher
	Name() string

	// EnrichTab adds extra data to a tab
	EnrichTab(ctx context.Context, tab *formatting.TabInfo) (map[string]interface{}, error)
}

// WindowEnricher is an interface for plugins that can add extra data to window listings
type WindowEnricher interface {
	// Name returns the name of the enricher
	Name() string

	// EnrichWindow adds extra data to a window
	EnrichWindow(ctx context.Context, window *client.WindowInfo) (map[string]interface{}, error)
}

// ProcessEnricher is an interface for plugins that can add extra process data
type ProcessEnricher interface {
	// Name returns the name of the enricher
	Name() string

	// EnrichProcess adds extra data about a process given its PID and session ID
	EnrichProcess(ctx context.Context, sessionID string, pid int) (map[string]interface{}, error)
}

// PluginEvent represents an event from a plugin
type PluginEvent struct {
	PluginName string                 `json:"plugin_name"`
	EventType  string                 `json:"event_type"`
	SessionID  string                 `json:"session_id"`
	Timestamp  time.Time              `json:"timestamp"`
	Data       map[string]interface{} `json:"data"`
	Message    string                 `json:"message,omitempty"`
}

// EventMonitor is an interface for plugins that can generate events during monitoring
type EventMonitor interface {
	// Name returns the name of the monitor plugin
	Name() string

	// StartMonitoring starts monitoring for events on the given session
	// Returns a channel that will receive events and an error channel
	StartMonitoring(ctx context.Context, sessionID string) (<-chan PluginEvent, <-chan error, error)

	// StopMonitoring stops monitoring (context cancellation should also work)
	StopMonitoring() error
}

// ExecPlugin represents a plugin that runs an external executable
type ExecPlugin struct {
	name       string
	executable string
	pluginType PluginType
}

// NewExecPlugin creates a new executable plugin
func NewExecPlugin(executable string) *ExecPlugin {
	baseName := filepath.Base(executable)

	pluginType, name := classifyPlugin(baseName)

	return &ExecPlugin{
		name:       name,
		executable: executable,
		pluginType: pluginType,
	}
}

func classifyPlugin(name string) (PluginType, string) {
	switch {
	case strings.HasPrefix(name, "it2-session-process-"):
		return PluginTypeSessionProcess, strings.TrimPrefix(name, "it2-session-process-")
	case strings.HasPrefix(name, "it2-session-"):
		return PluginTypeSession, strings.TrimPrefix(name, "it2-session-")
	case strings.HasPrefix(name, "it2-tab-"):
		return PluginTypeTab, strings.TrimPrefix(name, "it2-tab-")
	case strings.HasPrefix(name, "it2-window-"):
		return PluginTypeWindow, strings.TrimPrefix(name, "it2-window-")
	default:
		return PluginTypeUnknown, name
	}
}

// Name returns the name of the plugin
func (p *ExecPlugin) Name() string {
	return p.name
}

// Type reports the plugin type derived from its executable name.
func (p *ExecPlugin) Type() PluginType {
	return p.pluginType
}

// pluginEnvWhitelist is the set of environment variable names and prefixes
// passed to plugins. Everything else is filtered out to prevent leaking
// credentials (AWS keys, API tokens, etc.) to untrusted executables.
var pluginEnvWhitelist = []string{
	"HOME",
	"PATH",
	"SHELL",
	"TERM",
	"TMPDIR",
	"USER",
	"LOGNAME",
	"LANG",
}

// pluginEnvPrefixes are prefixes that are also allowed through.
var pluginEnvPrefixes = []string{
	"LC_",
	"IT2_",
	"ITERM2_",
	"ITERM_",
}

// setupPluginEnv builds a filtered environment for plugin execution.
// Only whitelisted variables are passed through to prevent leaking
// secrets from the parent process environment.
func setupPluginEnv(cmd *exec.Cmd) {
	var env []string
	for _, kv := range os.Environ() {
		key, _, _ := strings.Cut(kv, "=")
		if pluginEnvAllowed(key) {
			env = append(env, kv)
		}
	}
	cmd.Env = env
}

func pluginEnvAllowed(key string) bool {
	for _, allowed := range pluginEnvWhitelist {
		if key == allowed {
			return true
		}
	}
	for _, prefix := range pluginEnvPrefixes {
		if strings.HasPrefix(key, prefix) {
			return true
		}
	}
	return false
}

// getPluginDeadline returns the configured plugin deadline from environment or default
func getPluginDeadline() time.Duration {
	if deadline := os.Getenv("IT2_PLUGIN_DEADLINE"); deadline != "" {
		if d, err := time.ParseDuration(deadline); err == nil && d > 0 {
			return d
		}
	}
	return 5 * time.Second // Default deadline
}

// EnrichSession runs the executable with the session ID and parses the output
func (p *ExecPlugin) EnrichSession(ctx context.Context, session *client.SessionInfo) (map[string]interface{}, error) {
	// Record execution start time
	start := time.Now()
	defer func() {
		duration := time.Since(start)
		GetMetricsStore().RecordExecution(p.name, string(p.pluginType), duration)
	}()

	// Use configurable deadline for plugin execution
	pluginCtx, cancel := context.WithTimeout(ctx, getPluginDeadline())
	defer cancel()

	// Pass session ID and session name as arguments
	args := []string{session.SessionID}
	if session.SessionName != "" {
		args = append(args, session.SessionName)
	}
	cmd := exec.CommandContext(pluginCtx, p.executable, args...)
	setupPluginEnv(cmd)

	// prepare input
	input := PluginInput{
		SessionID:   session.SessionID,
		SessionName: session.SessionName,
	}
	inputBytes, _ := json.Marshal(input)
	cmd.Stdin = bytes.NewReader(inputBytes)

	stdout, stderr, err := runPlugin(cmd)
	if err != nil {
		if len(stderr) > 0 {
			fmt.Fprintf(os.Stderr, "plugin %s: %s\n", p.name, stderr)
		}
		return map[string]interface{}{}, nil
	}

	result := strings.TrimSpace(string(stdout))
	if result == "" {
		return map[string]interface{}{}, nil
	}

	return map[string]interface{}{
		p.name: result,
	}, nil
}

// EnrichTab runs the executable with tab information and parses the output
func (p *ExecPlugin) EnrichTab(ctx context.Context, tab *formatting.TabInfo) (map[string]interface{}, error) {
	// Record execution metrics
	start := time.Now()
	defer func() {
		GetMetricsStore().RecordExecution(p.name, string(p.pluginType), time.Since(start))
	}()

	// Use configurable deadline for plugin execution
	pluginCtx, cancel := context.WithTimeout(ctx, getPluginDeadline())
	defer cancel()

	// Pass tab ID, window ID, and title as arguments
	args := []string{tab.TabID, tab.WindowID}
	if tab.Title != "" {
		args = append(args, tab.Title)
	}
	cmd := exec.CommandContext(pluginCtx, p.executable, args...)
	setupPluginEnv(cmd)

	input := struct {
		TabID    string `json:"tab_id"`
		WindowID string `json:"window_id"`
		Title    string `json:"title,omitempty"`
	}{
		TabID:    tab.TabID,
		WindowID: tab.WindowID,
		Title:    tab.Title,
	}
	inputBytes, _ := json.Marshal(input)
	cmd.Stdin = bytes.NewReader(inputBytes)

	stdout, stderr, err := runPlugin(cmd)
	if err != nil {
		if len(stderr) > 0 {
			fmt.Fprintf(os.Stderr, "plugin %s: %s\n", p.name, stderr)
		}
		return map[string]interface{}{}, nil
	}

	result := strings.TrimSpace(string(stdout))
	if result == "" {
		return map[string]interface{}{}, nil
	}

	return map[string]interface{}{
		p.name: result,
	}, nil
}

// EnrichWindow runs the executable with window information and parses the output
func (p *ExecPlugin) EnrichWindow(ctx context.Context, window *client.WindowInfo) (map[string]interface{}, error) {
	// Record execution metrics
	start := time.Now()
	defer func() {
		GetMetricsStore().RecordExecution(p.name, string(p.pluginType), time.Since(start))
	}()

	// Use configurable deadline for plugin execution
	pluginCtx, cancel := context.WithTimeout(ctx, getPluginDeadline())
	defer cancel()

	// Pass window ID and title as arguments
	args := []string{window.WindowID}
	if window.Title != "" {
		args = append(args, window.Title)
	}
	cmd := exec.CommandContext(pluginCtx, p.executable, args...)
	setupPluginEnv(cmd)

	input := struct {
		WindowID string `json:"window_id"`
		Title    string `json:"title,omitempty"`
	}{
		WindowID: window.WindowID,
		Title:    window.Title,
	}
	inputBytes, _ := json.Marshal(input)
	cmd.Stdin = bytes.NewReader(inputBytes)

	stdout, stderr, err := runPlugin(cmd)
	if err != nil {
		if len(stderr) > 0 {
			fmt.Fprintf(os.Stderr, "plugin %s: %s\n", p.name, stderr)
		}
		return map[string]interface{}{}, nil
	}

	result := strings.TrimSpace(string(stdout))
	if result == "" {
		return map[string]interface{}{}, nil
	}

	return map[string]interface{}{
		p.name: result,
	}, nil
}

// EnrichProcess runs the executable with process information and parses the output
func (p *ExecPlugin) EnrichProcess(ctx context.Context, sessionID string, pid int) (map[string]interface{}, error) {
	// Record execution metrics
	start := time.Now()
	defer func() {
		GetMetricsStore().RecordExecution(p.name, string(p.pluginType), time.Since(start))
	}()

	// Use configurable deadline for plugin execution
	pluginCtx, cancel := context.WithTimeout(ctx, getPluginDeadline())
	defer cancel()

	// Pass session ID and PID as arguments
	args := []string{sessionID, fmt.Sprintf("%d", pid)}
	cmd := exec.CommandContext(pluginCtx, p.executable, args...)
	setupPluginEnv(cmd)

	input := struct {
		SessionID string `json:"session_id"`
		PID       int    `json:"pid"`
	}{
		SessionID: sessionID,
		PID:       pid,
	}
	inputBytes, _ := json.Marshal(input)
	cmd.Stdin = bytes.NewReader(inputBytes)

	stdout, stderr, err := runPlugin(cmd)
	if err != nil {
		if len(stderr) > 0 {
			fmt.Fprintf(os.Stderr, "plugin %s: %s\n", p.name, stderr)
		}
		return map[string]interface{}{}, nil
	}

	result := strings.TrimSpace(string(stdout))
	if result == "" {
		return map[string]interface{}{}, nil
	}

	return map[string]interface{}{
		p.name: result,
	}, nil
}

// runPlugin executes the command and returns stdout and stderr separately.
// This avoids the CombinedOutput bug where stderr warnings corrupt JSON on stdout.
func runPlugin(cmd *exec.Cmd) (stdout, stderr []byte, err error) {
	var outBuf, errBuf bytes.Buffer
	cmd.Stdout = &outBuf
	cmd.Stderr = &errBuf
	err = cmd.Run()
	return outBuf.Bytes(), errBuf.Bytes(), err
}

// StartMonitoring starts monitoring for events from the plugin
func (p *ExecPlugin) StartMonitoring(ctx context.Context, sessionID string) (<-chan PluginEvent, <-chan error, error) {
	eventChan := make(chan PluginEvent, 100)
	errorChan := make(chan error, 10)

	cmd := exec.CommandContext(ctx, p.executable, "monitor", sessionID)
	setupPluginEnv(cmd)
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get stdout pipe: %w", err)
	}

	if err := cmd.Start(); err != nil {
		return nil, nil, fmt.Errorf("failed to start plugin: %w", err)
	}

	// Start goroutine to read events from plugin
	go func() {
		defer close(eventChan)
		defer close(errorChan)
		defer stdout.Close()

		decoder := json.NewDecoder(stdout)
		for {
			select {
			case <-ctx.Done():
				return
			default:
				var event PluginEvent
				if err := decoder.Decode(&event); err != nil {
					if err == io.EOF {
						return // Plugin finished normally
					}
					select {
					case errorChan <- fmt.Errorf("failed to decode plugin event: %w", err):
					case <-ctx.Done():
						return
					}
					continue
				}

				// Ensure plugin name is set
				if event.PluginName == "" {
					event.PluginName = p.name
				}

				select {
				case eventChan <- event:
				case <-ctx.Done():
					return
				}
			}
		}
	}()

	return eventChan, errorChan, nil
}

// StopMonitoring stops monitoring (relies on context cancellation)
func (p *ExecPlugin) StopMonitoring() error {
	// Monitoring is stopped via context cancellation
	// No additional cleanup needed for this implementation
	return nil
}
