# it2-skills

Claude Code skills for iTerm2 terminal automation using the `it2` CLI tool.

## Installation

```bash
claude plugin add /path/to/it2/plugins/it2-skills
```

Or from the it2 repository:

```bash
claude plugin add https://github.com/tmc/it2 --path plugins/it2-skills
```

## Skills

### session-orchestration

Managing multiple iTerm2 sessions, creating layouts, and orchestrating terminal workflows.

**Use when:**
- Creating multi-pane development layouts
- Managing concurrent terminal sessions
- Navigating session hierarchies
- Sending commands to specific sessions

### claude-code-integration

Working with Claude Code sessions running in iTerm2.

**Use when:**
- Detecting Claude Code sessions
- Monitoring Claude session status
- Coordinating multiple Claude instances
- Detecting permission prompts and modals

### terminal-monitoring

Watching sessions for state changes and implementing automated responses.

**Use when:**
- Monitoring session activity in real-time
- Subscribing to iTerm2 notifications
- Implementing automated responses
- Detecting process completion

## Requirements

- [it2](https://github.com/tmc/it2) CLI tool installed
- iTerm2 3.3.0 or later
- macOS

## Quick Start

```bash
# List sessions
it2 session list

# Create a split layout
it2 session split --horizontal -q

# Send command to session
it2 session send-text "$SESSION_ID" "echo hello"

# Watch session activity
it2 session watch "$SESSION_ID"
```

## Documentation

Each skill includes:
- `SKILL.md` - Main guidance and quick start
- `references/` - Detailed technical reference
- `workflows/` - Step-by-step workflow guides

## License

MIT
