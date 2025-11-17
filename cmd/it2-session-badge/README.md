# it2-session-badge

Updates iTerm2 session badges based on session events using Go templates.

## Features

- **Go template support** - Flexible badge formatting with custom functions
- **Rate limiting** - Prevents excessive updates (configurable interval)
- **Event monitoring** - Tracks plugin events from session artifacts
- **Session-aware** - Access to session ID, event count, timestamps, etc.

## Usage

```bash
# Update badge for current session
it2-session-badge

# Update with custom template
it2-session-badge -template='{{.SessionID | trunc 8}}\n{{.EventCount}} events'

# Update specific session
it2-session-badge -session <session-id>

# Set minimum update interval
it2-session-badge -min-interval 10s

# Verbose output
it2-session-badge -verbose
```

## Template Data

The badge template has access to:

```go
type SessionData struct {
    SessionID      string    // Full session ID
    EventCount     int       // Number of plugin events
    LastEventTime  time.Time // Time of last event
    WorkingDir     string    // Current working directory
    Hostname       string    // System hostname
}
```

## Template Functions

### trunc
Truncate string to N characters:
```
{{.SessionID | trunc 8}}  // "9FCB9E39"
```

### duration
Human-readable time since:
```
{{.LastEventTime | duration}}  // "2m30s ago"
```

### shortTime
Format time as HH:MM:SS:
```
{{.LastEventTime | shortTime}}  // "14:32:15"
```

## Examples

### Minimal - Session ID only
```bash
it2-session-badge -template='{{.SessionID | trunc 8}}'
```

### Event counter
```bash
it2-session-badge -template='{{.SessionID | trunc 8}}\n{{.EventCount}} events'
```

### With timestamp
```bash
it2-session-badge -template='{{.SessionID | trunc 8}}\n{{.EventCount}} events\n{{.LastEventTime | shortTime}}'
```

### Full status
```bash
it2-session-badge -template='{{.SessionID | trunc 8}}\n{{.WorkingDir}}\n{{.EventCount}} events ({{.LastEventTime | duration}})'
```

## Rate Limiting

Updates are rate-limited using a timestamp file:
- Location: `~/.it2/sessions/$SESSION_ID/.last-badge-update`
- Default interval: 5 seconds
- Prevents excessive iTerm2 API calls

## Integration

### Manual trigger
```bash
it2-session-badge
```

### From plugin
Add to your plugin to update badge after changes:
```bash
#!/bin/bash
# ... your plugin logic ...
it2-session-badge -min-interval 2s
```

### Periodic updates
Run in background with watch:
```bash
watch -n 10 'it2-session-badge -min-interval 5s'
```

### Hook integration
Add to `.claude/hooks.json`:
```json
{
  "hooks": {
    "PostToolUse": [{
      "hooks": [{
        "type": "command",
        "command": "it2-session-badge -min-interval 5s"
      }]
    }]
  }
}
```

## Event Monitoring

By default, monitors: `~/.it2/sessions/$SESSION_ID/artifacts/plugin-events.ndjson`

Override with `-events` flag:
```bash
it2-session-badge -events /custom/path/events.ndjson
```
