#!/bin/bash

set -e

# Parse flags
USE_GIT_ROOT=false
while [[ $# -gt 0 ]]; do
    case $1 in
        --git-root|-g)
            USE_GIT_ROOT=true
            shift
            ;;
        -h|--help)
            echo "Usage: $0 [--git-root|-g]"
            echo ""
            echo "Consolidate iTerm2 sessions by working directory into vertical splits."
            echo ""
            echo "Options:"
            echo "  --git-root, -g  Group sessions by nearest git repository root instead of exact CWD"
            echo "  --help, -h      Show this help message"
            exit 0
            ;;
        *)
            echo "Unknown option: $1"
            exit 1
            ;;
    esac
done

if $USE_GIT_ROOT; then
    echo "=== Consolidating sessions by Git repository root ==="
else
    echo "=== Consolidating sessions by CWD and ordering by creation time ==="
fi

# Function to find git root for a directory
find_git_root() {
    local dir="$1"
    if [[ ! -d "$dir" ]]; then
        echo "$dir"
        return
    fi
    local git_root
    git_root=$(git -C "$dir" rev-parse --show-toplevel 2>/dev/null) || git_root="$dir"
    echo "$git_root"
}

# Get all sessions as JSON
it2 session list --format=json > /tmp/sessions_raw.json

if $USE_GIT_ROOT; then
    # Build a mapping of session ID -> git root
    echo "{}" > /tmp/git_roots.json
    jq -r '.[].SessionID' /tmp/sessions_raw.json | while read -r sid; do
        cwd=$(jq -r --arg sid "$sid" '.[] | select(.SessionID == $sid) | .WorkingDirectory' /tmp/sessions_raw.json)
        git_root=$(find_git_root "$cwd")
        # Update the JSON with git root
        jq --arg sid "$sid" --arg root "$git_root" '. + {($sid): $root}' /tmp/git_roots.json > /tmp/git_roots.json.tmp
        mv /tmp/git_roots.json.tmp /tmp/git_roots.json
    done

    # Add GitRoot field to each session and group by it
    jq --slurpfile roots /tmp/git_roots.json '
      map(. + {GitRoot: $roots[0][.SessionID]}) |
      group_by(.GitRoot) |
      map(
        select(length > 1) |
        {
          cwd: .[0].GitRoot,
          sessions: (. | sort_by(.ShellPID) | map({id: .SessionID, short: .ShortID, pid: .ShellPID, original_cwd: .WorkingDirectory}))
        }
      )
    ' /tmp/sessions_raw.json > /tmp/session_groups.json
else
    # Get all sessions grouped by CWD, sorted by PID within each group
    jq '
      group_by(.WorkingDirectory) |
      map(
        select(length > 1) |
        {
          cwd: .[0].WorkingDirectory,
          sessions: (. | sort_by(.ShellPID) | map({id: .SessionID, short: .ShortID, pid: .ShellPID}))
        }
      )
    ' /tmp/sessions_raw.json > /tmp/session_groups.json
fi

if $USE_GIT_ROOT; then
    echo "Session groups by Git root:"
else
    echo "Session groups by CWD:"
fi
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

    # Move all other sessions to vertical splits with anchor (maintaining time order)
    # Using vertical splits to maximize usable width and avoid crowding
    # Note: iTerm2's session move doesn't support custom split fractions,
    # it uses equal splits. For better control, you may need to manually adjust
    # pane sizes after running this script.
    for ((i=1; i<${#sessions[@]}; i++)); do
        sid="${sessions[$i]}"
        echo "Moving session $i: $sid to vertical split with $anchor"

        it2 session move "$sid" "$anchor" --vertical --quiet || {
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
