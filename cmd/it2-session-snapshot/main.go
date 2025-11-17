package main

import (
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
)

var (
	sessionID = flag.String("session", "", "iTerm2 session ID (uses $ITERM_SESSION_ID if not provided)")
	parallel  = flag.Bool("parallel", true, "Run snapshot plugins in parallel")
	verbose   = flag.Bool("verbose", false, "Verbose output")
)

func main() {
	flag.Parse()

	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func run() error {
	// Get session ID
	sid := *sessionID
	if sid == "" && flag.NArg() > 0 {
		sid = flag.Arg(0)
	}
	if sid == "" {
		sid = os.Getenv("ITERM_SESSION_ID")
	}
	if sid == "" {
		return fmt.Errorf("session ID required (provide as argument, -session flag, or set ITERM_SESSION_ID)")
	}

	// Discover snapshot plugins
	plugins, err := discoverSnapshotPlugins()
	if err != nil {
		return fmt.Errorf("failed to discover snapshot plugins: %w", err)
	}

	if len(plugins) == 0 {
		if *verbose {
			fmt.Println("No snapshot plugins found")
		}
		return nil
	}

	if *verbose {
		fmt.Printf("Found %d snapshot plugin(s):\n", len(plugins))
		for _, p := range plugins {
			fmt.Printf("  - %s\n", filepath.Base(p))
		}
	}

	// Run plugins
	if *parallel {
		return runPluginsParallel(plugins, sid)
	}
	return runPluginsSequential(plugins, sid)
}

// discoverSnapshotPlugins finds all it2-session-snapshot-* executables in PATH
func discoverSnapshotPlugins() ([]string, error) {
	var plugins []string
	seen := make(map[string]bool)

	// Get PATH
	pathEnv := os.Getenv("PATH")
	if pathEnv == "" {
		return plugins, nil
	}

	// Search each directory in PATH
	for _, dir := range strings.Split(pathEnv, string(os.PathListSeparator)) {
		entries, err := os.ReadDir(dir)
		if err != nil {
			continue
		}

		for _, entry := range entries {
			if entry.IsDir() {
				continue
			}

			name := entry.Name()
			if !strings.HasPrefix(name, "it2-session-snapshot-") {
				continue
			}

			// Check if executable
			fullPath := filepath.Join(dir, name)
			info, err := os.Stat(fullPath)
			if err != nil {
				continue
			}
			if info.Mode()&0111 == 0 {
				continue
			}

			// Avoid duplicates (first in PATH wins)
			if seen[name] {
				continue
			}
			seen[name] = true

			plugins = append(plugins, fullPath)
		}
	}

	return plugins, nil
}

// runPluginsSequential runs plugins one at a time
func runPluginsSequential(plugins []string, sessionID string) error {
	for _, plugin := range plugins {
		if err := runPlugin(plugin, sessionID); err != nil {
			return fmt.Errorf("%s failed: %w", filepath.Base(plugin), err)
		}
	}
	return nil
}

// runPluginsParallel runs plugins concurrently
func runPluginsParallel(plugins []string, sessionID string) error {
	var wg sync.WaitGroup
	errChan := make(chan error, len(plugins))

	for _, plugin := range plugins {
		wg.Add(1)
		go func(p string) {
			defer wg.Done()
			if err := runPlugin(p, sessionID); err != nil {
				errChan <- fmt.Errorf("%s failed: %w", filepath.Base(p), err)
			}
		}(plugin)
	}

	wg.Wait()
	close(errChan)

	// Collect any errors
	var errs []error
	for err := range errChan {
		errs = append(errs, err)
	}

	if len(errs) > 0 {
		return fmt.Errorf("snapshot plugins failed: %v", errs)
	}

	return nil
}

// runPlugin executes a single snapshot plugin
func runPlugin(pluginPath, sessionID string) error {
	if *verbose {
		fmt.Printf("Running %s...\n", filepath.Base(pluginPath))
	}

	cmd := exec.Command(pluginPath, sessionID)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Env = append(os.Environ(), "ITERM_SESSION_ID="+sessionID)

	return cmd.Run()
}
