package client

import (
	"reflect"
	"testing"

	pb "github.com/tmc/it2/proto"
)

func TestSessionInfo_Struct(t *testing.T) {
	session := &SessionInfo{
		SessionID:   "session-123",
		WindowID:    "window-456",
		TabID:       "tab-789",
		WindowTitle: "My Window",
		TabTitle:    "My Tab",
		SessionName: "My Session",
		PluginData:  map[string]interface{}{"key": "value"},
	}

	// Test all fields are set correctly
	if session.SessionID != "session-123" {
		t.Errorf("Expected SessionID to be 'session-123', got '%s'", session.SessionID)
	}
	if session.WindowID != "window-456" {
		t.Errorf("Expected WindowID to be 'window-456', got '%s'", session.WindowID)
	}
	if session.TabID != "tab-789" {
		t.Errorf("Expected TabID to be 'tab-789', got '%s'", session.TabID)
	}
	if session.WindowTitle != "My Window" {
		t.Errorf("Expected WindowTitle to be 'My Window', got '%s'", session.WindowTitle)
	}
	if session.TabTitle != "My Tab" {
		t.Errorf("Expected TabTitle to be 'My Tab', got '%s'", session.TabTitle)
	}
	if session.SessionName != "My Session" {
		t.Errorf("Expected SessionName to be 'My Session', got '%s'", session.SessionName)
	}
	if session.PluginData["key"] != "value" {
		t.Errorf("Expected PluginData['key'] to be 'value', got '%v'", session.PluginData["key"])
	}
}

func TestExtractSessionsFromNode_NilNode(t *testing.T) {
	sessions := extractSessionsFromNode(nil, "window-1", 1, "tab-1", "")
	if sessions != nil {
		t.Error("Expected nil sessions for nil node")
	}
}

func TestExtractSessionsFromNode_EmptyNode(t *testing.T) {
	node := &pb.SplitTreeNode{}
	sessions := extractSessionsFromNode(node, "window-1", 1, "tab-1", "")
	if len(sessions) != 0 {
		t.Errorf("Expected 0 sessions for empty node, got %d", len(sessions))
	}
}

func TestExtractSessionsFromNode_WithSession(t *testing.T) {
	sessionID := "session-123"
	sessionTitle := "Test Session"

	session := &pb.SessionSummary{
		UniqueIdentifier: &sessionID,
		Title:            &sessionTitle,
	}

	link := &pb.SplitTreeNode_SplitTreeLink{
		Child: &pb.SplitTreeNode_SplitTreeLink_Session{
			Session: session,
		},
	}

	node := &pb.SplitTreeNode{
		Links: []*pb.SplitTreeNode_SplitTreeLink{link},
	}

	sessions := extractSessionsFromNode(node, "window-1", 1, "tab-1", "")

	if len(sessions) != 1 {
		t.Fatalf("Expected 1 session, got %d", len(sessions))
	}

	s := sessions[0]
	if s.SessionID != "session-123" {
		t.Errorf("Expected SessionID to be 'session-123', got '%s'", s.SessionID)
	}
	if s.WindowID != "window-1" {
		t.Errorf("Expected WindowID to be 'window-1', got '%s'", s.WindowID)
	}
	if s.TabID != "tab-1" {
		t.Errorf("Expected TabID to be 'tab-1', got '%s'", s.TabID)
	}
	if s.SessionName != "Test Session" {
		t.Errorf("Expected SessionName to be 'Test Session', got '%s'", s.SessionName)
	}
}

func TestExtractSessionsFromNode_WithNestedNode(t *testing.T) {
	sessionID := "session-456"
	sessionTitle := "Nested Session"

	session := &pb.SessionSummary{
		UniqueIdentifier: &sessionID,
		Title:            &sessionTitle,
	}

	nestedLink := &pb.SplitTreeNode_SplitTreeLink{
		Child: &pb.SplitTreeNode_SplitTreeLink_Session{
			Session: session,
		},
	}

	nestedNode := &pb.SplitTreeNode{
		Links: []*pb.SplitTreeNode_SplitTreeLink{nestedLink},
	}

	parentLink := &pb.SplitTreeNode_SplitTreeLink{
		Child: &pb.SplitTreeNode_SplitTreeLink_Node{
			Node: nestedNode,
		},
	}

	parentNode := &pb.SplitTreeNode{
		Links: []*pb.SplitTreeNode_SplitTreeLink{parentLink},
	}

	sessions := extractSessionsFromNode(parentNode, "window-1", 1, "tab-1", "")

	if len(sessions) != 1 {
		t.Fatalf("Expected 1 session from nested node, got %d", len(sessions))
	}

	s := sessions[0]
	if s.SessionID != "session-456" {
		t.Errorf("Expected SessionID to be 'session-456', got '%s'", s.SessionID)
	}
	if s.SessionName != "Nested Session" {
		t.Errorf("Expected SessionName to be 'Nested Session', got '%s'", s.SessionName)
	}
}

func TestExtractSessionsFromNode_WithMultipleSessions(t *testing.T) {
	// Create two sessions
	sessionID1 := "session-1"
	sessionTitle1 := "First Session"
	session1 := &pb.SessionSummary{
		UniqueIdentifier: &sessionID1,
		Title:            &sessionTitle1,
	}

	sessionID2 := "session-2"
	sessionTitle2 := "Second Session"
	session2 := &pb.SessionSummary{
		UniqueIdentifier: &sessionID2,
		Title:            &sessionTitle2,
	}

	// Create links for both sessions
	link1 := &pb.SplitTreeNode_SplitTreeLink{
		Child: &pb.SplitTreeNode_SplitTreeLink_Session{
			Session: session1,
		},
	}

	link2 := &pb.SplitTreeNode_SplitTreeLink{
		Child: &pb.SplitTreeNode_SplitTreeLink_Session{
			Session: session2,
		},
	}

	node := &pb.SplitTreeNode{
		Links: []*pb.SplitTreeNode_SplitTreeLink{link1, link2},
	}

	sessions := extractSessionsFromNode(node, "window-1", 1, "tab-1", "")

	if len(sessions) != 2 {
		t.Fatalf("Expected 2 sessions, got %d", len(sessions))
	}

	// Verify both sessions are present (order may vary)
	sessionIDs := make([]string, len(sessions))
	for i, s := range sessions {
		sessionIDs[i] = s.SessionID
	}

	expectedIDs := []string{"session-1", "session-2"}
	if !reflect.DeepEqual(sessionIDs, expectedIDs) && !reflect.DeepEqual(sessionIDs, []string{"session-2", "session-1"}) {
		t.Errorf("Expected session IDs to be %v or %v, got %v", expectedIDs, []string{"session-2", "session-1"}, sessionIDs)
	}
}

func TestExtractSessionsFromNode_NilSession(t *testing.T) {
	link := &pb.SplitTreeNode_SplitTreeLink{
		Child: &pb.SplitTreeNode_SplitTreeLink_Session{
			Session: nil, // nil session should be skipped
		},
	}

	node := &pb.SplitTreeNode{
		Links: []*pb.SplitTreeNode_SplitTreeLink{link},
	}

	sessions := extractSessionsFromNode(node, "window-1", 1, "tab-1", "")

	if len(sessions) != 0 {
		t.Errorf("Expected 0 sessions for nil session, got %d", len(sessions))
	}
}
