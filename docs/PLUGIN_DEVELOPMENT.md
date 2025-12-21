# it2 Plugin Development Guide

This guide covers developing plugins for the it2 command-line interface for iTerm2 automation.

> **See also:** [TAXONOMY.md](TAXONOMY.md) for the authoritative definitions of plugins, hooks, and tools.

## Overview

**Plugins** are external executables that extend it2 functionality. They are discovered automatically from your PATH and invoked programmatically by it2.

Plugins can:

1. **Enrich listings** - Add metadata to session, tab, and window listings
2. **Automate responses** - Monitor and respond to patterns
3. **Categorize entities** - Apply context-specific logic for better organization

### Plugins vs Hooks vs Tools

| Concept | Definition | Invocation |
|---------|------------|------------|
| **Plugin** | Extends it2, auto-discovered | By it2 (`it2 plugin <name>`) |
| **Hook** | Responds to external tool events | By Claude Code, Gemini CLI, etc. |
| **Tool** | Standalone CLI program | Directly by user |

Plugins are **it2-facing**: it2 discovers and invokes them. Hooks are **external-tool-facing**: Claude Code or other tools invoke them based on their configuration.

## Plugin Architecture

```
it2 (core)
  ↓ Discovery
Plugin Registry (internal/plugins/)
  ↓ Execution
External Executables (it2-*)
  ↓ Integration
Enhanced Output
```

### Plugin Types

it2 recognizes three types of plugins based on executable naming:

1. **Session Enrichers** (`it2-session-*`) - Add data to session listings
2. **Tab Enrichers** (`it2-tab-*`) - Add data to tab listings
3. **Window Enrichers** (`it2-window-*`) - Add data to window listings
4. **Generic Plugins** (`it2-*`) - Apply to all entity types

### Capabilities

Plugins declare capabilities that determine what they can do:

**Enrichment Capabilities:**
- `session` - Adds data to session listings
- `tab` - Adds data to tab listings
- `window` - Adds data to window listings
- `process` - Adds process inspection data

**Automation Capabilities:**
- `suggest` - Returns recommendations (e.g., "send 'continue'")
- `execute` - Performs actions (sends keystrokes, modifies state)

Capabilities are inferred from the plugin name or declared explicitly in a manifest. See [plugin-manifest-schema.md](plugin-manifest-schema.md) for manifest format.

### Trust Levels

Plugins are assigned trust levels based on their source:

| Level | Source | Permissions |
|-------|--------|-------------|
| **Core** | Embedded in it2 binary | Full access |
| **Verified** | Reproducible build verified | As declared in manifest |
| **Community** | Has manifest, not verified | As declared in manifest |
| **Untrusted** | No manifest | Minimal (session ID only) |

## Quick Start

### 1. Create a Simple Session Plugin

Create an executable named `it2-session-example`:

```bash
#!/bin/bash
# File: it2-session-example

SESSION_ID="$1"
SESSION_NAME="$2"

# Example: Show if session is running a specific tool
if pgrep -f "vim.*$SESSION_ID" >/dev/null; then
    echo "vim-active"
elif pgrep -f "git.*$SESSION_ID" >/dev/null; then
    echo "git-active"
else
    echo "shell"
fi
```

Make it executable:
```bash
chmod +x it2-session-example
```

Add to PATH:
```bash
export PATH="$PATH:$(pwd)"
```

### 2. Test Your Plugin

```bash
it2 session list
```

Your plugin output will appear in the session listing as additional metadata.

## Plugin Interface

### Session Enrichers

**Input Arguments:**
- `$1`: Session ID (required)
- `$2`: Session Name (optional, empty if not set)

**Expected Output:**
- Single line string describing the session state
- Empty output is ignored
- Exit code 0 for success (non-zero ignored, not treated as error)

**Example:**
```bash
#!/bin/bash
SESSION_ID="$1"
SESSION_NAME="$2"

# Check if this is a Claude Code session
if [ "$SESSION_NAME" = "Claude Code" ] || echo "$SESSION_NAME" | grep -q "claude"; then
    echo "claude-session"
fi
```

### Tab Enrichers

**Input Arguments:**
- `$1`: Tab ID (required)
- `$2`: Window ID (required)
- `$3`: Tab Title (optional)

**Expected Output:**
- Single line string describing tab state
- Empty output is ignored

### Window Enrichers

**Input Arguments:**
- `$1`: Window ID (required)
- `$2`: Window Title (optional)

**Expected Output:**
- Single line string describing window state
- Empty output is ignored

## Advanced Examples

### Multi-Purpose Plugin

Create `it2-dev-tools` that works across all entity types:

```bash
#!/bin/bash
# Universal development tool detector

if [ $# -eq 1 ]; then
    # Window enricher mode (1 arg = window ID)
    WINDOW_ID="$1"
    # Check if any session in this window is running dev tools
    if it2 session list --format json | jq -r ".[] | select(.window_id == \"$WINDOW_ID\") | .session_name" | grep -E "(vim|code|git)" >/dev/null; then
        echo "development-window"
    fi
elif [ $# -eq 2 ]; then
    # Tab enricher mode (2 args = tab ID, window ID)
    TAB_ID="$1"
    WINDOW_ID="$2"
    # Check sessions in this specific tab
    if it2 session list --format json | jq -r ".[] | select(.tab_id == \"$TAB_ID\") | .session_name" | grep -E "(vim|code)" >/dev/null; then
        echo "coding-tab"
    fi
elif [ $# -ge 2 ]; then
    # Session enricher mode (2+ args = session ID, session name, ...)
    SESSION_ID="$1"
    SESSION_NAME="$2"

    # Detailed session analysis
    if echo "$SESSION_NAME" | grep -q "vim"; then
        echo "vim-session"
    elif echo "$SESSION_NAME" | grep -q "git"; then
        echo "git-session"
    elif pgrep -f "node.*$SESSION_ID" >/dev/null; then
        echo "node-dev"
    fi
fi
```

### Configuration-Based Plugin

Create `it2-session-classifier` that reads from config:

```bash
#!/bin/bash
CONFIG_FILE="$HOME/.it2/plugins/classifier.yaml"

SESSION_ID="$1"
SESSION_NAME="$2"

if [ ! -f "$CONFIG_FILE" ]; then
    exit 0
fi

# Parse YAML config (requires yq or similar)
if command -v yq >/dev/null; then
    PATTERNS=$(yq eval '.patterns[]' "$CONFIG_FILE" 2>/dev/null)
    while IFS= read -r pattern; do
        if echo "$SESSION_NAME" | grep -qE "$pattern"; then
            LABEL=$(yq eval ".patterns | to_entries | .[] | select(.value == \"$pattern\") | .key" "$CONFIG_FILE")
            echo "$LABEL"
            exit 0
        fi
    done <<< "$PATTERNS"
fi
```

With config file `~/.it2/plugins/classifier.yaml`:
```yaml
patterns:
  development: "(vim|code|emacs|nvim)"
  git: "(git|gh|hub)"
  database: "(mysql|postgres|redis|mongo)"
  containers: "(docker|kubectl|k8s)"
  cloud: "(aws|gcp|azure|terraform)"
```

## Plugin Discovery

it2 discovers plugins by:

1. Scanning all directories in `$PATH`
2. Looking for executable files with names starting with `it2-`
3. Determining type based on naming pattern:
   - `it2-session-*` → Session enricher
   - `it2-tab-*` → Tab enricher
   - `it2-window-*` → Window enricher
   - `it2-*` → Generic (applies to all types)

## Integration Points

### Session Listings

Plugins add data to the `PluginData` field in session listings:

```json
{
  "session_id": "sess123",
  "session_name": "vim-work",
  "plugin_data": {
    "example": "vim-active",
    "dev-tools": "vim-session"
  }
}
```

### Tab Listings

Similar integration for tab metadata:

```json
{
  "tab_id": "tab456",
  "sessions": [...],
  "plugin_data": {
    "dev-tools": "coding-tab"
  }
}
```

## Best Practices

### Performance
- Plugins have a 5-second timeout - keep them fast
- Cache expensive operations when possible
- Exit early for irrelevant sessions/tabs/windows

### Error Handling
- Non-zero exit codes are ignored, not treated as errors
- Empty output is ignored
- Use stderr for debugging (visible with `ITERM2_DEBUG=1`)

### Naming Conventions
- Use descriptive plugin names: `it2-session-git-status` not `it2-git`
- Use hyphens for multi-word names: `it2-session-docker-status`
- Be specific about what the plugin does

### Output Formatting
- Keep output concise (single line preferred)
- Use consistent terminology
- Consider localization if relevant

### Configuration
- Store config in `~/.it2/plugins/`
- Use standard formats (YAML, JSON, TOML)
- Provide sensible defaults
- Document configuration options

## Debugging

### Enable Debug Mode
```bash
export ITERM2_DEBUG=1
it2 session list
```

### Test Plugin Directly
```bash
./it2-session-example "session-id" "session-name"
echo $?  # Should be 0
```

### Common Issues
1. **Plugin not discovered**: Check it's in PATH and executable
2. **No output**: Verify plugin returns non-empty string
3. **Timeout**: Plugin takes longer than 5 seconds
4. **Wrong arguments**: Check argument handling for your plugin type

## Distribution

### Packaging
- Create a directory with your plugin and README
- Include configuration examples
- Add installation script if needed

### Example Structure
```
my-plugin/
├── it2-session-mystatus
├── it2-tab-mystatus
├── README.md
├── config.yaml.example
└── install.sh
```

### Installation Methods

**Method 1: PATH Installation**
```bash
cp it2-session-mystatus /usr/local/bin/
chmod +x /usr/local/bin/it2-session-mystatus
```

**Method 2: Plugin Directory**
```bash
mkdir -p ~/.it2/plugins/mystatus
cp it2-session-mystatus ~/.it2/plugins/mystatus/
export PATH="$PATH:$HOME/.it2/plugins/mystatus"
```

## Examples Repository

See the `examples/` directory for working plugin implementations:

- **claude-monitoring/**: Complete Claude Code session monitoring
- **vim-monitoring/**: Advanced vim session detection and assistance
- **tmux-monitoring/**: tmux session integration
- **shell-monitoring/**: Shell state detection

## Using the CLI

### List Available Plugins

```bash
it2 plugin list
```

### Run a Plugin

```bash
# Via it2 (automatic session ID injection)
it2 plugin <name> [args...]

# Or directly (if in PATH)
it2-session-example <session-id> [args...]
```

The `it2 plugin` wrapper provides:
- Automatic session ID injection from `$ITERM_SESSION_ID`
- Plugin discovery information
- Consistent error handling

### Install Claude Code Hooks

```bash
# Install to project settings (.claude/settings.json)
it2 plugin claude-code-hook --install

# Install to global settings (~/.claude/settings.json)
it2 plugin claude-code-hook --install --scope global
```

## Plugin Registry (Future)

Future versions may include:
- Plugin registry with `it2 plugin install <name>`
- Automatic updates
- Plugin marketplace
- Dependency management

## Contributing

To contribute plugins or improvements:

1. Test thoroughly with various session types
2. Follow naming conventions
3. Include comprehensive documentation
4. Add examples to the examples directory
5. Submit pull requests with test cases

For complex plugins, consider:
- Unit tests for plugin logic
- Integration tests with it2
- Performance benchmarks
- Cross-platform compatibility