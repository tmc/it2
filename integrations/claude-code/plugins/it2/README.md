# it2 Terminal Automation Plugin

A comprehensive Claude Code plugin providing iTerm2 terminal automation and AI-assisted development workflows using the it2 CLI.

## Overview

This plugin provides three essential agents for:
- **iTerm2 Automation**: Comprehensive session management, monitoring, and control
- **Claude Code Integration**: Configuration management and workflow automation
- **Session Analysis**: Understanding and documenting work patterns

## Installation

### From GitHub

```bash
# Add the plugin
/plugin add tmc/it2 --source integrations/claude-code/plugins/it2

# Or install directly
/plugin install it2@github:tmc/it2
```

### From Local Repository

```bash
# Add plugin from local path
/plugin add /Volumes/tmc/go/src/github.com/tmc/it2/integrations/claude-code/plugins/it2

# Or using relative path if you're in the it2 directory
/plugin add ./integrations/claude-code/plugins/it2
```

## Available Agents

### iterm2-terminal-automation (v2.0.0)
Comprehensive iTerm2 automation specialist covering 150+ it2 commands across 15 categories:
- Session, tab, window, and app management
- Event-driven monitoring and notifications (14 notification types)
- Variable management and broadcast domains
- Text operations, buffer manipulation, and search
- Profile management and workspace arrangements

**Use when**: You need to automate iTerm2 sessions, monitor terminal states, or orchestrate multi-session workflows.

### claude-code-automation (v1.0.0)
Complete Claude Code configuration and session automation:
- Multi-scope settings management (user/project/local)
- Automated session creation and control
- Hooks and rules configuration
- Session state monitoring and recovery
- Slash command automation (/status, /statusline, etc.)

**Use when**: You need to configure Claude Code programmatically, automate Claude sessions, or manage settings across scopes.

### session-work-analyzer (v1.0.0)
Analyze session activity to understand and document work patterns:
- Extract work patterns from session buffers
- Generate agent definitions from observed workflows
- Characterize session activities and tool usage
- Create specialized agents based on real usage patterns

**Use when**: You want to understand what work is happening in a session or create agents based on actual workflow patterns.

## Usage Examples

### Automate iTerm2 Sessions

```bash
# The agent handles all the complexity
/task Create three iTerm2 sessions for backend, frontend, and testing with proper badging
```

### Configure Claude Code

```bash
# Let the agent configure settings
/task Configure project settings for Go development with hooks for fmt and vet
```

### Analyze Session Work

```bash
# Understand what's happening in a session
/task Analyze session ABC123 and create an agent definition for the work being done
```

## Requirements

- iTerm2 with websocket server enabled
- it2 CLI installed: `go install github.com/tmc/it2@latest`
- Claude Code for AI agent functionality
- jq for JSON processing (recommended)

## Integration Guides

See the comprehensive integration guides in the docs directory:
- [Claude Code Integration](../../../docs/CLAUDE_CODE_INTEGRATION.md)
- [Gemini CLI Integration](../../../docs/GEMINI_CLI_INTEGRATION.md)
- [OpenAI Codex Integration](../../../docs/OPENAI_CODEX_INTEGRATION.md)

## Support

- **Issues**: https://github.com/tmc/it2/issues
- **Documentation**: See docs/ directory
- **it2 CLI Documentation**: https://github.com/tmc/it2

## License

MIT - See LICENSE file for details
