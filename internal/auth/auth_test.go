package auth

import (
	"os"
	"testing"
)

func TestHasAuthentication(t *testing.T) {
	// Test when no auth is set
	os.Unsetenv("ITERM2_COOKIE")
	os.Unsetenv("ITERM2_KEY")
	if HasAuthentication() {
		t.Error("Expected HasAuthentication to return false when no env vars set")
	}

	// Test when auth is set
	os.Setenv("ITERM2_COOKIE", "test-cookie")
	defer os.Unsetenv("ITERM2_COOKIE")
	if !HasAuthentication() {
		t.Error("Expected HasAuthentication to return true when ITERM2_COOKIE is set")
	}
}

func TestClearAuthentication(t *testing.T) {
	os.Setenv("ITERM2_COOKIE", "test-cookie")
	os.Setenv("ITERM2_KEY", "test-key")

	ClearAuthentication()

	if os.Getenv("ITERM2_COOKIE") != "" {
		t.Error("Expected ITERM2_COOKIE to be cleared")
	}
	if os.Getenv("ITERM2_KEY") != "" {
		t.Error("Expected ITERM2_KEY to be cleared")
	}
}
