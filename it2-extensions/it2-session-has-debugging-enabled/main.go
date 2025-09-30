package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"

	"github.com/tmc/it2/internal/client"
	"github.com/tmc/it2/internal/cmdutil"
	"github.com/tmc/it2/internal/connect"
)

func main() {
	var (
		jsonOutput = flag.Bool("json", false, "Output result as JSON")
		timeout    = flag.Duration("timeout", 5*time.Second, "Timeout for operations")
	)
	flag.Parse()

	if len(flag.Args()) != 1 {
		fmt.Fprintf(os.Stderr, "Usage: %s [flags] <session-id>\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "Check if a Node.js process in a session has debugging/inspector enabled\n")
		flag.PrintDefaults()
		os.Exit(1)
	}

	sessionID := flag.Args()[0]

	ctx, cancel := cmdutil.CreateContext(*timeout)
	defer cancel()

	c, err := connect.ConnectClient(ctx)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to connect: %v\n", err)
		os.Exit(1)
	}
	defer c.Close()

	// Resolve session ID if needed
	sessionID, err = c.ResolveSessionID(ctx, sessionID)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to resolve session ID: %v\n", err)
		os.Exit(1)
	}

	// Get the PID of the session
	pid, err := getSessionPID(ctx, c, sessionID)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to get session PID: %v\n", err)
		os.Exit(1)
	}

	// Check if the process has debugging enabled
	debugInfo := getNodeDebugInfo(pid)

	if *jsonOutput {
		result := map[string]interface{}{
			"session_id":        sessionID,
			"pid":              pid,
			"has_debugging":    debugInfo.HasDebugging,
			"inspector_port":   debugInfo.InspectorPort,
			"debug_flags":      debugInfo.DebugFlags,
		}
		json.NewEncoder(os.Stdout).Encode(result)
	} else {
		if debugInfo.HasDebugging {
			fmt.Printf("Session %s (PID: %d) has Node.js debugging enabled\n", sessionID, pid)
			if debugInfo.InspectorPort > 0 {
				fmt.Printf("  Inspector port: %d\n", debugInfo.InspectorPort)
			}
			if len(debugInfo.DebugFlags) > 0 {
				fmt.Printf("  Debug flags: %s\n", strings.Join(debugInfo.DebugFlags, ", "))
			}
		} else {
			fmt.Printf("Session %s (PID: %d) does NOT have Node.js debugging enabled\n", sessionID, pid)
		}
	}

	if !debugInfo.HasDebugging {
		os.Exit(1)
	}
}

type DebugInfo struct {
	HasDebugging  bool
	InspectorPort int
	DebugFlags    []string
}

// getSessionPID gets the PID of a session by trying multiple methods
func getSessionPID(ctx context.Context, c *client.Client, sessionID string) (int, error) {
	// First try using shell command
	echoCmd := "echo $$\n"
	if err := c.SendText(ctx, sessionID, echoCmd); err != nil {
		return 0, fmt.Errorf("failed to send PID command: %w", err)
	}

	// Wait a moment for command to execute
	time.Sleep(150 * time.Millisecond)

	// Get the buffer to see the output
	bufferResp, err := c.GetBuffer(ctx, sessionID, 10)
	if err != nil {
		return 0, fmt.Errorf("failed to get buffer: %w", err)
	}

	// Parse the output to find the PID
	for _, lineContent := range bufferResp.GetContents() {
		line := strings.TrimSpace(lineContent.GetText())
		if line == "" {
			continue
		}

		// Look for numeric PID output
		if pid, err := strconv.Atoi(line); err == nil && pid > 0 {
			return pid, nil
		}
	}

	return 0, fmt.Errorf("could not extract PID from shell output")
}

// getNodeDebugInfo checks if a Node.js process has debugging enabled
func getNodeDebugInfo(pid int) DebugInfo {
	info := DebugInfo{}

	// Get the full command line for the process
	cmd := exec.Command("ps", "-p", strconv.Itoa(pid), "-o", "command=")
	output, err := cmd.Output()
	if err != nil {
		return info
	}

	fullCommand := string(output)

	// Check for various debugging flags and patterns
	debugPatterns := []string{
		"--inspect",
		"--inspect-brk",
		"--debug",
		"--debug-brk",
		"--debug-port",
		"--inspect-port",
	}

	var debugFlags []string
	for _, pattern := range debugPatterns {
		if strings.Contains(fullCommand, pattern) {
			info.HasDebugging = true
			debugFlags = append(debugFlags, pattern)
		}
	}
	info.DebugFlags = debugFlags

	// Try to extract inspector port
	info.InspectorPort = extractInspectorPort(fullCommand)

	// Also check for listening on common debug ports using lsof
	if !info.HasDebugging {
		info.HasDebugging = checkDebugPorts(pid)
	}

	return info
}

// extractInspectorPort tries to find the inspector port from command line
func extractInspectorPort(cmdline string) int {
	// Look for patterns like --inspect=9229 or --inspect-port=9229
	parts := strings.Fields(cmdline)
	for _, part := range parts {
		if strings.HasPrefix(part, "--inspect=") {
			portStr := strings.TrimPrefix(part, "--inspect=")
			if port, err := strconv.Atoi(portStr); err == nil {
				return port
			}
		}
		if strings.HasPrefix(part, "--inspect-port=") {
			portStr := strings.TrimPrefix(part, "--inspect-port=")
			if port, err := strconv.Atoi(portStr); err == nil {
				return port
			}
		}
		if strings.HasPrefix(part, "--debug-port=") {
			portStr := strings.TrimPrefix(part, "--debug-port=")
			if port, err := strconv.Atoi(portStr); err == nil {
				return port
			}
		}
	}

	// Check if lsof shows the process listening on default debug ports
	debugPorts := []int{9229, 5858, 9222}
	for _, port := range debugPorts {
		cmd := exec.Command("lsof", "-p", strconv.Itoa(os.Getpid()), "-a", "-i", fmt.Sprintf(":%d", port))
		if err := cmd.Run(); err == nil {
			return port
		}
	}

	return 0
}

// checkDebugPorts checks if the process is listening on common Node.js debug ports
func checkDebugPorts(pid int) bool {
	debugPorts := []int{9229, 5858, 9222}

	for _, port := range debugPorts {
		// Use lsof to check if this PID is listening on the debug port
		cmd := exec.Command("lsof", "-p", strconv.Itoa(pid), "-a", "-i", fmt.Sprintf(":%d", port))
		if err := cmd.Run(); err == nil {
			return true
		}
	}

	return false
}