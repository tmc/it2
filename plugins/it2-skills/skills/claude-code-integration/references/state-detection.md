# Claude Code State Detection

Techniques for detecting and interpreting Claude Code session states.

## State Categories

### Idle

Session is at shell prompt, Claude Code not running or waiting for input.

Indicators:
- `$` or `%` prompt visible at bottom
- No spinning indicator
- Screen stable

```bash
STATE=$(it2 session get-state "$SID" --format json | jq -r '.state')
if [ "$STATE" = "idle" ]; then
  echo "Ready for input"
fi
```

### Working

Claude is actively processing.

Indicators:
- Spinning/progress indicator visible
- Screen updating
- No prompt at bottom

### Waiting

Claude needs user input or approval.

Indicators:
- Question mark indicator
- Permission prompt visible
- Modal dialog present

```bash
if it2 session has-modal "$SID"; then
  echo "Claude waiting for approval"
fi
```

## Detailed State Analysis

```bash
# Full state breakdown
it2 session get-state "$SID" --format json
```

Returns:
```json
{
  "state": "working|idle|waiting",
  "has_modal": false,
  "foreground_process": "claude",
  "screen_stable": true,
  "last_change": "2024-01-15T10:30:00Z"
}
```

## Screen Content Patterns

Claude Code TUI elements to detect:

| Pattern | Meaning |
|---------|---------|
| `●` | Working indicator |
| `?` | Awaiting input |
| `✓` | Task complete |
| `╭─` | Message box start |
| `╰─` | Message box end |
| `Tool:` | Tool execution |

## Reliable Detection Script

```bash
#!/bin/bash
SID="$1"

SCREEN=$(it2 session get-screen "$SID")
HAS_MODAL=$(it2 session has-modal "$SID" && echo "true" || echo "false")

if [ "$HAS_MODAL" = "true" ]; then
  echo "waiting:modal"
elif echo "$SCREEN" | grep -q '●'; then
  echo "working"
elif echo "$SCREEN" | grep -q '\?'; then
  echo "waiting:input"
elif echo "$SCREEN" | grep -qE '^\$|^%'; then
  echo "idle"
else
  echo "unknown"
fi
```

## Stability Detection

Wait for screen to stabilize:

```bash
wait_stable() {
  local SID="$1"
  local PREV=""
  local CURR=""

  while true; do
    CURR=$(it2 session get-screen "$SID" | md5)
    if [ "$CURR" = "$PREV" ]; then
      break
    fi
    PREV="$CURR"
    sleep 0.5
  done
}
```
