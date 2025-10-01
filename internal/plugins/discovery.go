package plugins

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/tmc/it2/internal/embedded"
)

// DiscoverPlugins finds all it2-* executables in PATH and categorizes them
func DiscoverPlugins() ([]SessionEnricher, []TabEnricher, []WindowEnricher, []ProcessEnricher, error) {
	var sessionPlugins []SessionEnricher
	var tabPlugins []TabEnricher
	var windowPlugins []WindowEnricher
	var processPlugins []ProcessEnricher
	seen := make(map[string]bool) // Track seen plugin names to avoid duplicates

	// Get PATH environment variable
	pathEnv := os.Getenv("PATH")
	if pathEnv == "" {
		return sessionPlugins, tabPlugins, windowPlugins, processPlugins, nil
	}

	// Split PATH into directories
	paths := strings.Split(pathEnv, string(os.PathListSeparator))

	// Look for executables starting with "it2-"
	for _, dir := range paths {
		entries, err := os.ReadDir(dir)
		if err != nil {
			// Skip directories we can't read
			continue
		}

		for _, entry := range entries {
			if entry.IsDir() {
				continue
			}

			name := entry.Name()
			if !strings.HasPrefix(name, "it2-") {
				continue
			}

			fullPath := filepath.Join(dir, name)

			// Check if it's executable
			info, err := os.Stat(fullPath)
			if err != nil {
				continue
			}

			// Check if file is executable (Unix-like systems)
			if info.Mode()&0111 != 0 {
				// Skip if we've already seen this plugin name
				if seen[name] {
					continue
				}
				seen[name] = true

				plugin := NewExecPlugin(fullPath)

				// Add to appropriate plugin lists based on type
				if plugin.pluginType == "session" || plugin.pluginType == "generic" {
					sessionPlugins = append(sessionPlugins, plugin)
				}
				if plugin.pluginType == "tab" || plugin.pluginType == "generic" {
					tabPlugins = append(tabPlugins, plugin)
				}
				if plugin.pluginType == "window" || plugin.pluginType == "generic" {
					windowPlugins = append(windowPlugins, plugin)
				}
				if plugin.pluginType == "session-process" {
					processPlugins = append(processPlugins, plugin)
				}
			}
		}
	}

	return sessionPlugins, tabPlugins, windowPlugins, processPlugins, nil
}

// DiscoverEventMonitors finds all it2-* executables that support event monitoring
func DiscoverEventMonitors() ([]EventMonitor, error) {
	var eventMonitors []EventMonitor
	seen := make(map[string]bool) // Track seen plugin names to avoid duplicates

	// Get PATH environment variable
	pathEnv := os.Getenv("PATH")
	if pathEnv == "" {
		return eventMonitors, nil
	}

	// Split PATH into directories
	paths := strings.Split(pathEnv, string(os.PathListSeparator))

	// Look for executables starting with "it2-"
	for _, dir := range paths {
		entries, err := os.ReadDir(dir)
		if err != nil {
			// Skip directories we can't read
			continue
		}

		for _, entry := range entries {
			if entry.IsDir() {
				continue
			}

			name := entry.Name()
			if !strings.HasPrefix(name, "it2-") {
				continue
			}

			fullPath := filepath.Join(dir, name)

			// Check if it's executable
			info, err := os.Stat(fullPath)
			if err != nil {
				continue
			}

			// Check if file is executable (Unix-like systems)
			if info.Mode()&0111 != 0 {
				// Skip if we've already seen this plugin name
				if seen[name] {
					continue
				}
				seen[name] = true

				plugin := NewExecPlugin(fullPath)

				// All session plugins can potentially be event monitors
				if plugin.pluginType == "session" || plugin.pluginType == "generic" {
					eventMonitors = append(eventMonitors, plugin)
				}
			}
		}
	}

	return eventMonitors, nil
}

// Registry holds discovered plugins
type Registry struct {
	sessionEnrichers []SessionEnricher
	tabEnrichers     []TabEnricher
	windowEnrichers  []WindowEnricher
	processEnrichers []ProcessEnricher
}

// NewRegistry creates a new plugin registry
func NewRegistry() *Registry {
	return &Registry{
		sessionEnrichers: []SessionEnricher{},
		tabEnrichers:     []TabEnricher{},
		windowEnrichers:  []WindowEnricher{},
		processEnrichers: []ProcessEnricher{},
	}
}

// DiscoverAndRegister discovers and registers all plugins
func (r *Registry) DiscoverAndRegister() error {
	sessionPlugins, tabPlugins, windowPlugins, processPlugins, err := DiscoverPlugins()
	if err != nil {
		return err
	}
	r.sessionEnrichers = sessionPlugins
	r.tabEnrichers = tabPlugins
	r.windowEnrichers = windowPlugins
	r.processEnrichers = processPlugins
	return nil
}

// GetEnrichers returns all registered session enrichers (backward compatibility)
func (r *Registry) GetEnrichers() []SessionEnricher {
	return r.sessionEnrichers
}

// GetSessionEnrichers returns all registered session enrichers
func (r *Registry) GetSessionEnrichers() []SessionEnricher {
	return r.sessionEnrichers
}

// GetTabEnrichers returns all registered tab enrichers
func (r *Registry) GetTabEnrichers() []TabEnricher {
	return r.tabEnrichers
}

// GetWindowEnrichers returns all registered window enrichers
func (r *Registry) GetWindowEnrichers() []WindowEnricher {
	return r.windowEnrichers
}

// AddSessionEnricher manually adds a session enricher to the registry
func (r *Registry) AddSessionEnricher(enricher SessionEnricher) {
	r.sessionEnrichers = append(r.sessionEnrichers, enricher)
}

// AddTabEnricher manually adds a tab enricher to the registry
func (r *Registry) AddTabEnricher(enricher TabEnricher) {
	r.tabEnrichers = append(r.tabEnrichers, enricher)
}

// AddWindowEnricher manually adds a window enricher to the registry
func (r *Registry) AddWindowEnricher(enricher WindowEnricher) {
	r.windowEnrichers = append(r.windowEnrichers, enricher)
}

// GetProcessEnrichers returns all registered process enrichers
func (r *Registry) GetProcessEnrichers() []ProcessEnricher {
	return r.processEnrichers
}

// AddProcessEnricher manually adds a process enricher to the registry
func (r *Registry) AddProcessEnricher(enricher ProcessEnricher) {
	r.processEnrichers = append(r.processEnrichers, enricher)
}
