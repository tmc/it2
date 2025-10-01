# Terminal Hyperlinks Design (OSC 8)

**Feature**: Make session/tab/window IDs clickable in `it2` list output
**Priority**: P1 - Quick Win
**Effort**: 1-2 days
**Impact**: Transforms static list output into interactive navigation interface

## Overview

Implement OSC 8 escape sequences to make session, tab, and window IDs clickable in terminal output. Clicking an ID will focus/activate that session/tab/window instantly.

## Technical Background

### OSC 8 Specification

OSC 8 is the standard escape sequence for hyperlinks in terminal emulators:

```
\e]8;;[URL]\e\\[DISPLAY_TEXT]\e]8;;\e\\
```

Where:
- `\e]8;;` - Start hyperlink
- `[URL]` - The URL to open (can be custom scheme)
- `\e\\` - String terminator (ST)
- `[DISPLAY_TEXT]` - The visible text
- `\e]8;;\e\\` - End hyperlink

### iTerm2 Support

iTerm2 fully supports OSC 8 since version 3.1 (2017):
- Clickable links with Cmd+click
- Custom URL schemes via URL handlers
- Hover highlighting
- 2083 byte URL limit

### Custom URL Scheme

We'll use `it2://` custom URL scheme:

```
it2://session/<session-id>     - Activate session
it2://tab/<tab-id>              - Activate tab
it2://window/<window-id>        - Focus window
it2://profile/<profile-name>    - Apply profile
it2://arrangement/<name>        - Restore arrangement
```

## Implementation Plan

### Phase 1: Core Hyperlink Support (Day 1)

**1. Create hyperlink formatting package**

File: `internal/formatting/hyperlinks.go`

```go
package formatting

import (
    "fmt"
    "os"
    "strings"
)

// HyperlinkMode controls when hyperlinks are emitted
type HyperlinkMode int

const (
    HyperlinkAuto HyperlinkMode = iota  // Auto-detect based on terminal
    HyperlinkAlways                     // Always emit escape sequences
    HyperlinkNever                      // Never emit escape sequences
)

// Hyperlink wraps text in OSC 8 escape sequences
func Hyperlink(url, text string) string {
    if url == "" || text == "" {
        return text
    }
    return fmt.Sprintf("\x1b]8;;%s\x1b\\%s\x1b]8;;\x1b\\", url, text)
}

// SessionLink creates a hyperlink to activate a session
func SessionLink(sessionID string) string {
    return Hyperlink(fmt.Sprintf("it2://session/%s", sessionID), sessionID)
}

// TabLink creates a hyperlink to activate a tab
func TabLink(tabID string) string {
    return Hyperlink(fmt.Sprintf("it2://tab/%s", tabID), tabID)
}

// WindowLink creates a hyperlink to focus a window
func WindowLink(windowID string) string {
    return Hyperlink(fmt.Sprintf("it2://window/%s", windowID), windowID)
}

// ShouldUseHyperlinks determines if hyperlinks should be used
func ShouldUseHyperlinks(mode HyperlinkMode) bool {
    switch mode {
    case HyperlinkAlways:
        return true
    case HyperlinkNever:
        return false
    case HyperlinkAuto:
        return isITerm2() && isatty(os.Stdout.Fd())
    default:
        return false
    }
}

func isITerm2() bool {
    term := os.Getenv("TERM_PROGRAM")
    return term == "iTerm.app"
}

func isatty(fd uintptr) bool {
    // Use syscall to check if fd is a terminal
    // Implementation depends on platform
    return true // Simplified for example
}
```

**2. Add global hyperlink flag**

File: `cmd/it2/main.go`

```go
var (
    hyperlinkMode string // "auto", "always", "never"
)

func init() {
    rootCmd.PersistentFlags().StringVar(&hyperlinkMode,
        "hyperlinks", "auto",
        "Enable terminal hyperlinks (auto|always|never)")
}
```

**3. Update session list command**

File: `internal/cmd/session/list.go`

```go
func formatSessionID(id string) string {
    if formatting.ShouldUseHyperlinks(getHyperlinkMode()) {
        return formatting.SessionLink(id)
    }
    return id
}

// In table rendering:
table.AddRow(
    formatSessionID(session.ID),
    session.Name,
    session.State,
    // ... other columns
)
```

### Phase 2: URL Handler Registration (Day 1-2)

**1. Create URL handler script**

File: `scripts/install-url-handler.sh`

```bash
#!/bin/bash
# Install it2:// URL handler for iTerm2

HANDLER_PLIST="$HOME/Library/Application Support/iTerm2/Scripts/AutoLaunch/it2-url-handler.py"

mkdir -p "$(dirname "$HANDLER_PLIST")"

cat > "$HANDLER_PLIST" << 'EOF'
#!/usr/bin/env python3
import iterm2
import subprocess
import sys
from urllib.parse import urlparse

async def main(connection):
    app = await iterm2.async_get_app(connection)

    @iterm2.RPC
    async def handle_it2_url(url):
        """Handle it2:// URL scheme"""
        parsed = urlparse(url)

        if parsed.scheme != "it2":
            return

        parts = parsed.path.strip('/').split('/')
        if len(parts) < 2:
            return

        resource_type = parts[0]  # session, tab, window, etc.
        resource_id = parts[1]

        # Call it2 CLI to activate the resource
        subprocess.run(['it2', resource_type, 'activate', resource_id])

    await handle_it2_url.async_register(connection)

iterm2.run_forever(main)
EOF

chmod +x "$HANDLER_PLIST"

echo "✅ Installed it2:// URL handler"
echo "Restart iTerm2 for changes to take effect"
```

**Alternative: Direct URL handler via AppleScript**

Register with macOS URL handling system:

```bash
#!/bin/bash
# Register it2:// scheme with Launch Services

cat > /tmp/it2-handler.plist << EOF
<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
    <key>CFBundleURLTypes</key>
    <array>
        <dict>
            <key>CFBundleURLName</key>
            <string>it2 Protocol</string>
            <key>CFBundleURLSchemes</key>
            <array>
                <string>it2</string>
            </array>
        </dict>
    </array>
</dict>
</plist>
EOF

# Register with Launch Services
/System/Library/Frameworks/CoreServices.framework/Frameworks/LaunchServices.framework/Support/lsregister -f /tmp/it2-handler.plist
```

**2. Create standalone URL handler binary**

File: `cmd/it2-url-handler/main.go`

```go
package main

import (
    "fmt"
    "net/url"
    "os"
    "os/exec"
    "strings"
)

func main() {
    if len(os.Args) < 2 {
        fmt.Fprintln(os.Stderr, "Usage: it2-url-handler <it2://...>")
        os.Exit(1)
    }

    rawURL := os.Args[1]
    u, err := url.Parse(rawURL)
    if err != nil {
        fmt.Fprintf(os.Stderr, "Invalid URL: %v\n", err)
        os.Exit(1)
    }

    if u.Scheme != "it2" {
        fmt.Fprintf(os.Stderr, "Invalid scheme: %s (expected it2)\n", u.Scheme)
        os.Exit(1)
    }

    parts := strings.Split(strings.Trim(u.Path, "/"), "/")
    if len(parts) < 2 {
        fmt.Fprintf(os.Stderr, "Invalid path: %s\n", u.Path)
        os.Exit(1)
    }

    resourceType := parts[0]  // session, tab, window, etc.
    resourceID := parts[1]

    // Execute: it2 <resource-type> activate <resource-id>
    cmd := exec.Command("it2", resourceType, "activate", resourceID)
    cmd.Stdout = os.Stdout
    cmd.Stderr = os.Stderr

    if err := cmd.Run(); err != nil {
        fmt.Fprintf(os.Stderr, "Failed to activate %s %s: %v\n",
            resourceType, resourceID, err)
        os.Exit(1)
    }
}
```

### Phase 3: Extended Support (Day 2)

**1. Add hyperlinks to tab list**

```go
// internal/cmd/tab/list.go
func formatTabID(id string) string {
    if formatting.ShouldUseHyperlinks(getHyperlinkMode()) {
        return formatting.TabLink(id)
    }
    return id
}
```

**2. Add hyperlinks to window list**

```go
// internal/cmd/window/list.go
func formatWindowID(id string) string {
    if formatting.ShouldUseHyperlinks(getHyperlinkMode()) {
        return formatting.WindowLink(id)
    }
    return id
}
```

**3. Add hyperlinks to error messages**

```go
// internal/cmdutil/errors.go
func SessionNotFoundError(sessionID string) error {
    if formatting.ShouldUseHyperlinks(getHyperlinkMode()) {
        return fmt.Errorf("session not found: %s\n\nRun %s to see available sessions",
            sessionID,
            formatting.Hyperlink("it2://command/session/list", "it2 session list"))
    }
    return fmt.Errorf("session not found: %s\n\nRun 'it2 session list' to see available sessions",
        sessionID)
}
```

## Usage Examples

### Basic Session List with Hyperlinks

```bash
$ it2 session list
ID                                                    NAME     STATE
w0t0p0:A1B2C3D4-1234-5678-9ABC-DEF012345678          bash     ACTIVE
w0t0p1:B2C3D4E5-2345-6789-ABCD-EF0123456789          vim      IDLE
w0t1p0:C3D4E5F6-3456-789A-BCDE-F01234567890          claude   ACTIVE
```

With hyperlinks enabled, each ID is clickable (Cmd+click):
- Click `A1B2C3D4...` → Activates that session
- Visual indicator on hover
- Works exactly like web links

### Disabling Hyperlinks

```bash
# Never use hyperlinks
it2 session list --hyperlinks=never

# Always use hyperlinks (even when piping)
it2 session list --hyperlinks=always | less -R

# Auto-detect (default)
it2 session list --hyperlinks=auto
```

### Piping with Hyperlinks

Hyperlinks are automatically disabled when piping unless `--hyperlinks=always`:

```bash
# No hyperlinks (piped output)
it2 session list | grep ACTIVE

# Force hyperlinks (for tools that support them)
it2 session list --hyperlinks=always | less -R
```

## Advanced Features (Future)

### 1. Rich Hover Tooltips

Add parameters to OSC 8 for tooltips:

```
\e]8;id=sess123:tooltip=Session: bash (pid 12345);;it2://session/A1B2...\e\\
```

### 2. Multi-Action Links

Support multiple actions via URL parameters:

```
it2://session/A1B2?action=activate&split=vertical
it2://session/A1B2?action=send-text&text=ls%20-la
```

### 3. Clickable File Paths

Detect file paths in output and make them clickable:

```go
// Auto-detect file paths in error messages
func detectFilePath(text string) string {
    // If text looks like a file path, make it clickable
    if strings.HasPrefix(text, "/") || strings.HasPrefix(text, "~") {
        return formatting.Hyperlink("file://"+text, text)
    }
    return text
}
```

### 4. Interactive Selection

Support for selecting from list:

```bash
# Display numbered hyperlinks for selection
it2 session list --select
1. [Session A1B2...] bash  - ACTIVE
2. [Session B2C3...] vim   - IDLE
3. [Session C3D4...] claude - ACTIVE

Click or type number to activate session.
```

## Testing Strategy

### Unit Tests

```go
func TestHyperlink(t *testing.T) {
    tests := []struct {
        url  string
        text string
        want string
    }{
        {
            url:  "it2://session/abc123",
            text: "abc123",
            want: "\x1b]8;;it2://session/abc123\x1b\\abc123\x1b]8;;\x1b\\",
        },
    }

    for _, tt := range tests {
        got := Hyperlink(tt.url, tt.text)
        if got != tt.want {
            t.Errorf("Hyperlink(%q, %q) = %q, want %q",
                tt.url, tt.text, got, tt.want)
        }
    }
}
```

### Integration Tests

```go
func TestSessionListHyperlinks(t *testing.T) {
    // Set up test environment
    os.Setenv("TERM_PROGRAM", "iTerm.app")

    // Run session list with hyperlinks=always
    output := runCommand("it2", "session", "list", "--hyperlinks=always")

    // Verify OSC 8 sequences present
    if !strings.Contains(output, "\x1b]8;;it2://session/") {
        t.Error("Expected OSC 8 hyperlinks in output")
    }
}
```

### Manual Testing Checklist

- [ ] Hyperlinks appear correctly in iTerm2
- [ ] Cmd+click activates correct session/tab/window
- [ ] Hover shows hyperlink indicator
- [ ] Works with `--hyperlinks=always`
- [ ] Disabled with `--hyperlinks=never`
- [ ] Auto-detection works correctly
- [ ] Hyperlinks disabled when piping output
- [ ] URL handler installed and working
- [ ] Error messages with suggestions are clickable
- [ ] Works across different iTerm2 versions

## Documentation Updates

### README.md

Add "Interactive Output" section:

```markdown
## Interactive Output

When using iTerm2, `it2` list commands output clickable hyperlinks:

```bash
it2 session list  # Session IDs are clickable!
```

Cmd+click any session ID to instantly activate that session.

### Configuration

```bash
# Disable hyperlinks
it2 session list --hyperlinks=never

# Force hyperlinks (for compatible pagers)
it2 session list --hyperlinks=always | less -R
```

### URL Handler Setup

Install the URL handler for full functionality:

```bash
make install-url-handler
```
```

### QUICKSTART.md

Add interactive navigation example:

```markdown
### Interactive Navigation

List sessions and click IDs to activate:

```bash
it2 session list
# Cmd+click any session ID to switch to it!
```

This works for tabs and windows too:

```bash
it2 tab list     # Click tab IDs
it2 window list  # Click window IDs
```
```

## Installation

Add to `Makefile`:

```makefile
.PHONY: install-url-handler
install-url-handler:
	@echo "Installing it2:// URL handler..."
	./scripts/install-url-handler.sh
	@echo "URL handler installed. Restart iTerm2 for changes to take effect."

.PHONY: install-full
install-full: install install-url-handler
	@echo "Full installation complete!"
```

## Compatibility

### Supported Terminals

- ✅ iTerm2 (3.1+)
- ✅ WezTerm
- ✅ Hyper
- ⚠️ Terminal.app (no OSC 8 support - graceful degradation)
- ⚠️ Other terminals - auto-detection disables hyperlinks

### Graceful Degradation

When hyperlinks aren't supported:
- IDs display as plain text
- All functionality works normally
- No visual artifacts or corruption
- Clean fallback behavior

## Performance Considerations

### Minimal Overhead

OSC 8 sequences add ~40 bytes per hyperlink:
```
Base ID: "A1B2C3D4-1234-5678-9ABC-DEF012345678" = 36 bytes
With OSC 8: "\e]8;;it2://session/A1B2...\e\\A1B2...\e]8;;\e\\" = ~76 bytes
Overhead: ~40 bytes per ID
```

For 100 sessions: ~4KB additional output (negligible)

### Rendering Performance

iTerm2 handles OSC 8 efficiently:
- No noticeable latency
- Hover detection is fast
- Clickable links don't impact scrolling

## Security Considerations

### URL Validation

Always validate URLs before opening:

```go
func isValidIT2URL(rawURL string) bool {
    u, err := url.Parse(rawURL)
    if err != nil {
        return false
    }

    // Only allow it2:// scheme
    if u.Scheme != "it2" {
        return false
    }

    // Validate path structure
    parts := strings.Split(strings.Trim(u.Path, "/"), "/")
    if len(parts) < 2 {
        return false
    }

    // Whitelist resource types
    validTypes := map[string]bool{
        "session": true,
        "tab": true,
        "window": true,
    }

    return validTypes[parts[0]]
}
```

### No Arbitrary Command Execution

URL handler only supports predefined actions:
- ✅ `it2://session/<id>` → `it2 session activate <id>`
- ✅ `it2://tab/<id>` → `it2 tab activate <id>`
- ❌ `it2://exec/rm%20-rf%20/` → Rejected

No shell interpretation, only safe API calls.

## Success Metrics

- [ ] Hyperlinks work in iTerm2 with Cmd+click
- [ ] Auto-detection correctly identifies terminal support
- [ ] URL handler activates correct session/tab/window
- [ ] Graceful degradation in unsupported terminals
- [ ] Zero visual artifacts or output corruption
- [ ] Performance impact < 5ms for 100-item lists
- [ ] Documentation updated with examples
- [ ] User feedback is positive (clickable lists are intuitive)

## Future Enhancements

1. **Extended URL schemes**: Support for more actions via parameters
2. **Hover tooltips**: Rich information on hover
3. **Multi-select**: Click multiple items with Cmd modifier
4. **Visual indicators**: Show which sessions have active hyperlinks
5. **Clickable file paths**: Auto-detect and link file paths in output
6. **Integration examples**: Show hybrid workflows (CLI + clickable links)

## References

- [iTerm2 OSC 8 Documentation](https://iterm2.com/documentation-escape-codes.html)
- [iTerm2 Hyperlinks Feature](https://iterm2.com/feature-reporting/Hyperlinks_in_Terminal_Emulators.html)
- [OSC 8 Specification (Egmont)](https://gist.github.com/egmontkob/eb114294efbcd5adb1944c9f3cb5feda)
- [Terminal Hyperlink Adoption](https://github.com/Alhadis/OSC8-Adoption)
