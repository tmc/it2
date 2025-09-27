package client

import (
	"context"
	"fmt"

	pb "github.com/tmc/it2/proto"
)

type SessionInfo struct {
	SessionID       string
	ShortID         string // First 8 characters of SessionID for easier reference
	ParentSessionID string // Parent session ID if this is a split pane
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
	CurrentCommand string // Current command from GetPrompt
	ExitCode       uint32 // Last command exit status
	PromptState    string // Shell prompt state (AT_COMMAND_LINE, IN_COMMAND, etc.)
	CommandCount   int32  // Number of commands executed
	JobPID         int32  // Process ID of current job
}

func (c *Client) ListSessions(ctx context.Context) ([]*SessionInfo, error) {
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

	var sessions []*SessionInfo
	for _, window := range listResp.GetWindows() {
		windowID := window.GetWindowId()
		windowNumber := window.GetNumber()
		for _, tab := range window.GetTabs() {
			tabID := tab.GetTabId()
			sessions = append(sessions, extractSessionsFromNode(tab.GetRoot(), windowID, windowNumber, tabID, "")...)
		}
	}

	// Add buried sessions
	for _, buried := range listResp.GetBuriedSessions() {
		sessions = append(sessions, &SessionInfo{
			SessionID:   buried.GetUniqueIdentifier(),
			SessionName: buried.GetTitle(),
		})
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

			// Remaining sessions become children of the first
			for i := 1; i < len(directSessions); i++ {
				directSessions[i].ParentSessionID = parentSession.SessionID
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

// populateJobInfo adds job information to sessions using GetPrompt and GetVariable
func (c *Client) populateJobInfo(ctx context.Context, sessions []*SessionInfo) {
	for _, session := range sessions {
		// Get prompt information
		if promptResp, err := c.GetPrompt(ctx, session.SessionID); err == nil {
			session.CurrentCommand = promptResp.GetCommand()
			session.ExitCode = promptResp.GetExitStatus()
			session.PromptState = promptResp.GetPromptState().String()
			// CommandCount would need to be tracked separately, for now set to 0
			session.CommandCount = 0
		}

		// Try to get job PID from session variables
		if pidStr, err := c.GetVariableWithScope(ctx, "session", session.SessionID, "jobPid"); err == nil && pidStr != "" {
			// Parse PID if it's a valid number
			var pid int32
			if n, parseErr := fmt.Sscanf(pidStr, "%d", &pid); parseErr == nil && n == 1 {
				session.JobPID = pid
			}
		}
	}
}
