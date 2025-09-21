package client

import (
	"context"
	"fmt"
	"strings"
	"time"

	pb "github.com/tmc/it2/proto"
)

type JobInfo struct {
	JobID   string
	Status  string
	Command string
}

// GetJobs gets the running jobs in a session by sending the 'jobs' command
func (c *Client) GetJobs(ctx context.Context, sessionID string) ([]*JobInfo, error) {
	// Send the jobs command
	jobsCmd := "jobs -l\n"
	if err := c.SendText(ctx, sessionID, jobsCmd); err != nil {
		return nil, fmt.Errorf("failed to send jobs command: %w", err)
	}

	// Wait a moment for command to execute
	time.Sleep(100 * time.Millisecond)

	// Get the buffer to see the output
	bufferResp, err := c.GetBuffer(ctx, sessionID, 20) // Get last 20 lines
	if err != nil {
		return nil, fmt.Errorf("failed to get buffer: %w", err)
	}

	return parseJobsOutput(bufferResp), nil
}

func parseJobsOutput(bufferResp *pb.GetBufferResponse) []*JobInfo {
	var jobs []*JobInfo

	// Combine all buffer lines into text
	var lines []string
	for _, lineContent := range bufferResp.GetContents() {
		if lineContent.GetText() != "" {
			lines = append(lines, lineContent.GetText())
		}
	}

	// Look for jobs output (lines that start with [jobnum])
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "[") && strings.Contains(line, "]") {
			if job := parseJobLine(line); job != nil {
				jobs = append(jobs, job)
			}
		}
	}

	return jobs
}

func parseJobLine(line string) *JobInfo {
	// Example job line: "[1]+  12345 Running                 vim file.txt &"
	// or: "[2]-  12346 Stopped                 sleep 100"

	parts := strings.Fields(line)
	if len(parts) < 4 {
		return nil
	}

	// Extract job ID from [1]+ or [1]-
	jobIDPart := parts[0]
	if !strings.HasPrefix(jobIDPart, "[") || !strings.Contains(jobIDPart, "]") {
		return nil
	}

	jobID := strings.Trim(strings.Split(jobIDPart, "]")[0], "[")

	// Skip PID (parts[1])
	status := parts[2]

	// Combine remaining parts as command
	command := strings.Join(parts[3:], " ")

	return &JobInfo{
		JobID:   jobID,
		Status:  status,
		Command: command,
	}
}

// GetProcessInfo gets basic process information from the session title
func (c *Client) GetProcessInfo(sessionInfo *SessionInfo) *ProcessInfo {
	if sessionInfo.SessionName == "" {
		return nil
	}

	// Parse title format: "✳ Git Workflow (-bash) — /"
	title := sessionInfo.SessionName

	var process, directory string

	// Look for pattern: (command) — directory
	if idx := strings.Index(title, " ("); idx != -1 {
		afterParen := title[idx+2:]
		if endIdx := strings.Index(afterParen, ")"); endIdx != -1 {
			process = afterParen[:endIdx]

			// Look for directory after " — "
			remaining := afterParen[endIdx+1:]
			if dirIdx := strings.Index(remaining, " — "); dirIdx != -1 {
				directory = remaining[dirIdx+3:]
			}
		}
	}

	return &ProcessInfo{
		Command:   process,
		Directory: directory,
		SessionID: sessionInfo.SessionID,
	}
}

type ProcessInfo struct {
	Command   string
	Directory string
	SessionID string
}
