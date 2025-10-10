# Claude Code Integration with it2

A comprehensive guide to using Claude Code with the it2 CLI through our plugin.

**Last Updated**: 2025-10-09
**Quick Start**: Install the it2 plugin and access 3 essential agents for iTerm2 automation

## Table of Contents

- [Introduction](#introduction)
- [Quick Start](#quick-start)
- [Plugin Marketplace](#plugin-marketplace)
- [Available Plugins & Agents](#available-plugins--agents)
- [Installation](#installation)
- [Usage Examples](#usage-examples)
- [Automation Workflows](#automation-workflows)
- [Advanced Topics](#advanced-topics)
- [Troubleshooting](#troubleshooting)
- [Reference](#reference)
- [See Also](#see-also)
- [Contributing](#contributing)

---

## Introduction

### What is the it2 Marketplace?

The it2 repository provides a Claude Code marketplace with 4 essential plugins:

- **core**: iTerm2 terminal automation with iterm2-terminal-automation agent
- **claude-automation**: Claude Code configuration and session management
- **claude-plugin-development**: Tools for developing Claude Code plugins and marketplaces
- **claude-agent-development**: Specialized agents for creating and testing Claude Code agents

### Why Use This Marketplace?

- **Pre-built Expertise**: Agents with deep knowledge of it2 CLI, iTerm2, and Claude Code
- **Production-Ready**: Battle-tested patterns and best practices built-in
- **Composable**: Mix and match plugins for complex workflows
- **Maintained**: Updated alongside it2 CLI and Claude Code development

---

## Quick Start

### Install the Marketplace (30 seconds)

```bash
# Launch Claude Code
claude

# Add the marketplace from GitHub
/plugin marketplace add tmc/it2

# Or from local path
/plugin marketplace add /path/to/it2

# Browse available plugins
/plugin

# Install all plugins
# The format is: /plugin install it2@<full-plugin-directory-name>
/plugin install it2@it2-core
/plugin install it2@it2-claude-automation
/plugin install it2@it2-claude-plugin-development
/plugin install it2@it2-claude-agent-development
```

That's it! You now have access to iTerm2 automation, Claude Code configuration, and agent development tools.

### Your First Automation

```bash
# Use the agent you just installed
> Create three iTerm2 sessions for backend, frontend, and testing with badges

# Claude will use the iterm2-terminal-automation agent to:
# - Create sessions with proper splits
# - Set descriptive badges
# - Configure session titles
# - Track session IDs for you
```

---

## Manual Automation (Without Agents)

You can automate iTerm2 and Claude Code sessions manually using it2 CLI commands, but **we strongly recommend using the plugin marketplace** for a better experience.

### Why Use the Marketplace Instead?

| Manual Commands | Marketplace Agents |
|----------------|-------------------|
| ❌ Need to know 150+ commands | ✅ Natural language requests |
| ❌ Write complex bash scripts | ✅ Agents handle complexity |
| ❌ Manual error handling | ✅ Built-in error recovery |
| ❌ Maintain your own code | ✅ Updated with it2 releases |
| ❌ Learn notification APIs | ✅ Event-driven patterns included |

### Basic Manual Commands

If you need to use it2 directly:

```bash
# Create a session and launch Claude
SESSION_ID=$(it2 session split --vertical --quiet)
it2 session send-text "$SESSION_ID" "cd ~/project && claude"

# Wait for Claude to be ready
sleep 5
while [ "$(it2-session-has-no-queued-claude-messages "$SESSION_ID")" != "true" ]; do
    sleep 2
done

# Send a command
it2 session send-text "$SESSION_ID" "analyze the code"

# Get session output
it2 text get-buffer "$SESSION_ID" --lines 20

# Set badge for identification
it2 badge set "$SESSION_ID" "🔵 Analysis"
```

### Common Manual Patterns

**Multi-session setup**:
```bash
# Create three sessions
BACKEND=$(it2 session split --horizontal --quiet)
FRONTEND=$(it2 session split --horizontal --quiet)
TESTING=$(it2 session split --horizontal --quiet)

# Set titles and badges
it2 session set-title "$BACKEND" "Backend"
it2 badge set "$BACKEND" "🔴 Backend"

# Launch Claude in each
for SID in "$BACKEND" "$FRONTEND" "$TESTING"; do
    it2 session send-text "$SID" "cd ~/project && claude"
    sleep 3
done
```

**Event-driven monitoring** (advanced):
```bash
# Subscribe to screen updates (no polling!)
it2 notification monitor --type screen --session "$SESSION_ID" --format json | \
while read -r event; do
    # Check for modals
    screen=$(it2 text get-screen "$SESSION_ID")
    if echo "$screen" | grep -q "Do you want to proceed"; then
        it2 session send-key "$SESSION_ID" "1"
        it2 session send-key "$SESSION_ID" "Return"
    fi
done
```

### Recommendation: Use the Marketplace

Instead of writing and maintaining these scripts, install the marketplace:

```bash
# One-time setup
/plugin marketplace add tmc/it2
/plugin install it2@it2-core

# Then just ask Claude
> Create three sessions for backend, frontend, and testing with proper badges
```

The agents handle all the complexity above (and much more) automatically.

---

## Plugin Marketplace

### Marketplace Structure

The it2 marketplace is located at `.claude-plugin/` in the repository:

```
.claude-plugin/
└── marketplace.json    # Plugin manifest
```

### Adding the Marketplace

**Individual Users** (recommended):

Use the `/plugin marketplace add` command:

```bash
# From GitHub (recommended)
/plugin marketplace add tmc/it2

# From local repository
/plugin marketplace add /path/to/it2

# From marketplace.json URL
/plugin marketplace add https://raw.githubusercontent.com/tmc/it2/main/.claude-plugin/marketplace.json
```

**Teams** (automatic installation):

Add to your repository's `.claude/settings.json`:
```json
{
  "extraKnownMarketplaces": {
    "it2": {
      "source": {
        "source": "github",
        "repo": "tmc/it2"
      }
    }
  }
}
```

Team members will have the marketplace available automatically when they trust your repository.

### Verify Installation

```bash
/plugin marketplace list
# Output: • it2
```

---

## Available Plugins & Agents

### Plugin 1: core

**iTerm2 terminal automation tools**

#### iterm2-terminal-automation

**Comprehensive iTerm2 automation covering 150+ it2 commands**

**Categories covered**:
- Session, tab, window management
- Event-driven monitoring (14 notification types)
- Variable management and state persistence
- Broadcast domains for multi-session control
- Text operations and buffer manipulation

**Use when**: You need any iTerm2 automation capability

**Example**:
```bash
> Set up a development environment with 3 sessions: backend on port 8080,
  frontend on 3000, and database monitoring
```

**Commands**: `/it2-split`

---

### Plugin 2: claude-automation

**Claude Code configuration and session automation**

#### claude-code-automation

**Complete Claude Code configuration and workflow automation**

**Features**:
- Multi-scope settings (user/project/local)
- Session orchestration
- Hooks and rules configuration
- MCP server integration

**Use when**: You need to configure or automate Claude Code workflows

**Example**:
```bash
> Configure Claude Code to auto-approve Bash(git) and Read operations
  for this project
```

---

### Plugin 3: claude-plugin-development

**Tools for developing Claude Code plugins and marketplaces**

#### plugin-creator

**Create new Claude Code plugins**

**Features**:
- Plugin scaffolding
- Manifest generation
- Documentation templates
- Validation helpers

**Use when**: You need to create a new Claude Code plugin

**Example**:
```bash
> Create a new plugin for Docker automation with agents for container management
```

#### marketplace-manager

**Manage Claude Code marketplaces**

**Features**:
- Marketplace configuration
- Plugin registration
- Version management
- Distribution helpers

**Use when**: You need to manage a plugin marketplace

**Commands**: `/claude-plugin-validate`, `/claude-plugin-validate-all`

---

### Plugin 4: claude-agent-development

**Specialized agents for creating and testing Claude Code agents**

#### agent-creator

**Create new Claude Code agents from scratch**

**Features**:
- CLI tool-based agent creation
- Documentation research
- Capability validation
- Agent definition generation

**Use when**: You need to create new specialized agents

**Example**:
```bash
> Create an agent for the kubectl CLI that specializes in debugging pod issues
```

#### agent-definition-improver

**Review and improve existing agent definitions**

**Features**:
- Capability grounding
- Over-promise detection
- Realistic scope definition
- Documentation improvement

**Use when**: You need to refine agent definitions

#### agent-from-command-line-tool-builder

**Build agents around CLI tools**

**Features**:
- Tool capability exploration
- Help documentation analysis
- Agent scaffolding for tools
- Best practice patterns

**Use when**: You want to wrap a CLI tool in an agent

#### agent-output-comparator

**Test agent consistency across runs**

**Features**:
- Output comparison
- Behavioral validation
- Regression detection
- Quality assurance

**Use when**: You need to test agent reliability

#### agent-suggestor

**Suggest new agents based on workflow patterns**

**Features**:
- Session analysis
- Pattern recognition
- Agent recommendations
- Gap identification

**Use when**: You want suggestions for helpful agents

#### agent-testing-and-evaluation

**Validate and test agent definitions**

**Features**:
- Isolated testing environment
- Validation scenarios
- Performance evaluation
- Quality metrics

**Use when**: You need to test agent behavior

---

## Installation

### Recommended: Use CLI Commands

For individual users, use the interactive `/plugin` commands:

```bash
# 1. Add the marketplace
/plugin marketplace add tmc/it2

# 2. Browse and select plugins interactively
/plugin

# Or install specific plugins directly
/plugin install it2@it2-core
/plugin install it2@it2-claude-automation
/plugin install it2@it2-claude-plugin-development
/plugin install it2@it2-claude-agent-development
```

**Why use CLI commands?**
- Interactive browsing and discovery
- No manual file editing
- Immediate feedback on installation
- Easier to update and manage

### Team Setup: Automatic Installation

For teams, add marketplace and plugins to `.claude/settings.json` in your repository:

```json
{
  "extraKnownMarketplaces": {
    "it2": {
      "source": {
        "source": "github",
        "repo": "tmc/it2"
      }
    }
  },
  "enabledPlugins": [
    "it2@it2-core",
    "it2@it2-claude-automation",
    "it2@it2-claude-plugin-development",
    "it2@it2-claude-agent-development"
  ]
}
```

**How it works:**
- Team members clone the repository
- When they trust the repository folder, plugins install automatically
- No manual setup required for team members
- Consistent configuration across the team

### Verify Installation

```bash
# List installed agents
/plugin list

# Test an agent
> List all iTerm2 sessions using it2
```

---

## Usage Examples

### Example 1: Multi-Session Development Environment

```bash
> Using the iterm2-terminal-automation agent, create a development
  environment with:
  - Backend session running on port 8080
  - Frontend session running on port 3000
  - Testing session for running tests
  - Database monitoring session

  Set badges to identify each session and use broadcast domain to send
  'git status' to all sessions
```

The agent will:
1. Create 4 sessions with proper splits
2. Set descriptive badges (🔴 Backend, 🔵 Frontend, 🟢 Testing, 🟡 Database)
3. Navigate each to the correct directory
4. Set up broadcast domain
5. Execute synchronized commands

### Example 2: Configure Claude Code Automation

```bash
> Using the claude-code-automation agent, configure this project to:
  - Auto-approve git read operations (status, diff, log)
  - Auto-approve Read, Grep, and Glob operations
  - Set up project-specific settings for Python development
  - Add git pre-commit hooks
```

The agent will:
1. Create .claude/settings.json with appropriate scopes
2. Configure auto-approve patterns
3. Set up language-specific configurations
4. Install and configure hooks

### Example 3: Create Custom Agent

```bash
> Use agent-creator to create a new agent for kubectl that specializes in
  debugging pod issues. Include commands for logs, describe, events, and
  common troubleshooting patterns.
```

The agent-creator will:
1. Research kubectl documentation
2. Identify relevant commands and patterns
3. Generate agent definition with examples
4. Test the agent definition
5. Save to your project

### Example 4: Develop and Test Plugins

```bash
> Use plugin-creator to scaffold a new plugin for Docker automation, then
  use marketplace-manager to add it to a local marketplace for testing.
```

The agents will:
1. Create plugin directory structure
2. Generate plugin.json manifest
3. Create documentation templates
4. Add to marketplace configuration
5. Validate the plugin structure

---

## Automation Workflows

### Workflow 1: Parallel Testing Across Projects

Using `iterm2-terminal-automation`:

```bash
> Create sessions for these 3 projects and run tests in parallel:
  - ~/projects/backend
  - ~/projects/frontend
  - ~/projects/api

  Monitor for completion and report results
```

### Workflow 2: Plugin Development Pipeline

Using agent development plugins:

```bash
> Create a new plugin for Kubernetes operations:
  1. Use agent-creator to create kubectl debugging agent
  2. Use agent-creator to create helm deployment agent
  3. Use plugin-creator to scaffold the plugin structure
  4. Use marketplace-manager to add to local marketplace
  5. Use agent-testing-and-evaluation to validate agents

  Package everything for distribution
```

### Workflow 3: Automated Development Environment

Using `claude-code-automation` and `iterm2-terminal-automation`:

```bash
> Set up an automated development environment:
  1. Create sessions for backend, frontend, and testing
  2. Configure Claude Code auto-approvals per session
  3. Set up session badges and titles for identification
  4. Use broadcast domain for synchronized git operations
  5. Monitor sessions for completion
```

---

## Advanced Topics

### Event-Driven Session Monitoring

The `iterm2-terminal-automation` agent supports iTerm2's notification system for real-time monitoring:

```bash
> Using iterm2-terminal-automation, monitor all Claude Code sessions and:
  - Track session state changes in real-time
  - Detect when sessions are idle or processing
  - Get notifications when commands complete
  - React to screen updates without polling

  Use iTerm2's notification API for event-driven monitoring
```

**How it works**:
- Uses `it2 notification` commands
- Subscribes to session events (screen, layout, alert, etc.)
- Reacts instantly to state changes
- No sleep/polling overhead

### Multi-Agent Orchestration

Combine agents from different plugins for complex workflows:

```bash
> Orchestrate these agents:
  1. iterm2-terminal-automation: Create 3 sessions with specific layouts
  2. claude-code-automation: Configure auto-approvals for each session
  3. agent-creator: Generate project-specific automation agents
  4. plugin-creator: Package agents into a team plugin

  Coordinate execution and save configuration for team use
```

### Session State Persistence

Using variable management:

```bash
> For each active iTerm2 session, save:
  - Working directory
  - Session role (backend/frontend/testing)
  - Current git branch
  - Last command status

  Store in session variables and restore after restart
```

### Custom Agent Development

Create project-specific agents:

```bash
> Use agent-creator to create a "deployment-automator" agent that:
  - Knows our deployment process
  - Can deploy to staging/production
  - Runs smoke tests
  - Sends Slack notifications

  Base it on the kubectl and helm CLIs
```

---

## Troubleshooting

### Marketplace Not Found

```bash
# Check if .claude-plugin/marketplace.json exists
ls -la /path/to/it2/.claude-plugin/

# Validate marketplace
claude plugin validate /path/to/it2/.claude-plugin/marketplace.json

# Re-add marketplace
/plugin marketplace remove it2
/plugin marketplace add tmc/it2
```

### Agent Not Available

```bash
# List available agents
/plugin

# Check installed agents
/plugin list

# Reinstall specific plugin
/plugin install it2@it2-core
```

### Agent Not Working as Expected

```bash
# Check agent version
/plugin list

# Review agent documentation
> Show me the documentation for the iterm2-terminal-automation agent

# Test with simple command
> Using iterm2-terminal-automation, list all sessions
```

### GitHub Access Issues

If installing from GitHub marketplace:

```bash
# Use local path instead
/plugin marketplace add tmc/it2
```

---

## Reference

### Command Quick Reference

```bash
# Marketplace management
/plugin marketplace add tmc/it2
/plugin marketplace list
/plugin marketplace update it2
/plugin marketplace remove it2

# Plugin management
/plugin                                    # Browse available
/plugin list                              # List installed
/plugin install <name>@<marketplace>      # Install
/plugin uninstall <name>                  # Uninstall

# Using agents
> Using <agent-name>, do <task>
```

### Agent Selection Guide

| Task | Agent | Plugin |
|------|-------|--------|
| iTerm2 automation | iterm2-terminal-automation | it2-core |
| Configure Claude Code | claude-code-automation | it2-claude-automation |
| Create new agent | agent-creator | it2-claude-agent-development |
| Improve agent | agent-definition-improver | it2-claude-agent-development |
| Wrap CLI tool | agent-from-command-line-tool-builder | it2-claude-agent-development |
| Test agents | agent-testing-and-evaluation | it2-claude-agent-development |
| Suggest agents | agent-suggestor | it2-claude-agent-development |
| Create plugin | plugin-creator | it2-claude-plugin-development |
| Manage marketplace | marketplace-manager | it2-claude-plugin-development |

### Prerequisites

- **iTerm2** with websocket server enabled
- **it2 CLI** installed: `go install github.com/tmc/it2@latest`
- **Claude Code** v2.0.12 or later
- **jq** for JSON processing (recommended)
- **Git** for GitHub marketplace access

### Environment Variables

```bash
# Current iTerm2 session
echo $ITERM_SESSION_ID

# iTerm2 authentication
echo $ITERM2_COOKIE
echo $ITERM2_KEY
```

### File Locations

```bash
# Marketplace configuration
.claude-plugin/marketplace.json    # Plugin manifest
```

For Claude Code configuration file locations, see the [official documentation](https://docs.claude.com/en/docs/claude-code/plugins).

---

## See Also

- **[it2 Repository](https://github.com/tmc/it2)** - Main it2 CLI repository
- **[it2 Automation Best Practices](../.claude/it2-automation.md)** - Core automation patterns
- **[Claude Code Plugins](https://docs.claude.com/en/docs/claude-code/plugins)** - Official plugin documentation
- **[Claude Code Marketplaces](https://docs.claude.com/en/docs/claude-code/plugin-marketplaces)** - Marketplace guide
- **[Gemini CLI Integration](./GEMINI_CLI_INTEGRATION.md)** - Alternative AI CLI tool
- **[OpenAI Codex Integration](./OPENAI_CODEX_INTEGRATION.md)** - OpenAI CLI integration

---

## Contributing

### Adding New Agents

1. Create agent definition following the [plugin manifest schema](https://docs.claude.com/en/docs/claude-code/plugins-reference)
2. Add entry to `.claude-plugin/marketplace.json`
2. Add plugin to `integration/claude-code/plugins/*`

### Reporting Issues

- **it2 Issues**: https://github.com/tmc/it2/issues
- **Claude Code**: https://support.claude.com/

---

**Last Updated**: 2025-10-09
**Marketplace Version**: 1.0.0
**Total Plugins**: 4
**Total Agents**: 10
