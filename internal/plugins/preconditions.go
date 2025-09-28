package plugins

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

// PreConditionResult represents the result of a pre-condition check
type PreConditionResult struct {
	Success bool
	Message string
	Error   error
}

// PreConditionChecker is an interface for checking pre-conditions
type PreConditionChecker interface {
	// Check runs the pre-condition check
	Check(ctx context.Context, args []string) (PreConditionResult, error)

	// Name returns the name of the pre-condition checker
	Name() string

	// Description returns a description of what this checker validates
	Description() string
}

// ExecPreConditionChecker runs an external executable to check conditions
type ExecPreConditionChecker struct {
	name        string
	executable  string
	description string
}

// NewExecPreConditionChecker creates a new executable pre-condition checker
func NewExecPreConditionChecker(executable string) *ExecPreConditionChecker {
	baseName := filepath.Base(executable)

	// Clean up the name for display and registration
	name := baseName

	// Remove common prefixes to get the core name
	prefixes := []string{"it2-session-", "it2-check-", "it2-"}
	for _, prefix := range prefixes {
		if strings.HasPrefix(name, prefix) {
			name = strings.TrimPrefix(name, prefix)
			break
		}
	}

	return &ExecPreConditionChecker{
		name:        name,
		executable:  executable,
		description: fmt.Sprintf("Pre-condition check via %s", baseName),
	}
}

// Name returns the name of the checker
func (c *ExecPreConditionChecker) Name() string {
	return c.name
}

// Description returns the description of the checker
func (c *ExecPreConditionChecker) Description() string {
	return c.description
}

// Check runs the executable and interprets the result
func (c *ExecPreConditionChecker) Check(ctx context.Context, args []string) (PreConditionResult, error) {
	cmd := exec.CommandContext(ctx, c.executable, args...)
	output, err := cmd.CombinedOutput()

	// Trim the output
	outputStr := strings.TrimSpace(string(output))

	// Check if the command execution failed
	if err != nil {
		// Check if it was a timeout
		if ctx.Err() == context.DeadlineExceeded {
			return PreConditionResult{
				Success: false,
				Message: fmt.Sprintf("Pre-condition check '%s' timed out", c.name),
				Error:   fmt.Errorf("timeout waiting for %s", c.name),
			}, nil
		}

		// Command failed - this is treated as condition not met
		return PreConditionResult{
			Success: false,
			Message: fmt.Sprintf("Pre-condition '%s' not met: %s", c.name, outputStr),
			Error:   nil, // Not an error in the check itself, just condition not met
		}, nil
	}

	// Interpret the output
	// Convention: "true" or exit 0 means success, "false" or non-zero exit means failure
	success := false
	message := outputStr

	switch strings.ToLower(outputStr) {
	case "true", "yes", "ok", "success", "1":
		success = true
		message = fmt.Sprintf("Pre-condition '%s' satisfied", c.name)
	case "false", "no", "fail", "failure", "0":
		success = false
		if message == "false" || message == "0" {
			message = fmt.Sprintf("Pre-condition '%s' not satisfied", c.name)
		}
	default:
		// If output is empty and exit code is 0, consider it success
		if outputStr == "" && err == nil {
			success = true
			message = fmt.Sprintf("Pre-condition '%s' satisfied", c.name)
		} else {
			// Otherwise, consider any output with exit 0 as success with message
			success = (err == nil)
			if message == "" {
				if success {
					message = fmt.Sprintf("Pre-condition '%s' satisfied", c.name)
				} else {
					message = fmt.Sprintf("Pre-condition '%s' not satisfied", c.name)
				}
			}
		}
	}

	return PreConditionResult{
		Success: success,
		Message: message,
		Error:   nil,
	}, nil
}

// PreConditionManager manages and executes pre-condition checks
type PreConditionManager struct {
	checkers map[string]PreConditionChecker
}

// NewPreConditionManager creates a new pre-condition manager
func NewPreConditionManager() *PreConditionManager {
	return &PreConditionManager{
		checkers: make(map[string]PreConditionChecker),
	}
}

// Register adds a new pre-condition checker
func (m *PreConditionManager) Register(checker PreConditionChecker) {
	m.checkers[checker.Name()] = checker
}

// RegisterExecutable registers an executable as a pre-condition checker
func (m *PreConditionManager) RegisterExecutable(executable string) error {
	// Check if the executable exists and is executable
	info, err := os.Stat(executable)
	if err != nil {
		return fmt.Errorf("pre-condition plugin not found: %s", executable)
	}

	// Check if it's executable
	if info.Mode()&0111 == 0 {
		return fmt.Errorf("pre-condition plugin not executable: %s", executable)
	}

	checker := NewExecPreConditionChecker(executable)
	m.Register(checker)
	return nil
}

// WaitForCondition waits for a condition to be met, with retries and timeout
func (m *PreConditionManager) WaitForCondition(ctx context.Context, conditionName string, args []string, timeout time.Duration, pollInterval time.Duration) (PreConditionResult, error) {
	// First check if the provided name is a direct path
	if strings.Contains(conditionName, "/") || strings.Contains(conditionName, "\\") {
		// It's a path - register it directly
		if err := m.RegisterExecutable(conditionName); err == nil {
			// Extract the base name for lookup
			checker := NewExecPreConditionChecker(conditionName)
			conditionName = checker.Name()
		}
	}

	checker, exists := m.checkers[conditionName]
	if !exists {
		// Try to find it as an executable in PATH or common locations
		executable := m.findExecutable(conditionName)
		if executable != "" {
			if err := m.RegisterExecutable(executable); err != nil {
				return PreConditionResult{
					Success: false,
					Message: fmt.Sprintf("Failed to load pre-condition plugin '%s': %v", conditionName, err),
					Error:   err,
				}, err
			}
			checker = m.checkers[conditionName]
			if checker == nil {
				return PreConditionResult{
					Success: false,
					Message: fmt.Sprintf("Failed to initialize pre-condition plugin '%s'", conditionName),
					Error:   fmt.Errorf("failed to initialize plugin: %s", conditionName),
				}, fmt.Errorf("failed to initialize plugin: %s", conditionName)
			}
		} else {
			return PreConditionResult{
				Success: false,
				Message: fmt.Sprintf("Pre-condition '%s' not found", conditionName),
				Error:   fmt.Errorf("unknown pre-condition: %s", conditionName),
			}, fmt.Errorf("unknown pre-condition: %s", conditionName)
		}
	}

	// Safety check for nil checker
	if checker == nil {
		return PreConditionResult{
			Success: false,
			Message: fmt.Sprintf("Pre-condition checker '%s' is not initialized", conditionName),
			Error:   fmt.Errorf("nil checker: %s", conditionName),
		}, fmt.Errorf("nil checker: %s", conditionName)
	}

	// Create a timeout context
	timeoutCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	// Try immediately first
	result, err := checker.Check(timeoutCtx, args)
	if err != nil {
		return result, err
	}
	if result.Success {
		return result, nil
	}

	// If not successful, start polling
	ticker := time.NewTicker(pollInterval)
	defer ticker.Stop()

	for {
		select {
		case <-timeoutCtx.Done():
			// Timeout reached
			return PreConditionResult{
				Success: false,
				Message: fmt.Sprintf("Timeout waiting for pre-condition '%s': %s", conditionName, result.Message),
				Error:   fmt.Errorf("timeout waiting for condition '%s'", conditionName),
			}, nil

		case <-ticker.C:
			// Try again
			result, err = checker.Check(timeoutCtx, args)
			if err != nil {
				return result, err
			}
			if result.Success {
				return result, nil
			}
		}
	}
}

// CheckCondition checks a condition once without waiting
func (m *PreConditionManager) CheckCondition(ctx context.Context, conditionName string, args []string) (PreConditionResult, error) {
	checker, exists := m.checkers[conditionName]
	if !exists {
		// Try to find it as an executable
		executable := m.findExecutable(conditionName)
		if executable != "" {
			if err := m.RegisterExecutable(executable); err != nil {
				return PreConditionResult{
					Success: false,
					Message: fmt.Sprintf("Failed to load pre-condition plugin '%s': %v", conditionName, err),
					Error:   err,
				}, err
			}
			checker = m.checkers[conditionName]
		} else {
			return PreConditionResult{
				Success: false,
				Message: fmt.Sprintf("Pre-condition '%s' not found", conditionName),
				Error:   fmt.Errorf("unknown pre-condition: %s", conditionName),
			}, fmt.Errorf("unknown pre-condition: %s", conditionName)
		}
	}

	return checker.Check(ctx, args)
}

// ListConditions returns a list of all registered conditions
func (m *PreConditionManager) ListConditions() []string {
	conditions := make([]string, 0, len(m.checkers))
	for name := range m.checkers {
		conditions = append(conditions, name)
	}
	return conditions
}

// findExecutable tries to find an executable for a condition name
func (m *PreConditionManager) findExecutable(conditionName string) string {
	// Try different naming patterns
	patterns := []string{
		conditionName,
		"it2-" + conditionName,
		"it2-session-" + conditionName,
		"it2-check-" + conditionName,
	}

	// Look in common locations
	searchPaths := []string{
		".", // Current directory
		"./plugins",
		"./internal/plugins/scripts",
		os.Getenv("HOME") + "/bin",
		os.Getenv("HOME") + "/.local/bin",
		os.Getenv("HOME") + "/go/bin",
		"/usr/local/bin",
		"/opt/it2/plugins",
	}

	// Also check PATH
	if path := os.Getenv("PATH"); path != "" {
		searchPaths = append(searchPaths, filepath.SplitList(path)...)
	}

	for _, pattern := range patterns {
		for _, dir := range searchPaths {
			executable := filepath.Join(dir, pattern)
			if info, err := os.Stat(executable); err == nil && info.Mode()&0111 != 0 {
				return executable
			}
		}

		// Also try using 'which' command
		if output, err := exec.Command("which", pattern).Output(); err == nil {
			if executable := strings.TrimSpace(string(output)); executable != "" {
				return executable
			}
		}
	}

	return ""
}

// Global pre-condition manager instance
var globalManager = NewPreConditionManager()

// GetGlobalManager returns the global pre-condition manager
func GetGlobalManager() *PreConditionManager {
	return globalManager
}

// WaitForCondition is a convenience function using the global manager
func WaitForCondition(ctx context.Context, conditionName string, args []string, timeout time.Duration) (PreConditionResult, error) {
	return globalManager.WaitForCondition(ctx, conditionName, args, timeout, 500*time.Millisecond)
}

// CheckCondition is a convenience function using the global manager
func CheckCondition(ctx context.Context, conditionName string, args []string) (PreConditionResult, error) {
	return globalManager.CheckCondition(ctx, conditionName, args)
}
