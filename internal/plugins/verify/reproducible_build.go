// Package verify implements reproducible build verification for it2 plugins.
//
// Layer 1.5: Verify that a plugin binary matches its claimed source code
// by performing a reproducible build and comparing hashes.
//
// This catches supply chain attacks where:
// - Author's build environment is compromised (XZ backdoor scenario)
// - Binary is modified during distribution
// - Different users receive different binaries
package verify

import (
	"crypto/sha256"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// BuildSource describes where to get plugin source code.
type BuildSource struct {
	Repo      string            // Git repository URL
	Commit    string            // Specific commit hash
	GoVersion string            // Required Go version
	BuildCmd  string            // Build command (must be deterministic)
	BuildEnv  map[string]string // Environment variables for build
}

// BinaryInfo describes the expected plugin binary.
type BinaryInfo struct {
	SHA256    string // Expected SHA256 hash
	SizeBytes int64  // Expected file size
	Platform  string // Expected platform (e.g., "darwin/arm64")
}

// VerificationResult contains the outcome of reproducible build verification.
type VerificationResult struct {
	Success      bool   // True if hashes match
	LocalHash    string // Hash of locally built binary
	ClaimedHash  string // Hash from manifest
	ErrorMessage string // Error details if verification failed
}

// VerifyReproducibleBuild performs a reproducible build verification.
//
// Process:
// 1. Clone source repository at specific commit
// 2. Build binary with deterministic flags
// 3. Hash the built binary
// 4. Compare with claimed hash from manifest
// 5. Return verification result
//
// This catches compromised build environments by ensuring the binary
// matches what can be built from the declared source.
func VerifyReproducibleBuild(source BuildSource, binary BinaryInfo) (*VerificationResult, error) {
	result := &VerificationResult{
		ClaimedHash: binary.SHA256,
	}

	// Create temporary directory for verification build
	tmpDir, err := os.MkdirTemp("", "it2-verify-*")
	if err != nil {
		return nil, fmt.Errorf("create temp dir: %w", err)
	}
	defer os.RemoveAll(tmpDir)

	// Step 1: Clone repository
	repoDir := filepath.Join(tmpDir, "repo")
	if err := cloneRepo(source.Repo, source.Commit, repoDir); err != nil {
		result.ErrorMessage = fmt.Sprintf("clone failed: %v", err)
		return result, nil // Return result, not error (allow caller to handle)
	}

	// Step 2: Verify Go version
	if err := verifyGoVersion(source.GoVersion); err != nil {
		result.ErrorMessage = fmt.Sprintf("Go version mismatch: %v", err)
		return result, nil
	}

	// Step 3: Build with deterministic flags
	binaryPath := filepath.Join(tmpDir, "plugin")
	if err := buildDeterministic(repoDir, binaryPath, source.BuildCmd, source.BuildEnv); err != nil {
		result.ErrorMessage = fmt.Sprintf("build failed: %v", err)
		return result, nil
	}

	// Step 4: Hash the built binary
	localHash, err := hashFile(binaryPath)
	if err != nil {
		result.ErrorMessage = fmt.Sprintf("hash failed: %v", err)
		return result, nil
	}
	result.LocalHash = localHash

	// Step 5: Compare hashes
	result.Success = (localHash == binary.SHA256)
	if !result.Success {
		result.ErrorMessage = fmt.Sprintf("hash mismatch: local=%s claimed=%s", localHash, binary.SHA256)
	}

	return result, nil
}

// cloneRepo clones a git repository at a specific commit.
func cloneRepo(repoURL, commit, destDir string) error {
	// Clone with depth=1 for speed (we only need one commit)
	cmd := exec.Command("git", "clone", "--depth=1", repoURL, destDir)
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("git clone failed: %w\n%s", err, output)
	}

	// Checkout specific commit
	cmd = exec.Command("git", "checkout", commit)
	cmd.Dir = destDir
	if _, err := cmd.CombinedOutput(); err != nil {
		// If depth=1 didn't get the commit, fetch full history
		fetchCmd := exec.Command("git", "fetch", "--unshallow")
		fetchCmd.Dir = destDir
		fetchCmd.Run()

		// Try checkout again
		cmd = exec.Command("git", "checkout", commit)
		cmd.Dir = destDir
		if output, err := cmd.CombinedOutput(); err != nil {
			return fmt.Errorf("git checkout %s failed: %w\n%s", commit, err, output)
		}
	}

	return nil
}

// verifyGoVersion checks that the current Go version matches requirements.
func verifyGoVersion(required string) error {
	cmd := exec.Command("go", "version")
	output, err := cmd.Output()
	if err != nil {
		return fmt.Errorf("go version check failed: %w", err)
	}

	versionStr := string(output)
	if !strings.Contains(versionStr, required) {
		return fmt.Errorf("require Go %s, have: %s", required, versionStr)
	}

	return nil
}

// buildDeterministic builds a Go binary with deterministic flags.
//
// Key flags for reproducibility:
// - CGO_ENABLED=0: No CGO (eliminates C toolchain variations)
// - SOURCE_DATE_EPOCH=1: Fixed timestamp for deterministic builds
// - -trimpath: Remove absolute paths from binary
// - -ldflags='-s -w': Strip debug info and symbol tables
// - -buildvcs=false: Don't embed VCS info (varies by git state)
func buildDeterministic(srcDir, outputPath, buildCmd string, buildEnv map[string]string) error {
	// Clear Go build cache to ensure clean build
	cleanCmd := exec.Command("go", "clean", "-cache", "-modcache")
	cleanCmd.Dir = srcDir
	cleanCmd.Run() // Ignore errors, not critical

	// Build environment: start with clean slate
	env := []string{
		"CGO_ENABLED=0",
		"SOURCE_DATE_EPOCH=1",
		"PATH=" + os.Getenv("PATH"),
		"HOME=" + os.Getenv("HOME"),
	}

	// Add manifest-specified environment
	for k, v := range buildEnv {
		env = append(env, k+"="+v)
	}

	// Execute build command
	// Expected format: "go build -trimpath -ldflags='-s -w' -buildvcs=false -o plugin"
	cmd := exec.Command("sh", "-c", buildCmd)
	cmd.Dir = srcDir
	cmd.Env = env

	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("build command failed: %w\n%s", err, output)
	}

	// Move built binary to output path
	builtBinary := filepath.Join(srcDir, "plugin")
	if err := os.Rename(builtBinary, outputPath); err != nil {
		return fmt.Errorf("move binary: %w", err)
	}

	return nil
}

// hashFile computes SHA256 hash of a file.
func hashFile(path string) (string, error) {
	f, err := os.Open(path)
	if err != nil {
		return "", fmt.Errorf("open file: %w", err)
	}
	defer f.Close()

	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		return "", fmt.Errorf("read file: %w", err)
	}

	return fmt.Sprintf("%x", h.Sum(nil)), nil
}

// QuickVerify performs a fast hash check without rebuilding.
// Use this to verify downloaded binaries before first execution.
func QuickVerify(binaryPath string, expectedHash string) (bool, error) {
	actualHash, err := hashFile(binaryPath)
	if err != nil {
		return false, err
	}
	return actualHash == expectedHash, nil
}
