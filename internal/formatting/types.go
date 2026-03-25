package formatting

import (
	"time"

	"github.com/tmc/it2/internal/client"
)

const (
	colorReset      = "\033[0m"
	colorBold       = "\033[1m"
	colorBrightCyan = "\033[1;36m" // Bright cyan for path
	colorGreen      = "\033[1;32m" // Bright green for current session
	colorCyan       = "\033[36m"   // Regular cyan for ancestors
)

// Formatter handles formatting of iTerm2 data structures for display.
type Formatter struct {
	format            string
	columns           []string
	sortBy            string
	sortReverse       bool
	quiet             bool
	hyperlinks        bool // Enable OSC 8 terminal hyperlinks
	wide              bool // Show full IDs instead of short IDs
	includeEmptyLines bool // Include trailing empty lines in buffer output
}

// New creates a new formatter with the specified format.
func New(format string) *Formatter {
	return &Formatter{format: format}
}

// NewWithOptions creates a new formatter with column and sort options.
func NewWithOptions(format string, columns []string, sortBy string, sortReverse bool, quiet bool) *Formatter {
	// Disable hyperlinks by default to avoid conflicts with user's Semantic History
	return NewWithHyperlinks(format, columns, sortBy, sortReverse, quiet, false)
}

// NewWithHyperlinks creates a new formatter with full options including hyperlinks.
func NewWithHyperlinks(format string, columns []string, sortBy string, sortReverse bool, quiet bool, hyperlinks bool) *Formatter {
	return &Formatter{
		format:      format,
		columns:     columns,
		sortBy:      sortBy,
		sortReverse: sortReverse,
		quiet:       quiet,
		hyperlinks:  hyperlinks,
	}
}

// GetFormat returns the current format string.
func (f *Formatter) GetFormat() string {
	return f.format
}

// SetWide sets whether to show full IDs instead of short IDs.
func (f *Formatter) SetWide(wide bool) {
	f.wide = wide
}

// SetIncludeEmptyLines sets whether to include trailing empty lines in buffer output.
func (f *Formatter) SetIncludeEmptyLines(include bool) {
	f.includeEmptyLines = include
}

// WindowInfo represents window information for formatting.
type WindowInfo struct {
	WindowID     string                 `json:"window_id"`
	Name         string                 `json:"name,omitempty"`
	Title        string                 `json:"title,omitempty"`
	Frame        string                 `json:"frame,omitempty"`
	Fullscreen   string                 `json:"fullscreen,omitempty"`
	Miniaturized string                 `json:"miniaturized,omitempty"`
	TabCount     int                    `json:"tab_count"`
	PluginData   map[string]interface{} `json:"plugin_data,omitempty"`
}

// TabInfo represents tab information for formatting.
type TabInfo struct {
	TabID        string                 `json:"tab_id"`
	WindowID     string                 `json:"window_id"`
	WindowNumber int                    `json:"window_number"`
	Title        string                 `json:"title,omitempty"`
	Active       bool                   `json:"active"`
	Position     int                    `json:"position"`
	Sessions     []*client.SessionInfo  `json:"sessions,omitempty"`
	PluginData   map[string]interface{} `json:"plugin_data,omitempty"`
}

// Annotation represents an annotation for formatting.
type Annotation struct {
	ID        string    `json:"id"`
	Text      string    `json:"text"`
	Line      int       `json:"line,omitempty"`
	Column    int       `json:"column,omitempty"`
	Timestamp time.Time `json:"timestamp"`
	Type      string    `json:"type,omitempty"`
}

// Trigger represents a trigger for formatting.
type Trigger struct {
	ID      string    `json:"id"`
	Pattern string    `json:"pattern"`
	Action  string    `json:"action"`
	Enabled bool      `json:"enabled"`
	Created time.Time `json:"created"`
}

// TreeNode represents a session in the tree hierarchy.
type TreeNode struct {
	SessionID      string
	ShortID        string
	Name           string
	ParentID       string
	SplitVertical  *bool // Direction of split from parent (nil if no parent, true=vertical, false=horizontal)
	Children       []*TreeNode
	CurrentCommand string
	ShellPID       int32
	JobPID         int32
	PromptState    string
	PluginData     map[string]interface{}
}

// SplitTreeNode represents a node in the panel hierarchy (can be a session or a split container).
type SplitTreeNode struct {
	IsSession      bool                   // True if this is a session, false if it's a split container
	IsVertical     bool                   // Direction of split (only relevant if IsSession=false)
	SessionID      string                 // Only set if IsSession=true
	ShortID        string                 // Only set if IsSession=true
	Name           string                 // Only set if IsSession=true
	CurrentCommand string                 // Only set if IsSession=true
	ShellPID       int32                  // Only set if IsSession=true
	JobPID         int32                  // Only set if IsSession=true
	PromptState    string                 // Only set if IsSession=true
	PluginData     map[string]interface{} // Only set if IsSession=true
	Children       []*SplitTreeNode       // Child nodes (either sessions or more splits)
}
