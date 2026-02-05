---
name: claude-code-integration
description: Use when working with Claude Code sessions in iTerm2 - detecting Claude sessions, checking their status, linking sessions, or coordinating multiple Claude instances.
allowed-tools:
  - Bash
  - Read
  - Grep
---

# Claude Code Integration

Guidance for interacting with Claude Code sessions running in iTerm2.

## When to Use

- Detecting which iTerm2 sessions are running Claude Code
- Checking the status of Claude Code sessions (idle, working, waiting for input)
- Coordinating multiple Claude Code sessions
- Monitoring Claude Code session activity
- Setting up multi-agent Claude Code workflows

## When NOT to Use

- For general terminal automation (use session-orchestration skill instead)
- When not working with Claude Code specifically
- For Claude API interactions (this is for the CLI tool in terminals)

## Quick Start

```bash
# List sessions and identify Claude Code instances
it2 session list --format json | jq '.[] | select(.name | contains("claude"))'

# Get Claude-specific session info
it2 session claude "$SESSION_ID"

# Get rich Claude Code status
it2 session claude-status "$SESSION_ID"

# Check if session has a modal dialog (permission prompt, etc.)
it2 session has-modal "$SESSION_ID"
```

## Detecting Claude Code Sessions

### By Process Name

```bash
# Find sessions running Claude
it2 session list --format json | jq -r '.[] |
  select(.foreground_job.command | test("claude"; "i")) | .session_id'
```

### By Session State

```bash
# Get comprehensive state analysis
it2 session get-state "$SID"
```

### Claude Session Indicators

A session is likely running Claude Code if:
- Process name contains "claude" or "node" (Claude Code's runtime)
- Screen content shows Claude Code UI elements (╭─, tool use blocks, etc.)
- Session has specific environment variables set

## Session Status Detection

### Modal Detection

Claude Code shows modal dialogs for permissions, confirmations, etc.

```bash
# Check for modal
if it2 session has-modal "$SID"; then
  echo "Session has pending modal"
fi
```

### Activity Detection

```bash
# Check if session is actively working
it2 session is-active "$SID"

# Suggest intervention action
it2 session suggest-action "$SID"
```

## Session Linking

Link iTerm2 sessions to their Claude Code session files:

```bash
# Get Claude session link info
it2 session claude "$SID"

# This reveals:
# - Session file path
# - Working directory
# - Session metadata
```

## Multi-Claude Workflows

### Setting Up Parallel Claude Sessions

```bash
# Create splits for multiple Claude instances
MAIN=$ITERM_SESSION_ID
RESEARCH=$(it2 session split --vertical -q)
IMPLEMENT=$(it2 session split --horizontal -q)

# Badge each for identification
it2 session set-badge "$MAIN" "$(echo $MAIN | cut -c1-8)\nMain"
it2 session set-badge "$RESEARCH" "$(echo $RESEARCH | cut -c1-8)\nResearch"
it2 session set-badge "$IMPLEMENT" "$(echo $IMPLEMENT | cut -c1-8)\nImplement"

# Start Claude in each
it2 session send-text "$RESEARCH" "claude --continue"
it2 session send-text "$IMPLEMENT" "claude"
```

### Monitoring Multiple Sessions

```bash
# Watch all sessions for state changes
it2 session watch --all
```

## Screen Content Analysis

### Getting Claude's Output

```bash
# Get visible screen
it2 session get-screen "$SID"

# Get more context from buffer
it2 session get-buffer "$SID" --lines 200
```

### Parsing Claude UI Elements

Claude Code's TUI has recognizable patterns:
- `╭─` / `╰─` - Message box borders
- `●` - Working indicator
- `?` - Question/prompt indicator
- Tool use blocks with specific formatting

## Best Practices

1. **Always check for modals** before sending input to Claude sessions
2. **Use badges** with session ID prefix for visual tracking
3. **Monitor state** before automated interactions
4. **Verify screen content** matches expected state before proceeding
5. **Use quiet flags** (`-q`) when capturing session IDs for scripts
6. **Include response hints** when messaging other sessions - always tell them how to respond back (see [workflows/multi-agent.md](workflows/multi-agent.md#inter-session-communication))

## Advanced Usage

See [references/state-detection.md](references/state-detection.md) for detailed state analysis.
See [references/ui-patterns.md](references/ui-patterns.md) for parsing Claude Code output.
See [workflows/multi-agent.md](workflows/multi-agent.md) for coordinating multiple Claude instances.
