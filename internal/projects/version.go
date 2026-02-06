package projects

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"runtime/debug"
	"time"
)

// VersionInfo captures version and build metadata about the it2 binary.
// This is written to project directories to track what version created/modified them.
type VersionInfo struct {
	// Core version info
	Version   string `json:"version,omitempty"` // From build tags or module version
	GoVersion string `json:"go_version"`        // e.g., "go1.22.0"
	OS        string `json:"os"`                // e.g., "darwin"
	Arch      string `json:"arch"`              // e.g., "arm64"

	// Build metadata (from debug.BuildInfo)
	ModulePath    string            `json:"module_path,omitempty"`    // e.g., "github.com/tmc/it2"
	ModuleVersion string            `json:"module_version,omitempty"` // e.g., "v0.1.0" or "(devel)"
	VCS           string            `json:"vcs,omitempty"`            // e.g., "git"
	VCSRevision   string            `json:"vcs_revision,omitempty"`   // Git commit hash
	VCSTime       string            `json:"vcs_time,omitempty"`       // Commit timestamp
	VCSModified   bool              `json:"vcs_modified,omitempty"`   // True if uncommitted changes
	BuildSettings map[string]string `json:"build_settings,omitempty"` // CGO, tags, etc.

	// Runtime context
	WrittenAt time.Time `json:"written_at"`
	WrittenBy string    `json:"written_by,omitempty"` // Hostname or user context
}

// GetVersionInfo returns the current binary's version info.
func GetVersionInfo() *VersionInfo {
	info := &VersionInfo{
		GoVersion: runtime.Version(),
		OS:        runtime.GOOS,
		Arch:      runtime.GOARCH,
		WrittenAt: time.Now(),
	}

	// Try to get hostname
	if host, err := os.Hostname(); err == nil {
		info.WrittenBy = host
	}

	// Get build info from runtime/debug
	if bi, ok := debug.ReadBuildInfo(); ok {
		info.ModulePath = bi.Main.Path
		info.ModuleVersion = bi.Main.Version

		// Parse build settings for VCS info
		info.BuildSettings = make(map[string]string)
		for _, setting := range bi.Settings {
			switch setting.Key {
			case "vcs":
				info.VCS = setting.Value
			case "vcs.revision":
				info.VCSRevision = setting.Value
			case "vcs.time":
				info.VCSTime = setting.Value
			case "vcs.modified":
				info.VCSModified = setting.Value == "true"
			default:
				// Store other interesting settings
				if setting.Key == "-tags" || setting.Key == "CGO_ENABLED" ||
					setting.Key == "GOAMD64" || setting.Key == "GOARM" {
					info.BuildSettings[setting.Key] = setting.Value
				}
			}
		}
	}

	return info
}

// ShortVersion returns a human-readable short version string.
func (v *VersionInfo) ShortVersion() string {
	version := v.ModuleVersion
	if version == "" || version == "(devel)" {
		if v.VCSRevision != "" {
			version = v.VCSRevision[:min(8, len(v.VCSRevision))]
			if v.VCSModified {
				version += "-dirty"
			}
		} else {
			version = "unknown"
		}
	}
	return version
}

const versionFileName = ".it2-version.json"

// WriteVersionInfo writes version info to a project directory.
// This allows tracking which version of it2 created/modified the project.
func WriteVersionInfo(projectPath string) error {
	dir, err := ProjectDir(projectPath)
	if err != nil {
		return err
	}

	info := GetVersionInfo()
	data, err := json.MarshalIndent(info, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal version info: %w", err)
	}

	versionPath := filepath.Join(dir, versionFileName)
	if err := os.WriteFile(versionPath, data, 0644); err != nil {
		return fmt.Errorf("write version info: %w", err)
	}

	return nil
}

// ReadVersionInfo reads version info from a project directory.
// Returns nil if no version file exists.
func ReadVersionInfo(projectPath string) (*VersionInfo, error) {
	projectsDir, err := ProjectsDir()
	if err != nil {
		return nil, err
	}

	name := PathToProjectName(projectPath)
	versionPath := filepath.Join(projectsDir, name, versionFileName)

	data, err := os.ReadFile(versionPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("read version info: %w", err)
	}

	var info VersionInfo
	if err := json.Unmarshal(data, &info); err != nil {
		return nil, fmt.Errorf("parse version info: %w", err)
	}

	return &info, nil
}

// EnsureVersionInfo writes version info if it doesn't exist or is outdated.
// "Outdated" means the VCS revision has changed.
func EnsureVersionInfo(projectPath string) error {
	current := GetVersionInfo()
	existing, err := ReadVersionInfo(projectPath)
	if err != nil {
		return err
	}

	// Write if no existing version or if revision changed
	if existing == nil || existing.VCSRevision != current.VCSRevision {
		return WriteVersionInfo(projectPath)
	}

	return nil
}
