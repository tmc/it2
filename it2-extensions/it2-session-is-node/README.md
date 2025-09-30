# it2-session-is-node

A plugin for it2 that checks if a session's process is running Node.js.

## Usage

```bash
# Basic usage
./it2-session-is-node <session-id>

# JSON output
./it2-session-is-node --json <session-id>

# With custom timeout
./it2-session-is-node --timeout 10s <session-id>
```

## Build

```bash
go build
```

## Description

This plugin:
1. Gets the PID of the specified session by running `echo $$` in the session
2. Checks if that PID is running a Node.js process by examining the process name and command line
3. Returns exit code 0 if it's a Node.js process, 1 otherwise

The plugin detects various Node.js-related processes including:
- node
- nodejs
- npm
- npx
- yarn
- pnpm
- bun
- deno

## Use Cases

- Diagnosing stuck Claude Code sessions
- Identifying which sessions are running Node.js applications
- Automation scripts that need to operate on Node.js sessions only