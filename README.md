# it2

[![Go Reference](https://pkg.go.dev/badge/github.com/tmc/it2.svg)](https://pkg.go.dev/github.com/tmc/it2)
[![Go Report Card](https://goreportcard.com/badge/github.com/tmc/it2)](https://goreportcard.com/report/github.com/tmc/it2)
[![License: MIT](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)

Programmatic control of iTerm2 — sessions, splits, text I/O, buffer reading,
state detection, and real-time events over the Unix socket API.

it2 turns iTerm2 into an orchestration layer for terminal workflows. Split
panes, send commands, read output, detect state — from shell scripts, CI
pipelines, or AI agents like Claude Code and Gemini CLI.

## Install

Requires Go 1.21+ and iTerm2 3.3.0+.

```bash
go install github.com/tmc/it2/cmd/it2@latest
it2 auth check   # iTerm2 will prompt to allow API access
```

## Quick Start

```bash
it2 session list                          # list all sessions
it2 session split -q                      # split pane, print new session ID
it2 session send-text $SID "make test"    # send text to a session
it2 session get-screen $SID               # read visible screen content
it2 session get-buffer $SID --lines 100   # read scrollback buffer
```

## Agent Integration

it2 detects and interacts with AI agents running in terminal sessions.
Built-in support for Claude Code; extensible to any agent via plugins.

### State Detection

Check what a session is doing without reading the full buffer:

```bash
it2 session is-active $SID                # exit 0 if agent is working
it2 session has-modal $SID                # exit 0 if modal dialog detected
it2 session suggest-action $SID           # recommend: continue, approve, wait
it2 session claude-status $SID            # human-readable Claude Code status
it2 session get-state $SID --agent=auto   # full state as structured output
```

### Auto-Respond

Watch a session and automatically respond to prompts:

```bash
it2 session autorespond $SID                     # default patterns
it2 session autorespond $SID --pattern "Continue?"  # custom pattern
```

The embedded `claude-auto-approve` plugin checks safety heuristics before
approving any modal. See [Claude Plugins Reference](docs/CLAUDE_PLUGINS_REFERENCE.md)
for hook setup and event logging.

## Multi-Agent Workflows

it2 supports multi-agent patterns where sessions split, communicate, and
read each other's output.

### Split and Delegate

```bash
# Create a worker session and send it a task
WORKER=$(it2 session split -q --vertical)
it2 session set-badge "$WORKER" "$(echo $WORKER | cut -c1-8)\nWorker"
it2 session send-text "$WORKER" "cd /project && make lint"
```

### Cross-Session Communication

```bash
# Read a session's output
it2 session get-buffer "$WORKER" --lines 50

# Wait for prompt, then send next command
it2 session send-text --require is-at-prompt "$WORKER" "make test"

# Skip delivery confirmation for speed
it2 session send-text --skip-confirm "$WORKER" "echo done"
```

### Prime Protocol

Output inter-session communication guidelines for an agent:

```bash
it2 prime | it2 session send-text $SID -f -
```

This sends a protocol document that teaches agents how to respond via
`it2 session send-text`, include session IDs, and format messages.

### Broadcast

Send the same input to multiple sessions simultaneously:

```bash
it2 broadcast set $SID1 $SID2 $SID3   # create broadcast domain
it2 broadcast send "git pull"           # send to all sessions in domain
it2 broadcast clear                     # tear down broadcast domain
```

## Plugin System

Plugins are executables on your PATH matching `it2-session-*`, `it2-tab-*`,
or `it2-window-*`. They receive session context as JSON on stdin and return
enrichment data on stdout.

```bash
it2 plugin list           # show discovered plugins with source and hash
it2 plugin run <name>     # run a plugin directly
```

Embedded plugins:

| Plugin | Type | Purpose |
|--------|------|---------|
| `is-at-prompt` | session | Detect shell prompt (exit 0 = at prompt) |
| `claude-has-modal` | session | Detect Claude Code modal dialogs |
| `claude-suggest-action` | session | Recommend intervention for stuck sessions |
| `claude-auto-approve` | session | Auto-approve safe operations |
| `claude-is-safe-operation` | session | Check if an operation is safe to auto-approve |
| `has-no-queued-claude-messages` | session | Check for pending message queue |
| `is-claude.sh` | session-process | Detect if session runs Claude Code |
| `tree.sh` | session-process | Show process hierarchy for a session |

PATH plugins shadow embedded plugins of the same name.
See [docs/TAXONOMY.md](docs/TAXONOMY.md) for the full plugin model.

## Configuration

### Environment Variables

| Variable | Description |
|----------|-------------|
| `ITERM_SESSION_ID` | Current session ID (set by iTerm2) |
| `ITERM2_COOKIE` | Auth cookie (auto-requested if not set) |
| `ITERM2_KEY` | Auth key (auto-requested if not set) |
| `IT2_PLUGIN_PATHS` | Additional plugin search directories |
| `IT2_PLUGIN_DEADLINE` | Plugin execution timeout (default 5s) |
| `ITERM2_DEBUG` | Enable debug logging (`1`) |

### Global Flags

```
--format string      Output format: table, text, json, yaml (default "table")
--timeout duration   Connection timeout (default 5s)
```

### Session IDs

Sessions accept UUIDs, short prefixes, or full `w0t1p12:UUID` format.
When omitted, most commands use `$ITERM_SESSION_ID` (the current session).

## Links

- [Go Reference](https://pkg.go.dev/github.com/tmc/it2)
- [iTerm2 API Documentation](https://iterm2.com/documentation-api.html)
- [Shell Integration Guide](https://iterm2.com/documentation-shell-integration.html)
- [Claude Plugins Reference](docs/CLAUDE_PLUGINS_REFERENCE.md)

Run `it2 help [command]` for detailed usage on any command.
