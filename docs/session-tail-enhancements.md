# Session Tail Enhancement Ideas

Based on analysis of iTerm2's API, here are potential improvements for `it2 session tail`.

## 1. Event-Driven Architecture

### Available Notifications

```
NOTIFY_ON_KEYSTROKE = 1           // User input events
NOTIFY_ON_SCREEN_UPDATE = 2       // Screen content changed ⭐
NOTIFY_ON_PROMPT = 3              // Shell prompt events ⭐
NOTIFY_ON_CUSTOM_ESCAPE_SEQUENCE = 5  // Custom OSC sequences
NOTIFY_ON_VARIABLE_CHANGE = 12    // Session variables changed
```

### Prompt Notifications

`NOTIFY_ON_PROMPT` provides structured command tracking:

```protobuf
message PromptNotification {
  optional string session = 1;
  oneof event {
    PromptNotificationPrompt prompt = 2;        // New prompt shown
    PromptNotificationCommandStart command_start = 3;  // Command started
    PromptNotificationCommandEnd command_end = 4;      // Command ended (with status)
  }
  optional string unique_prompt_id = 5;
}
```

**Use cases:**
- `tail --commands-only` - Show only command output, skip prompts
- `tail --with-status` - Show exit status after each command
- `tail --failed-only` - Only show output from failed commands (status != 0)
- `tail --command-boundaries` - Add visual separators between commands

## 2. Advanced Filtering

### Content Filters

```bash
# Only show lines matching pattern
it2 session tail -f --grep "ERROR|WARN"

# Exclude lines matching pattern
it2 session tail -f --grep-v "DEBUG"

# Case-insensitive matching
it2 session tail -f --grep "error" -i

# Show context around matches
it2 session tail -f --grep "ERROR" -A 3 -B 2
```

### Smart Filtering with Prompts

```bash
# Only show output from failed commands
it2 session tail -f --failed-commands

# Only show commands that take >5 seconds
it2 session tail -f --slow-commands 5s

# Skip interactive output (vim, less, etc)
it2 session tail -f --non-interactive
```

## 3. Coordinate-Based Efficient Fetching

Instead of fetching last N lines, use absolute coordinates:

```go
type TailState struct {
    LastSeenY int64  // Last processed line coordinate
}

// On notification:
// Fetch using windowed_coord_range: {start: {y: lastSeenY+1}, end: {y: maxInt}}
// Only fetches truly new lines
```

**Benefits:**
- No duplicate output
- Works correctly with scrollback
- Minimal data transfer

## 4. Output Formatting Enhancements

### Timestamps

```bash
# Add timestamps to each line
it2 session tail -f --timestamps

# Show relative time since start
it2 session tail -f --relative-time

# Show time since last output
it2 session tail -f --delta-time
```

### Decorations

```bash
# Add session name to each line (for multi-session tail)
it2 session tail -f --show-session

# Color-code by session
it2 session tail sess1 sess2 -f --color-by-session

# Show command that produced output
it2 session tail -f --show-command
```

## 5. Multi-Session Tailing

```bash
# Tail multiple sessions, interleaved
it2 session tail sess1 sess2 sess3 -f

# Multiplex output side-by-side (with terminal splitting)
it2 session tail sess1 sess2 -f --split

# Follow all sessions in current window
it2 session tail --all-in-window -f

# Auto-tail new sessions as they're created
it2 session tail --follow-new -f
```

Implementation: Subscribe to `NOTIFY_ON_NEW_SESSION` and `NOTIFY_ON_TERMINATE_SESSION`

## 6. Performance Optimizations

### Batch Notifications

Instead of processing each screen update immediately, batch them:

```go
type NotificationBatcher struct {
    updates chan string  // Session IDs
    ticker  *time.Ticker
}

// Collect notifications for 100ms, then process once
```

**Benefits:**
- Reduces API calls during rapid output
- Better performance for fast-scrolling sessions
- Lower CPU usage

### Smart Polling Fallback

```bash
# Start with notifications, fall back to polling if connection lost
it2 session tail -f --adaptive

# Adjust polling interval based on activity
it2 session tail -f --adaptive-interval
```

## 7. Output Buffering & Replay

```bash
# Keep last 1000 lines in memory, allow scrollback
it2 session tail -f --buffer 1000

# Save all output to file while tailing
it2 session tail -f --tee output.log

# Replay last N seconds of buffer
it2 session tail -f --replay 30s
```

## 8. Conditional Actions

### Triggers

```bash
# Run command when pattern matches
it2 session tail -f --on-match "ERROR" --exec "notify-send 'Error detected'"

# Alert when command fails
it2 session tail -f --on-fail "osascript -e 'beep'"

# Pause tail when pattern seen (like less/more)
it2 session tail -f --pause-on "Press ENTER"
```

### Custom Escape Sequences

Subscribe to `NOTIFY_ON_CUSTOM_ESCAPE_SEQUENCE`:

```bash
# Trigger on custom OSC sequences from running programs
it2 session tail -f --on-escape "my-app" --exec "handle_event.sh"
```

## 9. Window/Coordinate Range Queries

Use `WindowedCoordRange` for precise fetching:

```protobuf
message WindowedCoordRange {
  optional CoordRange coord_range = 1;  // Line range
  optional Range columns = 2;            // Column range
}
```

**Use cases:**
- `--columns 0-80` - Only show first 80 columns (trim wide output)
- `--crop` - Auto-detect and crop to terminal width
- Extract specific columns for structured logs

## 10. Integration Features

### JSON Output

```bash
# Output as JSON stream for piping to jq
it2 session tail -f --json | jq 'select(.status != 0)'

# Include metadata
{
  "timestamp": "2025-10-01T03:45:00Z",
  "session": "ABC123",
  "line": "Error: connection failed",
  "command": "curl https://api.example.com",
  "exit_status": null  // still running
}
```

### Webhooks

```bash
# POST lines to webhook
it2 session tail -f --webhook https://example.com/logs

# Only POST on match
it2 session tail -f --grep "ERROR" --webhook https://example.com/alerts
```

## Implementation Priority

1. **High Priority:**
   - [ ] Event-driven with NOTIFY_ON_SCREEN_UPDATE
   - [ ] Coordinate-based fetching
   - [ ] Basic grep filtering
   - [ ] Multi-session tail

2. **Medium Priority:**
   - [ ] NOTIFY_ON_PROMPT integration
   - [ ] Command-aware filtering
   - [ ] Timestamps
   - [ ] Output buffering

3. **Low Priority:**
   - [ ] Triggers and webhooks
   - [ ] JSON output mode
   - [ ] Advanced decorations
   - [ ] Column range queries

## References

- iTerm2 API: `/Volumes/tmc/go/src/github.com/gnachman/iTerm2/proto/api.proto`
- Notification types: Lines 869-887
- Prompt notifications: Lines 1058-1079
- Line ranges: Lines 1321-1332
- Python examples: `/Volumes/tmc/go/src/github.com/gnachman/iTerm2/api/library/python/iterm2/docs/examples/`
