# iTerm2 Notification Types

Complete reference for subscribable iTerm2 notifications.

## Notification Categories

### Session Events

| Type | Description | Data |
|------|-------------|------|
| `session_ended` | Session terminated | session_id |
| `session_created` | New session created | session_id, tab_id |
| `focus_changed` | Session focus changed | old_session, new_session |

### Screen Events

| Type | Description | Data |
|------|-------------|------|
| `screen_update` | Screen content changed | session_id, lines_changed |
| `cursor_moved` | Cursor position changed | session_id, x, y |
| `selection_changed` | Text selection changed | session_id, text |

### Input Events

| Type | Description | Data |
|------|-------------|------|
| `keystroke` | Key pressed | session_id, key, modifiers |
| `prompt` | Shell prompt event | session_id, command |

### Variable Events

| Type | Description | Data |
|------|-------------|------|
| `variable_changed` | Variable value changed | scope, name, old, new |

## Subscribing

### Basic Subscribe

```bash
# Subscribe to screen updates
it2 subscribe screen_update --session "$SID"

# Subscribe globally (all sessions)
it2 subscribe keystroke
```

### With Filters

```bash
# Only specific session
it2 notification monitor --type screen_update --session "$SID"

# JSON output for parsing
it2 subscribe session_ended --format json
```

## Event Payload Examples

### keystroke

```json
{
  "type": "keystroke",
  "session_id": "ABC123...",
  "timestamp": "2024-01-15T10:30:00Z",
  "data": {
    "key": "a",
    "modifiers": ["ctrl"],
    "raw_code": 1
  }
}
```

### screen_update

```json
{
  "type": "screen_update",
  "session_id": "ABC123...",
  "timestamp": "2024-01-15T10:30:00Z",
  "data": {
    "lines_changed": [1, 5, 10],
    "cursor": {"x": 0, "y": 5}
  }
}
```

### prompt

```json
{
  "type": "prompt",
  "session_id": "ABC123...",
  "timestamp": "2024-01-15T10:30:00Z",
  "data": {
    "command": "git status",
    "working_directory": "/path/to/project",
    "exit_status": 0
  }
}
```

## Processing Events

```bash
# Parse events with jq
it2 subscribe screen_update --format json | while read event; do
  SID=$(echo "$event" | jq -r '.session_id')
  LINES=$(echo "$event" | jq -r '.data.lines_changed | length')
  echo "Session $SID: $LINES lines changed"
done
```

## Unsubscribing

```bash
# List active subscriptions
it2 notification status

# Unsubscribe by ID
it2 notification unsubscribe "$SUBSCRIPTION_ID"
```
