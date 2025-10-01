package plugins

import (
	"crypto/sha256"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/tmc/it2/internal/embedded"
)

// PluginMetadata contains detailed information about a discovered plugin
type PluginMetadata struct {
	Name       string
	Path       string
	Source     string // "PATH", "custom", or "embedded"
	SHA256     string
	Duplicates int // Count of other plugins with same name
}

// DiscoverPlugins finds all it2-* executables in PATH and categorizes them
// Search order (highest to lowest priority):
//  1. Directories from PATH environment variable (highest priority)
//  2. Directories from IT2_PLUGIN_PATHS (--plugin-path flag, middle priority)
//  3. Embedded plugins directory (lowest priority, fallback)
func DiscoverPlugins() ([]SessionEnricher, []TabEnricher, []WindowEnricher, []ProcessEnricher, error) {
	var sessionPlugins []SessionEnricher
	var tabPlugins []TabEnricher
	var windowPlugins []WindowEnricher
	var processPlugins []ProcessEnricher
	seen := make(map[string]bool) // Track seen plugin names to avoid duplicates

	var paths []string

	// Priority 1: User's PATH directories (highest priority)
	pathEnv := os.Getenv("PATH")
	if pathEnv != "" {
		paths = append(paths, strings.Split(pathEnv, string(os.PathListSeparator))...)
	}

	// Priority 2: Additional plugin paths from --plugin-path flag (middle priority)
	pluginPathsEnv := os.Getenv("IT2_PLUGIN_PATHS")
	if pluginPathsEnv != "" {
		paths = append(paths, strings.Split(pluginPathsEnv, string(os.PathListSeparator))...)
	}

	// Priority 3: Embedded plugins directory (lowest priority, fallback)
	pluginsDir, err := embedded.ExtractPlugins()
	if err == nil {
		paths = append(paths, pluginsDir)
	}

	if len(paths) == 0 {
		return sessionPlugins, tabPlugins, windowPlugins, processPlugins, nil
	}

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
// Search order (highest to lowest priority):
//  1. Directories from PATH environment variable (highest priority)
//  2. Directories from IT2_PLUGIN_PATHS (--plugin-path flag, middle priority)
//  3. Embedded plugins directory (lowest priority, fallback)
func DiscoverEventMonitors() ([]EventMonitor, error) {
	var eventMonitors []EventMonitor
	seen := make(map[string]bool) // Track seen plugin names to avoid duplicates

	var paths []string

	// Priority 1: User's PATH directories (highest priority)
	pathEnv := os.Getenv("PATH")
	if pathEnv != "" {
		paths = append(paths, strings.Split(pathEnv, string(os.PathListSeparator))...)
	}

	// Priority 2: Additional plugin paths from --plugin-path flag (middle priority)
	pluginPathsEnv := os.Getenv("IT2_PLUGIN_PATHS")
	if pluginPathsEnv != "" {
		paths = append(paths, strings.Split(pluginPathsEnv, string(os.PathListSeparator))...)
	}

	// Priority 3: Embedded plugins directory (lowest priority, fallback)
	pluginsDir, err := embedded.ExtractPlugins()
	if err == nil {
		paths = append(paths, pluginsDir)
	}

	if len(paths) == 0 {
		return eventMonitors, nil
	}

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

// DiscoverPluginMetadata returns detailed metadata about all discovered plugins
func DiscoverPluginMetadata() ([]PluginMetadata, error) {
	var metadata []PluginMetadata
	nameCount := make(map[string]int) // Track duplicates

	var paths []string
	var pathSources []string // Track which source each path comes from

	// Priority 1: User's PATH directories (highest priority)
	pathEnv := os.Getenv("PATH")
	if pathEnv != "" {
		pathDirs := strings.Split(pathEnv, string(os.PathListSeparator))
		for _, dir := range pathDirs {
			paths = append(paths, dir)
			pathSources = append(pathSources, "PATH")
		}
	}

	// Priority 2: Additional plugin paths from --plugin-path flag
	pluginPathsEnv := os.Getenv("IT2_PLUGIN_PATHS")
	if pluginPathsEnv != "" {
		pluginDirs := strings.Split(pluginPathsEnv, string(os.PathListSeparator))
		for _, dir := range pluginDirs {
			paths = append(paths, dir)
			pathSources = append(pathSources, "custom")
		}
	}

	// Priority 3: Embedded plugins directory
	pluginsDir, err := embedded.ExtractPlugins()
	if err == nil {
		paths = append(paths, pluginsDir)
		pathSources = append(pathSources, "embedded")
	}

	// First pass: collect all plugins and count duplicates
	allPlugins := make(map[string][]PluginMetadata) // name -> list of metadata

	for i, dir := range paths {
		entries, err := os.ReadDir(dir)
		if err != nil {
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
			info, err := os.Stat(fullPath)
			if err != nil || info.Mode()&0111 == 0 {
				continue
			}

			// Calculate SHA256
			f, err := os.Open(fullPath)
			if err != nil {
				continue
			}
			h := sha256.New()
			if _, err := io.Copy(h, f); err != nil {
				f.Close()
				continue
			}
			f.Close()
			sha := fmt.Sprintf("%x", h.Sum(nil))[:16] // First 16 chars

			// Determine plugin name (remove prefix)
			plugin := NewExecPlugin(fullPath)
			pluginName := plugin.Name()

			meta := PluginMetadata{
				Name:   pluginName,
				Path:   fullPath,
				Source: pathSources[i],
				SHA256: sha,
			}

			allPlugins[pluginName] = append(allPlugins[pluginName], meta)
			nameCount[pluginName]++
		}
	}

	// Second pass: assign duplicate counts and collect first occurrence of each
	seen := make(map[string]bool)
	for _, dir := range paths {
		entries, err := os.ReadDir(dir)
		if err != nil {
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
			info, err := os.Stat(fullPath)
			if err != nil || info.Mode()&0111 == 0 {
				continue
			}

			plugin := NewExecPlugin(fullPath)
			pluginName := plugin.Name()

			// Only include first occurrence (highest priority)
			if seen[pluginName] {
				continue
			}
			seen[pluginName] = true

			// Find this plugin's metadata
			for _, meta := range allPlugins[pluginName] {
				if meta.Path == fullPath {
					meta.Duplicates = nameCount[pluginName] - 1 // Don't count itself
					metadata = append(metadata, meta)
					break
				}
			}
		}
	}

	return metadata, nil
}
