package client

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"

	pb "github.com/tmc/it2/proto"
)

type SessionInfo struct {
	SessionID       string
	ShortID         string // First 8 characters of SessionID for easier reference
	ParentSessionID string // Parent session ID if this is a split pane
	SplitVertical   *bool  // Direction of split from parent (nil if no parent, true=vertical, false=horizontal)
	WindowID        string
	WindowNumber    int32 // iTerm2 window index
	TabID           string
	WindowTitle     string
	TabTitle        string
	SessionName     string
	PluginData      map[string]interface{} // Additional data from plugins
	Frame           *pb.Frame              // Frame coordinates if available
	GridSize        *pb.Size               // Grid size if available

	// Command information from shell integration
	CurrentCommand   string // Current command from GetPrompt
	ExitCode         uint32 // Last command exit status
	PromptState      string // Shell prompt state (AT_COMMAND_LINE, IN_COMMAND, etc.)
	CommandCount     int32  // Number of commands executed
	ShellPID         int32  // Process ID of the shell
	JobPID           int32  // Process ID of current job
	WorkingDirectory string // Current working directory from session.path
}

// ListSessionsOptions controls optional session listing behaviour.
type ListSessionsOptions struct {
	IncludeBuried bool
}

// ListSessionsRaw returns the raw protobuf response with full window/tab tree
func (c *Client) ListSessionsRaw(ctx context.Context) (*pb.ListSessionsResponse, error) {
	msg := &pb.ClientOriginatedMessage{
		Submessage: &pb.ClientOriginatedMessage_ListSessionsRequest{
			ListSessionsRequest: &pb.ListSessionsRequest{},
		},
	}

	response, err := c.SendRequest(ctx, msg)
	if err != nil {
		return nil, err
	}

	listResp := response.GetListSessionsResponse()
	if listResp == nil {
		return nil, fmt.Errorf("unexpected response type")
	}

	return listResp, nil
}

func (c *Client) ListSessions(ctx context.Context) ([]*SessionInfo, error) {
	return c.listSessions(ctx, true)
}

// ListSessionsWithOptions returns sessions using the provided options.
func (c *Client) ListSessionsWithOptions(ctx context.Context, opts ListSessionsOptions) ([]*SessionInfo, error) {
	return c.listSessions(ctx, opts.IncludeBuried)
}

func (c *Client) listSessions(ctx context.Context, includeBuried bool) ([]*SessionInfo, error) {
	listResp, err := c.ListSessionsRaw(ctx)
	if err != nil {
		return nil, err
	}

	var sessions []*SessionInfo
	for _, window := range listResp.GetWindows() {
		windowID := window.GetWindowId()
		windowNumber := window.GetNumber()
		for _, tab := range window.GetTabs() {
			tabID := tab.GetTabId()
			sessions = append(sessions, extractSessionsFromNode(tab.GetRoot(), windowID, windowNumber, tabID, "")...)
		}
	}

	if includeBuried {
		for _, buried := range listResp.GetBuriedSessions() {
			sessions = append(sessions, &SessionInfo{
				SessionID:   buried.GetUniqueIdentifier(),
				SessionName: buried.GetTitle(),
			})
		}
	}

	// Populate job information for all sessions
	c.populateJobInfo(ctx, sessions)

	return sessions, nil
}

func extractSessionsFromNode(node *pb.SplitTreeNode, windowID string, windowNumber int32, tabID string, parentSessionID string) []*SessionInfo {
	if node == nil {
		return nil
	}

	var sessions []*SessionInfo
	links := node.GetLinks()

	// Get split direction for this node
	var splitVertical *bool
	if node.Vertical != nil {
		v := node.GetVertical()
		splitVertical = &v
	}

	// If this node has multiple children (it's a split), establish parent-child relationships
	if len(links) > 1 {
		var directSessions []*SessionInfo
		var childSessions []*SessionInfo

		// Separate direct sessions from child nodes
		for _, link := range links {
			switch child := link.GetChild().(type) {
			case *pb.SplitTreeNode_SplitTreeLink_Session:
				session := child.Session
				if session != nil {
					sessionID := session.GetUniqueIdentifier()
					shortID := sessionID
					if len(sessionID) > 8 {
						shortID = sessionID[:8]
					}
					info := &SessionInfo{
						SessionID:       sessionID,
						ShortID:         shortID,
						ParentSessionID: parentSessionID,
						SplitVertical:   splitVertical,
						WindowID:        windowID,
						WindowNumber:    windowNumber,
						TabID:           tabID,
						SessionName:     session.GetTitle(),
						Frame:           session.GetFrame(),
						GridSize:        session.GetGridSize(),
					}
					directSessions = append(directSessions, info)
				}
			case *pb.SplitTreeNode_SplitTreeLink_Node:
				// Recursively process child nodes
				childSessions = append(childSessions, extractSessionsFromNode(child.Node, windowID, windowNumber, tabID, parentSessionID)...)
			}
		}

		// If we have multiple direct sessions in this split, make the first one parent of the rest
		if len(directSessions) > 1 {
			// First session becomes the parent
			parentSession := directSessions[0]
			sessions = append(sessions, parentSession)

			// Remaining sessions become children of the first, with split direction
			for i := 1; i < len(directSessions); i++ {
				directSessions[i].ParentSessionID = parentSession.SessionID
				directSessions[i].SplitVertical = splitVertical
				sessions = append(sessions, directSessions[i])
			}
		} else {
			// Add any single direct sessions as-is
			sessions = append(sessions, directSessions...)
		}

		// Add all child sessions from recursive calls
		sessions = append(sessions, childSessions...)
	} else {
		// Single child - process normally
		for _, link := range links {
			switch child := link.GetChild().(type) {
			case *pb.SplitTreeNode_SplitTreeLink_Session:
				session := child.Session
				if session != nil {
					sessionID := session.GetUniqueIdentifier()
					shortID := sessionID
					if len(sessionID) > 8 {
						shortID = sessionID[:8]
					}
					info := &SessionInfo{
						SessionID:       sessionID,
						ShortID:         shortID,
						ParentSessionID: parentSessionID,
						SplitVertical:   splitVertical,
						WindowID:        windowID,
						WindowNumber:    windowNumber,
						TabID:           tabID,
						SessionName:     session.GetTitle(),
						Frame:           session.GetFrame(),
						GridSize:        session.GetGridSize(),
					}
					sessions = append(sessions, info)
				}
			case *pb.SplitTreeNode_SplitTreeLink_Node:
				// Recursively process child nodes
				sessions = append(sessions, extractSessionsFromNode(child.Node, windowID, windowNumber, tabID, parentSessionID)...)
			}
		}
	}

	return sessions
}

// populateJobInfo adds job information to sessions using GetPrompt and GetVariable.
// Uses limited concurrency to avoid overwhelming the WebSocket connection.
func (c *Client) populateJobInfo(ctx context.Context, sessions []*SessionInfo) {
	// Limit concurrent session info requests to avoid WebSocket contention.
	// Each session makes 3 API calls, so 10 concurrent sessions = ~30 in-flight requests.
	const maxConcurrency = 10

	var wg sync.WaitGroup
	sem := make(chan struct{}, maxConcurrency)

	for _, session := range sessions {
		wg.Add(1)
		go func(s *SessionInfo) {
			defer wg.Done()
			sem <- struct{}{}        // Acquire semaphore
			defer func() { <-sem }() // Release semaphore
			c.populateSessionJobInfo(ctx, s)
		}(session)
	}
	wg.Wait()
}

// populateSessionJobInfo populates job info for a single session.
// Uses batched API calls for performance: 3 calls instead of 5.
func (c *Client) populateSessionJobInfo(ctx context.Context, session *SessionInfo) {
	var wg sync.WaitGroup
	wg.Add(3) // 3 API calls: GetPrompt, ListPrompts, GetMultipleVariables

	// Get prompt information
	go func() {
		defer wg.Done()
		if promptResp, err := c.GetPrompt(ctx, session.SessionID); err == nil {
			session.CurrentCommand = promptResp.GetCommand()
			session.ExitCode = promptResp.GetExitStatus()
			session.PromptState = promptResp.GetPromptState().String()
		}
	}()

	// Get command count by listing all prompts for this session
	go func() {
		defer wg.Done()
		if listResp, err := c.ListPrompts(ctx, session.SessionID); err == nil {
			session.CommandCount = int32(len(listResp.GetUniquePromptId()))
		}
	}()

	// Get all variables in a single batched call (pid, jobPid, path)
	go func() {
		defer wg.Done()
		vars, err := c.GetMultipleVariablesWithScope(ctx, "session", session.SessionID, []string{"pid", "jobPid", "path"})
		if err != nil {
			return
		}

		// Parse shell PID
		if pidStr, ok := vars["pid"]; ok && pidStr != "" {
			var pid int32
			if n, parseErr := fmt.Sscanf(pidStr, "%d", &pid); parseErr == nil && n == 1 {
				session.ShellPID = pid
			}
		}

		// Parse job PID
		if pidStr, ok := vars["jobPid"]; ok && pidStr != "" {
			var pid int32
			if n, parseErr := fmt.Sscanf(pidStr, "%d", &pid); parseErr == nil && n == 1 {
				session.JobPID = pid
			}
		}

		// Parse working directory (JSON-escaped)
		if path, ok := vars["path"]; ok && path != "" {
			var unescaped string
			if err := json.Unmarshal([]byte(path), &unescaped); err == nil {
				session.WorkingDirectory = unescaped
			} else {
				session.WorkingDirectory = path
			}
		}
	}()

	wg.Wait()
}
