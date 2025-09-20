# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

This is `it2`, a Go command-line interface for interacting with iTerm2's API via WebSocket connection using protocol buffers.

## Common Development Commands

### Build and Install
```bash
# Install the binary to $GOPATH/bin
go install

# Build using Makefile
make build

# Run tests
make test
go test -v ./...

# Run linter
make lint

# Format code
go fmt ./...
make fmt
```

### Protocol Buffer Generation
```bash
# Generate protobuf code (uses go generate)
make proto
go generate ./...

# Manual generation (if needed)
protoc --go_out=. ./proto/api.proto
```

### Running the CLI
```bash
# Basic usage - lists sessions (auto-requests auth from iTerm2)
it2 session list

# With debug output
ITERM2_DEBUG=1 it2 session list

# Send text to a session
it2 session send-text <session-id> "echo Hello"

# Create a new tab
it2 tab create "Default" <window-id>
```

## Architecture

### Authentication
- The client automatically requests authentication from iTerm2 via AppleScript when connecting
- Located in `internal/client/client.go` - `requestAuth()` method
- Falls back from Unix socket (`~/Library/Application Support/iTerm2/private/socket`) to TCP if needed
- Authentication credentials are obtained dynamically on each connection

### Command Structure
- Commands are organized in `cmd/` directory with subpackages for each command group
- Each command group has a `NewCommand()` function that returns a `cobra.Command`
- Main entry point is `main.go` which assembles all commands

### Client Communication
- WebSocket client in `internal/client/`
- Protocol buffer definitions in `proto/api.proto`
- Sessions functionality split into `internal/client/sessions.go`
- Formatting utilities in `internal/formatting/`

### Key Components
- **Client** (`internal/client/client.go`): Handles WebSocket connection, authentication, and message routing
- **Commands** (`cmd/*/`): Cobra-based CLI commands organized by iTerm2 feature area
- **Proto** (`proto/`): Protocol buffer definitions and generated Go code
- **Auth** (`internal/auth/`): Authentication utilities for requesting iTerm2 credentials

## Implementation Status

Many commands have TODO placeholders. Implemented features include:
- Session: list, close, activate, restart, split, send-text
- Tab: list, create, close, activate, move
- Window: list, create, close, activate
- App: focus
- Auth: request, check

## Testing Individual Features

```bash
# Test a single package
go test -v ./internal/client

# Run with specific test
go test -v -run TestListSessions ./internal/client
```

## Debug Mode

Set `ITERM2_DEBUG=1` environment variable to see detailed connection and message logs.