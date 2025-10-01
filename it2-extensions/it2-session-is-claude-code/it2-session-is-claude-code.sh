#!/bin/bash
# it2-session-is-claude-code
# Plugin to detect Claude Code sessions by checking if the main process is claude

# Get session ID - use argument if provided, otherwise use current session
SESSION_ID="${1:-$(it2 session current 2>/dev/null)}"

# Debug output if ITERM2_DEBUG is set
if [ "$ITERM2_DEBUG" = "1" ]; then
    echo "[DEBUG] it2-session-is-claude-code: SESSION_ID=$SESSION_ID" >&2
fi

# If no session ID, cannot determine
if [ -z "$SESSION_ID" ]; then
    if [ "$ITERM2_DEBUG" = "1" ]; then
        echo "[DEBUG] No session ID available" >&2
    fi
    exit 0
fi

# Get the PID for this session using get-pid
SESSION_PID=$(it2 session get-pid "$SESSION_ID" 2>/dev/null)

if [ "$ITERM2_DEBUG" = "1" ]; then
    echo "[DEBUG] Got PID from get-pid: $SESSION_PID" >&2
fi

# If no PID found, cannot determine
if [ -z "$SESSION_PID" ]; then
    if [ "$ITERM2_DEBUG" = "1" ]; then
        echo "[DEBUG] Could not get PID from session" >&2
    fi
    exit 0
fi

# Get the process name for the PID
PROCESS_NAME=$(ps -p "$SESSION_PID" -o comm= 2>/dev/null)

if [ "$ITERM2_DEBUG" = "1" ]; then
    echo "[DEBUG] Process name for PID $SESSION_PID: $PROCESS_NAME" >&2
fi

# Check if the process name is claude (or contains claude in the path)
if [[ "$PROCESS_NAME" == *"claude"* ]]; then
    echo "✅"
    exit 0
fi

# Check parent process chain (up to 3 levels)
CURRENT_PID="$SESSION_PID"
for i in 1 2 3; do
    PARENT_PID=$(ps -p "$CURRENT_PID" -o ppid= 2>/dev/null | tr -d ' ')
    if [ -z "$PARENT_PID" ] || [ "$PARENT_PID" = "1" ]; then
        break
    fi

    PARENT_NAME=$(ps -p "$PARENT_PID" -o comm= 2>/dev/null)

    if [ "$ITERM2_DEBUG" = "1" ]; then
        echo "[DEBUG] Parent $i PID $PARENT_PID: $PARENT_NAME" >&2
    fi

    if [[ "$PARENT_NAME" == *"claude"* ]]; then
        echo "✅"
        exit 0
    fi

    CURRENT_PID="$PARENT_PID"
done

# Not a Claude session
exit 1
