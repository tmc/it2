---
name: session-orchestration
description: Use when managing multiple iTerm2 sessions, creating split layouts, navigating session hierarchies, or orchestrating terminal workflows across panes.
allowed-tools:
  - Bash
  - Read
  - Grep
---

# Session Orchestration

Guidance for managing iTerm2 sessions using the `it2` CLI tool.

## When to Use

- Creating multi-pane terminal layouts for development workflows
- Managing multiple concurrent terminal sessions
- Navigating and organizing sessions across windows and tabs
- Sending commands to specific sessions programmatically
- Setting up reproducible terminal environments

## When NOT to Use

- For simple single-session terminal operations (just use the terminal directly)
- When iTerm2 is not the active terminal emulator
- For remote server management (use SSH directly, then optionally it2 for local orchestration)

## Quick Start

```bash
# List all sessions
it2 session list

# Get current session ID
it2 session current

# Create a horizontal split (new pane below)
it2 session split --horizontal -q

# Create a vertical split (new pane to the right)
it2 session split --vertical -q

# Send text to a session
it2 session send-text "$SESSION_ID" "echo hello"
```

## Core Concepts

### Session Identification

Sessions are identified by UUIDs (e.g., `A81F7ED5-3F4B-4504-8F14-6C98F9422F72`). Use the first 8 characters as shorthand for display purposes.

```bash
# Get current session
CURRENT_SID=$ITERM_SESSION_ID

# List sessions with JSON output for parsing
it2 session list --format json
```

### Session Hierarchy

Sessions exist within a hierarchy: **Window → Tab → Session (Pane)**

```bash
# Look up parent tab
it2 session lookup tab "$SID"

# Look up parent window
it2 session lookup window "$SID"

# Find sibling sessions (same split root)
it2 session lookup siblings "$SID"
```

### Split Operations

The `-q` flag returns only the new session ID (quiet mode), essential for scripting.

```bash
# Create split and capture new session ID
NEW_SID=$(it2 session split --horizontal -q)

# Send command to new session
it2 session send-text "$NEW_SID" "cd /path/to/project && npm start"
```

## Layout Patterns

### Development Layout (3 panes)

```bash
# Start from main session
MAIN=$ITERM_SESSION_ID

# Create editor pane (right)
EDITOR=$(it2 session split --vertical -q)

# Create log pane (bottom of main)
LOGS=$(it2 session split --horizontal -q)

# Set up each pane
it2 session send-text "$EDITOR" "vim ."
it2 session send-text "$LOGS" "tail -f logs/*.log"
```

### Badge Sessions for Identification

Always include the session ID prefix in badges for automation:

```bash
it2 session set-badge "$SID" "$(echo $SID | cut -c1-8)\nServer"
```

## Session State Inspection

```bash
# Get comprehensive session info
it2 session get-info "$SID" --format json

# Check if session has shell integration
it2 session has-shell-integration "$SID"

# Get current screen contents
it2 session get-screen "$SID"

# Get scrollback buffer (last 100 lines)
it2 session get-buffer "$SID" --lines 100
```

## Best Practices

1. **Always use `-q` flag** when capturing session IDs from split operations
2. **Check session dimensions** before splitting: `it2 session get-info --format json | jq .grid_size`
3. **Use badges** to label sessions for visual identification
4. **Verify target session state** before sending commands: check screen contents first
5. **Chain commands** with `&&` or `;` when sending multiple commands

## Advanced Usage

See [references/hierarchy.md](references/hierarchy.md) for session hierarchy navigation.
See [references/layouts.md](references/layouts.md) for complex layout recipes.
See [workflows/multi-project.md](workflows/multi-project.md) for multi-project setups.
