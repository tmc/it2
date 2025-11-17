package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func run() error {
	// Get home directory
	home, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("failed to get home directory: %w", err)
	}

	// Work directly in ~/.claude
	claudeDir := filepath.Join(home, ".claude")

	// Check if .claude directory exists
	if _, err := os.Stat(claudeDir); os.IsNotExist(err) {
		// No .claude directory, nothing to snapshot
		return nil
	}

	// Initialize git repo if needed
	gitDir := filepath.Join(claudeDir, ".git")
	if _, err := os.Stat(gitDir); os.IsNotExist(err) {
		if err := initGitRepo(claudeDir); err != nil {
			return fmt.Errorf("failed to initialize git repo: %w", err)
		}
	}

	// Create git commit with simple message
	if err := gitCommit(claudeDir, "claude: snapshot"); err != nil {
		return fmt.Errorf("failed to commit snapshot: %w", err)
	}

	return nil
}

// initGitRepo initializes a new git repository
func initGitRepo(dir string) error {
	cmd := exec.Command("git", "init")
	cmd.Dir = dir
	return cmd.Run()
}

// gitCommit creates a git commit with 'git add . && git commit -m'
func gitCommit(dir, message string) error {
	// Add all files
	addCmd := exec.Command("git", "add", ".")
	addCmd.Dir = dir
	if err := addCmd.Run(); err != nil {
		return err
	}

	// Check if there are changes
	statusCmd := exec.Command("git", "status", "--porcelain")
	statusCmd.Dir = dir
	output, err := statusCmd.Output()
	if err != nil {
		return err
	}

	// No changes, nothing to commit
	if len(output) == 0 {
		return nil
	}

	// Create commit
	commitCmd := exec.Command("git", "commit", "-m", message)
	commitCmd.Dir = dir
	return commitCmd.Run()
}
