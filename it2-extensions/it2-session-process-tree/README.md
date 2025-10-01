# it2-session-process-tree

Example process enricher plugin that visualizes the process tree for a session's process.

## Plugin Type

This is a **process enricher** plugin (`it2-session-process-*`).

## How It Works

Process enricher plugins are called with two arguments:
1. **Session ID**: The iTerm2 session identifier
2. **PID**: The process ID to inspect

The plugin outputs process tree information that can be used by `it2 session process-list` and other process inspection commands.

## Example Usage

```bash
# Call the plugin directly
./it2-session-process-tree.sh "ABC123" "12345"
# Output: node → /usr/bin/login → iTerm2

# Plugins are automatically discovered when in PATH
# They can enrich process-list output:
it2 session process-list --format=json
```

## Creating Your Own Process Plugin

To create a custom process enricher plugin:

1. Name it `it2-session-process-<name>` (or `it2-session-process-<name>.sh`)
2. Make it executable (`chmod +x`)
3. Place it in your PATH
4. Accept two arguments: `<session-id> <pid>`
5. Output your process information to stdout

Example template:

```bash
#!/bin/bash
# it2-session-process-custom

SESSION_ID="$1"
PID="$2"

# Your custom process inspection logic here
if [ -n "$PID" ]; then
    # Output your custom process data
    echo "Custom process info for PID $PID"
fi
```

## Plugin Discovery

The it2 plugin system automatically discovers executables matching these patterns:
- `it2-session-process-*` - Process enricher plugins
- `it2-session-*` - Session enricher plugins
- `it2-tab-*` - Tab enricher plugins
- `it2-window-*` - Window enricher plugins

Process plugins are invoked for PIDs discovered in sessions and can add custom metadata, visualizations, or analysis.
