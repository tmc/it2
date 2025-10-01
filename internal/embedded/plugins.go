package embedded

import (
	"crypto/sha256"
	"embed"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"runtime/debug"
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

	return pluginsDir, nil
}

// getBuildHash returns a hash of the build info to use as plugin directory name
func getBuildHash() string {
	info, ok := debug.ReadBuildInfo()
	if !ok {
		// Fallback to a default if build info not available
		return "dev"
	}

	// Create hash from build info
	h := sha256.New()
	h.Write([]byte(info.Main.Path))
	h.Write([]byte(info.Main.Version))

	// Include VCS info if available
	for _, setting := range info.Settings {
		if setting.Key == "vcs.revision" || setting.Key == "vcs.time" {
			h.Write([]byte(setting.Value))
		}
	}

	return fmt.Sprintf("%x", h.Sum(nil))[:12]
}
