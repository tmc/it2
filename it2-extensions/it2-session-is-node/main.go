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
		fmt.Fprintf(os.Stderr, "Check if a session's process is running Node.js\n")
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

	// Check if the PID is running Node.js
	isNode := isNodeProcess(pid)

	if *jsonOutput {
		result := map[string]interface{}{
			"session_id": sessionID,
			"pid":        pid,
			"is_node":    isNode,
		}
		json.NewEncoder(os.Stdout).Encode(result)
	} else {
		if isNode {
			fmt.Printf("Session %s (PID: %d) is running Node.js\n", sessionID, pid)
		} else {
			fmt.Printf("Session %s (PID: %d) is NOT running Node.js\n", sessionID, pid)
		}
	}

	if !isNode {
		os.Exit(1)
	}
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

// isNodeProcess checks if a given PID is running a Node.js process
func isNodeProcess(pid int) bool {
	// Use ps to get process info
	cmd := exec.Command("ps", "-p", strconv.Itoa(pid), "-o", "comm=")
	output, err := cmd.Output()
	if err != nil {
		return false
	}

	processName := strings.TrimSpace(string(output))

	// Check for various Node.js process names
	nodeNames := []string{"node", "nodejs", "npm", "npx", "yarn", "pnpm", "bun", "deno"}
	for _, name := range nodeNames {
		if strings.Contains(strings.ToLower(processName), name) {
			return true
		}
	}

	// Also check the full command line for Node.js indicators
	cmd = exec.Command("ps", "-p", strconv.Itoa(pid), "-o", "command=")
	output, err = cmd.Output()
	if err != nil {
		return false
	}

	fullCommand := strings.ToLower(string(output))
	for _, name := range nodeNames {
		if strings.Contains(fullCommand, name) {
			return true
		}
	}

	return false
}
