# it2-session-has-debugging-enabled

A plugin for it2 that checks if a Node.js process in a session has debugging/inspector enabled.

## Usage

```bash
# Basic usage
./it2-session-has-debugging-enabled <session-id>

# JSON output
./it2-session-has-debugging-enabled --json <session-id>

# With custom timeout
./it2-session-has-debugging-enabled --timeout 10s <session-id>
```

## Build

```bash
go build
```

## Description

This plugin:
1. Gets the PID of the specified session by running `echo $$` in the session
2. Checks if that PID is running a Node.js process with debugging enabled
3. Returns exit code 0 if debugging is enabled, 1 otherwise

The plugin detects various debugging configurations:
- `--inspect` flag
- `--inspect-brk` flag
- `--debug` flag (legacy)
- `--debug-brk` flag (legacy)
- `--debug-port` flag
- `--inspect-port` flag
- Processes listening on common debug ports (9229, 5858, 9222)

## Output

When using `--json`, the output includes:
- `session_id`: The resolved session ID
- `pid`: The process ID
- `has_debugging`: Boolean indicating if debugging is enabled
- `inspector_port`: The inspector port if detected (0 if not found)
- `debug_flags`: Array of detected debug flags

## Use Cases

- Diagnosing stuck Claude Code sessions that might have debugging enabled
- Identifying Node.js processes that can be debugged
- Automation scripts for development environments
- Monitoring Node.js applications in production