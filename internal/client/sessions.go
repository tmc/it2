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
	CurrentCommand  string // Current command from GetPrompt
	ExitCode        uint32 // Last command exit status
	PromptState     string // Shell prompt state (AT_COMMAND_LINE, IN_COMMAND, etc.)
	CommandCount    int32  // Number of commands executed
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

	// Populate command information for all sessions
	c.populateCommandInfo(ctx, sessions)

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
		var allSessions []*SessionInfo

		// First, collect all sessions from all children
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
					allSessions = append(allSessions, info)
				}
			case *pb.SplitTreeNode_SplitTreeLink_Node:
				// Recursively process child nodes
				childSessions := extractSessionsFromNode(child.Node, windowID, windowNumber, tabID, parentSessionID)
				allSessions = append(allSessions, childSessions...)
			}
		}

		// If this is a top-level split (no parent), make the first session the parent of the rest
		if parentSessionID == "" && len(allSessions) > 1 {
			// First session becomes the parent
			parentSession := allSessions[0]
			sessions = append(sessions, parentSession)

			// Remaining sessions become children
			for i := 1; i < len(allSessions); i++ {
				allSessions[i].ParentSessionID = parentSession.SessionID
				sessions = append(sessions, allSessions[i])
			}
		} else {
			// If we already have a parent or only one session, add all as-is
			sessions = append(sessions, allSessions...)
		}
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

// populateCommandInfo adds command information to sessions using GetPrompt
func (c *Client) populateCommandInfo(ctx context.Context, sessions []*SessionInfo) {
	for _, session := range sessions {
		if promptResp, err := c.GetPrompt(ctx, session.SessionID); err == nil {
			session.CurrentCommand = promptResp.GetCommand()
			session.ExitCode = promptResp.GetExitStatus()
			session.PromptState = promptResp.GetPromptState().String()
			// CommandCount would need to be tracked separately, for now set to 0
			session.CommandCount = 0
		}
	}
}
