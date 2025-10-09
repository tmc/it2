# Claude Code Integrations for it2

This directory contains Claude Code plugins for the it2 CLI, providing comprehensive iTerm2 automation and session management capabilities.

## Available Plugins

### [it2-core](plugins/it2-core/)

**Core iTerm2 terminal automation with essential slash commands**

- **iterm2-terminal-automation** agent - 150+ it2 commands for comprehensive terminal control
- **/it2-split** command - Quick Claude session creation in a new split

**Install**: `/plugin add tmc/it2 --source integrations/claude-code/plugins/it2-core`

### [it2-claude-automation](plugins/it2-claude-automation/)

**Claude Code configuration and workflow automation**

- **claude-code-automation** agent - Manage settings, sessions, hooks, and rules
- **session-work-analyzer** agent - Analyze and document session patterns

**Install**: `/plugin add tmc/it2 --source integrations/claude-code/plugins/it2-claude-automation`

### [it2-claude-irc](plugins/it2-claude-irc/)

**IRC-style session management commands**

- **/join** - Create/switch to Claude sessions
- **/msg** - Send commands to specific sessions
- **/who** - List all active Claude sessions
- **/part** - Close sessions gracefully
- **/topic** - Set/view session badges

**Install**: `/plugin add tmc/it2 --source integrations/claude-code/plugins/it2-claude-irc`

## Quick Start

### Install All Plugins

```bash
# Core automation
/plugin add tmc/it2 --source integrations/claude-code/plugins/it2-core

# Claude automation
/plugin add tmc/it2 --source integrations/claude-code/plugins/it2-claude-automation

# IRC-style commands
/plugin add tmc/it2 --source integrations/claude-code/plugins/it2-claude-irc
```

### Quick Examples

```bash
# Create a new Claude split (from it2-core)
/it2-split

# Join a named session (from it2-claude-irc)
/join backend-dev

# See all active sessions (from it2-claude-irc)
/who

# Use the automation agent (from it2-core)
> Create three sessions for backend, frontend, and testing with badges

# Configure Claude Code (from it2-claude-automation)
> Set up this project with auto-approval for Read and git operations
```

## Installation Options

### From GitHub (Recommended)

```bash
/plugin add tmc/it2 --source integrations/claude-code/plugins/PLUGIN_NAME
```

### From Local Repository

```bash
/plugin add /path/to/it2/integrations/claude-code/plugins/PLUGIN_NAME
```

### Team Configuration

Add to `.claude/settings.json`:

```json
{
  "enabledPlugins": [
    "it2-core@github:tmc/it2",
    "it2-claude-automation@github:tmc/it2",
    "it2-claude-irc@github:tmc/it2"
  ]
}
```

## Plugin Comparison

| Feature | it2-core | it2-claude-automation | it2-claude-irc |
|---------|----------|---------------------|----------------|
| iTerm2 automation | ✅ Full (150+ commands) | ❌ | ❌ |
| Claude configuration | ❌ | ✅ Full | ❌ |
| Session analysis | ❌ | ✅ Pattern recognition | ❌ |
| Quick session creation | ✅ /it2-split | ❌ | ✅ /join |
| IRC-style commands | ❌ | ❌ | ✅ Full suite |
| Multi-session messaging | ❌ | ❌ | ✅ /msg |
| Session listing | ✅ (via agent) | ❌ | ✅ /who |

## Common Workflows

### Workflow 1: Multi-Session Development

```bash
# Use IRC-style commands for quick setup
/join backend
/join frontend
/join testing

# Use core agent for complex automation
> Set badges for all three sessions and create a broadcast domain
```

### Workflow 2: Automated Project Setup

```bash
# Create new session
/it2-split

# Configure the project
> Using claude-code-automation, configure this project with:
  - Auto-approval for Read, Glob, Grep
  - Git hooks for fmt and vet
  - Project-local settings for Go development
```

### Workflow 3: Session Monitoring and Control

```bash
# See what's running
/who

# Send commands to specific sessions
/msg ABC123 run the tests
/msg DEF456 analyze the architecture

# Check on them using the automation agent
> Show me the current status of all Claude sessions
```

## Requirements

- **iTerm2** with websocket server enabled
- **it2 CLI**: `go install github.com/tmc/it2@latest`
- **Claude Code** v2.0.12 or later
- **jq** for JSON processing (recommended)

## Documentation

- [Claude Code Integration Guide](../../docs/CLAUDE_CODE_INTEGRATION.md) - Comprehensive integration documentation
- [it2 CLI Documentation](https://github.com/tmc/it2) - Main it2 repository
- [Claude Code Plugins](https://docs.claude.com/en/docs/claude-code/plugins) - Official plugin documentation

## Contributing

To add a new plugin or improve existing ones:

1. Create your plugin in `plugins/your-plugin-name/`
2. Follow the plugin.json schema
3. Add comprehensive README.md
4. Test with Claude Code
5. Submit a pull request

## Support

- **Issues**: https://github.com/tmc/it2/issues
- **Discussions**: https://github.com/tmc/it2/discussions

## License

MIT - See LICENSE file for details
