# it2-core Plugin

Core iTerm2 terminal automation for the it2 CLI with essential slash commands.

## Overview

The it2-core plugin provides the fundamental iTerm2 automation agent and basic slash commands for common terminal workflows.

## Features

### Agents

- **iterm2-terminal-automation** (v2.0.0) - Comprehensive iTerm2 automation covering 150+ it2 commands across 15 categories including session management, monitoring, text operations, and more.

### Slash Commands

- **/it2-split** - Create an iTerm2 vertical split, navigate to the current directory, and launch Claude Code

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

**/it2-split** - Quickly create a new Claude session:

```bash
/it2-split
```

This will:
1. Create a vertical split in iTerm2
2. Navigate to your current working directory
3. Launch Claude Code
4. Set a badge showing the session ID

## Requirements

- iTerm2 with websocket server enabled
- it2 CLI installed: `go install github.com/tmc/it2@latest`
- Claude Code

## License

MIT
