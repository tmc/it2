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
	Type       PluginType
	Path       string
	Source     string // "PATH", "custom", or "embedded"
	SHA256     string
	Duplicates int // Count of other plugins with same name and type
}

// DiscoverPlugins finds plugin executables with supported prefixes in PATH and categorizes them
// Search order (highest to lowest priority):
//  1. Directories from PATH environment variable (highest priority)
//  2. Directories from IT2_PLUGIN_PATHS (--plugin-path flag, middle priority)
//  3. Embedded plugins directory (lowest priority, fallback)
func DiscoverPlugins() ([]SessionEnricher, []TabEnricher, []WindowEnricher, []ProcessEnricher, error) {
	var sessionPlugins []SessionEnricher
	var tabPlugins []TabEnricher
	var windowPlugins []WindowEnricher
	var processPlugins []ProcessEnricher
	seen := make(map[string]bool) // Track seen plugin (type,name) pairs to avoid duplicates

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

	// Look for executables starting with "it2-" and having a supported subtype prefix
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
			fullPath := filepath.Join(dir, name)

			// Check if it's executable
			info, err := os.Stat(fullPath)
			if err != nil {
				continue
			}

			// Check if file is executable (Unix-like systems)
			if info.Mode()&0111 != 0 {
				if !strings.HasPrefix(name, "it2-") {
					continue
				}

				plugin := NewExecPlugin(fullPath)
				if plugin.Type() == PluginTypeUnknown {
					continue
				}

				key := discoveryKey(plugin.Type(), plugin.Name())
				if seen[key] {
					continue
				}
				seen[key] = true

				// Add to appropriate plugin lists based on type
				switch plugin.Type() {
				case PluginTypeSession:
					sessionPlugins = append(sessionPlugins, plugin)
				case PluginTypeTab:
					tabPlugins = append(tabPlugins, plugin)
				case PluginTypeWindow:
					windowPlugins = append(windowPlugins, plugin)
				case PluginTypeSessionProcess:
					processPlugins = append(processPlugins, plugin)
				}
			}
		}
	}

	return sessionPlugins, tabPlugins, windowPlugins, processPlugins, nil
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

// GetTabEnrichers returns all registered tab enrichers
func (r *Registry) GetTabEnrichers() []TabEnricher {
	return r.tabEnrichers
}

// GetWindowEnrichers returns all registered window enrichers
func (r *Registry) GetWindowEnrichers() []WindowEnricher {
	return r.windowEnrichers
}

// GetProcessEnrichers returns all registered process enrichers
func (r *Registry) GetProcessEnrichers() []ProcessEnricher {
	return r.processEnrichers
}

// DiscoverPluginMetadata returns detailed metadata about all discovered plugins
func DiscoverPluginMetadata() ([]PluginMetadata, error) {
	var metadata []PluginMetadata
	nameCount := make(map[string]int) // Track duplicates per (type,name)

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
	allPlugins := make(map[string][]PluginMetadata) // key -> list of metadata

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
			if plugin.Type() == PluginTypeUnknown {
				continue
			}
			pluginName := plugin.Name()
			key := discoveryKey(plugin.Type(), pluginName)

			meta := PluginMetadata{
				Name:   pluginName,
				Type:   plugin.Type(),
				Path:   fullPath,
				Source: pathSources[i],
				SHA256: sha,
			}

			allPlugins[key] = append(allPlugins[key], meta)
			nameCount[key]++
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
			fullPath := filepath.Join(dir, name)
			info, err := os.Stat(fullPath)
			if err != nil || info.Mode()&0111 == 0 {
				continue
			}

			plugin := NewExecPlugin(fullPath)
			if plugin.Type() == PluginTypeUnknown {
				continue
			}
			pluginName := plugin.Name()
			key := discoveryKey(plugin.Type(), pluginName)

			// Only include first occurrence (highest priority)
			if seen[key] {
				continue
			}
			seen[key] = true

			// Find this plugin's metadata
			for _, meta := range allPlugins[key] {
				if meta.Path == fullPath {
					meta.Duplicates = nameCount[key] - 1 // Don't count itself
					metadata = append(metadata, meta)
					break
				}
			}
		}
	}

	return metadata, nil
}

func discoveryKey(t PluginType, name string) string {
	return string(t) + ":" + name
}
