#!/bin/bash
# it2-session-process-is-claude
# Plugin to check if a process is running Claude Code
#
# Called with: <session-id> <pid>
#
# Outputs: "true" if the process is Claude Code, empty otherwise

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

# Check if the process is claude
if [[ "$PROC_NAME" == *"claude"* ]]; then
    echo "true"
fi
