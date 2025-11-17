# Software Attribution System

## Problem

We track WHO wrote code (git commits) but not HOW it was produced:
- Which AI assisted? (Claude, GPT, Gemini)
- Which tools ran? (linters, formatters, generators)
- What was the workflow? (commands, edits, timing)

Can't answer: "How was this code actually created?"

## Vision

Full provenance chain from idea → artifact.

Capture:
- Human inputs (commands, prompts)
- AI contributions (which model, what it generated)
- Tool outputs (plugins, scripts, external programs)
- Session context (files open, cursor position, timing)

Verify:
- Cryptographically prove "this artifact came from this workflow"
- Attribute specific lines to specific agents/tools
- Replay session to see how code evolved

## Use Cases

### 1. Code Review

```bash
$ git show abc123 --attribution
commit abc123
Author: Alice <alice@example.com>
Date: 2025-11-17

Add user authentication

Attribution:
  auth.go lines 1-45: Claude Code (model: sonnet-4.5)
  auth.go lines 46-78: Alice (manual edit)
  auth_test.go: it2-plugin-testgen v1.2.3
```

### 2. Audit Trail

```bash
$ it2 session export abc123 --format attribution-log
Session: abc123 (2025-11-17 14:32:15 - 14:47:03)
Working dir: /proj
Files modified: auth.go, auth_test.go

Timeline:
14:32:18 Alice: $ claude
14:32:23 Claude: Read auth.go (243 lines)
14:33:45 Claude: Write auth.go (added lines 1-45)
14:34:02 Alice: Edited auth.go line 47
14:35:11 Alice: $ it2-plugin-testgen auth.go
14:35:18 it2-plugin-testgen: Write auth_test.go (78 lines)
```

### 3. Reproducibility

```bash
$ it2 session replay abc123 --verify
Replaying session abc123...
✓ Claude Code v2.0.42 available
✓ it2-plugin-testgen v1.2.3 (hash: def456)
✓ Files match recorded state

Generated auth.go: hash matches ✓
Generated auth_test.go: hash matches ✓
```

## Design

### Event Stream

Every action produces event:

```go
type Event struct {
    Timestamp   time.Time
    SessionID   string
    Actor       Actor      // Human, AI, Plugin, Tool
    Action      Action     // Read, Write, Execute, Prompt
    Target      string     // File path, command, etc.
    Content     []byte     // What changed
    Signature   []byte     // Cryptographic signature
}

type Actor struct {
    Type    ActorType  // Human, AI, Plugin
    ID      string     // User ID, model name, plugin name
    Version string     // For AI/plugins
}
```

Events written to append-only log: `~/.it2/sessions/{id}/events.log`

### Attribution Graph

Artifacts link back to events:

```
auth.go:1-45 → Event{Claude Code, Write, ...}
auth.go:46-78 → Event{Alice, Write, ...}
auth_test.go:1-78 → Event{it2-plugin-testgen, Write, ...}
```

Stored as git notes:
```bash
$ git notes --ref=attribution show abc123
{
  "files": {
    "auth.go": {
      "1-45": {"actor": "claude-code", "session": "abc123", "event": 42},
      "46-78": {"actor": "alice", "session": "abc123", "event": 58}
    }
  }
}
```

### Session DVR

Record terminal for replay:

```
~/.it2/sessions/{id}/
├── events.log          # Event stream
├── screen.recording    # Screen captures
├── timing.log          # Timing data
└── metadata.json       # Session metadata
```

Replay engine reconstructs session:
- Parse events
- Replay commands with timing
- Show file diffs inline
- Highlight AI vs human contributions

### Security

Sensitive data (credentials, tokens) must be redacted:

```go
type Redactor interface {
    Redact(content []byte) []byte
}

// Redact common patterns
- Environment variables matching *_TOKEN, *_KEY, *_PASSWORD
- Command arguments to ssh, curl -u, etc.
- File contents matching secret patterns
```

Redacted events still recorded but content replaced with `[REDACTED]`.

### Verification

Events signed with user's key:
```go
signature = Sign(SHA256(event), userPrivateKey)
```

On replay:
```bash
$ it2 session verify abc123
✓ Event 1: signature valid (alice@example.com)
✓ Event 2: signature valid (claude-code@anthropic.com)
...
✓ All 127 events verified
```

Tampered events detected:
```bash
✗ Event 42: signature invalid (content modified)
```

## Implementation

Phase 1: Event capture
```
internal/attribution/
├── event.go       # Event types
├── logger.go      # Append-only log
├── capture.go     # Hook into it2 commands
└── redactor.go    # Sensitive data redaction
```

Phase 2: Attribution graph
```
internal/attribution/
├── graph.go       # Build attribution graph
├── git.go         # Git notes integration
└── query.go       # Query attribution
```

Phase 3: Session DVR
```
internal/dvr/
├── recorder.go    # Record sessions
├── player.go      # Replay sessions
└── export.go      # Export formats
```

## Open Questions

1. Storage: Event logs grow large. Rotation strategy? Compression?
2. Privacy: How to share attributed code without leaking workflow?
3. Performance: Event logging overhead in tight loops?
4. Git integration: git notes or custom git trailers or separate .attribution files?
5. Cross-session attribution: Code copied from session A to session B?
