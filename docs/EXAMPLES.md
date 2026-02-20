# it2 CLI Examples

This document provides practical examples of using the it2 CLI for real-world automation and terminal management tasks.

## Basic Session Management

### List and inspect sessions
```bash
# List all sessions with details
it2 session list

# List sessions in JSON format for scripting
it2 session list --format json

# Get detailed info about current session
it2 session get-info

# Get detailed info about specific session
it2 session get-info "C3D91F33-3805-47E2-A3F6-B8AED6EC2209"

# Show current session ID
it2 session current
```

### Send commands to sessions
```bash
# Send text to current session
it2 session send-text "echo 'Hello from it2!'"

# Send text to specific session (supports partial IDs)
it2 session send-text 85F9 "ls -la"

# Send text with full UUID (still works)
it2 session send-text "85F9DD0D-1A75-4793-861D-8F5C62166AB7" "ls -la"

# Send complex commands with quotes
it2 session send-text 'find . -name "*.go" | head -5'

# Send with template wrapping for structured messages
it2 session send-text --template '<msg from="{{.ShortID}}">{{.Content}}</msg>' "status update"
```

## Tab and Window Management

### Create and manage tabs
```bash
# Create a new tab with Default profile
it2 tab create "Default"

# Create tab with a specific profile
it2 tab create "Development"

# Create tab with badge
it2 tab create "Default" --badge "API Server"

# Create tab with initial command
it2 tab create "Default" --command "cd ~/projects && npm run dev"

# List all tabs
it2 tab list

# Close specific tab
it2 tab close "tab-id-here"
```

### Window operations
```bash
# List all windows
it2 window list

# Create new window
it2 window create "Default"

# Focus specific window
it2 window focus "window-id-here"
```

## Split Pane Operations

### Create split layouts
```bash
# Split current pane vertically (side by side)
it2 session split --vertical

# Split current pane horizontally (top/bottom)
it2 session split --horizontal

# Split with specific profile
it2 session split --vertical --profile "Development"

# Show split layout structure
it2 tab splits
```

### Development environment setup
```bash
#!/bin/bash
# Create a development environment with multiple panes

echo "Setting up development environment..."

# Create main development tab
TAB_ID=$(it2 tab create "Development" --badge "Dev" --format json | jq -r '.id')

# Split for editor (left) and terminal (right)
it2 session split --vertical --profile "Development"

# Get the new session ID (right pane)
RIGHT_PANE=$(it2 session list --format json | jq -r '.[-1].id')

# Start editor in left pane (original session)
it2 session send-text "cd ~/project && code ."

# Start dev server in right pane
it2 session send-text "$RIGHT_PANE" "cd ~/project && npm run dev"

# Split right pane horizontally for logs
it2 session send-text "$RIGHT_PANE" "C-c"  # Stop if running
SPLIT_CMD="it2 session split --horizontal --session $RIGHT_PANE"
eval "$SPLIT_CMD"

# Get the bottom pane ID
BOTTOM_PANE=$(it2 session list --format json | jq -r '.[-1].id')

# Show logs in bottom pane
it2 session send-text "$BOTTOM_PANE" "cd ~/project && tail -f logs/app.log"

echo "Development environment ready!"
```

## Text Buffer Operations

### Capture and search content
```bash
# Get current screen contents
it2 text get-screen

# Get buffer contents (scrollback)
it2 text get-buffer --lines 100

# Get specific line range
it2 text get-contents --first-line 10 --num-lines 20

# Search for patterns in buffer
it2 text search "error" --session-id "session-id"

# Find text with regex
it2 text find "ERROR.*failed" --regex
```

### Cursor and selection
```bash
# Get cursor position
it2 text get-cursor

# Set cursor position
it2 text set-cursor --x 10 --y 5

# Select text range
it2 text select --start-x 0 --start-y 10 --end-x 20 --end-y 10

# Get current selection
it2 selection get
```

## Profile Management

### List and inspect profiles
```bash
# List all profiles
it2 profile list

# Get profile details
it2 profile get "Default"

# Get profile in YAML format
it2 profile get "Development" --format yaml

# Get specific profile property
it2 profile get-property "Default" "background_color"
```

### Modify profiles
```bash
# Set profile property
it2 profile set-property "Development" "background_color" '{"red": 0.1, "green": 0.1, "blue": 0.1}'

# Apply profile to current session
it2 session set-property "profile" "Development"

# Get current session's profile
it2 session get-property "profile"
```

## Shell Integration Features

### Command history and prompts
```bash
# List command history with exit codes
it2 prompt list

# List last 10 commands
it2 prompt list --max-results 10

# Search command history
it2 prompt search "git commit"

# Get current prompt info
it2 prompt get
```

### Job monitoring and shell state
```bash
# List running jobs in session
it2 job list

# Monitor job changes
it2 job monitor

# Check if shell is ready for commands
it2 shell state

# Wait for shell to be ready before automation
it2 shell wait-for-prompt && it2 session send-text "make deploy"
```

## Real-time Monitoring

### Subscribe to events
```bash
# Monitor all notifications
it2 notification monitor

# Monitor specific event types
it2 notification monitor --type session
it2 notification monitor --type keystroke
it2 notification monitor --type variable

# List available notification types
it2 notification list-types
```

### Variable monitoring
```bash
# Monitor variable changes
it2 variable monitor session user.current_directory

# Get variable value
it2 variable get user.current_directory

# Set variable
it2 variable set user.custom_status "working"
```

## Scripting and Automation

### Batch operations
```bash
#!/bin/bash
# Connect to multiple servers in separate tabs

servers=("web01" "web02" "db01" "cache01")

for server in "${servers[@]}"; do
    echo "Connecting to $server..."

    # Create tab for this server
    TAB_ID=$(it2 tab create "SSH" --badge "$server" --format json | jq -r '.id')

    # Connect via SSH
    it2 session send-text "ssh $server"

    # Wait a moment between connections
    sleep 1
done

echo "Connected to ${#servers[@]} servers"
```

### Environment restoration
```bash
#!/bin/bash
# Save and restore development environment

# Save current arrangement
TIMESTAMP=$(date +%Y%m%d_%H%M%S)
it2 arrangement save "dev_env_$TIMESTAMP"

# List saved arrangements
it2 arrangement list

# Restore specific arrangement
# it2 arrangement restore "dev_env_20240101_120000"
```

### Automated testing setup
```bash
#!/bin/bash
# Set up automated testing environment

echo "Setting up test environment..."

# Create test tab
TEST_TAB=$(it2 tab create "Testing" --badge "Tests" --format json | jq -r '.id')

# Split into multiple panes for different test types
it2 session split --horizontal
UNIT_TESTS=$(it2 session list --format json | jq -r '.[-1].id')

it2 session split --vertical --session "$UNIT_TESTS"
E2E_TESTS=$(it2 session list --format json | jq -r '.[-1].id')

# Run different test suites in each pane
it2 session send-text "npm run test:watch"
it2 session send-text "$UNIT_TESTS" "npm run test:unit:watch"
it2 session send-text "$E2E_TESTS" "npm run test:e2e:watch"

echo "Test environment ready with 3 test runners"
```

## Advanced Features

### Broadcast domains
```bash
# Set up broadcast to multiple sessions
it2 broadcast set "dev-servers" session1 session2 session3

# Send command to all sessions in broadcast group
it2 broadcast send "dev-servers" "sudo systemctl restart nginx"

# List broadcast domains
it2 broadcast list

# Clear broadcast domain
it2 broadcast clear "dev-servers"
```

### Color and appearance
```bash
# List available color presets
it2 color preset list

# Apply color preset to profile
it2 color preset apply "Solarized Dark" "Development"

# Export color preset
it2 color preset export "My Theme" ~/my-theme.itermcolors
```

### tmux integration
```bash
# List tmux connections
it2 tmux list-connections

# List tmux windows
it2 tmux list-windows

# Send command to tmux session
it2 tmux send "connection-id" "new-session -d -s work"
```

## Troubleshooting and Debugging

### Connection and authentication
```bash
# Check authentication status
it2 auth check

# Request new authentication
it2 auth request

# Debug connection with verbose output
ITERM2_DEBUG=1 it2 session list

# Test connection
it2 app version
```

### Session debugging
```bash
# Get comprehensive session info
it2 session get-info --properties --prompt

# Extract specific properties for scripts
FRAME=$(it2 session get-info --extract "frame.coords")
echo "Session frame: $FRAME"

# Check session properties
it2 session get-property grid_size
it2 session get-property buried
```

## Integration Examples

### Git workflow automation
```bash
#!/bin/bash
# Automated git workflow with multiple panes

# Create git workflow tab
it2 tab create "Git" --badge "Git Flow"

# Split for git status (left) and commands (right)
it2 session split --vertical

STATUS_PANE=$(it2 session current)
it2 session send-text "watch -n 2 git status --short"

# Get command pane ID
CMD_PANE=$(it2 session list --format json | jq -r '.[-1].id')

# Split command pane for logs
it2 session split --horizontal --session "$CMD_PANE"
LOG_PANE=$(it2 session list --format json | jq -r '.[-1].id')

# Show git log in bottom pane
it2 session send-text "$LOG_PANE" "git log --oneline --graph -10"

echo "Git workflow setup complete"
```

### System monitoring dashboard
```bash
#!/bin/bash
# Create system monitoring dashboard

it2 tab create "Monitor" --badge "System"

# CPU usage (top-left)
it2 session send-text "htop"

# Split vertically for memory/disk
it2 session split --vertical
MEM_PANE=$(it2 session list --format json | jq -r '.[-1].id')
it2 session send-text "$MEM_PANE" "watch -n 5 'free -h && echo && df -h'"

# Split horizontally for network
it2 session split --horizontal --session "$MEM_PANE"
NET_PANE=$(it2 session list --format json | jq -r '.[-1].id')
it2 session send-text "$NET_PANE" "sudo nethogs"

echo "System monitoring dashboard ready"
```

## Tips and Best Practices

1. **Use JSON output for scripting**: Always use `--format json` when processing output in scripts
2. **Save session IDs**: Store session IDs in variables for complex automation
3. **Use badges**: Tag tabs with badges to identify their purpose
4. **Leverage arrangements**: Save common layouts for quick restoration
5. **Monitor with care**: Be selective about notifications to avoid noise
6. **Handle errors**: Check exit codes in automation scripts
7. **Use Shell Integration**: Install iTerm2 Shell Integration for advanced features

## Environment Variables

Useful environment variables for scripting:

```bash
# Current session (automatically set by iTerm2)
echo $ITERM_SESSION_ID

# Enable debug output
export ITERM2_DEBUG=1

# Use custom WebSocket URL
export ITERM2_URL="ws://localhost:1912"
```

## Common Patterns

### Session cleanup
```bash
# Find and close idle sessions (using quiet mode)
it2 session list -q | \
  while read session_id; do
    # Use partial ID matching for convenience
    it2 session close "$session_id"
  done

# Or with JSON filtering
it2 session list --format json | \
  jq -r '.[] | select(.name | contains("idle")) | .id' | \
  while read session_id; do
    it2 session close "$session_id"
  done
```

### Batch profile updates
```bash
# Update all profiles with a property
it2 profile list --format json | \
  jq -r '.[]' | \
  while read profile; do
    it2 profile set-property "$profile" "cursor_blink" true
  done
```

This examples file demonstrates the power and flexibility of the it2 CLI for terminal automation and management tasks.