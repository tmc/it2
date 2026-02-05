# Shell Integration

iTerm2 shell integration provides rich metadata about commands and prompts.

## Overview

Shell integration enables:
- Command history with timing and exit codes
- Prompt detection for precise command boundaries
- Working directory tracking
- Mark navigation

## Checking Status

```bash
# Check if session has shell integration
it2 session has-shell-integration "$SID"

# Get shell integration details
it2 session get-info "$SID" --format json | jq '.shell_integration'
```

## Enabling Shell Integration

```bash
# Enable for a profile
it2 profile shell-integration enable "Default"

# Check status
it2 profile shell-integration status "Default"
```

Manual installation per shell:

**bash:**
```bash
curl -L https://iterm2.com/shell_integration/bash -o ~/.iterm2_shell_integration.bash
echo 'source ~/.iterm2_shell_integration.bash' >> ~/.bashrc
```

**zsh:**
```bash
curl -L https://iterm2.com/shell_integration/zsh -o ~/.iterm2_shell_integration.zsh
echo 'source ~/.iterm2_shell_integration.zsh' >> ~/.zshrc
```

## Command History

### Listing Recent Commands

```bash
# Get prompts/commands
it2 prompt list

# Search history
it2 prompt search "git commit"
```

### Prompt Metadata

```bash
# Get prompt details for session
it2 session prompt "$SID" --format json
```

Returns:
```json
{
  "command": "npm test",
  "working_directory": "/home/user/project",
  "start_time": "2024-01-15T10:30:00Z",
  "end_time": "2024-01-15T10:30:15Z",
  "exit_status": 0,
  "duration_ms": 15000
}
```

## Working Directory Tracking

With shell integration, iTerm2 tracks the current directory:

```bash
# Get current working directory
it2 session get-info "$SID" --format json | jq -r '.working_directory'
```

## Event Subscriptions

Shell integration enables prompt events:

```bash
# Subscribe to prompt events
it2 subscribe prompt --session "$SID" --format json | while read event; do
  CMD=$(echo "$event" | jq -r '.data.command')
  EXIT=$(echo "$event" | jq -r '.data.exit_status')
  echo "Command: $CMD (exit: $EXIT)"
done
```

## Use Cases

### Command Completion Detection

```bash
# Wait for specific command to complete
wait_for_command() {
  local SID="$1"
  local CMD="$2"

  it2 subscribe prompt --session "$SID" --format json | while read event; do
    if echo "$event" | jq -r '.data.command' | grep -q "$CMD"; then
      echo "$event" | jq -r '.data.exit_status'
      break
    fi
  done
}
```

### Error Detection

```bash
# Alert on non-zero exit codes
it2 subscribe prompt --format json | while read event; do
  EXIT=$(echo "$event" | jq -r '.data.exit_status')
  if [ "$EXIT" != "0" ]; then
    CMD=$(echo "$event" | jq -r '.data.command')
    it2 notification "Command failed: $CMD (exit $EXIT)"
  fi
done
```

## Limitations

- Only works with supported shells (bash, zsh, fish)
- Must be installed per shell configuration
- Some commands may not trigger prompt events (e.g., subshells)
