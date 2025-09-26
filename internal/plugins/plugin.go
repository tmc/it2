package plugins

import (
	"context"
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
	if strings.HasPrefix(baseName, "it2-session-") {
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

// EnrichSession runs the executable with the session ID and parses the output
func (p *ExecPlugin) EnrichSession(ctx context.Context, session *client.SessionInfo) (map[string]interface{}, error) {
	// Use a longer timeout for plugins to ensure they can complete
	pluginCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	// Pass session ID and session name as arguments
	args := []string{session.SessionID}
	if session.SessionName != "" {
		args = append(args, session.SessionName)
	}
	cmd := exec.CommandContext(pluginCtx, p.executable, args...)
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
	// Use a longer timeout for plugins to ensure they can complete
	pluginCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	// Pass tab ID, window ID, and title as arguments
	args := []string{tab.TabID, tab.WindowID}
	if tab.Title != "" {
		args = append(args, tab.Title)
	}
	cmd := exec.CommandContext(pluginCtx, p.executable, args...)
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
	// Use a longer timeout for plugins to ensure they can complete
	pluginCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	// Pass window ID and title as arguments
	args := []string{window.WindowID}
	if window.Title != "" {
		args = append(args, window.Title)
	}
	cmd := exec.CommandContext(pluginCtx, p.executable, args...)
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
