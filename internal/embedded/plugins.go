package embedded

import (
	"embed"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"runtime/debug"
	"time"
)

//go:embed plugins/*
var pluginsFS embed.FS

// GetPluginsDir returns the path where embedded plugins should be extracted
// Uses build hash to ensure plugins match the binary version
func GetPluginsDir() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed to get home directory: %w", err)
	}

	buildHash := getBuildHash()
	pluginsDir := filepath.Join(home, ".it2", "plugins", buildHash)
	return pluginsDir, nil
}

// ExtractPlugins extracts embedded plugins to ~/.it2/plugins/{build-hash}
// Returns the directory path where plugins were extracted
func ExtractPlugins() (string, error) {
	pluginsDir, err := GetPluginsDir()
	if err != nil {
		return "", err
	}

	// Create plugins directory if it doesn't exist
	if err := os.MkdirAll(pluginsDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create plugins directory: %w", err)
	}

	// Check if plugins are already extracted by looking for a marker file
	markerFile := filepath.Join(pluginsDir, ".extracted")
	if _, err := os.Stat(markerFile); err == nil {
		// Already extracted
		return pluginsDir, nil
	}

	// Extract all plugins
	err = fs.WalkDir(pluginsFS, "plugins", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if d.IsDir() {
			return nil
		}

		// Read embedded file
		content, err := pluginsFS.ReadFile(path)
		if err != nil {
			return fmt.Errorf("failed to read embedded file %s: %w", path, err)
		}

		// Get relative path (remove "plugins/" prefix)
		relPath, err := filepath.Rel("plugins", path)
		if err != nil {
			return err
		}

		// Write to destination
		destPath := filepath.Join(pluginsDir, relPath)
		if err := os.WriteFile(destPath, content, 0755); err != nil {
			return fmt.Errorf("failed to write plugin %s: %w", destPath, err)
		}

		return nil
	})

	if err != nil {
		return "", fmt.Errorf("failed to extract plugins: %w", err)
	}

	// Create marker file to indicate successful extraction
	if err := os.WriteFile(markerFile, []byte("extracted"), 0644); err != nil {
		return "", fmt.Errorf("failed to create marker file: %w", err)
	}

	// Clean up old plugin directories (keep only current one)
	cleanupOldPluginDirs(pluginsDir)

	return pluginsDir, nil
}

// cleanupOldPluginDirs removes old plugin directories except recent ones
// Keeps the current directory plus any accessed in the last 7 days to support
// concurrent use of different it2 versions
func cleanupOldPluginDirs(currentDir string) {
	pluginsBase := filepath.Dir(currentDir)
	entries, err := os.ReadDir(pluginsBase)
	if err != nil {
		return // Silently fail cleanup
	}

	currentHash := filepath.Base(currentDir)
	now := time.Now()
	sevenDaysAgo := now.Add(-7 * 24 * time.Hour)

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		if entry.Name() == currentHash {
			continue // Don't remove current directory
		}

		oldDir := filepath.Join(pluginsBase, entry.Name())

		// Check last access time of the directory
		info, err := os.Stat(oldDir)
		if err != nil {
			continue
		}

		// Only remove directories not accessed in the last 7 days
		// This allows concurrent use of different it2 versions
		if info.ModTime().Before(sevenDaysAgo) {
			os.RemoveAll(oldDir) // Ignore errors
		}
	}
}

// getBuildHash returns a version-aware directory name for plugin extraction
// Format uses module version from BuildInfo which includes timestamp and commit
// Example: v0.0.0-20251001042908-d843f51e2780+dirty
func getBuildHash() string {
	info, ok := debug.ReadBuildInfo()
	if !ok {
		// Fallback to a default if build info not available
		return "dev"
	}

	// Use the full module version which includes pseudo-version format
	if info.Main.Version != "" && info.Main.Version != "(devel)" {
		return info.Main.Version
	}

	// Fallback to VCS revision if no version
	for _, setting := range info.Settings {
		if setting.Key == "vcs.revision" {
			revision := setting.Value
			if len(revision) > 8 {
				return revision[:8]
			}
			return revision
		}
	}

	// Fallback to dev if no VCS info
	return "dev"
}
