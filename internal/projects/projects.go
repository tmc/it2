// Package projects provides helpers for managing project-specific directories
// and data, following a similar pattern to Claude Code's ~/.claude/projects layout.
//
// Project directories are named by converting absolute paths to directory names
// where slashes are replaced with dashes. For example:
//
//	/Users/tmc/code/myapp → -Users-tmc-code-myapp
//
// This allows storing session data, artifacts, and metadata per-project.
package projects

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// BaseDir returns the base directory for it2 data (~/.it2).
func BaseDir() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("get home dir: %w", err)
	}
	return filepath.Join(home, ".it2"), nil
}

// ProjectsDir returns the projects directory (~/.it2/projects).
func ProjectsDir() (string, error) {
	base, err := BaseDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(base, "projects"), nil
}

// PathToProjectName converts an absolute path to a project directory name.
// Slashes are replaced with dashes, preserving the leading dash for absolute paths.
//
// Examples:
//
//	/Users/tmc/code/myapp    → -Users-tmc-code-myapp
//	/Volumes/tmc/go/src/...  → -Volumes-tmc-go-src-...
//	relative/path            → relative-path
func PathToProjectName(path string) string {
	// Clean the path first
	path = filepath.Clean(path)

	// Replace path separators with dashes
	name := strings.ReplaceAll(path, string(filepath.Separator), "-")

	return name
}

// ProjectNameToPath converts a project directory name back to the original path.
func ProjectNameToPath(name string) string {
	// Replace dashes with path separators
	path := strings.ReplaceAll(name, "-", string(filepath.Separator))
	return path
}

// ProjectDir returns the directory for a specific project path.
// Creates the directory if it doesn't exist.
func ProjectDir(projectPath string) (string, error) {
	projectsDir, err := ProjectsDir()
	if err != nil {
		return "", err
	}

	// Convert path to project name
	name := PathToProjectName(projectPath)
	dir := filepath.Join(projectsDir, name)

	// Create directory if needed
	if err := os.MkdirAll(dir, 0700); err != nil {
		return "", fmt.Errorf("create project dir: %w", err)
	}

	return dir, nil
}

// CurrentProjectDir returns the directory for the current working directory.
func CurrentProjectDir() (string, error) {
	cwd, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("get cwd: %w", err)
	}
	return ProjectDir(cwd)
}

// SessionInfo represents metadata about a session within a project.
type SessionInfo struct {
	ID           string    `json:"id"`
	Name         string    `json:"name,omitempty"`
	StartedAt    time.Time `json:"started_at"`
	LastActiveAt time.Time `json:"last_active_at"`
	ITerm2SID    string    `json:"iterm2_sid,omitempty"`
	Badge        string    `json:"badge,omitempty"`
	Role         string    `json:"role,omitempty"` // e.g., "research", "implement", "oracle"
}

// SessionsIndex manages the index of sessions for a project.
type SessionsIndex struct {
	Sessions []SessionInfo `json:"sessions"`
	path     string
}

// LoadSessionsIndex loads or creates a sessions index for a project.
func LoadSessionsIndex(projectPath string) (*SessionsIndex, error) {
	dir, err := ProjectDir(projectPath)
	if err != nil {
		return nil, err
	}

	indexPath := filepath.Join(dir, "sessions-index.json")
	idx := &SessionsIndex{
		Sessions: []SessionInfo{},
		path:     indexPath,
	}

	data, err := os.ReadFile(indexPath)
	if err != nil {
		if os.IsNotExist(err) {
			return idx, nil
		}
		return nil, fmt.Errorf("read sessions index: %w", err)
	}

	if err := json.Unmarshal(data, &idx.Sessions); err != nil {
		return nil, fmt.Errorf("parse sessions index: %w", err)
	}

	return idx, nil
}

// Save writes the sessions index to disk.
func (idx *SessionsIndex) Save() error {
	data, err := json.MarshalIndent(idx.Sessions, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal sessions index: %w", err)
	}

	if err := os.WriteFile(idx.path, data, 0600); err != nil {
		return fmt.Errorf("write sessions index: %w", err)
	}

	return nil
}

// AddSession adds a session to the index.
func (idx *SessionsIndex) AddSession(session SessionInfo) {
	// Update if exists
	for i, s := range idx.Sessions {
		if s.ID == session.ID {
			idx.Sessions[i] = session
			return
		}
	}
	// Add new
	idx.Sessions = append(idx.Sessions, session)
}

// GetSession returns a session by ID.
func (idx *SessionsIndex) GetSession(id string) *SessionInfo {
	for i := range idx.Sessions {
		if idx.Sessions[i].ID == id {
			return &idx.Sessions[i]
		}
	}
	return nil
}

// UpdateLastActive updates the last active time for a session.
func (idx *SessionsIndex) UpdateLastActive(id string) {
	for i := range idx.Sessions {
		if idx.Sessions[i].ID == id {
			idx.Sessions[i].LastActiveAt = time.Now()
			return
		}
	}
}

// ListProjects returns all project directories.
func ListProjects() ([]string, error) {
	projectsDir, err := ProjectsDir()
	if err != nil {
		return nil, err
	}

	entries, err := os.ReadDir(projectsDir)
	if err != nil {
		if os.IsNotExist(err) {
			return []string{}, nil
		}
		return nil, fmt.Errorf("read projects dir: %w", err)
	}

	var projects []string
	for _, entry := range entries {
		if entry.IsDir() {
			projects = append(projects, ProjectNameToPath(entry.Name()))
		}
	}

	return projects, nil
}

// SessionArtifactsDir returns the directory for session artifacts.
func SessionArtifactsDir(projectPath, sessionID string) (string, error) {
	dir, err := ProjectDir(projectPath)
	if err != nil {
		return "", err
	}

	artifactsDir := filepath.Join(dir, sessionID)
	if err := os.MkdirAll(artifactsDir, 0700); err != nil {
		return "", fmt.Errorf("create artifacts dir: %w", err)
	}

	return artifactsDir, nil
}
