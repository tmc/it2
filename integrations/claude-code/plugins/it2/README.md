# it2 Terminal Automation Plugin

A comprehensive Claude Code plugin providing iTerm2 terminal automation and AI-assisted development workflows using the it2 CLI.

## Overview

This plugin provides a session analysis agent. For iTerm2 automation, see the `it2-core` plugin. For Claude Code configuration, see the `it2-claude-automation` plugin.

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

### session-work-analyzer
Analyze session activity to understand and document work patterns:
- Extract work patterns from session buffers
- Generate agent definitions from observed workflows
- Characterize session activities and tool usage
- Create specialized agents based on real usage patterns

**Use when**: You want to understand what work is happening in a session or create agents based on actual workflow patterns.

## Usage Examples

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
