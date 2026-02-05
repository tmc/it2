---
name: it2
description: Control iTerm2 terminals. Use when splitting panes, sending commands to sessions, checking session output, setting badges, creating layouts, or coordinating multiple terminals.
allowed-tools:
  - Bash
  - Read
---

# it2 - iTerm2 CLI Automation

## Essential Commands

```bash
# List sessions
it2 session list

# Get current session ID
echo $ITERM_SESSION_ID

# Split pane (-q returns new session ID)
NEW=$(it2 session split --horizontal -q)  # pane below
NEW=$(it2 session split --vertical -q)    # pane right

# Send command to session
it2 session send-text "$SID" "your command here"

# Read session content
it2 session get-screen "$SID"           # visible screen
it2 session get-buffer "$SID" --lines 50  # scrollback

# Set badge (always include session ID prefix)
it2 session set-badge "$SID" "$(echo $SID | cut -c1-8)\nLabel"
```

## Discovering Sessions

```bash
# List all sessions
it2 session list

# With JSON for scripting
it2 session list --format=json | jq -r '.[] | "\(.session_id[:8]) \(.cwd)"'
```

## Before Sending Commands

1. Check target session state first: `it2 session get-screen "$SID"`
2. Verify it's not in vim/modal/busy state
3. Use `it2 session list` to find session IDs

## Before Splitting

Check dimensions to plan layout:
```bash
it2 session get-info --format=json | jq .grid_size
```

## Common Patterns

### Dev Layout (main + logs + server)
```bash
MAIN=$ITERM_SESSION_ID
LOGS=$(it2 session split --horizontal -q)
SERVER=$(it2 session split --vertical -q)
it2 session send-text "$LOGS" "tail -f *.log"
it2 session send-text "$SERVER" "npm start"
```

### Send to Another Claude Session
```bash
it2 session send-text "$TARGET_SID" "your message"
```

### Monitor a Session
```bash
while true; do it2 session get-screen "$SID"; sleep 5; done
```

## When Things Fail

- **Session not found**: Re-run `it2 session list` - sessions close, IDs change
- **Send-text not delivered**: Session is at modal/vim/busy - check screen first
- **Split fails**: Terminal too small - check grid_size, need ~80x24 minimum per pane
