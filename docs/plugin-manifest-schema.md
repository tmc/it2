# Plugin Manifest Schema

## Overview

Plugin manifest defines permissions, verification data, and sandbox constraints for it2 plugins.

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

## Caching Strategy

- Verified plugins: 60s TTL
- Unverified plugins: 30s TTL
- Failed plugins: exponential backoff (5s, 10s, 20s, ...)
- Cache invalidation: on explicit refresh flag

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
