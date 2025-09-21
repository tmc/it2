package client

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	pb "github.com/tmc/it2/proto"
)

// KeyBinding represents a keyboard binding configuration
type KeyBinding struct {
	Key         string `json:"key"`
	Action      string `json:"action"`
	Profile     string `json:"profile,omitempty"`
	Global      bool   `json:"global,omitempty"`
	Description string `json:"description,omitempty"`
}

// FindResult represents a text search result
type FindResult struct {
	Line     int64  `json:"line"`
	Column   int32  `json:"column"`
	Text     string `json:"text"`
	Context  string `json:"context"`
	LineText string `json:"line_text"`
}

// ReplaceResult represents a text replacement result
type ReplaceResult struct {
	Line         int64  `json:"line"`
	Column       int32  `json:"column"`
	OriginalText string `json:"original_text"`
	NewText      string `json:"new_text"`
	Success      bool   `json:"success"`
}

// HighlightResult represents a text highlighting result
type HighlightResult struct {
	Line        int64  `json:"line"`
	Column      int32  `json:"column"`
	Text        string `json:"text"`
	Color       string `json:"color"`
	Duration    int32  `json:"duration"`
	HighlightID string `json:"highlight_id"`
}

// BindKey creates or updates a keyboard binding
func (c *Client) BindKey(ctx context.Context, key, action, profile string, global bool, description string) error {
	// This would need to use preferences or profile property requests
	// Since iTerm2's API doesn't directly support keyboard binding management,
	// we'll simulate this by setting profile properties

	if profile == "" && !global {
		// Get current session's profile
		var err error
		profile, err = c.getCurrentProfile(ctx)
		if err != nil {
			return fmt.Errorf("failed to get current profile: %w", err)
		}
	}

	keyProperty := fmt.Sprintf("keyboard_binding_%s", normalizeKeyName(key))
	bindingData := map[string]interface{}{
		"action":      action,
		"description": description,
	}

	jsonValue, err := json.Marshal(bindingData)
	if err != nil {
		return fmt.Errorf("failed to marshal binding data: %w", err)
	}

	var msg *pb.ClientOriginatedMessage
	if global {
		// Use preferences for global bindings
		msg = &pb.ClientOriginatedMessage{
			Submessage: &pb.ClientOriginatedMessage_PreferencesRequest{
				PreferencesRequest: &pb.PreferencesRequest{
					Requests: []*pb.PreferencesRequest_Request{
						{
							Request: &pb.PreferencesRequest_Request_SetPreferenceRequest{
								SetPreferenceRequest: &pb.PreferencesRequest_Request_SetPreference{
									Key:       &keyProperty,
									JsonValue: stringPtr(string(jsonValue)),
								},
							},
						},
					},
				},
			},
		}
	} else {
		// Use profile properties for profile-specific bindings
		msg = &pb.ClientOriginatedMessage{
			Submessage: &pb.ClientOriginatedMessage_SetProfilePropertyRequest{
				SetProfilePropertyRequest: &pb.SetProfilePropertyRequest{
					Target: &pb.SetProfilePropertyRequest_GuidList_{
						GuidList: &pb.SetProfilePropertyRequest_GuidList{
							Guids: []string{profile},
						},
					},
					Key:       &keyProperty,
					JsonValue: stringPtr(string(jsonValue)),
				},
			},
		}
	}

	response, err := c.SendRequest(ctx, msg)
	if err != nil {
		return fmt.Errorf("failed to bind key: %w", err)
	}

	if global {
		if resp := response.GetPreferencesResponse(); resp != nil {
			if len(resp.Results) > 0 {
				if setResult := resp.Results[0].GetSetPreferenceResult(); setResult != nil {
					if setResult.GetStatus() != pb.PreferencesResponse_Result_SetPreferenceResult_OK {
						return fmt.Errorf("bind key failed: %v", setResult.GetStatus())
					}
				}
			}
		}
	} else {
		if resp := response.GetSetProfilePropertyResponse(); resp != nil {
			if resp.GetStatus() != pb.SetProfilePropertyResponse_OK {
				return fmt.Errorf("bind key failed: %v", resp.GetStatus())
			}
		}
	}

	return nil
}

// UnbindKey removes a keyboard binding
func (c *Client) UnbindKey(ctx context.Context, key, profile string, global bool) error {
	if profile == "" && !global {
		var err error
		profile, err = c.getCurrentProfile(ctx)
		if err != nil {
			return fmt.Errorf("failed to get current profile: %w", err)
		}
	}

	keyProperty := fmt.Sprintf("keyboard_binding_%s", normalizeKeyName(key))

	var msg *pb.ClientOriginatedMessage
	if global {
		// Remove global preference
		msg = &pb.ClientOriginatedMessage{
			Submessage: &pb.ClientOriginatedMessage_PreferencesRequest{
				PreferencesRequest: &pb.PreferencesRequest{
					Requests: []*pb.PreferencesRequest_Request{
						{
							Request: &pb.PreferencesRequest_Request_SetPreferenceRequest{
								SetPreferenceRequest: &pb.PreferencesRequest_Request_SetPreference{
									Key:       &keyProperty,
									JsonValue: stringPtr("null"), // Set to null to remove
								},
							},
						},
					},
				},
			},
		}
	} else {
		// Remove from profile
		msg = &pb.ClientOriginatedMessage{
			Submessage: &pb.ClientOriginatedMessage_SetProfilePropertyRequest{
				SetProfilePropertyRequest: &pb.SetProfilePropertyRequest{
					Target: &pb.SetProfilePropertyRequest_GuidList_{
						GuidList: &pb.SetProfilePropertyRequest_GuidList{
							Guids: []string{profile},
						},
					},
					Key:       &keyProperty,
					JsonValue: stringPtr("null"),
				},
			},
		}
	}

	response, err := c.SendRequest(ctx, msg)
	if err != nil {
		return fmt.Errorf("failed to unbind key: %w", err)
	}

	if global {
		if resp := response.GetPreferencesResponse(); resp != nil {
			if len(resp.Results) > 0 {
				if setResult := resp.Results[0].GetSetPreferenceResult(); setResult != nil {
					if setResult.GetStatus() != pb.PreferencesResponse_Result_SetPreferenceResult_OK {
						return fmt.Errorf("unbind key failed: %v", setResult.GetStatus())
					}
				}
			}
		}
	} else {
		if resp := response.GetSetProfilePropertyResponse(); resp != nil {
			if resp.GetStatus() != pb.SetProfilePropertyResponse_OK {
				return fmt.Errorf("unbind key failed: %v", resp.GetStatus())
			}
		}
	}

	return nil
}

// ListKeyBindings lists all keyboard bindings for a profile or globally
func (c *Client) ListKeyBindings(ctx context.Context, profile string, global bool) ([]interface{}, error) {
	// This is a simplified implementation - in reality, we would need to
	// enumerate all keyboard binding properties
	return []interface{}{}, fmt.Errorf("ListKeyBindings not fully implemented - would require enumerating all properties")
}

// ExportKeyBindings exports keyboard bindings to a file
func (c *Client) ExportKeyBindings(ctx context.Context, filename, profile string, global bool) error {
	bindings, err := c.ListKeyBindings(ctx, profile, global)
	if err != nil {
		return fmt.Errorf("failed to list bindings: %w", err)
	}

	data, err := json.MarshalIndent(bindings, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal bindings: %w", err)
	}

	if filename == "" {
		fmt.Print(string(data))
		return nil
	}

	if err := os.WriteFile(filename, data, 0644); err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}

	return nil
}

// ImportKeyBindings imports keyboard bindings from a file
func (c *Client) ImportKeyBindings(ctx context.Context, filename, profile string, global, merge bool) error {
	data, err := os.ReadFile(filename)
	if err != nil {
		return fmt.Errorf("failed to read file: %w", err)
	}

	var bindings []KeyBinding
	if err := json.Unmarshal(data, &bindings); err != nil {
		return fmt.Errorf("failed to unmarshal bindings: %w", err)
	}

	// If not merging, clear existing bindings first
	if !merge {
		// This would require implementing clear functionality
		// For now, we'll just proceed with adding the new bindings
	}

	for _, binding := range bindings {
		targetProfile := profile
		targetGlobal := global

		if binding.Profile != "" {
			targetProfile = binding.Profile
		}
		if binding.Global {
			targetGlobal = true
		}

		if err := c.BindKey(ctx, binding.Key, binding.Action, targetProfile, targetGlobal, binding.Description); err != nil {
			return fmt.Errorf("failed to bind key %s: %w", binding.Key, err)
		}
	}

	return nil
}

// Helper functions

func (c *Client) getCurrentProfile(ctx context.Context) (string, error) {
	// Get current session's profile - this would require getting the active session
	// and then getting its profile property
	sessions, err := c.ListSessions(ctx)
	if err != nil {
		return "", fmt.Errorf("failed to list sessions: %w", err)
	}

	// Find the active session (simplified - would need to check focus)
	if len(sessions) > 0 {
		// This is simplified - would need proper active session detection
		return "default", nil
	}

	return "default", nil
}

func normalizeKeyName(key string) string {
	// Normalize key names for consistent property naming
	key = strings.ToLower(key)
	key = strings.ReplaceAll(key, "+", "_")
	key = strings.ReplaceAll(key, " ", "_")
	return key
}

// Text search and manipulation methods

// FindText searches for text patterns in the session buffer
func (c *Client) FindText(ctx context.Context, sessionID, pattern string, regex, caseSensitive, wholeWord bool, maxResults int32) ([]*FindResult, error) {
	// Get buffer contents first
	resp, err := c.GetBufferWithStyles(ctx, sessionID, 1000, false) // Get substantial buffer
	if err != nil {
		return nil, fmt.Errorf("failed to get buffer: %w", err)
	}

	var results []*FindResult
	resultCount := int32(0)

	// Search through the buffer contents
	for i, lineContent := range resp.Contents {
		if resultCount >= maxResults {
			break
		}

		text := lineContent.GetText()
		if text == "" {
			continue
		}

		// Apply search options
		searchText := text
		searchPattern := pattern

		if !caseSensitive {
			searchText = strings.ToLower(searchText)
			searchPattern = strings.ToLower(searchPattern)
		}

		var matches []int
		if regex {
			// For now, we'll do simple substring search
			// In a full implementation, this would use proper regex
			if idx := strings.Index(searchText, searchPattern); idx != -1 {
				matches = []int{idx}
			}
		} else {
			if wholeWord {
				// Simple whole word matching
				words := strings.Fields(searchText)
				for j, word := range words {
					if word == searchPattern {
						// Calculate position
						pos := 0
						for k := 0; k < j; k++ {
							pos += len(words[k]) + 1 // +1 for space
						}
						matches = append(matches, pos)
					}
				}
			} else {
				// Substring search
				pos := 0
				for {
					idx := strings.Index(searchText[pos:], searchPattern)
					if idx == -1 {
						break
					}
					matches = append(matches, pos+idx)
					pos += idx + 1
				}
			}
		}

		// Create results for matches on this line
		for _, matchPos := range matches {
			if resultCount >= maxResults {
				break
			}

			result := &FindResult{
				Line:     int64(i),
				Column:   int32(matchPos),
				Text:     pattern,
				Context:  getContext(text, matchPos, len(pattern)),
				LineText: text,
			}
			results = append(results, result)
			resultCount++
		}
	}

	return results, nil
}

// ReplaceText replaces text patterns in the session buffer
func (c *Client) ReplaceText(ctx context.Context, sessionID, pattern, replacement string, regex, caseSensitive, wholeWord bool, maxReplacements int32, confirm bool) ([]*ReplaceResult, error) {
	// First find the text
	finds, err := c.FindText(ctx, sessionID, pattern, regex, caseSensitive, wholeWord, maxReplacements)
	if err != nil {
		return nil, fmt.Errorf("failed to find text: %w", err)
	}

	var results []*ReplaceResult

	// Note: This is a simplified implementation
	// Real text replacement in a terminal would be complex and might not be practical
	// through the iTerm2 API, as it would require:
	// 1. Positioning the cursor at each match
	// 2. Selecting the text
	// 3. Typing the replacement
	// This could interfere with running programs

	for _, find := range finds {
		if confirm {
			// In a real implementation, this would prompt the user
			fmt.Printf("Replace '%s' with '%s' at line %d, column %d? (y/n): ",
				find.Text, replacement, find.Line, find.Column)
			// For now, we'll assume yes
		}

		result := &ReplaceResult{
			Line:         find.Line,
			Column:       find.Column,
			OriginalText: find.Text,
			NewText:      replacement,
			Success:      false, // Set to false since we're not actually replacing
		}
		results = append(results, result)
	}

	return results, fmt.Errorf("text replacement not fully implemented - would require complex cursor manipulation")
}

// HighlightText creates temporary highlighting for text patterns
func (c *Client) HighlightText(ctx context.Context, sessionID, pattern string, regex, caseSensitive, wholeWord bool, color, background string, duration int32) ([]*HighlightResult, error) {
	// Find the text first
	finds, err := c.FindText(ctx, sessionID, pattern, regex, caseSensitive, wholeWord, 100)
	if err != nil {
		return nil, fmt.Errorf("failed to find text: %w", err)
	}

	var results []*HighlightResult

	// Note: This is a placeholder implementation
	// Real highlighting would require using iTerm2's annotation or marking features
	// which might not be directly available through the current API

	for i, find := range finds {
		result := &HighlightResult{
			Line:        find.Line,
			Column:      find.Column,
			Text:        find.Text,
			Color:       color,
			Duration:    duration,
			HighlightID: fmt.Sprintf("highlight_%d", i),
		}
		results = append(results, result)
	}

	return results, fmt.Errorf("text highlighting not fully implemented - would require annotation API")
}

// ClearHighlights removes all text highlights
func (c *Client) ClearHighlights(ctx context.Context, sessionID string) error {
	return fmt.Errorf("clear highlights not implemented - would require annotation API")
}

// GetSelectionText gets the text content of the current selection
func (c *Client) GetSelectionText(ctx context.Context, sessionID string) (string, error) {
	selection, err := c.GetSelection(ctx, sessionID)
	if err != nil {
		return "", fmt.Errorf("failed to get selection: %w", err)
	}

	if selection == nil || len(selection.SubSelections) == 0 {
		return "", nil
	}

	// This is a simplified implementation
	// In reality, we would need to:
	// 1. Get the buffer contents for the selected range
	// 2. Extract the text from the selected coordinates
	// 3. Handle multi-line selections properly

	return "", fmt.Errorf("get selection text not fully implemented - would require coordinate-based text extraction")
}

// SetSelectionRange sets selection using explicit coordinates
func (c *Client) SetSelectionRange(ctx context.Context, sessionID string, startX int32, startY int64, endX int32, endY int64, mode string) error {
	// Convert mode string to SelectionMode enum
	var selectionMode pb.SelectionMode
	switch strings.ToLower(mode) {
	case "character":
		selectionMode = pb.SelectionMode_CHARACTER
	case "word":
		selectionMode = pb.SelectionMode_WORD
	case "line":
		selectionMode = pb.SelectionMode_LINE
	case "smart":
		selectionMode = pb.SelectionMode_SMART
	case "box":
		selectionMode = pb.SelectionMode_BOX
	case "whole-line":
		selectionMode = pb.SelectionMode_WHOLE_LINE
	default:
		selectionMode = pb.SelectionMode_CHARACTER
	}

	// Create the selection
	selection := &pb.Selection{
		SubSelections: []*pb.SubSelection{
			{
				WindowedCoordRange: &pb.WindowedCoordRange{
					CoordRange: &pb.CoordRange{
						Start: &pb.Coord{
							X: &startX,
							Y: &startY,
						},
						End: &pb.Coord{
							X: &endX,
							Y: &endY,
						},
					},
				},
				SelectionMode: &selectionMode,
				Connected:     boolPtr(true),
			},
		},
	}

	msg := &pb.ClientOriginatedMessage{
		Submessage: &pb.ClientOriginatedMessage_SelectionRequest{
			SelectionRequest: &pb.SelectionRequest{
				Request: &pb.SelectionRequest_SetSelectionRequest_{
					SetSelectionRequest: &pb.SelectionRequest_SetSelectionRequest{
						SessionId: &sessionID,
						Selection: selection,
					},
				},
			},
		},
	}

	response, err := c.SendRequest(ctx, msg)
	if err != nil {
		return fmt.Errorf("failed to set selection: %w", err)
	}

	if resp := response.GetSelectionResponse(); resp != nil {
		if resp.GetStatus() != pb.SelectionResponse_OK {
			return fmt.Errorf("set selection failed: %v", resp.GetStatus())
		}
	}

	return nil
}

// CopySelection copies the current selection to clipboard
func (c *Client) CopySelection(ctx context.Context, sessionID string) error {
	// This would typically be done through a menu item or key binding
	// For now, we'll try to invoke the copy function

	msg := &pb.ClientOriginatedMessage{
		Submessage: &pb.ClientOriginatedMessage_InvokeFunctionRequest{
			InvokeFunctionRequest: &pb.InvokeFunctionRequest{
				Context: &pb.InvokeFunctionRequest_Session_{
					Session: &pb.InvokeFunctionRequest_Session{
						SessionId: &sessionID,
					},
				},
				Invocation: stringPtr("copy:"),
			},
		},
	}

	response, err := c.SendRequest(ctx, msg)
	if err != nil {
		return fmt.Errorf("failed to copy selection: %w", err)
	}

	if resp := response.GetInvokeFunctionResponse(); resp != nil {
		if resp.GetError() != nil {
			return fmt.Errorf("copy selection failed: %v", resp.GetError().GetErrorReason())
		}
	}

	return nil
}

// Helper functions

func getContext(text string, pos, length int) string {
	contextSize := 20
	start := pos - contextSize
	if start < 0 {
		start = 0
	}
	end := pos + length + contextSize
	if end > len(text) {
		end = len(text)
	}

	context := text[start:end]
	if start > 0 {
		context = "..." + context
	}
	if end < len(text) {
		context = context + "..."
	}

	return context
}

// Helper functions using the shared implementations from helpers.go
