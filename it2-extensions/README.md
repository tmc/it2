# it2 Extensions

This directory contains extension plugins for it2 that help diagnose and manage sessions.

## Available Extensions

### it2-session-is-node

Checks if a session's process is running Node.js.

```bash
cd it2-session-is-node
go build
./it2-session-is-node <session-id>
```

### it2-session-has-debugging-enabled

Checks if a Node.js process in a session has debugging/inspector enabled.

```bash
cd it2-session-has-debugging-enabled
go build
./it2-session-has-debugging-enabled <session-id>
```

## Use Cases

These plugins are particularly useful for:

- **Diagnosing stuck Claude Code sessions**: Identify if Claude sessions are running Node.js and if debugging is enabled, which can cause sessions to hang
- **Development workflow automation**: Build scripts that operate only on Node.js sessions
- **Monitoring and debugging**: Quickly identify which sessions can be debugged

## Quick Test

```bash
# Build both plugins
cd it2-session-is-node && go build && cd ..
cd it2-session-has-debugging-enabled && go build && cd ..

# Test with a session ID (replace with actual session ID)
SESSION_ID="A357A339"

echo "Testing Node.js detection:"
./it2-session-is-node/it2-session-is-node --json $SESSION_ID

echo "Testing debugging detection:"
./it2-session-has-debugging-enabled/it2-session-has-debugging-enabled --json $SESSION_ID
```

## Installation

To install system-wide:

```bash
# Build and install to PATH
cd it2-session-is-node && go build && sudo cp it2-session-is-node /usr/local/bin/
cd ../it2-session-has-debugging-enabled && go build && sudo cp it2-session-has-debugging-enabled /usr/local/bin/
```