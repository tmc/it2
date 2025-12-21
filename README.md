# it2

[![Go Reference](https://pkg.go.dev/badge/github.com/tmc/it2.svg)](https://pkg.go.dev/github.com/tmc/it2)

A powerful command-line tool for controlling iTerm2 through its API. Automate terminal sessions, manage tabs and windows, monitor events, and integrate iTerm2 into your workflows.

## Installation

### Requirements

- **iTerm2** (macOS only, for now)
- **Go** 1.21+

### Install with Go

```bash
go install github.com/tmc/it2/cmd/it2@latest
```

Make sure `~/go/bin` is in your PATH:

```bash
export PATH="$HOME/go/bin:$PATH"
```

### Install from GitHub Releases

Download pre-built binaries from [GitHub Releases](https://github.com/tmc/it2/releases).

### Verify Installation

```bash
it2 auth check
```

The first time you run `it2`, iTerm2 will prompt you to allow API access. Click "Allow" to continue.

### Extensibility

it2 supports extensibility through **plugins** and **hooks**:

- **Plugins** — External executables (`it2-*`) discovered by it2 that enrich session listings or provide automation
- **Hooks** — Executables triggered by external tools (Claude Code, Gemini CLI) for event logging and automation

Plugins are automatically discovered from your PATH or embedded in the binary (extracted to `~/.it2/plugins/{version}/`).

```bash
# List available plugins
it2 plugin list

# Run a plugin
it2 plugin <name> [args...]

# Install Claude Code hooks
it2 plugin claude-code-hook --install
```

Available plugins (embedded):
- `it2-session-is-at-prompt` - Check if session is at a shell prompt
- `it2-session-claude-has-modal` - Check if Claude has a modal dialog
- `it2-session-claude-suggest-action` - Suggest interventions for stuck Claude sessions
- `it2-session-claude-auto-approve` - Auto-approve safe operations
- `it2-claude-code-hook` - Log Claude Code events to session artifacts

To override embedded plugins, place your own version in your `PATH`.

See [docs/TAXONOMY.md](docs/TAXONOMY.md) for the full extensibility model.

For more details on Claude Code integration, see [docs/CLAUDE_PLUGINS_REFERENCE.md](docs/CLAUDE_PLUGINS_REFERENCE.md).

## Quick Start

Try these commands to get started:

```bash
# List all terminal sessions
it2 session list

# Send text to the current session
it2 session send-text "echo Hello, iTerm2!"

# Create a new tab
it2 tab create "Default"

# Split current pane vertically
it2 session split --vertical

# Split current pane horizontally
it2 session split --horizontal
```

### Your First Automation Script

Create a development environment with one command:

```bash
# Create a new tab and split it into three panes
it2 tab create "Default"
it2 session send-text "cd ~/project && vim"
it2 session split --horizontal
it2 session send-text "npm run dev"
it2 session split --vertical
it2 session send-text "git status && git log --oneline -5"
```

## Common Use Cases

### Development Workflow

Setup a complete development environment:

```bash
# Create a new tab with editor, dev server, and git status
it2 tab create "Default"
it2 session send-text "cd ~/project && vim"
it2 session split --vertical
it2 session send-text "npm run dev"
it2 session split --horizontal
it2 session send-text "git status"
```

### Multi-Server Management

Connect to multiple servers at once:

```bash
for server in web1 web2 db1; do
    it2 tab create "Default"
    it2 session send-text "ssh $server"
done
```

### Broadcast Commands

Run the same command in all sessions:

```bash
for session in $(it2 session list --format json | jq -r '.[].id'); do
    it2 session send-text "$session" "git pull && git status"
done
```

### Save and Restore Layouts

Save your current window arrangement:

```bash
it2 arrangement save "my-dev-setup"

# Later, restore it
it2 arrangement list
it2 arrangement restore "my-dev-setup"
```

## Command Reference

### Essential Commands

**Sessions & Tabs:**
- `it2 session list` - List all terminal sessions
- `it2 session split [--vertical|--horizontal]` - Split current pane
- `it2 session send-text <text>` - Send text to session
- `it2 tab create <profile>` - Create new tab
- `it2 window create` - Create new window

**Text & Content:**
- `it2 text get-buffer` - Get session buffer content
- `it2 clipboard paste` - Paste from clipboard
- `it2 badge set <text>` - Set session badge

**Monitoring:**
- `it2 notification monitor --type <type>` - Monitor events
- `it2 variable monitor <scope> <name>` - Watch variable changes

### All Command Categories

<details>
<summary>Click to expand full command list</summary>

**Application Control:**
- `app` - Control iTerm2 application (focus, variables, properties)
- `auth` - Manage API authentication and connection
- `arrangement` - Save and restore window arrangements

**Session & Tab Management:**
- `session` - Create, list, close, split, activate, restart sessions
- `tab` - Create, list, close, move, activate tabs
- `window` - Create, list, close, focus windows

**Text & Content Operations:**
- `text` - Buffer operations, cursor control, search, selection
- `selection` - Control and retrieve text selection
- `badge` - Display informational badges on sessions
- `clipboard` - Manage clipboard operations
- `screen` - Screen capture utilities

**Shell Integration Features** (requires Shell Integration):
- `prompt` - Access command history, prompts, exit codes
- `job` - Monitor running processes and jobs
- `shell` - Detect shell state (ready/busy/tui) and wait for prompts

**Customization:**
- `profile` - Manage terminal profiles and their properties
- `color` - Import/export color presets and themes
- `variable` - Get/set iTerm2 variables with scope control

**Monitoring & Events:**
- `notification` - Subscribe to real-time iTerm2 events
- `broadcast` - Manage broadcast input domains
- `statusbar` - Configure status bar components

**Extensibility:**
- `plugin` - Manage and run it2 plugins

**Advanced:**
- `tmux` - Control tmux integration
- `completion` - Generate shell completion scripts
- `config` - Manage it2 configuration

</details>

## Configuration

### Global Flags

All commands support these flags:

```bash
--format string      # Output format: text, json, yaml (default "text")
--timeout duration   # Connection timeout (default 5s)
--url string         # API URL (auto-detects Unix socket)
```

### Environment Variables

```bash
ITERM_SESSION_ID     # Current session ID (auto-set by iTerm2)
ITERM2_COOKIE        # Auth cookie (auto-requested if not set)
ITERM2_KEY           # Auth key (auto-requested if not set)
ITERM2_DEBUG=1       # Enable debug logging
```

### Output Formats

Get data in different formats:

```bash
it2 session list                    # Human-readable text
it2 session list --format json      # JSON for scripting
it2 profile get "Default" --format yaml  # YAML for config
```

## Advanced Features

### Shell Integration

Enable advanced features by installing iTerm2 Shell Integration:

**Menu → iTerm2 → Install Shell Integration**

This unlocks:
- Command history access
- Exit code tracking
- Job monitoring
- Shell state detection

```bash
# Show command history with exit codes
it2 prompt list

# Search for failed commands
it2 prompt search "git" | grep "Exit: [^0]"

# Wait for shell to be ready
it2 shell wait-for-prompt && it2 session send-text "echo ready"
```

### Real-time Event Monitoring

Subscribe to live iTerm2 events:

```bash
# Monitor keystrokes
it2 notification monitor --type keystroke

# Watch for new sessions
it2 notification monitor --type session

# Monitor variable changes
it2 variable monitor session user.myvar
```

### Authentication

Authentication is automatic:
1. First connection requests credentials via AppleScript
2. iTerm2 prompts you to allow API access
3. Credentials cached for session duration

```bash
# Check auth status
it2 auth check

# Force re-authentication
it2 auth request
```

## Usage Tips

### Session Identification

Sessions can be referenced by:
- Full format: `w0t1p12:C3D91F33-3805-47E2-A3F6-B8AED6EC2209`
- UUID only: `C3D91F33-3805-47E2-A3F6-B8AED6EC2209`
- When omitted, uses `$ITERM_SESSION_ID` (current session)

### Scripting Best Practices

```bash
# Use JSON output for parsing
SESSIONS=$(it2 session list --format json)
echo "$SESSIONS" | jq -r '.[].id'

# Always quote variables
it2 session send-text "$SESSION_ID" "$COMMAND"

# Chain commands carefully
it2 session send-text "cd /path && pwd && ls"
```

## Troubleshooting

### Connection Problems

```bash
# Check if iTerm2 API is enabled
it2 auth check

# Enable debug output
ITERM2_DEBUG=1 it2 session list

# Verify API is enabled in iTerm2
# Preferences → General → Magic → Enable Python API
```

### Shell Integration Not Working

```bash
# Verify Shell Integration is installed
echo $ITERM_SESSION_ID  # Should show session ID

# Reinstall if needed
curl -L https://iterm2.com/shell_integration/install_shell_integration_and_utilities.sh | bash
```

### Session ID Not Found

```bash
# List all sessions to find the correct ID
it2 session list

# Use UUID format or full format
it2 session send-text C3D91F33-3805-47E2-A3F6-B8AED6EC2209 "echo test"
```

## Go Package

For using it2 as a Go library in your own projects:

```bash
go get github.com/tmc/it2
```

See documentation at [pkg.go.dev/github.com/tmc/it2](https://pkg.go.dev/github.com/tmc/it2)

## Resources

- **Project Repository**: [github.com/tmc/it2](https://github.com/tmc/it2)
- **iTerm2 API Docs**: [iterm2.com/documentation-api.html](https://iterm2.com/documentation-api.html)
- **Shell Integration Guide**: [iterm2.com/documentation-shell-integration.html](https://iterm2.com/documentation-shell-integration.html)

Get help for any command:
```bash
it2 help [command]
it2 [command] --help
```
