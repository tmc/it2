# it2 Extensibility Taxonomy

## Overview

This document defines a clear taxonomy for it2's extensibility mechanisms. The goal is to eliminate terminology confusion and provide consistent naming throughout the codebase.

## Current State Analysis

The codebase currently uses three terms inconsistently:

| Term | Current Usage | Files/Locations |
|------|---------------|-----------------|
| **Plugin** | External executables (`it2-*`) that enrich sessions/tabs/windows | `internal/plugins/`, `docs/PLUGIN_DEVELOPMENT.md`, `docs/plugin-manifest-schema.md` |
| **Extension** | Directory of standalone plugins (`it2-extensions/`), also used in `CLAUDE_EXTENSIONS_REFERENCE.md` | `it2-extensions/`, `docs/CLAUDE_EXTENSIONS_REFERENCE.md` |
| **Tool** | The `it2 tool` command that runs plugins directly | `internal/cmd/tool/` |

### Key Observations

1. **"Plugin"** is the most-used term in Go code (`internal/plugins/`)
2. **"Extension"** appears in docs referring to the same thing as plugins
3. **"Tool"** is a command wrapper that exposes plugins as subcommands
4. The manifest system (`manifest.go`) uses "plugin" terminology
5. `it2-extensions/` directory contains what the code calls "plugins"

---

## Proposed Taxonomy

### 1. **Plugin** (Primary Term)

**Definition**: An external executable that extends it2 functionality by receiving structured input and producing structured output.

**Characteristics**:
- Executable files with `it2-` prefix
- Discovered from `PATH`, `IT2_PLUGIN_PATHS`, or embedded
- Sandboxed with manifest-defined permissions
- Support verification via reproducible builds

**Subtypes** (by naming convention):
- `it2-session-*` → Session Enricher
- `it2-tab-*` → Tab Enricher
- `it2-window-*` → Window Enricher
- `it2-session-process-*` → Process Enricher

**Examples**:
- `it2-session-claude-has-modal`
- `it2-session-is-at-prompt`
- `it2-tab-git-status`

### 2. **Hook** (Event Handler)

**Definition**: A plugin or command that runs in response to events from external tools (Claude Code, etc).

**Characteristics**:
- Registered via external tool configuration (e.g., Claude Code's `settings.json`)
- Receives event data via stdin (JSON)
- Used for logging, automation, or side effects
- Not discovered by it2—configured externally

**Examples**:
- `it2-claude-code-hook` (logs Claude Code events to session artifacts)
- Custom notification scripts

**Key Distinction**: Hooks are plugins by implementation, but their lifecycle is managed by external tools, not it2.

### 3. **Tool** (Command Wrapper)

**Definition**: An it2 subcommand (`it2 tool <name>`) that provides direct access to plugins.

**Characteristics**:
- Wraps plugin execution with CLI ergonomics
- Provides `--install` and configuration options
- Passes through arguments to underlying plugin

**Note**: "Tool" is not a separate extensibility concept—it's how plugins are exposed via the CLI.

---

## Removed/Deprecated Terms

### **Extension** (DEPRECATED)

**Action**: Rename to "Plugin" throughout.

**Rationale**:
- "Extension" and "Plugin" mean the same thing in this codebase
- "Plugin" is already used in Go code (`internal/plugins/`)
- "Extension" adds confusion without adding clarity

**Migration**:
1. Rename `it2-extensions/` → `it2-plugins/` or merge into `internal/plugins/scripts/`
2. Rename `docs/CLAUDE_EXTENSIONS_REFERENCE.md` → `docs/CLAUDE_PLUGINS_REFERENCE.md`
3. Update all documentation to use "plugin" consistently

---

## Complete Taxonomy Table

| Term | Type | Description | Discovery | Example |
|------|------|-------------|-----------|---------|
| **Plugin** | Enricher | Adds data to session/tab/window listings | Auto-discovered from PATH | `it2-session-is-claude-code` |
| **Plugin** | Monitor | Generates events during monitoring | Auto-discovered from PATH | `it2-session-claude-suggest-action` |
| **Hook** | Event Handler | Responds to external tool events | Configured externally | `it2-claude-code-hook` |
| **Tool** | CLI Command | Exposes plugins via `it2 tool` | N/A (wrapper) | `it2 tool claude-code-hook` |

---

## Integration Points

### MCP (Model Context Protocol)

MCP servers are **not** plugins. They are a separate integration mechanism:

| Aspect | Plugin | MCP Server |
|--------|--------|------------|
| Discovery | `it2-*` in PATH | Claude Code settings |
| Protocol | stdin/stdout JSON | MCP protocol |
| Purpose | Enrich it2 data | Extend Claude capabilities |
| Lifecycle | it2-managed | Claude-managed |

**Recommendation**: Don't try to unify MCP and plugins. They serve different purposes.

### Status Line Enrichers

Status line enrichers are a **use case** of plugins, not a separate type:

```
Plugin (it2-session-*) → EnrichSession() → PluginData → Status Line Display
```

### Session Monitors

Session monitors are plugins that implement the `EventMonitor` interface:

```go
type EventMonitor interface {
    StartMonitoring(ctx, sessionID) (<-chan PluginEvent, <-chan error, error)
    StopMonitoring() error
}
```

---

## Directory Structure (Proposed)

```
it2/
├── internal/
│   └── plugins/              # Plugin system core
│       ├── discovery.go      # Plugin discovery
│       ├── enrichment.go     # Enrichment execution
│       ├── events.go         # Event logging
│       ├── manifest.go       # Manifest schema
│       ├── plugin.go         # Plugin interfaces
│       └── scripts/          # Built-in plugins (embedded)
│           ├── it2-session-claude-has-modal
│           ├── it2-session-is-at-prompt
│           └── ...
├── plugins/                  # External plugins (formerly it2-extensions/)
│   ├── it2-session-is-claude-code/
│   ├── it2-session-is-node/
│   └── ...
└── docs/
    ├── PLUGIN_DEVELOPMENT.md           # How to write plugins
    ├── PLUGIN_MANIFEST_SCHEMA.md       # Manifest reference
    └── CLAUDE_PLUGINS_REFERENCE.md     # Claude-specific plugins
```

---

## Naming Conventions

### Plugin Names

```
it2-<scope>-<name>
it2-<scope>-<context>-<name>
```

**Scopes**: `session`, `tab`, `window`, `session-process`

**Examples**:
- `it2-session-is-at-prompt` (session scope, general purpose)
- `it2-session-claude-has-modal` (session scope, Claude-specific)
- `it2-tab-git-status` (tab scope, git-specific)

### Hook Names

```
it2-<source>-hook
```

**Examples**:
- `it2-claude-code-hook`
- `it2-gemini-hook` (future)

---

## Migration Checklist

### Phase 1: Documentation
- [ ] Rename `docs/CLAUDE_EXTENSIONS_REFERENCE.md` → `docs/CLAUDE_PLUGINS_REFERENCE.md`
- [ ] Update all doc references from "extension" to "plugin"
- [ ] Add this taxonomy doc to `docs/`

### Phase 2: Directory Structure
- [ ] Rename `it2-extensions/` → `plugins/` (external plugins)
- [ ] Update README.md references

### Phase 3: Code Comments
- [ ] Grep for "extension" in code comments
- [ ] Update to use "plugin" consistently

### Phase 4: User-Facing
- [ ] Update `it2 plugins list` output
- [ ] Update help text in `it2 tool --help`

---

## Summary

| OLD | NEW | Notes |
|-----|-----|-------|
| Extension | Plugin | Same thing, standardize on "plugin" |
| Tool | Tool | Keep as CLI wrapper concept |
| Hook | Hook | Keep as external event handler concept |

**Single Source of Truth**: `internal/plugins/` is the authoritative implementation. All terminology should align with the interfaces defined there:
- `SessionEnricher`, `TabEnricher`, `WindowEnricher` → Plugin subtypes
- `EventMonitor` → Monitoring plugin
- `Manifest` → Plugin manifest
