package cmdutil

import (
	"context"
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/tmc/it2/internal/client"
	"github.com/tmc/it2/internal/formatting"
	"github.com/tmc/it2/internal/plugins"
	"github.com/tmc/it2/internal/scope"
	pb "github.com/tmc/it2/proto"
)

// SharedListOptions contains common options for list operations
type SharedListOptions struct {
	WindowID      string
	TabID         string
	SessionID     string
	ScopeFlag     string
	Format        string
	Columns       []string
	SortBy        string
	SortReverse   bool
	Quiet         bool
	NoHyperlinks  bool
	PluginPattern *regexp.Regexp
	SkipPlugins   bool
	IncludeBuried bool
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
	// Get raw response for tree format (need the full split tree structure)
	var rawResp interface{}
	if opts.Format == "tree" {
		resp, err := s.client.ListSessionsRaw(s.ctx)
		if err != nil {
			return fmt.Errorf("failed to list sessions: %w", err)
		}
		rawResp = resp
	}

	var sessions []*client.SessionInfo
	var err error
	if opts.IncludeBuried {
		sessions, err = s.client.ListSessionsWithOptions(s.ctx, client.ListSessionsOptions{IncludeBuried: true})
	} else {
		sessions, err = s.client.ListSessions(s.ctx)
	}
	if err != nil {
		return fmt.Errorf("failed to list sessions: %w", err)
	}

	// Apply scope filtering first
	scopedSessions := scope.Filter(sessions, opts.ScopeFlag)

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

	// Apply plugin enrichment - only if enabled and columns require it
	if !opts.SkipPlugins && plugins.ShouldRunPlugins(opts.Columns) {
		registry := plugins.NewRegistry()
		if err := registry.DiscoverAndRegister(); err == nil {
			enrichers := registry.GetEnrichers()
			if opts.PluginPattern != nil {
				enrichers = plugins.FilterEnrichersByPattern(enrichers, opts.PluginPattern)
			}
			activeEnrichers := plugins.FilterEnrichers(enrichers, opts.Columns)

			if len(activeEnrichers) > 0 {
				plugins.EnrichSessionsParallel(s.ctx, filteredSessions, activeEnrichers)

				// Save metrics after enrichment
				plugins.GetMetricsStore().Save()
			}
		}
	}

	// Format and output
	// Disable hyperlinks by default
	enableHyperlinks := false
	formatter := formatting.NewWithHyperlinks(opts.Format, opts.Columns, opts.SortBy, opts.SortReverse, opts.Quiet, enableHyperlinks)

	// Use raw tree structure for tree format
	if opts.Format == "tree" && rawResp != nil {
		if pbResp, ok := rawResp.(*pb.ListSessionsResponse); ok {
			return formatter.FormatSessionsWithRawTree(pbResp, filteredSessions)
		}
	}

	return formatter.FormatSessions(filteredSessions)
}

// ListTabs lists tabs with optional filtering by window
func (s *SharedListOperations) ListTabs(opts SharedListOptions) error {
	// Get sessions to build tab information
	sessions, err := s.client.ListSessions(s.ctx)
	if err != nil {
		return fmt.Errorf("failed to list sessions: %w", err)
	}

	// Get the currently selected tab to mark it as active
	selectedTabID := ""
	if focusInfo, err := s.client.GetFocus(s.ctx); err == nil {
		selectedTabID = focusInfo.TabId
	}

	// Build tab info from sessions
	tabInfos := s.buildTabInfoFromSessions(sessions, opts.WindowID, selectedTabID)

	// Apply plugin enrichment unless disabled
	if !opts.SkipPlugins {
		registry := plugins.NewRegistry()
		if err := registry.DiscoverAndRegister(); err == nil {
			enrichers := registry.GetTabEnrichers()
			plugins.EnrichTabs(s.ctx, tabInfos, enrichers)
		}
	}

	// Format and output
	// Disable hyperlinks by default
	enableHyperlinks := false
	formatter := formatting.NewWithHyperlinks(opts.Format, opts.Columns, opts.SortBy, opts.SortReverse, opts.Quiet, enableHyperlinks)
	return formatter.FormatTabs(tabInfos)
}

// ListWindows lists all windows
func (s *SharedListOperations) ListWindows(opts SharedListOptions) error {
	windows, err := s.client.ListWindows(s.ctx)
	if err != nil {
		return fmt.Errorf("failed to list windows: %w", err)
	}

	// Get sessions to calculate session counts per window
	sessions, err := s.client.ListSessions(s.ctx)
	if err == nil {
		// Count sessions per window
		sessionCounts := make(map[string]int)
		for _, session := range sessions {
			sessionCounts[session.WindowID]++
		}
		// Update window session counts
		for _, window := range windows {
			window.SessionCount = sessionCounts[window.WindowID]
		}
	}

	// Apply plugins to windows unless disabled
	if !opts.SkipPlugins {
		registry := plugins.NewRegistry()
		if err := registry.DiscoverAndRegister(); err == nil {
			windowEnrichers := registry.GetWindowEnrichers()
			plugins.EnrichWindows(s.ctx, windows, windowEnrichers)
		}
	}

	// Disable hyperlinks by default
	enableHyperlinks := false
	formatter := formatting.NewWithHyperlinks(opts.Format, opts.Columns, opts.SortBy, opts.SortReverse, opts.Quiet, enableHyperlinks)
	return formatter.FormatClientWindows(windows)
}

// buildTabInfoFromSessions builds tab information from session data
func (s *SharedListOperations) buildTabInfoFromSessions(sessions []*client.SessionInfo, windowID string, selectedTabID string) []*formatting.TabInfo {
	tabMap := make(map[string]*formatting.TabInfo)

	for _, session := range sessions {
		// Filter by window if specified
		if windowID != "" && session.WindowID != windowID {
			continue
		}

		// Create or update tab info
		key := session.WindowID + ":" + session.TabID
		if tab, exists := tabMap[key]; !exists {
			// Use TabTitle if available, otherwise use first session's name
			title := session.TabTitle
			if title == "" && session.SessionName != "" {
				title = session.SessionName
			}
			// Check if this tab is the selected one
			isActive := (session.TabID == selectedTabID)
			tabMap[key] = &formatting.TabInfo{
				WindowID:     session.WindowID,
				WindowNumber: int(session.WindowNumber),
				TabID:        session.TabID,
				Title:        title,
				Position:     0,       // Position needs to be determined differently
				Active:       isActive, // Set based on selected tab
				Sessions:     []*client.SessionInfo{session},
			}
		} else {
			// Add session to existing tab
			tab.Sessions = append(tab.Sessions, session)
			// Update title preference: TabTitle > first SessionName
			if tab.Title == "" {
				if session.TabTitle != "" {
					tab.Title = session.TabTitle
				} else if session.SessionName != "" {
					tab.Title = session.SessionName
				}
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
