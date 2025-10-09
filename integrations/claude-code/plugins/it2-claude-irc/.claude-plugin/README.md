# it2-claude-irc Plugin

IRC-style commands for managing Claude Code sessions with familiar chat-like commands.

## Overview

This plugin provides IRC-inspired slash commands for managing multiple Claude Code sessions, making it easy to create, switch between, message, and monitor Claude sessions using familiar IRC conventions.

## Features

### Slash Commands

- **/join** [session-name] - Create or switch to a Claude Code session (like IRC `/join #channel`)
- **/msg** <session-id> <message> - Send a message to a Claude Code session (like IRC `/msg user message`)
- **/who** - List all active Claude Code sessions (like IRC `/who`)
- **/part** [session-id] - Close a Claude Code session gracefully (like IRC `/part #channel`)
- **/topic** [session-id] [new-topic] - Set or view session badge/title (like IRC `/topic`)

## Installation

```bash
# Add the plugin from GitHub
/plugin add tmc/it2 --source integrations/claude-code/plugins/it2-claude-irc

# Or from local path
/plugin add /path/to/it2/integrations/claude-code/plugins/it2-claude-irc
```

## Usage Examples

### Join a Session

```bash
# Create a new Claude session
/join

# Create or switch to a named session
/join backend-dev
```

### Message a Session

```bash
# Send a command to a specific session
/msg ABC123 run the tests

# Send to a named session
/msg backend-dev analyze the API code
```

### List Active Sessions

```bash
# See all active Claude sessions
/who
```

Output example:
```
Session ID  Title           Status      Directory
--------    -----           ------      ---------
ABC12345    backend-dev     processing  ~/projects/api
DEF67890    frontend        idle        ~/projects/web
GHI24680    testing         waiting     ~/projects/tests
```

### Leave a Session

```bash
# Close the current session
/part

# Close a specific session
/part ABC123
```

### Set Session Topic

```bash
# View current session's topic
/topic

# Set a new topic for a session
/topic ABC123 Backend Development
```

## IRC Command Mapping

| IRC Command | it2-claude-irc Command | Description |
|-------------|----------------------|-------------|
| /join #channel | /join [name] | Create/switch to session |
| /msg user text | /msg <session> <text> | Send message to session |
| /who | /who | List active sessions |
| /part #channel | /part [session] | Leave/close session |
| /topic #channel [topic] | /topic [session] [topic] | View/set session topic |

## Use Cases

- **Multi-Session Workflows**: Manage multiple Claude sessions for different tasks (backend, frontend, testing, etc.)
- **Session Orchestration**: Send commands to specific sessions without switching context
- **Quick Overview**: See all active Claude sessions and their status at a glance
- **Session Labeling**: Organize sessions with descriptive topics/badges

## Requirements

- iTerm2 with websocket server enabled
- it2 CLI installed: `go install github.com/tmc/it2@latest`
- Claude Code

## Tips

- Use `/join` with meaningful names to organize your work
- Use `/who` regularly to see what Claude sessions are doing
- Set topics with `/topic` to keep track of session purposes
- Use `/msg` to queue commands in background sessions

## License

MIT
