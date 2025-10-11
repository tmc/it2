package plugins

import (
	"context"
	"regexp"
	"strings"
	"sync"

	"github.com/tmc/it2/internal/client"
	"github.com/tmc/it2/internal/formatting"
)

// ShouldRunPlugins determines if plugins should be executed based on requested columns.
// It returns true if:
//   - No columns are specified (run all plugins)
//   - Any non-standard column is requested
func ShouldRunPlugins(columns []string) bool {
	if len(columns) == 0 {
		return true // Run all plugins if no columns specified
	}

	// Standard columns that don't require plugins
	standardCols := map[string]bool{
		"id": true, "split from": true, "pid": true, "exit": true, "state": true,
		"window": true, "tab": true, "title": true, "command": true,
	}

	// Check if any plugin column is requested
	for _, col := range columns {
		colLower := strings.ToLower(col)
		if !standardCols[colLower] {
			return true
		}
	}

	return false
}

// FilterEnrichers filters enrichers based on requested plugin names from columns.
// If no columns are specified, all enrichers are returned.
func FilterEnrichers(enrichers []SessionEnricher, columns []string) []SessionEnricher {
	if len(columns) == 0 {
		return enrichers // Return all if no specific columns requested
	}

	// Build set of requested plugin names
	requestedPlugins := make(map[string]bool)
	for _, col := range columns {
		colLower := strings.ToLower(col)
		requestedPlugins[colLower] = true
	}

	// Filter enrichers based on requested plugins
	var activeEnrichers []SessionEnricher
	for _, enricher := range enrichers {
		pluginName := strings.ToLower(enricher.Name())
		if requestedPlugins[pluginName] {
			activeEnrichers = append(activeEnrichers, enricher)
		}
	}

	return activeEnrichers
}

// FilterEnrichersByPattern filters enrichers whose names match the provided regexp.
// If pattern is nil, the original slice is returned.
func FilterEnrichersByPattern(enrichers []SessionEnricher, pattern *regexp.Regexp) []SessionEnricher {
	if pattern == nil {
		return enrichers
	}

	var filtered []SessionEnricher
	for _, enricher := range enrichers {
		if pattern.MatchString(enricher.Name()) {
			filtered = append(filtered, enricher)
		}
	}
	return filtered
}

// EnrichSessionsParallel runs plugins concurrently for each session and collects results.
// It limits concurrency to prevent overwhelming the system (max 20 concurrent executions).
func EnrichSessionsParallel(ctx context.Context, sessions []*client.SessionInfo, enrichers []SessionEnricher) {
	if len(enrichers) == 0 {
		return
	}

	// Create a semaphore to limit concurrency (max 20 concurrent plugin executions)
	semaphore := make(chan struct{}, 20)

	// Use a WaitGroup to wait for all goroutines
	var wg sync.WaitGroup

	// Create a mutex per session for safe concurrent writes
	type sessionMutex struct {
		session *client.SessionInfo
		mu      sync.Mutex
	}

	sessionMutexes := make([]*sessionMutex, len(sessions))
	for i, session := range sessions {
		// Initialize plugin data map if needed
		if session.PluginData == nil {
			session.PluginData = make(map[string]interface{})
		}
		sessionMutexes[i] = &sessionMutex{session: session}
	}

	// Process each session with parallel enrichers
	for _, sm := range sessionMutexes {
		for _, enricher := range enrichers {
			wg.Add(1)
			go func(sm *sessionMutex, enr SessionEnricher) {
				defer wg.Done()

				// Acquire semaphore to limit concurrency
				semaphore <- struct{}{}
				defer func() { <-semaphore }()

				// Execute the enricher
				data, err := enr.EnrichSession(ctx, sm.session)
				if err == nil && len(data) > 0 {
					// Lock this session's mutex for safe writes
					sm.mu.Lock()
					for k, v := range data {
						sm.session.PluginData[k] = v
					}
					sm.mu.Unlock()
				}
			}(sm, enricher)
		}
	}

	// Wait for all enrichers to complete
	wg.Wait()
}

// EnrichTabs applies tab enrichers to a list of tabs.
func EnrichTabs(ctx context.Context, tabs []*formatting.TabInfo, enrichers []TabEnricher) {
	for _, tab := range tabs {
		if tab.PluginData == nil {
			tab.PluginData = make(map[string]interface{})
		}
		for _, enricher := range enrichers {
			pluginData, err := enricher.EnrichTab(ctx, tab)
			if err == nil {
				for k, v := range pluginData {
					tab.PluginData[k] = v
				}
			}
		}
	}
}

// EnrichWindows applies window enrichers to a list of windows.
func EnrichWindows(ctx context.Context, windows []*client.WindowInfo, enrichers []WindowEnricher) {
	for _, window := range windows {
		// Initialize plugin data map if needed
		if window.PluginData == nil {
			window.PluginData = make(map[string]interface{})
		}
		// Apply each window enricher
		for _, enricher := range enrichers {
			data, err := enricher.EnrichWindow(ctx, window)
			if err == nil && len(data) > 0 {
				for k, v := range data {
					window.PluginData[k] = v
				}
			}
		}
	}
}
