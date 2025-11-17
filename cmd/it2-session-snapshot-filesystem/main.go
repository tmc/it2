package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"time"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func run() error {
	// Get session ID
	var sessionID string
	if len(os.Args) >= 2 {
		sessionID = os.Args[1]
	} else {
		sessionID = os.Getenv("ITERM_SESSION_ID")
	}
	if sessionID == "" {
		return fmt.Errorf("session ID required")
	}

	// Get current working directory
	cwd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get current directory: %w", err)
	}

	// Get home directory for session artifacts
	home, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("failed to get home directory: %w", err)
	}

	// Create snapshot directory
	snapshotDir := filepath.Join(home, ".it2", "sessions", sessionID, "snapshots", "filesystem")
	if err := os.MkdirAll(snapshotDir, 0755); err != nil {
		return fmt.Errorf("failed to create snapshot directory: %w", err)
	}

	// Generate timestamp for filenames
	timestamp := time.Now().Format("20060102-150405")

	// Capture directory tree
	treeFile := filepath.Join(snapshotDir, fmt.Sprintf("tree-%s.txt", timestamp))
	if err := captureTree(cwd, treeFile); err != nil {
		// Tree might not be installed, that's okay
		_ = err
	}

	// Capture file listing with details
	lsFile := filepath.Join(snapshotDir, fmt.Sprintf("ls-%s.txt", timestamp))
	if err := captureLS(cwd, lsFile); err != nil {
		return fmt.Errorf("failed to capture file listing: %w", err)
	}

	// Capture git status if in a git repo
	if isGitRepo(cwd) {
		gitStatusFile := filepath.Join(snapshotDir, fmt.Sprintf("git-status-%s.txt", timestamp))
		if err := captureGitStatus(cwd, gitStatusFile); err != nil {
			// Git command might fail, that's okay
			_ = err
		}

		gitDiffFile := filepath.Join(snapshotDir, fmt.Sprintf("git-diff-%s.txt", timestamp))
		if err := captureGitDiff(cwd, gitDiffFile); err != nil {
			_ = err
		}
	}

	// Create a metadata file
	metadataFile := filepath.Join(snapshotDir, fmt.Sprintf("metadata-%s.txt", timestamp))
	if err := writeMetadata(metadataFile, sessionID, cwd); err != nil {
		return fmt.Errorf("failed to write metadata: %w", err)
	}

	return nil
}

// isGitRepo checks if directory is in a git repository
func isGitRepo(dir string) bool {
	cmd := exec.Command("git", "rev-parse", "--git-dir")
	cmd.Dir = dir
	err := cmd.Run()
	return err == nil
}

// captureTree captures directory tree structure
func captureTree(dir, outputFile string) error {
	f, err := os.Create(outputFile)
	if err != nil {
		return err
	}
	defer f.Close()

	cmd := exec.Command("tree", "-a", "-L", "3", "--gitignore")
	cmd.Dir = dir
	cmd.Stdout = f
	cmd.Stderr = f

	return cmd.Run()
}

// captureLS captures detailed file listing
func captureLS(dir, outputFile string) error {
	f, err := os.Create(outputFile)
	if err != nil {
		return err
	}
	defer f.Close()

	cmd := exec.Command("ls", "-laR")
	cmd.Dir = dir
	cmd.Stdout = f
	cmd.Stderr = f

	return cmd.Run()
}

// captureGitStatus captures git status
func captureGitStatus(dir, outputFile string) error {
	f, err := os.Create(outputFile)
	if err != nil {
		return err
	}
	defer f.Close()

	// Write status
	statusCmd := exec.Command("git", "status", "--verbose")
	statusCmd.Dir = dir
	statusCmd.Stdout = f
	statusCmd.Stderr = f
	if err := statusCmd.Run(); err != nil {
		return err
	}

	// Write branch info
	fmt.Fprintln(f, "\n=== Branch Info ===")
	branchCmd := exec.Command("git", "branch", "-vv")
	branchCmd.Dir = dir
	branchCmd.Stdout = f
	branchCmd.Stderr = f
	branchCmd.Run()

	// Write recent log
	fmt.Fprintln(f, "\n=== Recent Commits ===")
	logCmd := exec.Command("git", "log", "--oneline", "-10")
	logCmd.Dir = dir
	logCmd.Stdout = f
	logCmd.Stderr = f
	logCmd.Run()

	return nil
}

// captureGitDiff captures git diff
func captureGitDiff(dir, outputFile string) error {
	f, err := os.Create(outputFile)
	if err != nil {
		return err
	}
	defer f.Close()

	// Staged diff
	fmt.Fprintln(f, "=== Staged Changes ===")
	stagedCmd := exec.Command("git", "diff", "--cached")
	stagedCmd.Dir = dir
	stagedCmd.Stdout = f
	stagedCmd.Stderr = f
	stagedCmd.Run()

	// Unstaged diff
	fmt.Fprintln(f, "\n=== Unstaged Changes ===")
	unstagedCmd := exec.Command("git", "diff")
	unstagedCmd.Dir = dir
	unstagedCmd.Stdout = f
	unstagedCmd.Stderr = f
	unstagedCmd.Run()

	return nil
}

// writeMetadata writes snapshot metadata
func writeMetadata(outputFile, sessionID, cwd string) error {
	f, err := os.Create(outputFile)
	if err != nil {
		return err
	}
	defer f.Close()

	fmt.Fprintf(f, "Session ID: %s\n", sessionID)
	fmt.Fprintf(f, "Working Directory: %s\n", cwd)
	fmt.Fprintf(f, "Timestamp: %s\n", time.Now().Format(time.RFC3339))
	fmt.Fprintf(f, "Hostname: %s\n", getHostname())
	fmt.Fprintf(f, "User: %s\n", os.Getenv("USER"))

	return nil
}

// getHostname returns the system hostname
func getHostname() string {
	hostname, err := os.Hostname()
	if err != nil {
		return "unknown"
	}
	return hostname
}
