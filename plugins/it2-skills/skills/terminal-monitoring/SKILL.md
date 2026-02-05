---
name: terminal-monitoring
description: Use when monitoring iTerm2 session activity, watching for state changes, subscribing to notifications, or implementing automated responses to terminal events.
allowed-tools:
  - Bash
  - Read
  - Grep
---

# Terminal Monitoring

Guidance for monitoring iTerm2 sessions and responding to events.

## When to Use

- Watching sessions for state transitions (idle → active → complete)
- Subscribing to iTerm2 notifications (keystrokes, prompts, screen changes)
- Implementing automated responses to terminal events
- Tailing session output in real-time
- Detecting when processes complete or need attention

## When NOT to Use

- For one-time session queries (use session-orchestration instead)
- When you only need current state, not ongoing monitoring
- For simple command execution without feedback

## Quick Start

```bash
# Tail session output (like tail -f)
it2 session tail -f "$SESSION_ID"

# Watch session for state transitions
it2 session watch "$SESSION_ID"

# Monitor all sessions
it2 session watch --all

# Subscribe to notifications
it2 notification monitor --type keystroke
```

## Notification Types

### Available Notification Types

```bash
# List available notification types
it2 notification list-types
```

Common types:
- `keystroke` - Key presses in sessions
- `screen_update` - Screen content changes
- `prompt` - Shell prompt events (requires shell integration)
- `session_ended` - Session termination
- `variable_changed` - iTerm2 variable modifications

### Subscribing to Notifications

```bash
# Monitor keystrokes
it2 notification monitor --type keystroke

# Subscribe with JSON output for parsing
it2 subscribe screen_update --format json
```

## Session Watching

### Basic Watch

```bash
# Watch single session
it2 session watch "$SID"

# Watch with specific interval
it2 session watch "$SID" --interval 2s
```

### Watch Patterns

The watch command detects:
- State transitions (idle, working, waiting)
- Process changes
- Screen stability
- Modal dialogs

## Tail Mode

Follow session output like `tail -f`:

```bash
# Follow session output
it2 session tail -f "$SID"

# With line limit
it2 session tail -f "$SID" --lines 50
```

## Process Monitoring

### Job Monitoring

```bash
# Monitor jobs in a session
it2 job list "$SID"

# Watch for job completion
it2 job wait "$SID"
```

### Shell Integration Events

If shell integration is enabled:

```bash
# Check shell integration status
it2 session has-shell-integration "$SID"

# Get prompt metadata
it2 session prompt "$SID"

# Search command history
it2 prompt search "git commit"
```

## Automated Responses

### Autorespond

Set up automatic responses to patterns:

```bash
# Auto-respond to prompts
it2 session autorespond "$SID" --pattern "Continue?" --response "y"
```

### State-Based Actions

```bash
# Get suggested action
ACTION=$(it2 session suggest-action "$SID" --format json)

# Act based on state
STATE=$(it2 session get-state "$SID" --format json | jq -r '.state')
case "$STATE" in
  "idle") echo "Session ready for input" ;;
  "working") echo "Session busy" ;;
  "waiting") echo "Session needs attention" ;;
esac
```

## Screen Monitoring

### Detecting Changes

```bash
# Get current screen hash
HASH1=$(it2 session get-screen "$SID" | md5)
sleep 1
HASH2=$(it2 session get-screen "$SID" | md5)

if [ "$HASH1" != "$HASH2" ]; then
  echo "Screen changed"
fi
```

### Content Matching

```bash
# Wait for specific content
while ! it2 session get-screen "$SID" | grep -q "Complete"; do
  sleep 1
done
echo "Task completed"
```

## Event-Driven Monitoring

### Using Subscribe

```bash
# Subscribe to events and process
it2 subscribe screen_update --session "$SID" --format json | while read event; do
  # Process each event
  echo "$event" | jq .
done
```

## Best Practices

1. **Use shell integration** when available for richer monitoring
2. **Prefer event subscriptions** over polling for efficiency
3. **Set reasonable intervals** to avoid overwhelming the system
4. **Handle disconnections** gracefully in long-running monitors
5. **Use JSON output** for programmatic processing

## Advanced Usage

See [references/notifications.md](references/notifications.md) for all notification types.
See [references/shell-integration.md](references/shell-integration.md) for prompt/command tracking.
See [workflows/automated-testing.md](workflows/automated-testing.md) for CI/CD integration.
