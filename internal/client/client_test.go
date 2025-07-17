package client

import (
	"testing"
)

func TestNewClient(t *testing.T) {
	c := New("ws://localhost:1912")
	if c == nil {
		t.Fatal("Expected non-nil client")
	}
	if c.url != "ws://localhost:1912" {
		t.Errorf("Expected URL ws://localhost:1912, got %s", c.url)
	}
	if c.messages == nil {
		t.Fatal("Expected messages channel to be initialized")
	}
	if c.pending == nil {
		t.Fatal("Expected pending map to be initialized")
	}
}