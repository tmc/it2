#!/bin/bash

set -e

echo "=== Consolidating sessions by CWD and ordering by creation time ==="

# Get all sessions grouped by CWD, sorted by PID within each group
it2 session list --format=json | jq -r '
  group_by(.WorkingDirectory) |
  map(
    select(length > 1) |
    {
      cwd: .[0].WorkingDirectory,
      sessions: (. | sort_by(.ShellPID) | map({id: .SessionID, short: .ShortID, pid: .ShellPID}))
    }
  )
' > /tmp/session_groups.json

echo "Session groups by CWD:"
cat /tmp/session_groups.json | jq -r '.[] | "\(.cwd): \(.sessions | length) sessions"'
echo ""

# For each CWD group with multiple sessions
cat /tmp/session_groups.json | jq -c '.[]' | while read -r group; do
    cwd=$(echo "$group" | jq -r '.cwd')
    session_count=$(echo "$group" | jq -r '.sessions | length')

    echo "=== Processing: $cwd ($session_count sessions) ==="

    # Get sessions sorted by PID (oldest to newest)
    sessions=($(echo "$group" | jq -r '.sessions[].id'))

    if [ ${#sessions[@]} -lt 2 ]; then
        echo "Only one session, skipping..."
        continue
    fi

    # First session is the anchor (oldest)
    anchor="${sessions[0]}"
    echo "Anchor (oldest): $anchor"

    # Move all other sessions to split with anchor (maintaining time order)
    for ((i=1; i<${#sessions[@]}; i++)); do
        sid="${sessions[$i]}"
        echo "Moving session $i: $sid to split with $anchor"
        it2 session move "$sid" "$anchor" --quiet || {
            echo "Warning: Failed to move $sid, continuing..."
        }
    done

    echo ""
done

echo "=== Done consolidating sessions ==="
echo ""
echo "Final session distribution:"
it2 session list --format=json | jq -r '
  group_by(.WorkingDirectory) |
  map({
    cwd: .[0].WorkingDirectory,
    tab: .[0].TabID,
    count: length
  }) |
  .[] |
  "Tab \(.tab): \(.count) sessions in \(.cwd)"
'
