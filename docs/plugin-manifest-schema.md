# Plugin Manifest Schema

## Overview

Plugin manifest defines permissions, verification data, and sandbox constraints for it2 plugins.

> **See also:** [TAXONOMY.md](TAXONOMY.md) for plugin terminology and trust levels.

## Schema v1

```json
{
  "schema_version": "1",
  "name": "claude-status",
  "version": "1.0.0",
  "description": "Shows Claude Code session work status",
  "author": "tmc",
  "homepage": "https://github.com/tmc/it2-plugin-claude-status",

  "source": {
    "repo": "https://github.com/tmc/it2-plugin-claude-status",
    "commit": "a1b2c3d4e5f6789...",
    "go_version": "1.21.5",
    "build_command": "CGO_ENABLED=0 SOURCE_DATE_EPOCH=1 go build -trimpath -ldflags='-s -w' -buildvcs=false -o plugin",
    "build_env": {
      "CGO_ENABLED": "0",
      "SOURCE_DATE_EPOCH": "1"
    }
  },

  "binary": {
    "sha256": "deadbeef...",
    "size_bytes": 4194304,
    "platform": "darwin/arm64"
  },

  "capabilities": {
    "enrichment": ["session", "process"],
    "automation": ["suggest"]
  },

  "trust_level": "verified",

  "verification": {
    "reproducible_build": true,
    "transparency_log_url": "https://transparency.it2.dev/log/12345",
    "signed": true,
    "signature": "GPG_SIGNATURE_HERE"
  },

  "sandbox": {
    "level": "relaxed",
    "allowed_syscalls": ["read", "write", "open", "stat", "getpid", "exit"],
    "network": false,
    "filesystem_read": ["/tmp/claude-*.json"],
    "filesystem_write": [],
    "allow_exec": false
  },

  "inputs": {
    "session_id": true,
    "session_name": true,
    "working_directory": true,
    "screen_lines": 10,
    "environment": ["ITERM_SESSION_ID", "USER"]
  },

  "constraints": {
    "timeout_ms": 100,
    "max_output_bytes": 512
  },

  "caching": {
    "enabled": true,
    "ttl_seconds": 60,
    "cache_key": ["session_id"],
    "invalidate_on": ["session_change", "screen_change"]
  },

  "output": {
    "format": "json",
    "schema": {
      "icon": {"type": "string", "max_length": 4},
      "label": {"type": "string", "max_length": 20},
      "tooltip": {"type": "string", "max_length": 100}
    }
  }
}
```

## Capabilities

The `capabilities` field declares what the plugin can do:

### Enrichment Capabilities

| Capability | Description |
|------------|-------------|
| `session` | Adds data to session listings |
| `tab` | Adds data to tab listings |
| `window` | Adds data to window listings |
| `process` | Adds process inspection data |

### Automation Capabilities

| Capability | Description |
|------------|-------------|
| `suggest` | Returns recommendations (e.g., "send 'continue'") |
| `execute` | Performs actions (sends keystrokes, modifies state) |

Example:
```json
"capabilities": {
  "enrichment": ["session", "process"],
  "automation": ["suggest"]
}
```

## Trust Levels

The `trust_level` field indicates how the plugin was verified:

| Level | Description | Permissions |
|-------|-------------|-------------|
| `core` | Embedded in it2 binary | Full access |
| `verified` | Reproducible build verified | As declared in manifest |
| `community` | Has manifest, not verified | As declared in manifest |
| `untrusted` | No manifest | Minimal (session ID only) |

This field is typically set by it2 during verification, not by the plugin author.

## Sandbox Levels

### Relaxed (Verified Plugins Only)

- Can read limited screen lines
- Can read specific temp files
- Can access environment variables
- Cannot access network
- Cannot write files
- Cannot execute subprocesses
- Timeout: 100ms

### Strict (Unverified Plugins)

- Can only receive session ID
- Cannot read screen
- Cannot access filesystem
- Cannot access network
- Cannot execute subprocesses
- Timeout: 50ms

## Network and Exec Permissions

By default, plugins **cannot** make network requests or execute subprocesses. This ensures:
- Plugins cannot exfiltrate data
- Plugins cannot download and run malicious code
- Plugins cannot modify the system beyond their declared scope

### Requesting Network Access

Network access can be requested in the manifest but requires **elevated trust**:

```json
"sandbox": {
  "network": true,
  "allowed_hosts": ["api.github.com", "api.anthropic.com"],
  "allowed_ports": [443]
}
```

Network-enabled plugins:
- Must be **verified** (reproducible build)
- Are subject to additional review
- Must declare specific hosts (no wildcard)
- Are logged for audit purposes

### Requesting Exec Access

Subprocess execution is rarely needed and requires explicit justification:

```json
"sandbox": {
  "allow_exec": true,
  "allowed_executables": ["/usr/bin/git", "/usr/bin/ssh"]
}
```

Exec-enabled plugins:
- Must be **verified**
- Must declare specific executables (no shell access)
- Cannot execute arbitrary paths
- Are heavily sandboxed even when exec is allowed

## Output Caching

Plugin output can be cached to improve performance and reduce repeated executions.

### Cache Configuration

```json
"caching": {
  "enabled": true,
  "ttl_seconds": 60,
  "cache_key": ["session_id"],
  "invalidate_on": ["session_change", "screen_change"]
}
```

| Field | Description |
|-------|-------------|
| `enabled` | Whether caching is enabled (default: true for enrichment plugins) |
| `ttl_seconds` | Time-to-live in seconds (default: 60 for verified, 30 for unverified) |
| `cache_key` | Fields that form the cache key (e.g., `["session_id"]`, `["session_id", "working_directory"]`) |
| `invalidate_on` | Events that invalidate the cache |

### Cache Key Strategies

| Strategy | Use Case |
|----------|----------|
| `["session_id"]` | Output depends only on session identity |
| `["session_id", "screen_hash"]` | Output depends on current screen content |
| `["session_id", "working_directory"]` | Output depends on current directory |
| `[]` (empty) | Global cache (same output for all sessions) |

### Invalidation Events

| Event | Description |
|-------|-------------|
| `session_change` | Session state changes (new command, output) |
| `screen_change` | Screen content changes |
| `explicit` | Only invalidate on explicit refresh (e.g., `--refresh` flag) |
| `never` | Cache persists until TTL expires |

### Default Caching Behavior

| Trust Level | Default TTL | Default Invalidation |
|-------------|-------------|---------------------|
| Core | 60s | `session_change` |
| Verified | 60s | `session_change` |
| Community | 30s | `session_change` |
| Untrusted | 30s | `session_change` |

Plugins can opt out of caching with `"enabled": false`, but this impacts performance for frequently-polled enrichment plugins.

## Verification Process

### Layer 1.5: Reproducible Build Verification

```bash
# At install time:
1. Clone source at manifest.source.commit
2. Verify go.mod matches expected dependencies
3. Build with exact flags from manifest.source.build_command
4. Compare sha256(local_binary) == manifest.binary.sha256
5. If MATCH: mark as verified, enable relaxed sandbox
6. If MISMATCH: abort installation, alert user
```

### Security Properties

**What verification catches:**
- Compromised build environment (XZ-style backdoor)
- Binary doesn't match claimed source
- Supply chain attacks on distribution
- Targeted binary attacks (different users get different binaries)

**What verification doesn't catch:**
- Malicious source code (need code review)
- Malicious dependencies (need dependency audit)
- Author going rogue (need community flagging)

**Combined with sandboxing:**
- Even verified plugins are sandboxed (defense in depth)
- Unverified plugins are locked down (minimal privileges)
- All plugins have timeout + output limits (DoS prevention)

## Input Protocol

Plugins receive JSON on stdin:

```json
{
  "session_id": "abc123",
  "session_name": "Claude Code",
  "working_directory": "/path/to/project",
  "screen_lines": [
    "line 1",
    "line 2",
    "..."
  ],
  "environment": {
    "ITERM_SESSION_ID": "abc123",
    "USER": "tmc"
  }
}
```

Only fields declared in `manifest.inputs` are included.

## Output Protocol

Plugins write JSON to stdout:

```json
{
  "icon": "🚧",
  "label": "WIP",
  "tooltip": "Claude Code session has work in progress"
}
```

Output is validated against `manifest.output.schema` and sanitized before display.

## Error Handling

- Plugin timeout: enrichment omitted (graceful degradation)
- Plugin error: enrichment omitted (no cascading failures)
- Invalid output: enrichment omitted (bad data doesn't break display)
- Sandbox violation: logged, plugin auto-disabled after threshold

## Caching Strategy (Summary)

See [Output Caching](#output-caching) above for detailed configuration.

Default behavior:
- Verified plugins: 60s TTL
- Unverified plugins: 30s TTL
- Failed plugins: exponential backoff (5s, 10s, 20s, ...)
- Cache invalidation: on session change or explicit refresh flag

## Example Usage

```bash
# Install plugin (triggers verification)
$ it2 plugin install claude-status

Fetching manifest...
Cloning source from github.com/tmc/it2-plugin-claude-status@a1b2c3d4
Building from source (Go 1.21.5)...
Verifying reproducible build...
  Local:   deadbeef...
  Claimed: deadbeef...
  ✓ MATCH

✓ VERIFIED PLUGIN
  Sandbox: RELAXED (can read 10 screen lines, temp files)
  Timeout: 100ms

Install? [y/N] y
✓ Installed and enabled

# Use enrichment
$ it2 session list --enrich
[Enriched: claude-status (89ms, cached)]
ID      Name          Status
abc123  Claude Code   🚧 WIP
```

## Implementation Notes

### macOS: Seatbelt Profiles

Verified plugins use relaxed profile with limited filesystem access.
Unverified plugins use strict profile with no filesystem access.

### Linux: seccomp-bpf

Verified plugins whitelist standard syscalls minus network.
Unverified plugins whitelist minimal syscalls (read/write/exit only).

### Build Verification

Uses `go build` with deterministic flags. Entire build happens in isolated temp directory.
Build cache cleared before build to ensure reproducibility.

### Transparency Log

Optional but recommended. Provides merkle proof of binary hash at publication time.
Allows detection of binary changes after publication.
