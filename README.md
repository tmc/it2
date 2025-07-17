# iTerm2 Go CLI

A command-line interface for interacting with iTerm2's API, written in Go.

## Features

- List all iTerm2 sessions, tabs, and windows
- Send text to iTerm2 sessions
- Create new tabs
- Get session buffer content
- More features coming soon!

## Installation

```bash
go install github.com/gnachman/iTerm2/iterm2-go-cli@latest
```

Or clone the repository and build:

```bash
git clone https://github.com/gnachman/iTerm2/
cd iTerm2/iterm2-go-cli
go build
```

## Usage

Make sure iTerm2 is running with the WebSocket API enabled. You can check this in iTerm2 preferences under "Advanced" tab.

### List all sessions

```bash
iterm2-cli list
```

### Send text to a session

```bash
iterm2-cli send <session-id> "echo Hello, world!"
```

### Create a new tab

```bash
iterm2-cli create-tab "Default" <window-id>
```

### Get session buffer content

```bash
iterm2-cli buffer <session-id> 20  # Get the last 20 lines
```

## Configuration

You can configure the CLI using flags:

```bash
iterm2-cli --url="ws://localhost:1912" --timeout=10s list
```

## How the iTerm2 API Works

The iTerm2 API communicates over a WebSocket connection using protocol buffers. Key concepts:

1. **Sessions**: Individual terminal instances within tabs
2. **Tabs**: Collections of sessions
3. **Windows**: Collections of tabs

Each session, tab, and window has a unique identifier that can be used for API operations.

## Development

To generate the protocol buffer files from the API definition:

```bash
protoc --go_out=. ./proto/api.proto
```

## License

This project is licensed under the same license as iTerm2.