package client

import (
	"context"
	"fmt"

	pb "github.com/tmc/it2/proto"
)

type SessionInfo struct {
	SessionID    string
	WindowID     string
	WindowNumber int32  // iTerm2 window index
	TabID        string
	WindowTitle  string
	TabTitle     string
	SessionName  string
	PluginData   map[string]interface{} // Additional data from plugins
	Frame        *pb.Frame              // Frame coordinates if available
	GridSize     *pb.Size               // Grid size if available
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
			sessions = append(sessions, extractSessionsFromNode(tab.GetRoot(), windowID, windowNumber, tabID)...)
		}
	}

	// Add buried sessions
	for _, buried := range listResp.GetBuriedSessions() {
		sessions = append(sessions, &SessionInfo{
			SessionID:   buried.GetUniqueIdentifier(),
			SessionName: buried.GetTitle(),
		})
	}

	return sessions, nil
}

func extractSessionsFromNode(node *pb.SplitTreeNode, windowID string, windowNumber int32, tabID string) []*SessionInfo {
	if node == nil {
		return nil
	}

	var sessions []*SessionInfo

	// Process links which contain either sessions or child nodes
	for _, link := range node.GetLinks() {
		switch child := link.GetChild().(type) {
		case *pb.SplitTreeNode_SplitTreeLink_Session:
			session := child.Session
			if session != nil {
				info := &SessionInfo{
					SessionID:    session.GetUniqueIdentifier(),
					WindowID:     windowID,
					WindowNumber: windowNumber,
					TabID:        tabID,
					SessionName:  session.GetTitle(),
					Frame:        session.GetFrame(),
					GridSize:     session.GetGridSize(),
				}
				sessions = append(sessions, info)
			}
		case *pb.SplitTreeNode_SplitTreeLink_Node:
			// Recursively process child nodes
			sessions = append(sessions, extractSessionsFromNode(child.Node, windowID, windowNumber, tabID)...)
		}
	}

	return sessions
}
