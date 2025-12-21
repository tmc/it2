# it2 Extensibility Taxonomy

This document defines the taxonomy for it2's extensibility mechanisms. It serves as the authoritative reference for terminology used throughout the codebase and documentation.

## Overview

it2's extensibility model distinguishes between three concepts:

1. **Tools** — Standalone CLI programs for direct user invocation
2. **Plugins** — Executables discovered and exec'd by it2 programmatically
3. **Hooks** — Executables triggered by external tools (Claude Code, Gemini CLI, etc.)

These share implementation patterns but differ in invocation model and lifecycle.

---

## Core Concepts

### Tool

A **tool** is a standalone CLI program designed for direct user invocation from the command line.

**Characteristics:**
- Invoked directly by users (e.g., `it2`, `it2-convert`, `it2-migrate`)
- Self-contained with its own argument parsing and help
- Not discovered or managed by it2
- May use it2's libraries internally

**Examples:**
```
it2                    # The main CLI tool
it2-convert            # (hypothetical) Standalone format converter
it2-debug              # (hypothetical) Debugging utility
```

**Key distinction:** Tools are user-facing; plugins are it2-facing (though plugins *can* be run directly).

### Plugin

A **plugin** is an external executable that extends it2 functionality. Plugins are discovered automatically and managed by it2.

**Characteristics:**
- Executable files with `it2-` prefix
- Discovered from `PATH`, `IT2_PLUGIN_PATHS`, or embedded in the binary
- Receive structured input (arguments or JSON on stdin)
- Produce structured output (text or JSON on stdout)
- Sandboxed based on trust level

**Discovery Order (highest to lowest priority):**
1. Executables in `PATH`
2. Directories from `IT2_PLUGIN_PATHS` environment variable
3. Embedded plugins extracted from the it2 binary

**Example plugins:**
```
it2-session-is-claude-code      # Detects Claude Code sessions
it2-session-claude-has-modal    # Detects modal dialogs
it2-tab-git-status              # Shows git status for tabs
```

### Hook

A **hook** is an executable triggered by external tools in response to events. Hooks are configured externally, not discovered by it2.

**Characteristics:**
- Registered in external tool configuration (e.g., Claude Code's `settings.json`)
- Receive event data via stdin (JSON)
- Used for logging, automation, or side effects
- Lifecycle managed by the external tool, not it2

**Key distinction from plugins:** Hooks respond to events from external tools; plugins are invoked by it2 itself.

**Example hooks:**
```
it2-claude-code-hook    # Logs Claude Code events to session artifacts
it2-gemini-hook         # (future) Logs Gemini CLI events
```

**Hook configuration example (Claude Code):**
```json
{
  "hooks": {
    "PostToolUse": [{
      "hooks": [{"type": "command", "command": "it2-claude-code-hook"}]
    }]
  }
}
```

---

## Plugin Capabilities

Plugins declare their capabilities in a manifest. This enables capability-based discovery and appropriate sandboxing.

### Enrichment Capabilities

Enrichment plugins add data to it2 listings (sessions, tabs, windows).

| Capability | Naming Pattern | Description |
|------------|----------------|-------------|
| `session` | `it2-session-*` | Adds data to session listings |
| `tab` | `it2-tab-*` | Adds data to tab listings |
| `window` | `it2-window-*` | Adds data to window listings |
| `process` | `it2-session-process-*` | Adds process inspection data |

**Example manifest:**
```yaml
capabilities:
  enrichment:
    - session
    - process
```

### Automation Capabilities

Automation plugins can suggest or execute actions.

| Capability | Description |
|------------|-------------|
| `suggest` | Returns recommendations (e.g., "send 'continue'") |
| `execute` | Performs actions (sends keystrokes, modifies state) |

**Example:**
- `it2-session-claude-suggest-action` — Suggests interventions for Claude sessions
- `it2-session-claude-auto-approve` — Executes safe approvals

### Status Line Capabilities (Future)

| Capability | Description |
|------------|-------------|
| `component` | Renders in iTerm2 status bar |

---

## Trust Levels

Plugins are assigned trust levels that determine their permissions and sandbox constraints.

| Level | Source | Permissions | Sandbox |
|-------|--------|-------------|---------|
| **Core** | Embedded in it2 binary | Full access | None |
| **Verified** | Reproducible build matches manifest | As declared in manifest | Relaxed |
| **Community** | From PATH, has valid manifest | As declared in manifest | Strict |
| **Untrusted** | From PATH, no manifest | Minimal (session ID only) | Strict |

### Trust Level Details

**Core:**
- Ships with it2
- Full access to session data, screen content, environment
- No sandbox restrictions

**Verified:**
- Source available, reproducible build verified
- Permissions declared in manifest
- Relaxed sandbox (longer timeout, more filesystem access)

**Community:**
- Has manifest but not verified
- Permissions declared in manifest
- Strict sandbox (short timeout, limited access)

**Untrusted:**
- No manifest
- Receives only session ID
- Cannot access screen content or environment
- Strict sandbox with minimal permissions

### Verification Process

Verified plugins undergo reproducible build verification:

1. Clone source at specified commit
2. Build with exact flags from manifest
3. Compare SHA256 of built binary with manifest
4. If match: trust level = Verified
5. If mismatch: installation aborted

---

## Naming Conventions

### Plugin Names

```
it2-<scope>-<name>
it2-<scope>-<context>-<name>
```

**Components:**
- `it2-` — Required prefix for discovery
- `<scope>` — Target scope: `session`, `tab`, `window`, `session-process`
- `<context>` — Optional context: `claude`, `git`, `docker`, etc.
- `<name>` — Descriptive action: `has-modal`, `is-at-prompt`, `suggest-action`

**Examples:**
```
it2-session-is-at-prompt           # Session scope, general purpose
it2-session-claude-has-modal       # Session scope, Claude-specific
it2-session-claude-auto-approve    # Session scope, Claude automation
it2-tab-git-status                 # Tab scope, git-specific
it2-session-process-is-claude      # Process scope, Claude detection
```

### Hook Names

```
it2-<source>-hook
```

**Examples:**
```
it2-claude-code-hook    # Hook for Claude Code events
it2-gemini-hook         # Hook for Gemini CLI events
```

---

## CLI Commands

### `it2 plugin` — Run and manage plugins

The `it2 plugin` command provides access to discovered plugins:

```bash
# List available plugins
it2 plugin list

# Run a plugin directly via it2
it2 plugin <name> [args...]

# Install Claude Code hooks
it2 plugin claude-code-hook --install
```

### Running plugins directly

Plugins can also be invoked directly from the command line (without going through `it2 plugin`):

```bash
# Run directly (if in PATH)
it2-session-claude-has-modal <session-id>

# Equivalent via it2
it2 plugin claude-has-modal <session-id>
```

The `it2 plugin` wrapper provides:
- Automatic session ID injection from `$ITERM_SESSION_ID`
- Plugin discovery information
- Consistent error handling

### Note on `it2 tool` (renamed)

The previous `it2 tool` command has been renamed to `it2 plugin` for consistency with the taxonomy. The old command is deprecated and will warn once per day:

```bash
# Deprecated (warns once per day)
it2 tool list

# Use instead
it2 plugin list
```

---

## Not in Scope

### MCP (Model Context Protocol)

MCP servers are **not** part of it2's plugin taxonomy. They are a separate integration:

| Aspect | Plugin | MCP Server |
|--------|--------|------------|
| Discovery | `it2-*` in PATH | Claude Code settings |
| Protocol | stdin/stdout (args or JSON) | MCP protocol |
| Purpose | Extend it2 | Extend Claude capabilities |
| Lifecycle | it2-managed | Claude-managed |

it2 may provide hooks that facilitate MCP integration, but MCP servers themselves are external.

### iTerm2 Python API Scripts

iTerm2's native Python API scripts are separate from it2 plugins. They run within iTerm2's Python environment and use a different interface.

---

## Migration Guide

### Terminology Changes

| Old Term | New Term | Notes |
|----------|----------|-------|
| Extension | Plugin | "Extension" was used inconsistently |
| `it2 tool` (command) | `it2 plugin` | Command renamed to match taxonomy |
| `it2-extensions/` | `plugins/` | Directory rename (future) |

**Note:** "Tool" remains a valid concept for standalone CLI programs (like `it2` itself). Only the `it2 tool` *command* is renamed to `it2 plugin` because it operates on plugins.

### Command Migration

```bash
# Old (deprecated, warns once per day)
it2 tool list
it2 tool claude-code-hook --install

# New
it2 plugin list
it2 plugin claude-code-hook --install
```

### Deprecation Timeline

- **Current version:** `it2 tool` works with deprecation warning
- **Next major version:** `it2 tool` removed

---

## Quick Reference

| Term | Definition | Invocation | Lifecycle |
|------|------------|------------|-----------|
| Tool | Standalone CLI program | Direct by user | User-managed |
| Plugin | Extends it2, auto-discovered | By it2 (or direct) | it2-managed |
| Hook | Event handler for external tools | By external tool | External-managed |
| Enrichment | Plugin that adds data to listings | By capability | Per-request |
| Automation | Plugin that suggests/executes actions | By capability | Per-request |
| Core | Trusted plugin embedded in binary | Built-in | Always available |
| Verified | Plugin with reproducible build | PATH + manifest | Verified at install |
| Community | Plugin with manifest, not verified | PATH + manifest | Trust on use |
| Untrusted | Plugin without manifest | PATH only | Minimal permissions |

---

## Related Documentation

- [Plugin Development Guide](PLUGIN_DEVELOPMENT.md) — How to write plugins
- [Plugin Manifest Schema](plugin-manifest-schema.md) — Manifest format reference
- [Claude Plugins Reference](CLAUDE_PLUGINS_REFERENCE.md) — Claude-specific plugins
