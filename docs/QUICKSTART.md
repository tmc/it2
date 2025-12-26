# it2 Quick Start Guide

Get up and running with it2 CLI in minutes. This guide covers the essential commands you'll use most often.

## Prerequisites

1. **iTerm2 3.3.0+** installed on macOS
2. **Python API enabled**: iTerm2 → Preferences → General → Magic → Enable Python API
3. **it2 CLI installed**: `go install github.com/tmc/it2/cmd/it2@latest`

## First Steps

### 1. Check Connection
```bash
# Verify it2 can connect to iTerm2
it2 auth check

# If authentication fails, request access
it2 auth request
```

### 2. Basic Session Management
```bash
# See all your terminal sessions
it2 session list

# List only session IDs (quiet mode for scripting)
it2 session list -q

# Get your current session ID
it2 session current

# Send a command to current session
it2 session send-text "echo 'Hello from it2!'"

# Send to specific session using partial ID (4+ chars)
it2 session send-text 7AA9 "echo 'Using short ID!'"
```

## Essential Commands

### Session Operations
```bash
# List sessions with details
it2 session list

# Send text to current session
it2 session send-text "ls -la"

# Split current pane vertically (side by side)
it2 session split --vertical

# Split current pane horizontally (top/bottom)
it2 session split --horizontal

# Get comprehensive session information
it2 session get-info
```

### Tab Management
```bash
# Create new tab with Default profile
it2 tab create "Default"

# Create tab with badge label
it2 tab create "Default" --badge "Server"

# List all tabs
it2 tab list

# Show split pane layout
it2 tab splits
```

### Text and Buffer
```bash
# Get current screen contents
it2 text get-screen

# Get scrollback buffer
it2 text get-buffer

# Search buffer for text
it2 text search "error"

# Get cursor position
it2 text get-cursor
```

### Profile Management
```bash
# List available profiles
it2 profile list

# Get profile details
it2 profile get "Default"

# Apply different profile to current session
it2 profile session-set-property "background_color" '{"red": 0.1, "green": 0.1, "blue": 0.1}'
```

## Common Workflows

### 1. Development Environment Setup
```bash
# Create development tab
it2 tab create "Development" --badge "Dev"

# Split for code editor and terminal
it2 session split --vertical

# Start editor in left pane
it2 session send-text "code ."

# Use right pane for commands - get its session ID first
RIGHT_PANE=$(it2 session list --format json | jq -r '.[-1].id')
it2 session send-text "$RIGHT_PANE" "npm run dev"
```

### 2. Server Management
```bash
# Create tabs for different servers
for server in web01 web02 db01; do
    it2 tab create "SSH" --badge "$server"
    it2 session send-text "ssh $server"
done
```

### 3. Multi-pane Monitoring
```bash
# Create monitoring tab
it2 tab create "Monitoring" --badge "Monitor"

# System stats in main pane
it2 session send-text "htop"

# Split for logs
it2 session split --horizontal
LOG_PANE=$(it2 session list --format json | jq -r '.[-1].id')
it2 session send-text "$LOG_PANE" "tail -f /var/log/app.log"
```

### 4. Command History and Shell State Analysis
```bash
# Show recent commands with exit codes (requires Shell Integration)
it2 prompt list

# Search command history
it2 prompt search "git"

# Find failed commands
it2 prompt list | grep "Exit: [^0]"

# Check if shell is ready for input (requires Shell Integration)
it2 shell state

# Wait for shell to be ready before sending commands
it2 shell wait-for-prompt && it2 session send-text "make build"
```

## Output Formats

it2 supports multiple output formats for scripting:

```bash
# Human-readable (default)
it2 session list

# JSON for scripting
it2 session list --format json

# YAML for configuration
it2 profile get "Default" --format yaml

# Text for simple parsing
it2 session list --format text
```

## Useful Scripting Patterns

### Get session IDs
```bash
# Current session
CURRENT=$(it2 session current)

# Last created session
LAST=$(it2 session list --format json | jq -r '.[-1].id')

# All session IDs (using quiet mode)
SESSIONS=($(it2 session list -q))

# Or using JSON format
SESSIONS=($(it2 session list --format json | jq -r '.[].id'))

# Use partial IDs (4+ chars) instead of full UUIDs
it2 session send-text 7AA9 "echo 'Much easier!'"
```

### Conditional operations
```bash
# Check if session exists before operating
if it2 session get-info "$SESSION_ID" >/dev/null 2>&1; then
    it2 session send-text "$SESSION_ID" "echo 'Session exists'"
fi
```

### Error handling
```bash
# Capture and handle errors
if ! it2 session send-text "invalid-command" 2>/dev/null; then
    echo "Failed to send text to session"
fi
```

## Advanced Features (Quick Overview)

### Claude Code Session Monitoring

Monitor and automate Claude Code sessions with native commands:

```bash
# Human-readable status
it2 session claude-status

# JSON state for scripting
it2 session get-state --agent=claude --format=json

# Check if session is actively working (exit 0=active, 1=idle)
if it2 session is-active --agent=claude; then
    echo "Claude is working..."
fi

# Detect modals (exit 0=modal present, 1=no modal)
MODAL=$(it2 session has-modal)
echo "Modal type: $MODAL"  # none, approval, choice, confirmation, input

# Get suggested action
ACTION=$(it2 session suggest-action)
echo "Suggested: $ACTION"  # wait, continue, tab, modal:approve, none

# Watch state changes in real-time (NDJSON output)
it2 session watch --agent=claude

# Auto-execute suggested actions
it2 session watch --agent=claude --auto-act
```

### Session Filtering

```bash
# Filter by name or title (case-insensitive)
it2 session list --filter vim
it2 session list --filter "my project"
```

### Split with Working Directory

```bash
# Create split that starts in specific directory
it2 session split --horizontal --cwd /path/to/project
```

### Shell Integration Features
Requires iTerm2 Shell Integration to be installed:
```bash
it2 prompt list              # Command history
it2 job list                 # Running processes
it2 shell state              # Check if shell is ready/busy/tui
it2 shell wait-for-prompt    # Wait until shell is ready
it2 variable get             # iTerm2 variables
```

### Real-time Monitoring
```bash
it2 notification monitor --type session    # Watch session events
it2 variable monitor session user.pwd      # Watch directory changes
```

### Layouts and Arrangements
```bash
it2 arrangement save "my-layout"     # Save current window layout
it2 arrangement restore "my-layout"  # Restore saved layout
```

## Troubleshooting

### Connection Issues
```bash
# Enable debug output
ITERM2_DEBUG=1 it2 session list

# Check authentication
it2 auth check

# Re-authenticate if needed
it2 auth request
```

### Common Problems
- **"Connection refused"**: Enable Python API in iTerm2 preferences
- **"Authentication required"**: Run `it2 auth request`
- **"Session not found"**: Session may have closed, check `it2 session list`
- **Command not responding**: Try with shorter timeout: `--timeout 2s`

### send-text Delivery Issues

If you see "⚠ Text partially delivered" warnings:
```bash
# Skip verification for speed (no false positives)
it2 session send-text SESSION_ID --skip-confirm "command"

# Auto-retry on failures
it2 session send-text SESSION_ID --retry 3 --retry-delay 2s "command"

# Wait for session to be ready before sending
it2 session send-text SESSION_ID --require is-at-prompt,has-no-partial-input "command"

# Debug what's happening
IT2_DEBUG_DELIVERY=1 it2 session send-text SESSION_ID "test"
```

Exit codes: 0=success, 1=error, 2=partial (retry), 3=busy (retry), 4=modal

## Shell Integration Setup

For advanced features like command history and job monitoring:

1. **Install Shell Integration**:
   ```bash
   curl -L https://iterm2.com/shell_integration/install_shell_integration_and_utilities.sh | bash
   ```

2. **Restart your shell** or source the integration:
   ```bash
   source ~/.iterm2_shell_integration.bash  # or .zsh
   ```

3. **Verify it's working**:
   ```bash
   echo $ITERM_SESSION_ID  # Should show session ID
   it2 prompt list         # Should show command history
   ```

## Next Steps

- Read [EXAMPLES.md](./EXAMPLES.md) for comprehensive usage examples
- Check `it2 [command] --help` for detailed command documentation
- Explore the [main README](./README.md) for complete feature documentation
- Visit [iTerm2 API docs](https://iterm2.com/documentation-api.html) for background information

## Quick Reference Card

```bash
# Essential commands:
it2 session list                    # List all sessions
it2 session list --filter vim       # Filter by name/title
it2 session send-text "command"     # Execute command
it2 session split --vertical        # Split pane
it2 session split --cwd /path       # Split with working dir
it2 tab create "Profile"            # New tab
it2 text get-screen                 # Capture output
it2 auth check                      # Verify connection

# Claude Code monitoring:
it2 session claude-status           # Human-readable status
it2 session get-state --agent=claude  # JSON state
it2 session is-active               # Check if working
it2 session has-modal               # Detect modals
it2 session suggest-action          # Get suggested action
it2 session watch --agent=claude    # Real-time monitoring

# Pro tips:
it2 session get-info --json         # Detailed session data
it2 tab create "Default" --badge "Label"  # Labeled tabs
ITERM2_DEBUG=1 it2 [cmd]           # Debug mode
```

Happy automating!