package projects

import (
	"os"
	"path/filepath"
	"testing"
)

func TestPathToProjectName(t *testing.T) {
	tests := []struct {
		path string
		want string
	}{
		{"/Users/tmc/code/myapp", "-Users-tmc-code-myapp"},
		{"/Volumes/tmc/go/src/github.com/tmc/it2", "-Volumes-tmc-go-src-github.com-tmc-it2"},
		{"relative/path", "relative-path"},
		{"/private/tmp", "-private-tmp"},
	}

	for _, tt := range tests {
		got := PathToProjectName(tt.path)
		if got != tt.want {
			t.Errorf("PathToProjectName(%q) = %q, want %q", tt.path, got, tt.want)
		}
	}
}

func TestProjectDir(t *testing.T) {
	// Create temp base dir
	tmpHome, err := os.MkdirTemp("", "it2-test-home")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpHome)

	// Override home dir for test
	oldHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpHome)
	defer os.Setenv("HOME", oldHome)

	projectPath := "/tmp/test-project"
	dir, err := ProjectDir(projectPath)
	if err != nil {
		t.Fatalf("ProjectDir failed: %v", err)
	}

	expected := filepath.Join(tmpHome, ".it2", "projects", "-tmp-test-project")
	if dir != expected {
		t.Errorf("ProjectDir = %q, want %q", dir, expected)
	}

	// Verify directory was created
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		t.Error("ProjectDir did not create the directory")
	}
}

func TestVersionInfo(t *testing.T) {
	info := GetVersionInfo()

	if info.GoVersion == "" {
		t.Error("GoVersion should not be empty")
	}
	if info.OS == "" {
		t.Error("OS should not be empty")
	}
	if info.Arch == "" {
		t.Error("Arch should not be empty")
	}
	if info.WrittenAt.IsZero() {
		t.Error("WrittenAt should not be zero")
	}

	short := info.ShortVersion()
	if short == "" {
		t.Error("ShortVersion should not be empty")
	}
	t.Logf("Version info: %s (%s/%s) - short: %s", info.GoVersion, info.OS, info.Arch, short)
}

func TestWriteReadVersionInfo(t *testing.T) {
	// Create temp base dir
	tmpHome, err := os.MkdirTemp("", "it2-test-home")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpHome)

	// Override home dir for test
	oldHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpHome)
	defer os.Setenv("HOME", oldHome)

	projectPath := "/tmp/test-version-project"

	// Write version info
	if err := WriteVersionInfo(projectPath); err != nil {
		t.Fatalf("WriteVersionInfo failed: %v", err)
	}

	// Read it back
	info, err := ReadVersionInfo(projectPath)
	if err != nil {
		t.Fatalf("ReadVersionInfo failed: %v", err)
	}
	if info == nil {
		t.Fatal("ReadVersionInfo returned nil")
	}

	if info.GoVersion == "" {
		t.Error("Read GoVersion should not be empty")
	}

	t.Logf("Read version: %s, revision: %s", info.ShortVersion(), info.VCSRevision)
}

func TestSessionsIndex(t *testing.T) {
	// Create temp base dir
	tmpHome, err := os.MkdirTemp("", "it2-test-home")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpHome)

	// Override home dir for test
	oldHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpHome)
	defer os.Setenv("HOME", oldHome)

	projectPath := "/tmp/test-sessions-project"

	// Load (creates empty)
	idx, err := LoadSessionsIndex(projectPath)
	if err != nil {
		t.Fatalf("LoadSessionsIndex failed: %v", err)
	}

	// Add session
	idx.AddSession(SessionInfo{
		ID:   "test-session-1",
		Name: "Test Session",
		Role: "research",
	})

	// Save
	if err := idx.Save(); err != nil {
		t.Fatalf("Save failed: %v", err)
	}

	// Reload
	idx2, err := LoadSessionsIndex(projectPath)
	if err != nil {
		t.Fatalf("Reload failed: %v", err)
	}

	if len(idx2.Sessions) != 1 {
		t.Errorf("Expected 1 session, got %d", len(idx2.Sessions))
	}

	session := idx2.GetSession("test-session-1")
	if session == nil {
		t.Fatal("GetSession returned nil")
	}
	if session.Role != "research" {
		t.Errorf("Session role = %q, want %q", session.Role, "research")
	}
}
