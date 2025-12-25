# Claude Code Session Linking

Links Claude Code session IDs to iTerm2 sessions, enabling commands like:

```bash
# Get Claude session ID from an iTerm2 session
it2 session claude DE94CE8E

# Resume a Claude session running in another terminal
claude --resume $(it2 session claude DE94CE8E)
```

## Quick Setup

1. **Install the hook script** (already included in it2):

```bash
# Copy to a location in PATH
cp scripts/it2-claude-session-link ~/bin/
chmod +x ~/bin/it2-claude-session-link
```

2. **Add to Claude Code settings** (`~/.claude/settings.json`):

```json
{
  "hooks": {
    "SessionStart": [{
      "type": "command",
      "command": "it2-claude-session-link"
    }]
  }
}
```

3. **Rebuild it2** (for the `it2 session claude` command):

```bash
go install github.com/tmc/it2/cmd/it2@latest
```

## Usage

```bash
# Get Claude session ID for current terminal
it2 session claude

# Get Claude session ID for a specific iTerm2 session
it2 session claude DE94CE8E

# JSON output with full info
it2 session claude DE94CE8E --format json

# Use with claude --resume
claude --resume $(it2 session claude DE94CE8E)
```

## How It Works

1. Claude Code fires a `SessionStart` hook when a session begins
2. The hook receives JSON with `session_id` (Claude's UUID)
3. `it2-claude-session-link` extracts the ID and sets iTerm2 user variables:
   - `user.claude_session_id` - The Claude session UUID
   - `user.claude_transcript` - Path to session transcript file
   - `user.claude_last_event` - Most recent hook event
4. `it2 session claude` queries these variables

## Variables Set

| Variable | Description |
|----------|-------------|
| `user.claude_session_id` | Claude Code session UUID (e.g., `fa520651-161a-416c-a6e7-b06b5894e10a`) |
| `user.claude_transcript` | Full path to `.jsonl` transcript file |
| `user.claude_last_event` | Last hook event type (SessionStart, UserPromptSubmit, etc.) |

## Direct Variable Access

You can also query variables directly:

```bash
it2 session get-variable $SID user.claude_session_id
it2 session list-variables $SID | grep claude
```

## Troubleshooting

**No output from `it2 session claude`:**
- Ensure the hook is configured in `~/.claude/settings.json`
- Check that `it2-claude-session-link` is in your PATH
- Verify Claude Code started after hook was configured

**Test the hook manually:**
```bash
echo '{"session_id":"test-123","hook_event_name":"SessionStart"}' | it2-claude-session-link
it2 session get-variable $ITERM_SESSION_ID user.claude_session_id
```
