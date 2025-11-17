# Extensibility Architecture

## Problem

Confusion between "plugins", "extensions", and "tools" led to proliferation of `cmd/*` binaries.

## Solution: Clear Taxonomy

### Plugin

**Definition:** External binary that enriches it2 session/tab/window data.

**Characteristics:**
- Naming convention: `it2-session-*`, `it2-tab-*`, `it2-window-*`
- Location: Anywhere in `$PATH`
- Discovery: Scanned at runtime, must be explicitly enabled
- Trust: `LOCAL` or `COMMUNITY` (requires verification)
- Sandboxing: Always sandboxed (Seatbelt/Landlock)
- Manifest: Required (`manifest.json` declares permissions)

**Examples:**
```bash
# Community plugin (signed)
it2-session-git      # Adds git branch to session badge
it2-session-docker   # Shows container count
it2-tab-tmux         # Enriches tab with tmux pane info

# Local plugin (user script)
~/bin/it2-session-custom  # User's own enrichment
```

**How to use:**
```bash
# Discover available plugins
$ it2 plugin list

# Enable plugin
$ it2 plugin enable it2-session-git

# Enriched output
$ it2 session list
ABC123  main ✓  /workspace/it2  (git:next, docker:3)
```

**Distribution:**
- Separate repositories
- User's local ~/bin
- Package managers (brew, apt, etc.)

### Tool

**Definition:** Built-in utility exposed via `it2 tool <name>` subcommand.

**Characteristics:**
- Part of main `it2` binary (embedded)
- No external dependencies
- Trust: `FIRST_PARTY` (highest, auto-approved)
- No sandboxing (trusted code)
- No manifest needed
- Can be used standalone or as hooks

**Examples:**
```bash
# List available tools
$ it2 tool list
claude-code-hook   Hook for Claude Code events
session-snapshot   Capture session state

# Use directly
$ it2 tool claude-code-hook --install

# Use in hook
# .claude/hooks/user-prompt-submit.sh:
#!/bin/bash
it2 tool claude-code-hook "$@"
```

**Implementation:**
```go
// internal/cmd/tool/tool.go
var tools = map[string]Tool{
    "claude-code-hook": &ClaudeCodeHook{},
    "session-snapshot": &SessionSnapshot{},
}

// Exposed as: it2 tool <name> [args]
```

**Distribution:**
- Compiled into `it2` binary
- No separate installation
- Updated with `it2` releases

### Extension (Deprecated)

**Status:** Don't use this term. Use "Plugin" or "Tool" instead.

## Repository Organization

### What Goes in `cmd/`

**ONLY** the main binary:

```
cmd/
└── it2/           # Main binary
    └── main.go
```

Everything else removed from `cmd/`.

### Where Things Go Instead

**Built-in utilities** → Embedded in main binary via `it2 tool`:

```
internal/cmd/tool/
├── tool.go                  # Tool registry
├── claude_code_hook.go      # it2 tool claude-code-hook
└── session_snapshot.go      # it2 tool session-snapshot
```

**External plugins** → Separate repositories or user's PATH:

```
# Example plugin repos (separate from it2)
github.com/tmc/it2-session-git/
github.com/tmc/it2-session-docker/
~/bin/it2-session-custom

# NOT in it2 repo cmd/
```

**Example plugins** → `examples/` directory:

```
examples/plugins/
├── it2-session-simple/      # Minimal example
│   ├── main.go
│   └── manifest.json
└── it2-session-advanced/    # Full-featured example
    ├── main.go
    └── manifest.json
```

## Migration Plan

### Current State (Wrong)

```
cmd/
├── it2/                     ✓ Keep
├── it2-claude-code-hook/    ✗ Move to internal/cmd/tool/
├── it2-session-badge/       ✗ Move to examples/ or separate repo
├── it2-session-snapshot/    ✗ Move to internal/cmd/tool/
├── it2-session-snapshot-claude/      ✗ Move to examples/
└── it2-session-snapshot-filesystem/  ✗ Move to examples/
```

### Target State (Correct)

```
cmd/
└── it2/                     # Main binary only

internal/cmd/tool/
├── tool.go                  # Tool registry
├── claude_code_hook.go      # Built-in: it2 tool claude-code-hook
└── session_snapshot.go      # Built-in: it2 tool session-snapshot

examples/plugins/
├── it2-session-git/         # Example: git enrichment
├── it2-session-badge/       # Example: custom badges
└── it2-session-snapshot-*/  # Example: snapshot plugins
```

## Decision Tree

**When creating new functionality, ask:**

```
Is this core to it2 operation?
├─ YES → Add to main binary (cmd/it2/ or internal/)
└─ NO → Is it needed by most users?
       ├─ YES → Built-in tool (internal/cmd/tool/, exposed via it2 tool)
       └─ NO → External plugin (separate repo or examples/)
```

**Examples:**

| Functionality | Category | Reasoning |
|--------------|----------|-----------|
| `it2 session list` | Core | Fundamental operation |
| `it2 tool claude-code-hook` | Tool | Useful for integration, low complexity |
| `it2-session-git` | Plugin | Optional enrichment, external dependency (git) |
| `it2-session-docker` | Plugin | Optional enrichment, external dependency (docker) |
| Session snapshot orchestrator | Tool | Useful utility, no external deps |
| Individual snapshot plugins | Plugin | Optional, demonstrate plugin API |

## Benefits

### Clear Separation

- **Core** (`cmd/it2/`): Essential functionality
- **Tools** (`internal/cmd/tool/`): Built-in utilities, no installation
- **Plugins** (external): Optional enrichment, user/community maintained

### Simplified Distribution

```bash
# Install core
$ brew install it2

# Tools already available
$ it2 tool list

# Install plugins as needed
$ go install github.com/tmc/it2-session-git@latest
$ it2 plugin enable it2-session-git
```

### Reduced Maintenance

- Core team maintains: `it2` binary + built-in tools
- Community maintains: External plugins
- Clear boundary for contribution guidelines

## Implementation Notes

### Embedding Tools

```go
// internal/cmd/tool/tool.go
package tool

type Tool interface {
    Name() string
    Description() string
    Run(args []string) error
}

var registry = make(map[string]Tool)

func Register(t Tool) {
    registry[t.Name()] = t
}

func Get(name string) (Tool, bool) {
    t, ok := registry[name]
    return t, ok
}

// In cmd/it2/main.go:
import _ "github.com/tmc/it2/internal/cmd/tool" // Register all tools
```

### Plugin Discovery

```go
// internal/plugins/discovery.go
func DiscoverPlugins() []Plugin {
    var plugins []Plugin

    // Scan PATH for it2-session-*, it2-tab-*, it2-window-*
    pathDirs := filepath.SplitList(os.Getenv("PATH"))
    for _, dir := range pathDirs {
        entries, _ := os.ReadDir(dir)
        for _, e := range entries {
            if matchesPattern(e.Name()) {
                plugins = append(plugins, Plugin{
                    Name: e.Name(),
                    Path: filepath.Join(dir, e.Name()),
                })
            }
        }
    }

    return plugins
}
```

## FAQ

**Q: Can a tool become a plugin later?**

A: Yes. Start as tool for convenience, move to plugin when it gains external dependencies or becomes optional.

**Q: Can a plugin access other plugins?**

A: Not directly. Plugins are isolated. For composition, use the orchestrator pattern (like `it2 tool session-snapshot`).

**Q: Should CLIs distributed with it2 (like it2-claude-code-hook) be tools or plugins?**

A: If they're lightweight and useful to many users → Tool. If they have external dependencies or are optional → Plugin.

**Q: What about platform-specific functionality?**

A: Tools can have platform-specific code (build tags). Plugins are naturally platform-specific (distributed separately).

## References

- Plugin Security: `plugin-security.md`
- Plugin Architecture: `plugin-architecture.md`
- Plugin Execution: `plugin-execution.md`
