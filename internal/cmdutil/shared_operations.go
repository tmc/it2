package cmdutil

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"strings"
	"sync"

	"github.com/tmc/it2/internal/client"
	"github.com/tmc/it2/internal/formatting"
	"github.com/tmc/it2/internal/plugins"
)

// SharedListOptions contains common options for list operations
type SharedListOptions struct {
	WindowID     string
	TabID        string
	SessionID    string
	ScopeFlag    string
	Format       string
	Columns      []string
	SortBy       string
	SortReverse  bool
	Quiet        bool
	NoHyperlinks bool
}

// SharedListOperations provides shared listing functionality for both flat and hierarchical commands
type SharedListOperations struct {
	client *client.Client
	ctx    context.Context
}

// NewSharedListOperations creates a new instance with client and context
func NewSharedListOperations(client *client.Client, ctx context.Context) *SharedListOperations {
	return &SharedListOperations{
		client: client,
		ctx:    ctx,
	}
}

// ListSessions lists sessions with optional filtering by window and tab
func (s *SharedListOperations) ListSessions(opts SharedListOptions) error {
	sessions, err := s.client.ListSessions(s.ctx)
	if err != nil {
		return fmt.Errorf("failed to list sessions: %w", err)
	}

	// Apply scope filtering first
	scopedSessions := s.applyScopeFilter(sessions, opts.ScopeFlag)

	// Filter sessions based on options
	var filteredSessions []*client.SessionInfo
	for _, session := range scopedSessions {
		// Apply window filter if specified
		if opts.WindowID != "" && session.WindowID != opts.WindowID {
			continue
		}
		// Apply tab filter if specified
		if opts.TabID != "" && session.TabID != opts.TabID {
			continue
		}
		// Apply session filter if specified
		if opts.SessionID != "" && session.SessionID != opts.SessionID {
			continue
		}
		filteredSessions = append(filteredSessions, session)
	}

	// Apply plugin enrichment - only if plugin columns are requested or no columns specified
	shouldRunPlugins := len(opts.Columns) == 0 // Run all plugins if no columns specified
	if !shouldRunPlugins {
		// Check if any plugin column is requested
		for _, col := range opts.Columns {
			colLower := strings.ToLower(col)
			// Check if this looks like a plugin column (not a standard column)
			standardCols := map[string]bool{
				"id": true, "parent": true, "pid": true, "exit": true, "state": true,
				"window": true, "tab": true, "title": true, "command": true,
			}
			if !standardCols[colLower] {
				shouldRunPlugins = true
				break
			}
		}
	}

	if shouldRunPlugins {
		registry := plugins.NewRegistry()
		if err := registry.DiscoverAndRegister(); err == nil {
			enrichers := registry.GetEnrichers()

			// Build set of requested plugin names if columns are specified
			requestedPlugins := make(map[string]bool)
			if len(opts.Columns) > 0 {
				for _, col := range opts.Columns {
					colLower := strings.ToLower(col)
					requestedPlugins[colLower] = true
				}
			}

			// Filter enrichers based on requested plugins
			var activeEnrichers []plugins.SessionEnricher
			for _, enricher := range enrichers {
				if len(requestedPlugins) > 0 {
					pluginName := strings.ToLower(enricher.Name())
					if !requestedPlugins[pluginName] {
						continue
					}
				}
				activeEnrichers = append(activeEnrichers, enricher)
			}

			// Parallelize plugin execution across all sessions
			s.enrichSessionsParallel(filteredSessions, activeEnrichers)

			// Save metrics after enrichment
			plugins.GetMetricsStore().Save()
		}
	}

	// Format and output
	formatter := formatting.NewWithHyperlinks(opts.Format, opts.Columns, opts.SortBy, opts.SortReverse, opts.Quiet, !opts.NoHyperlinks)
	return formatter.FormatSessions(filteredSessions)
}

// ListTabs lists tabs with optional filtering by window
func (s *SharedListOperations) ListTabs(opts SharedListOptions) error {
	// Get sessions to build tab information
	sessions, err := s.client.ListSessions(s.ctx)
	if err != nil {
		return fmt.Errorf("failed to list sessions: %w", err)
	}

	// Build tab info from sessions
	tabInfos := s.buildTabInfoFromSessions(sessions, opts.WindowID)

	// Apply plugin enrichment
	registry := plugins.NewRegistry()
	if err := registry.DiscoverAndRegister(); err == nil {
		enrichers := registry.GetTabEnrichers()
		for _, tabInfo := range tabInfos {
			if tabInfo.PluginData == nil {
				tabInfo.PluginData = make(map[string]interface{})
			}
			for _, enricher := range enrichers {
				pluginData, _ := enricher.EnrichTab(s.ctx, tabInfo)
				for k, v := range pluginData {
					tabInfo.PluginData[k] = v
				}
			}
		}
	}

	// Format and output
	formatter := formatting.NewWithHyperlinks(opts.Format, opts.Columns, opts.SortBy, opts.SortReverse, opts.Quiet, !opts.NoHyperlinks)
	return formatter.FormatTabs(tabInfos)
}

// ListWindows lists all windows
func (s *SharedListOperations) ListWindows(opts SharedListOptions) error {
	windows, err := s.client.ListWindows(s.ctx)
	if err != nil {
		return fmt.Errorf("failed to list windows: %w", err)
	}

	// Apply plugins to windows
	registry := plugins.NewRegistry()
	if err := registry.DiscoverAndRegister(); err == nil {
		windowEnrichers := registry.GetWindowEnrichers()
		for _, window := range windows {
			// Initialize plugin data map if needed
			if window.PluginData == nil {
				window.PluginData = make(map[string]interface{})
			}
			// Apply each window enricher
			for _, enricher := range windowEnrichers {
				data, err := enricher.EnrichWindow(s.ctx, window)
				if err == nil && len(data) > 0 {
					for k, v := range data {
						window.PluginData[k] = v
					}
				}
			}
		}
	}

	formatter := formatting.NewWithHyperlinks(opts.Format, opts.Columns, opts.SortBy, opts.SortReverse, opts.Quiet, !opts.NoHyperlinks)
	return formatter.FormatClientWindows(windows)
}

// buildTabInfoFromSessions builds tab information from session data
func (s *SharedListOperations) buildTabInfoFromSessions(sessions []*client.SessionInfo, windowID string) []*formatting.TabInfo {
	tabMap := make(map[string]*formatting.TabInfo)

	for _, session := range sessions {
		// Filter by window if specified
		if windowID != "" && session.WindowID != windowID {
			continue
		}

		// Create or update tab info
		key := session.WindowID + ":" + session.TabID
		if _, exists := tabMap[key]; !exists {
			tabMap[key] = &formatting.TabInfo{
				WindowID: session.WindowID,
				TabID:    session.TabID,
				Title:    session.TabTitle,
				Position: 0, // Position needs to be determined differently
			}
		}
	}

	// Convert map to slice
	var result []*formatting.TabInfo
	for _, tabInfo := range tabMap {
		result = append(result, tabInfo)
	}

	return result
}

// ResolveWindowID converts a window index (like "0", "1") to the actual window UUID
// Returns the input unchanged if it's already a UUID or if resolution fails
func (s *SharedListOperations) ResolveWindowID(windowIDOrIndex string) (string, error) {
	// If it already looks like a UUID, return as-is
	if strings.HasPrefix(windowIDOrIndex, "pty-") && len(windowIDOrIndex) > 4 {
		return windowIDOrIndex, nil
	}

	// Try to parse as index
	index, err := strconv.Atoi(windowIDOrIndex)
	if err != nil {
		// Not a number, return as-is (might be a custom ID format)
		return windowIDOrIndex, nil
	}

	// Get all sessions to build window mapping
	sessions, err := s.client.ListSessions(s.ctx)
	if err != nil {
		return windowIDOrIndex, fmt.Errorf("failed to list sessions for window resolution: %w", err)
	}

	// Build index to UUID mapping
	windowIndexMap := make(map[int]string)
	for _, session := range sessions {
		windowIndexMap[int(session.WindowNumber)] = session.WindowID
	}

	// Look up the UUID for this index
	if uuid, exists := windowIndexMap[index]; exists {
		return uuid, nil
	}

	// Index not found, return original (might be valid in some contexts)
	return windowIDOrIndex, nil
}

// applyScopeFilter filters sessions based on scope settings
func (s *SharedListOperations) applyScopeFilter(sessions []*client.SessionInfo, scopeFlag string) []*client.SessionInfo {
	// Determine effective scope (flag overrides environment variable)
	effectiveScope := scopeFlag
	if effectiveScope == "" {
		effectiveScope = os.Getenv("IT2_SCOPE")
	}

	// If no scope set or scope is "none", return all sessions
	if effectiveScope == "" || effectiveScope == "none" {
		return sessions
	}

	// Get current session ID from environment
	currentSessionID := os.Getenv("ITERM_SESSION_ID")
	if currentSessionID != "" {
		currentSessionID = NormalizeSessionID(currentSessionID)
	}
	if currentSessionID == "" {
		// No current session, return all sessions
		return sessions
	}

	// Find the current session in the list
	var currentSession *client.SessionInfo
	for _, session := range sessions {
		if session.SessionID == currentSessionID {
			currentSession = session
			break
		}
	}

	if currentSession == nil {
		// Current session not found, return all sessions
		return sessions
	}

	// Apply scope filtering based on the effective scope
	var filteredSessions []*client.SessionInfo
	switch effectiveScope {
	case "window":
		// Filter to sessions in the same window
		for _, session := range sessions {
			if session.WindowID == currentSession.WindowID {
				filteredSessions = append(filteredSessions, session)
			}
		}
	case "tab":
		// Filter to sessions in the same tab
		for _, session := range sessions {
			if session.WindowID == currentSession.WindowID && session.TabID == currentSession.TabID {
				filteredSessions = append(filteredSessions, session)
			}
		}
	case "siblings":
		// Filter to sessions that share the same parent session
		for _, session := range sessions {
			if session.ParentSessionID == currentSession.ParentSessionID && session.ParentSessionID != "" {
				filteredSessions = append(filteredSessions, session)
			}
		}
	default:
		// Unknown scope, return all sessions
		return sessions
	}

	return filteredSessions
}

// enrichSessionsParallel runs plugins concurrently for each session and collects results
func (s *SharedListOperations) enrichSessionsParallel(sessions []*client.SessionInfo, enrichers []plugins.SessionEnricher) {
	if len(enrichers) == 0 {
		return
	}

	// Create a semaphore to limit concurrency (e.g., max 20 concurrent plugin executions)
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
			go func(sm *sessionMutex, enr plugins.SessionEnricher) {
				defer wg.Done()

				// Acquire semaphore to limit concurrency
				semaphore <- struct{}{}
				defer func() { <-semaphore }()

				// Execute the enricher
				data, err := enr.EnrichSession(s.ctx, sm.session)
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
