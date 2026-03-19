# it2-core Plugin

Core iTerm2 terminal automation for the it2 CLI with essential slash commands.

## Overview

The it2-core plugin provides the fundamental iTerm2 automation agent and basic slash commands for common terminal workflows.

## Features

### Agents

- **iterm2-terminal-automation** - iTerm2 automation specialist covering session management, monitoring, text operations, configuration, and advanced features using the it2 CLI.

### Slash Commands

- **/it2-split** - Create an iTerm2 split session and navigate to the current directory

## Installation

```bash
# Add the plugin from GitHub
/plugin add tmc/it2 --source integrations/claude-code/plugins/it2-core

# Or from local path
/plugin add /path/to/it2/integrations/claude-code/plugins/it2-core
```

## Usage

### Using the Agent

The iterm2-terminal-automation agent is automatically available for all iTerm2 automation tasks:

```bash
> Create three sessions for backend, frontend, and testing with badges
> Monitor session ABC123 for screen updates
> Set up a broadcast domain for all development sessions
```

### Using Slash Commands

**/it2-split** - Create a new split session:

```bash
/it2-split
```

This will:
1. Create a split in iTerm2 (direction auto-detected based on space)
2. Navigate the new session to your current working directory

## Requirements

- iTerm2 with websocket server enabled
- it2 CLI installed: `go install github.com/tmc/it2@latest`
- Claude Code

## License

MIT
