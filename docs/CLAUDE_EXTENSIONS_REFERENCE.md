# Claude Code Extensions for it2

## Overview

The it2 CLI includes a comprehensive suite of plugins specifically designed for automating and orchestrating Claude Code sessions. These plugins enable detection, monitoring, and intelligent interaction with Claude sessions.

## Extension Categories

### 1. Session Detection & Identification

#### `it2-session-process-is-claude`
**Location**: `internal/embedded/plugins/it2-session-process-is-claude.sh`
**Purpose**: Detect if a session process is running Claude Code

**Usage**:
```bash
it2-session-process-is-claude <session-id> <pid>
```

**Returns**:
- `"true"` if the process is Claude Code
- Empty string otherwise

**Implementation**: Checks process name using `ps` for "claude" substring

---

#### `it2-session-is-claude-code`
**Location**: `it2-extensions/it2-session-is-claude-code/it2-session-is-claude-code.sh`
**Purpose**: Detect Claude Code sessions by checking main process and parent chain

**Usage**:
```bash
it2-session-is-claude-code [session-id]
# Uses current session if no ID provided
```

**Returns**:
- `✅` if Claude session detected
- Exit code 0 on success, 1 otherwise

**Features**:
- Checks session PID and process name
- Traverses parent process chain (up to 3 levels)
- Debug mode available via `ITERM2_DEBUG=1`

**Example**:
```bash
# Check current session
it2-session-is-claude-code

# Check specific session
it2-session-is-claude-code A1B2C3D4

# With debug output
ITERM2_DEBUG=1 it2-session-is-claude-code A1B2C3D4
```

---

### 2. Modal Detection & Interaction

#### `it2-session-claude-has-modal`
**Location**: `internal/plugins/scripts/it2-session-claude-has-modal`
**Size**: 2.4K
**Purpose**: Detect modal dialogs in Claude sessions requiring user intervention

**Usage**:
```bash
it2-session-claude-has-modal <session-id>
```

**Returns**:
- `"approval"` - "Do you want to proceed?" prompts
- `"choice"` - Numbered option selections with ❯ selector
- `"confirmation"` - Y/n, y/N, [y/n] prompts
- `"input"` - Text input requests
- `"none"` - No modal detected

**Detection Patterns**:
1. **Approval prompts**: "Do you want to proceed?"
2. **Choice dialogs**: `❯ 1. Yes` with numbered options
3. **Confirmations**: `(Y/n)`, `(y/N)`, `[y/n]`
4. **Input prompts**: "Enter...", "Please enter", "Input:"
5. **Box dialogs**: `╭─╮` style with options
6. **Edit acceptance**: "⏵⏵ accept edits", "shift+tab to cycle"

**Example**:
```bash
MODAL=$(it2-session-claude-has-modal A1B2C3D4)
if [[ "$MODAL" == "approval" ]]; then
    echo "Claude is waiting for approval"
fi
```

---

#### `it2-session-claude-auto-approve`
**Location**: `internal/plugins/scripts/it2-session-claude-auto-approve`
**Size**: 2.2K
**Purpose**: Automatically approve safe Claude operations

**Usage**:
```bash
it2-session-claude-auto-approve <session-id> [--dry-run]
```

**Returns**:
- `"approved"` - Modal was auto-approved
- `"skipped"` - Modal unsafe or requires human review
- `"no-modal"` - No modal detected
- `"error"` - Unable to process
- `"approved-dry-run"` - Would approve (dry-run mode)

**Approval Actions by Modal Type**:
- **approval/choice**: Sends `1` + `Return` (selects first option)
- **confirmation**: Sends `y` (yes)
- **input**: Always skipped (too risky)

**Example**:
```bash
# Auto-approve if safe
it2-session-claude-auto-approve A1B2C3D4

# Dry run to see what would happen
it2-session-claude-auto-approve A1B2C3D4 --dry-run
```

---

#### `it2-session-claude-is-safe-operation`
**Location**: `internal/plugins/scripts/it2-session-claude-is-safe-operation`
**Size**: 4.2K
**Purpose**: Determine if a Claude operation is safe to auto-approve

**Usage**:
```bash
it2-session-claude-is-safe-operation <session-id>
```

**Returns**:
- `"safe"` - Operation can be safely auto-approved
- `"unsafe"` - Operation requires human review
- `"unknown"` - Operation cannot be determined

**Safe Operations**:
- **File reading**: `cat`, `head`, `tail`, `ls`, `find`, `grep`, `less`, `more`
- **Git read-only**: `git status`, `git log`, `git diff`, `git show`, `git ls-files`
- **Process inspection**: `ps`, `top`, `htop`, `lsof`, `netstat`, `pgrep`
- **System info**: `uname`, `whoami`, `pwd`, `date`, `uptime`, `df`, `du`, `env`
- **Package info**: `npm ls`, `pip list`, `brew list`, `go list`, `cargo tree`
- **Documentation**: `--help`, `-h`, `man`, `info`
- **Clipboard**: `pbcopy`, `copy`
- **it2 commands**: `get-buffer`, `get-screen`, `get-contents`

**Unsafe Operations**:
- **File modifications**: `rm`, `mv`, `cp`, `mkdir`, `chmod`, `chown`, `>`, `>>`
- **Git destructive**: `git add`, `git commit`, `git push`, `git merge`, `git rebase`, `git reset`
- **Network**: `curl`, `wget`, `ssh`, `scp`, `rsync`, `ftp`
- **Package installs**: `npm install`, `pip install`, `brew install`, `apt`, `yum`
- **System mods**: `sudo`, `su`, `kill`, `systemctl`, `mount`
- **Process control**: `&`, `nohup`, `screen`, `tmux`, `bg`, `fg`

**Context-Based Rules**:
- Operations in `/tmp`, `/var/tmp`, `*test*`, `*demo*`, `*example*` are more permissive
- Commands with `--dry-run`, `--simulate`, `--preview`, `--check` flags are safe

**Example**:
```bash
SAFETY=$(it2-session-claude-is-safe-operation A1B2C3D4)
if [[ "$SAFETY" == "safe" ]]; then
    it2-session-claude-auto-approve A1B2C3D4
else
    echo "Human review required"
fi
```

---

### 3. Session State Detection

#### `it2-session-is-at-empty-claude-prompt`
**Location**: `internal/plugins/scripts/it2-session-is-at-empty-claude-prompt`
**Size**: 3.7K
**Purpose**: Check if session is at an empty Claude prompt ready for input

**Usage**:
```bash
it2-session-is-at-empty-claude-prompt <session-id>
```

**Returns**:
- `"true"` - At empty Claude prompt
- `"false"` - Not at Claude prompt, has partial input, or busy

**Detection Logic**:
1. Checks for Claude prompt indicators (①②③④ badges, "claude" branding)
2. Verifies no modal dialogs present
3. Checks for empty input line (no partial commands)
4. Ensures Claude is not busy (no "thinking", "generating", etc.)

**Exclusions**:
- Modal dialogs present → `false`
- Choice dialogs (❯ selector) → `false`
- Thinking/generating indicators → `false`
- Partial command input → `false`

**Example**:
```bash
if [[ "$(it2-session-is-at-empty-claude-prompt A1B2C3D4)" == "true" ]]; then
    it2 session send-text A1B2C3D4 "continue"
fi
```

---

#### `it2-session-has-no-queued-claude-messages`
**Location**: `internal/plugins/scripts/it2-session-has-no-queued-claude-messages`
**Size**: 3.9K
**Purpose**: Check if there are no queued/pending messages in Claude session

**Usage**:
```bash
it2-session-has-no-queued-claude-messages <session-id>
```

**Returns**:
- `"true"` - No queued messages
- `"false"` - Messages queued or Claude is busy

**Detected Busy Indicators**:
1. **Processing**: "thinking...", "generating", "processing"
2. **Spinners**: `⠋⠙⠹⠸⠼⠴⠦⠧⠇⠏` characters
3. **Status messages**: "working on", "please wait", "one moment"
4. **Tool use**: `<tool_use>`, "Using tool", "Tool call", "Executing"
5. **Streaming**: `▌▍▎▏` characters
6. **Queue indicators**: "X pending", "X queued", "messages in queue"
7. **Waiting states**: "waiting for"
8. **Generation**: "Generating response", "Composing", "Drafting"
9. **Ellipsis**: Lines ending with `...`
10. **Approval dialogs**: "Do you want to proceed?"
11. **Pagination**: "continue", "show more", "load more" (when actively loading)

**Example**:
```bash
while [[ "$(it2-session-has-no-queued-claude-messages A1B2C3D4)" == "false" ]]; do
    echo "Waiting for Claude to finish..."
    sleep 1
done
echo "Claude is ready"
```

---

### 4. Intelligent Orchestration

#### `it2-session-claude-suggest-action`
**Location**: `internal/plugins/scripts/it2-session-claude-suggest-action`
**Size**: 5.2K
**Purpose**: Analyze Claude session context and suggest appropriate interventions

**Usage**:
```bash
it2-session-claude-suggest-action <session-id> [--execute]
```

**Returns**:
- `"continue"` - Session has incomplete todos, should send "continue"
- `"continue:milestone-breakdown"` - Vague todos need breakdown
- `"continue:test-build"` - Test/build todos ready to run
- `"tab"` - Pending edits to accept
- `"modal:approve"` - Safe modal ready for approval
- `"modal:approved"` - Modal was auto-approved (--execute mode)
- `"modal:review"` - Unsafe modal needs human review
- `"wait"` - Session actively working
- `"review:error"` - Error state needs attention
- `"none"` - No intervention needed

**Priority Decision Tree**:
1. **Modals** (highest priority) - Check for approval dialogs
2. **Edit acceptance** - Check for "⏵⏵ accept edits" prompts
3. **Active work** - Check for `⏺✳✽✶✢⚡🔄🔍` indicators or "Generating/Processing" text
4. **Incomplete todos** - Check for `☐` with specific patterns:
   - Vague monitoring → suggest milestone breakdown
   - Test/build todos → suggest running them
   - Generic → send "continue"
5. **Error states** - Check for "Error", "Failed", "Exception"
6. **Idle without todos** - No action needed

**Example**:
```bash
# Get suggestion
ACTION=$(it2-session-claude-suggest-action A1B2C3D4)
echo "Suggested action: $ACTION"

# Execute suggestion automatically
it2-session-claude-suggest-action A1B2C3D4 --execute
```

**Automation Loop Example**:
```bash
while true; do
    ACTION=$(it2-session-claude-suggest-action A1B2C3D4 --execute)
    echo "[$ACTION] at $(date)"

    case "$ACTION" in
        "modal:approved"|"continue:"*|"tab:sent")
            echo "Action taken"
            ;;
        "wait")
            sleep 2
            ;;
        "modal:review"|"review:error")
            echo "Human attention needed!"
            break
            ;;
        "none")
            echo "Session complete"
            break
            ;;
    esac
done
```

---

## Plugin Architecture

### Plugin Discovery
Plugins are loaded from multiple locations in priority order:
1. Command-line specified paths (`--plugin-path`)
2. External extensions (`it2-extensions/`)
3. Internal plugins (`internal/plugins/scripts/`)
4. Embedded plugins (`internal/embedded/plugins/`)

### Integration with it2 Commands
Plugins are automatically invoked by `it2 session list` for session property enrichment:

```bash
# The "is-claude-in-progress" column uses:
# - it2-session-process-is-claude (detection)
# - it2-session-has-no-queued-claude-messages (status)

it2 session list
# Shows 🚧 flag for Claude sessions with work in progress
```

### Plugin Development Guidelines

**Required Structure**:
```bash
#!/usr/bin/env bash
set -euo pipefail

SESSION_ID="${1:-}"
if [[ -z "$SESSION_ID" ]]; then
    echo "Error: Session ID required" >&2
    exit 1
fi

# Find it2 binary
IT2_BIN=""
if [[ -x "./it2" ]]; then
    IT2_BIN="./it2"
elif [[ -x "$HOME/go/bin/it2" ]]; then
    IT2_BIN="$HOME/go/bin/it2"
elif command -v it2 &> /dev/null; then
    IT2_BIN="$(command -v it2)"
fi

# Plugin logic here...
```

**Best Practices**:
- Always validate session ID input
- Use `it2 text get-screen` for screen content analysis
- Strip ANSI escape sequences: `sed 's/\x1b\[[0-9;]*m//g'`
- Return clear, parseable output
- Exit code 0 for success, 1 for failure
- Support `ITERM2_DEBUG=1` for debugging

---

## Real-World Usage Examples

### Example 1: Auto-Approve Safe Operations
```bash
#!/bin/bash
# Monitor Claude session and auto-approve safe operations

SESSION_ID="$1"

while true; do
    MODAL=$(it2-session-claude-has-modal "$SESSION_ID")

    if [[ "$MODAL" != "none" ]]; then
        RESULT=$(it2-session-claude-auto-approve "$SESSION_ID")
        echo "[$(date '+%H:%M:%S')] Modal $MODAL: $RESULT"
    fi

    sleep 1
done
```

### Example 2: Wait for Claude to Finish
```bash
#!/bin/bash
# Wait for Claude to complete current work

SESSION_ID="$1"

echo "Waiting for Claude to finish..."
while [[ "$(it2-session-has-no-queued-claude-messages "$SESSION_ID")" == "false" ]]; do
    sleep 2
done
echo "Claude is ready!"
```

### Example 3: Smart Claude Orchestration
```bash
#!/bin/bash
# Fully automated Claude orchestration

SESSION_ID="$1"

while true; do
    # Let suggest-action make the decision and execute it
    RESULT=$(it2-session-claude-suggest-action "$SESSION_ID" --execute)

    echo "[$(date '+%H:%M:%S')] $RESULT"

    case "$RESULT" in
        "modal:review"|"review:error")
            echo "⚠️  Human intervention required"
            exit 1
            ;;
        "none")
            echo "✅ Work complete"
            exit 0
            ;;
        "wait")
            sleep 3
            ;;
        *)
            sleep 1
            ;;
    esac
done
```

### Example 4: Multi-Claude Coordination
```bash
#!/bin/bash
# Coordinate multiple Claude sessions

CLAUDE_A="ABC12345"
CLAUDE_B="DEF67890"

# Send tasks to both Claudes
it2 session send-text "$CLAUDE_A" "Analyze the frontend code"
it2 session send-text "$CLAUDE_B" "Analyze the backend code"

# Monitor both sessions
while true; do
    A_DONE=$(it2-session-has-no-queued-claude-messages "$CLAUDE_A")
    B_DONE=$(it2-session-has-no-queued-claude-messages "$CLAUDE_B")

    # Auto-approve any safe modals
    it2-session-claude-auto-approve "$CLAUDE_A" 2>/dev/null
    it2-session-claude-auto-approve "$CLAUDE_B" 2>/dev/null

    if [[ "$A_DONE" == "true" ]] && [[ "$B_DONE" == "true" ]]; then
        echo "Both Claudes finished!"
        break
    fi

    sleep 2
done
```

---

## Performance Considerations

**Plugin Execution Time**:
- Modal detection: ~50-100ms (screen content analysis)
- Safety analysis: ~100-200ms (regex matching)
- Auto-approve: ~150-300ms (detection + approval action)
- Suggest-action: ~200-400ms (full decision tree)

**Caching Strategies**:
```bash
# Cache screen content for multiple checks
SCREEN_CONTENT=$(it2 text get-screen "$SESSION_ID")
echo "$SCREEN_CONTENT" | grep -q "Do you want to proceed?"
```

**Polling Recommendations**:
- Modal detection: 1-2 second intervals
- Queue status: 2-3 second intervals
- Suggest-action: 3-5 second intervals

---

## Troubleshooting

### Plugin Not Found
```bash
# Check plugin search paths
echo $IT2_PLUGIN_PATHS

# Verify plugin is executable
ls -la internal/plugins/scripts/it2-session-claude-*

# Add custom plugin path
it2 --plugin-path /path/to/plugins session list
```

### Debug Mode
```bash
# Enable debug output
ITERM2_DEBUG=1 it2-session-is-claude-code A1B2C3D4

# Shows:
# [DEBUG] it2-session-is-claude-code: SESSION_ID=A1B2C3D4
# [DEBUG] Got PID from get-pid: 12345
# [DEBUG] Process name for PID 12345: node
# [DEBUG] Parent 1 PID 12344: claude
```

### Common Issues

**"Session not found" errors**:
- Use full session IDs for variable commands
- Verify session exists: `it2 session list | grep <id>`

**Modals not detected**:
- Check screen content directly: `it2 text get-screen <id>`
- Modal patterns may need updating for new Claude UI

**Auto-approve not working**:
- Verify modal is detected: `it2-session-claude-has-modal <id>`
- Check safety level: `it2-session-claude-is-safe-operation <id>`
- Ensure operation is in safe list

---

## Extension Development

To create a new Claude-related plugin:

1. **Create the plugin file**:
```bash
touch internal/plugins/scripts/it2-session-claude-my-feature
chmod +x internal/plugins/scripts/it2-session-claude-my-feature
```

2. **Use the template**:
```bash
#!/usr/bin/env bash
set -euo pipefail

SESSION_ID="${1:-}"
if [[ -z "$SESSION_ID" ]]; then
    echo "Error: Session ID required" >&2
    exit 1
fi

# Find it2 binary
IT2_BIN=""
if [[ -x "./it2" ]]; then
    IT2_BIN="./it2"
elif command -v it2 &> /dev/null; then
    IT2_BIN="$(command -v it2)"
fi

# Your plugin logic here
SCREEN=$($IT2_BIN text get-screen "$SESSION_ID" 2>/dev/null || echo "")

# Return result
echo "result"
```

3. **Test the plugin**:
```bash
./internal/plugins/scripts/it2-session-claude-my-feature A1B2C3D4
```

4. **Integrate with it2** (if needed):
- Add to property calculation in `internal/cmd/session/list.go`
- Update documentation

---

## Summary

The Claude Code extension ecosystem for it2 provides:

- ✅ **Detection**: Identify Claude sessions by process
- ✅ **Modal handling**: Detect and auto-approve safe operations
- ✅ **State monitoring**: Track session readiness and queue status
- ✅ **Smart orchestration**: Intelligent intervention suggestions
- ✅ **Safety analysis**: Context-aware operation classification
- ✅ **Extensibility**: Clear plugin architecture for custom automation

These plugins enable sophisticated Claude Code automation while maintaining safety through multi-layered checks and context-aware decision making.
