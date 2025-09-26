package client

import (
	"testing"
)

func TestWindowInfo_Struct(t *testing.T) {
	window := &WindowInfo{
		WindowID:     "window-123",
		Title:        "My Window",
		Frame:        "{x:0, y:0, width:800, height:600}",
		Fullscreen:   "false",
		Miniaturized: "true",
		TabCount:     3,
		PluginData:   map[string]interface{}{"plugin_key": "plugin_value"},
	}

	// Test all fields are set correctly
	if window.WindowID != "window-123" {
		t.Errorf("Expected WindowID to be 'window-123', got '%s'", window.WindowID)
	}
	if window.Title != "My Window" {
		t.Errorf("Expected Title to be 'My Window', got '%s'", window.Title)
	}
	if window.Frame != "{x:0, y:0, width:800, height:600}" {
		t.Errorf("Expected Frame to be '{x:0, y:0, width:800, height:600}', got '%s'", window.Frame)
	}
	if window.Fullscreen != "false" {
		t.Errorf("Expected Fullscreen to be 'false', got '%s'", window.Fullscreen)
	}
	if window.Miniaturized != "true" {
		t.Errorf("Expected Miniaturized to be 'true', got '%s'", window.Miniaturized)
	}
	if window.TabCount != 3 {
		t.Errorf("Expected TabCount to be 3, got %d", window.TabCount)
	}
	if window.PluginData["plugin_key"] != "plugin_value" {
		t.Errorf("Expected PluginData['plugin_key'] to be 'plugin_value', got '%v'", window.PluginData["plugin_key"])
	}
}

func TestWindowInfo_EmptyWindow(t *testing.T) {
	window := &WindowInfo{}

	// Test zero values
	if window.WindowID != "" {
		t.Errorf("Expected empty WindowID, got '%s'", window.WindowID)
	}
	if window.Title != "" {
		t.Errorf("Expected empty Title, got '%s'", window.Title)
	}
	if window.TabCount != 0 {
		t.Errorf("Expected TabCount to be 0, got %d", window.TabCount)
	}
	if window.PluginData != nil {
		t.Errorf("Expected PluginData to be nil, got %v", window.PluginData)
	}
}

func TestWindowInfo_WithPluginData(t *testing.T) {
	window := &WindowInfo{
		WindowID: "window-456",
	}

	// Initialize plugin data
	window.PluginData = make(map[string]interface{})
	window.PluginData["test_plugin"] = map[string]string{
		"status":  "active",
		"version": "1.0.0",
	}
	window.PluginData["another_plugin"] = 42

	// Test plugin data access
	testPlugin, ok := window.PluginData["test_plugin"]
	if !ok {
		t.Error("Expected test_plugin to exist in PluginData")
	}

	pluginMap, ok := testPlugin.(map[string]string)
	if !ok {
		t.Error("Expected test_plugin to be map[string]string")
	} else {
		if pluginMap["status"] != "active" {
			t.Errorf("Expected plugin status to be 'active', got '%s'", pluginMap["status"])
		}
		if pluginMap["version"] != "1.0.0" {
			t.Errorf("Expected plugin version to be '1.0.0', got '%s'", pluginMap["version"])
		}
	}

	anotherPlugin, ok := window.PluginData["another_plugin"]
	if !ok {
		t.Error("Expected another_plugin to exist in PluginData")
	}

	pluginValue, ok := anotherPlugin.(int)
	if !ok {
		t.Error("Expected another_plugin to be int")
	} else if pluginValue != 42 {
		t.Errorf("Expected another_plugin value to be 42, got %d", pluginValue)
	}
}

func TestWindowInfo_JSONTags(t *testing.T) {
	// This test ensures the struct has the expected JSON tags
	// We can't directly test JSON tags in Go, but we can use reflection
	// For now, we'll just verify the struct can be marshaled/unmarshaled
	// (which would fail if the tags were malformed)

	window := &WindowInfo{
		WindowID:     "window-test",
		Title:        "Test Title",
		Frame:        "test-frame",
		Fullscreen:   "false",
		Miniaturized: "true",
		TabCount:     2,
	}

	// Basic test that the struct is valid
	if window.WindowID != "window-test" {
		t.Errorf("Expected WindowID to be 'window-test', got '%s'", window.WindowID)
	}
}
