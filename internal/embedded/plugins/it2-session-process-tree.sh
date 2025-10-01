#!/bin/bash
# it2-session-process-tree
# Example plugin to get process tree information for a session's process
#
# Called with: <session-id> <pid>
#
# Outputs: Process tree information (implementation can vary)

SESSION_ID="$1"
PID="$2"

if [ -z "$PID" ]; then
    exit 0
fi

# Get process name
PROC_NAME=$(ps -p "$PID" -o comm= 2>/dev/null)

if [ -z "$PROC_NAME" ]; then
    exit 0
fi

# Get parent processes (up to 3 levels)
TREE=""
CURRENT_PID="$PID"
for i in 1 2 3; do
    PARENT_PID=$(ps -p "$CURRENT_PID" -o ppid= 2>/dev/null | tr -d ' ')
    if [ -z "$PARENT_PID" ] || [ "$PARENT_PID" = "1" ]; then
        break
    fi

    PARENT_NAME=$(ps -p "$PARENT_PID" -o comm= 2>/dev/null)
    if [ -n "$PARENT_NAME" ]; then
        if [ -z "$TREE" ]; then
            TREE="$PARENT_NAME"
        else
            TREE="$TREE → $PARENT_NAME"
        fi
    fi

    CURRENT_PID="$PARENT_PID"
done

# Output tree if we found parents
if [ -n "$TREE" ]; then
    echo "$PROC_NAME → $TREE"
else
    echo "$PROC_NAME"
fi
