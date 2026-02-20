# Codebase Audit: main Branch vs Design Documents

Audit of the it2 codebase on `main` (commit 24c14b61) against the
architecture design documents on `next-planning`.

## P0: Critical Security Issues

### 1. Plugin Environment Leak

**File**: `internal/plugins/plugin.go:144-155`

```go
func setupPluginEnv(cmd *exec.Cmd) {
    cmd.Env = os.Environ()  // LEAKS ENTIRE ENVIRONMENT TO PLUGINS
    if cookie := os.Getenv("ITERM2_COOKIE"); cookie != "" {
        cmd.Env = append(cmd.Env, "ITERM2_COOKIE="+cookie)
    }
    if key := os.Getenv("ITERM2_KEY"); key != "" {
        cmd.Env = append(cmd.Env, "ITERM2_KEY="+key)
    }
}
```

Every plugin (discovered via PATH) receives the full parent environment:
AWS credentials, API tokens, SSH agent sockets, cloud credentials,
personal data. The SECURITY_ARCHITECTURE.md design identifies this as
the #1 finding. SECURE_UNIX_PLUGIN_ARCHITECTURE.md prescribes a
capability-based environment whitelist.

Called from: `EnrichSession()` (line 186), `EnrichTab()` (line 234),
`EnrichWindow()` (line 281), `EnrichProcess()` (line 323),
`StartMonitoring()` (line 356).

**Fix**: Replace `os.Environ()` with a whitelist: `HOME`, `PATH`,
`SHELL`, `TERM`, `USER`, `LANG`, `ITERM2_COOKIE`, `ITERM2_KEY`,
`IT2_*` prefix.

### 2. World-Readable Sensitive Files

All these locations use 0644 (world-readable) instead of 0600:

| Location | Line | Contents |
|----------|------|----------|
| `internal/plugins/events.go` | 46 | `~/.it2/plugin-events.ndjson` |
| `internal/plugins/events.go` | 97 | Session plugin events |
| `internal/plugins/events.go` | 114 | Overflow event files |
| `internal/config/config.go` | 89 | `~/.it2/config.json` |
| `internal/hints/hints.go` | 83 | `~/.it2/hints.json` |
| `internal/cmd/session/session_tag.go` | 187 | Session tags |
| `internal/projects/version.go` | 113 | Version files |
| `internal/cmd/artifact/artifact_add.go` | 201 | Artifact content |

Directory permissions:

| Location | Line | Perm | Should Be |
|----------|------|------|-----------|
| `internal/cmd/session/session_artifacts.go` | 129 | 0755 | 0700 |
| `internal/plugins/events.go` | 41 | 0755 | 0700 |

These files may contain session content, command history, plugin
execution data, and terminal output. On multi-user systems, any user
can read them.

## P1: Dead/Orphaned Code

### 3. `internal/projects/` Package (0 importers)

Files: `projects.go`, `version.go`, `projects_test.go`

Provides `ProjectDir()`, `SessionsIndex`, `VersionInfo`, etc. for
per-project session management. Zero files import this package.
Completely dead code that was merged but never wired up.

The `ProjectNameToPath()` function has a bug: it replaces ALL dashes
with path separators, so a project at `/my-app/src` would round-trip
incorrectly through `PathToProjectName` -> `ProjectNameToPath`.

### 4. Placeholder `event log` Command

**File**: `internal/cmd/event/event.go:210-254`

The `it2 event log` subcommand:
- Returns only "Event logging is not yet implemented"
- Has `TODO: Implement actual event logging` comment
- Takes a `--limit` flag that is never used
- Is hidden from help

The EVENT_SOURCING.md and SESSION_EVENT_JOURNALS.md designs completely
supersede this placeholder. It should be deleted.

### 5. Seven Hidden Experimental Commands

**File**: `internal/cmd/session/session.go:268-307`

All marked `Hidden = true` with comment "experimental, hidden":

| Command | Purpose | Status |
|---------|---------|--------|
| `session get-state` | Detect Claude Code state | Has implementation |
| `session is-active` | Check if Claude is active | Has implementation |
| `session has-modal` | Detect modal dialogs | Has implementation |
| `session suggest-action` | AI action suggestions | Has implementation |
| `session claude-status` | Rich Claude status | Has implementation |
| `session export-recording` | DVR export | Has implementation |
| `session claude` | Claude session linkage | Has implementation |

These are feature-branch code that landed on main without a decision to
either promote them to visible commands or remove them. The
SESSION_KNOWLEDGE_MINING.md design covers session state detection
properly.

## P2: Architecture Conflicts

### 6. Fragmented Event Storage (4 parallel paths)

The SESSION_EVENT_JOURNALS.md design unifies these into a single
`events.jsonl` per session. Currently there are 4 independent paths:

1. `~/.it2/plugin-events.ndjson` — global plugin events
2. `~/.it2/sessions/<id>/artifacts/plugin-events.ndjson` — per-session
3. `~/.it2/sessions/<id>/artifacts/.overflow/<pid>.ndjson` — overflow
4. `~/.it2/sessions/<id>/claude-code-events.ndjson` — hook events

Each has a different schema. No unified event envelope. No correlation
IDs. No cross-session discovery. The `RecordPluginEvent()` function in
`internal/plugins/events.go:51-86` dual-writes to paths 1 and 2.

The overflow mechanism (flock with 10ms timeout, spill to PID-named
file) is clever but creates fragmentation that requires periodic
`MergeOverflowFiles()` calls.

### 7. Prime Protocol vs Event Journals

**File**: `internal/prime/prime.go` + `prime_message.md`

The `it2 prime` command embeds a markdown inter-session communication
protocol. Only imported by `cmd/it2/main.go`. The content overlaps
significantly with the user's CLAUDE.md inter-session communication
section.

The SESSION_EVENT_JOURNALS.md design introduces `xref.*` events with
correlation IDs for structured cross-session communication. This is a
fundamentally different (and better) approach than the text-based
priming in prime_message.md.

Question: Should `prime` remain as a convenience for human-readable
protocol hints, or is it superseded by the journal-based approach?

### 8. cmdutil Framework Inconsistency (616 lines)

**Files**: `internal/cmdutil/standard.go` (203 lines),
`shared_operations.go` (294 lines), `templates.go` (119 lines)

The `StandardCommand`/`CommandTemplate` pattern is used by ~70 command
files. But many session commands use cobra directly. The design
documents don't reference this framework. Two concerns:

- Inconsistency: Some commands use the framework, some don't
- The framework adds indirection that makes commands harder to follow

Not blocking, but worth deciding: commit to the framework everywhere
or simplify back to direct cobra usage.

## P3: Stale Branches and Worktrees

35+ branches, 10 git worktrees. Several are superseded by the designs:

| Branch | Status | Superseded By |
|--------|--------|---------------|
| `feature/events-system` | Stale | EVENT_SOURCING.md |
| `feature/session-crosstalk` | Stale | SESSION_EVENT_JOURNALS.md |
| `feature/plugin-correlation` | Stale | SECURE_UNIX_PLUGIN_ARCHITECTURE.md |
| `feature/experimental-work` | Stale | Various designs |
| `next-next` | JSON-stdin protocol | Main already has this |
| `next` | 58 ahead, 1 behind main | Redundant |
| `omain` / `omain-1` | Duplicates | Typos? |

The `main` branch itself has 78 unpushed commits to GitHub.

## Questions for Reviewer

1. Should the P0 security fixes go directly onto `main` now, or wait
   for the Phase 0 implementation from the roadmap?

2. The `projects` package (P1 #3) — delete entirely, or is there a
   plan to wire it up that I'm missing?

3. The 7 hidden experimental commands (P1 #5) — several of these
   (`get-state`, `is-active`, `claude-status`) are actually useful for
   the knowledge mining design. Should they be kept but moved to a
   proper experimental namespace, or deleted and reimplemented later?

4. The `prime` protocol (P2 #7) — the event journal approach is
   superior for machine-to-machine communication. But `prime` serves
   a human-readable purpose (priming Claude sessions with protocol
   knowledge). Keep both? Merge into one?

5. The cmdutil framework (P2 #8) — 616 lines of abstraction used by
   70 files. Too embedded to remove easily. Standardize everything
   through it, or gradually simplify?

6. Branch cleanup — safe to delete `omain`, `omain-1`, `next-next`,
   and the superseded feature branches?
