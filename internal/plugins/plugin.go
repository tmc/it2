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

	"github.com/tmc/it2/internal/client"
	"github.com/tmc/it2/internal/formatting"
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
	pluginType string // "session", "tab", "window", or "generic"
}

// NewExecPlugin creates a new executable plugin
func NewExecPlugin(executable string) *ExecPlugin {
	baseName := filepath.Base(executable)

	// Determine plugin type and clean name based on prefix
	var pluginType, name string
	if strings.HasPrefix(baseName, "it2-session-process-") {
		pluginType = "session-process"
		name = strings.TrimPrefix(baseName, "it2-session-process-")
	} else if strings.HasPrefix(baseName, "it2-session-") {
		pluginType = "session"
		name = strings.TrimPrefix(baseName, "it2-session-")
	} else if strings.HasPrefix(baseName, "it2-tab-") {
		pluginType = "tab"
		name = strings.TrimPrefix(baseName, "it2-tab-")
	} else if strings.HasPrefix(baseName, "it2-window-") {
		pluginType = "window"
		name = strings.TrimPrefix(baseName, "it2-window-")
	} else if strings.HasPrefix(baseName, "it2-") {
		pluginType = "generic"
		name = strings.TrimPrefix(baseName, "it2-")
	} else {
		pluginType = "generic"
		name = baseName
	}

	return &ExecPlugin{
		name:       name,
		executable: executable,
		pluginType: pluginType,
	}
}

// Name returns the name of the plugin
func (p *ExecPlugin) Name() string {
	return p.name
}

// setupPluginEnv adds iTerm2 authentication credentials to the command environment
func setupPluginEnv(cmd *exec.Cmd) {
	// Inherit parent environment
	cmd.Env = os.Environ()

	// Add iTerm2 authentication if available
	if cookie := os.Getenv("ITERM2_COOKIE"); cookie != "" {
		cmd.Env = append(cmd.Env, "ITERM2_COOKIE="+cookie)
	}
	if key := os.Getenv("ITERM2_KEY"); key != "" {
		cmd.Env = append(cmd.Env, "ITERM2_KEY="+key)
	}
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
		GetMetricsStore().RecordExecution(p.name, duration)
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
	output, err := cmd.CombinedOutput() // Get both stdout and stderr
	if err != nil {
		// If the command fails, return empty data instead of error
		// This allows plugins to be optional
		// Debug output
		//fmt.Fprintf(os.Stderr, "Plugin %s error for session %s: %v, output: %s\n", p.name, session.SessionID, err, string(output))
		return map[string]interface{}{}, nil
	}

	// Parse the output - for now just return the trimmed output as a string
	result := strings.TrimSpace(string(output))

	// Debug output
	//fmt.Fprintf(os.Stderr, "Plugin %s returned: '%s' for session %s\n", p.name, result, session.SessionID)

	// Don't add empty results
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
		GetMetricsStore().RecordExecution(p.name, time.Since(start))
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
	output, err := cmd.CombinedOutput()
	if err != nil {
		return map[string]interface{}{}, nil
	}

	result := strings.TrimSpace(string(output))
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
		GetMetricsStore().RecordExecution(p.name, time.Since(start))
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
	output, err := cmd.CombinedOutput()
	if err != nil {
		return map[string]interface{}{}, nil
	}

	result := strings.TrimSpace(string(output))
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
		GetMetricsStore().RecordExecution(p.name, time.Since(start))
	}()

	// Use configurable deadline for plugin execution
	pluginCtx, cancel := context.WithTimeout(ctx, getPluginDeadline())
	defer cancel()

	// Pass session ID and PID as arguments
	args := []string{sessionID, fmt.Sprintf("%d", pid)}
	cmd := exec.CommandContext(pluginCtx, p.executable, args...)
	setupPluginEnv(cmd)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return map[string]interface{}{}, nil
	}

	result := strings.TrimSpace(string(output))
	if result == "" {
		return map[string]interface{}{}, nil
	}

	return map[string]interface{}{
		p.name: result,
	}, nil
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
