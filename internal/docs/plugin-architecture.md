# Plugin Architecture

## Goal

Extensible enrichment for `it2 session list` without compromising security or performance.

## Design

### Discovery

Plugins follow naming: `it2-session-*`, `it2-tab-*`, `it2-window-*`

Found via:
1. Embedded in binary (compiled in)
2. PATH scanning (external)

No auto-execution during discovery. Plugins must be explicitly enabled.

### Registration

```bash
# User discovers
$ it2 plugin list
it2-session-git      /usr/local/bin  (local, unverified)
it2-session-docker   /usr/local/bin  (community, verified)
it2-session-status   embedded        (first-party)

# User enables
$ it2 plugin enable it2-session-git
Trust level: LOCAL (strict sandbox)
Manifest: read 10 lines, timeout 100ms, no credentials
Hash: abc123... (unverified - enable anyway? y/N)
```

Enabled plugins stored in `~/.config/it2/plugins/enabled.json`

### Trust Determination

```go
// internal/plugins/trust.go
func DetermineTrust(plugin Plugin) TrustLevel {
    if plugin.Embedded {
        return FIRST_PARTY
    }
    if plugin.VerifiedPublisher {
        return COMMUNITY
    }
    return LOCAL
}
```

### Execution (MVP)

File-based cache with direct execution:

```bash
# Plugin runs on-demand, results cached to file
$ it2 session list
→ Reads cache: ~/.config/it2/cache/plugins.json
→ If stale (>30s): Execute plugins, update cache
→ Returns enriched session list
```

Post-MVP: Optional daemon for zero-latency execution.

### Manifest

Plugins declare permissions:

```json
{
  "name": "it2-session-git",
  "version": "1.0.0",
  "trust": "community",
  "permissions": {
    "read_buffer_lines": 10,
    "timeout_ms": 100,
    "include_credentials": false,
    "environment": {
      "allowed": ["HOME", "USER", "GIT_DIR"],
      "blocked": ["AWS_*", "*_TOKEN", "*_KEY", "*_SECRET"]
    },
    "filesystem": {
      "read": ["**/.git"],
      "write": []
    },
    "network": {
      "required": false,
      "allowed_hosts": [],
      "ports": []
    },
    "subprocess": {
      "allowed": true,
      "programs": ["/usr/bin/git"]
    }
  },
  "build": {
    "repo": "https://github.com/user/plugin",
    "commit": "abc123",
    "go_version": "1.21"
  }
}
```

Enforced at runtime. Plugin requesting more than declared → denied.

### Manifest Validation

Hard vs soft denials:

```go
// internal/plugins/manifest.go
func ValidateManifest(path string) (*Manifest, error) {
    data, err := os.ReadFile(path)
    if err != nil {
        return nil, HardDenial{err} // Unparseable = no override
    }

    var m Manifest
    if err := json.Unmarshal(data, &m); err != nil {
        return nil, HardDenial{err}
    }

    // Soft denials - user can override
    if m.Permissions.Network.Required && !userApproved(m) {
        return nil, SoftDenial{"network access requires approval"}
    }

    return &m, nil
}
```

### Communication

Plugin stdin: session ID + filtered buffer (if requested)

```json
{
  "session_id": "ABC123",
  "buffer_lines": ["line1", "line2"],
  "cwd": "/workspace"
}
```

Plugin stdout: JSON output (validated, sanitized)

```json
{
  "badge": "main",
  "status": "clean",
  "icon": "✓"
}
```

Output sanitization:
- Escape sequences filtered (SGR only)
- Max size: 4KB
- Control characters stripped
- Validated as JSON

### Anti-Spoofing

Host matching uses exact segment matching:

```go
// internal/plugins/network.go
func MatchHost(allowed, actual string) bool {
    // Prevent api.github.com.evil.com matching api.github.com
    if strings.HasPrefix(allowed, "*.") {
        // Wildcard: *.github.com matches api.github.com
        domain := allowed[2:]
        return strings.HasSuffix(actual, "."+domain) || actual == domain
    }
    // Exact match
    return allowed == actual
}
```

## Implementation

```
internal/plugins/
├── discovery.go     # Find plugins in PATH
├── registry.go      # Enabled plugins tracking
├── manifest.go      # Permission schemas, validation
├── trust.go         # Trust level determination
├── policy.go        # Deny rules engine
├── isolation.go     # Credential filtering
├── sandbox.go       # Platform-specific sandbox wrappers
├── executor.go      # Run plugins (sandboxed)
├── cache.go         # File-based result caching
└── audit.go         # Append-only audit log
```

Post-MVP:
```
├── daemon.go        # Background enrichment (optional)
└── approval.go      # Two-tier approval cache (optional)
```

## Error Handling

Plugin crashes trigger exponential backoff:

```go
// internal/plugins/executor.go
func (e *Executor) Run(plugin Plugin) (*Result, error) {
    failures := e.failures[plugin.Name]

    if failures >= 3 {
        // Auto-disable after 3 failures
        e.registry.Disable(plugin.Name)
        return nil, ErrAutoDisabled
    }

    result, err := e.execute(plugin)
    if err != nil {
        e.failures[plugin.Name]++
        backoff := time.Duration(1<<failures) * time.Second
        time.Sleep(backoff)
    } else {
        e.failures[plugin.Name] = 0
    }

    return result, err
}
```

## Cache Strategy

Simple file-based cache (MVP):

```go
// internal/plugins/cache.go
type Cache struct {
    Path string
    TTL  time.Duration // Default: 30s
}

func (c *Cache) Get(sessionID string) (*Result, bool) {
    data, err := os.ReadFile(c.Path)
    if err != nil {
        return nil, false
    }

    var cached CachedResult
    json.Unmarshal(data, &cached)

    if time.Since(cached.Timestamp) > c.TTL {
        return nil, false // Stale
    }

    return &cached.Result, true
}
```

Post-MVP: Two-tier cache for privilege escalation detection (see plugin-execution.md).

## Performance

Target: <100ms total for `it2 session list` with 10 plugins.

Optimizations:
- File cache avoids plugin execution on cache hit
- Parallel plugin execution (within sandbox limits)
- Timeout enforcement (default: 100ms per plugin)
- Auto-disable on repeated failures

Post-MVP daemon can pre-warm cache every 30s.

## Open Questions

1. ~~Daemon vs library?~~ → Start with file cache, add daemon post-MVP if needed
2. ~~Plugin output format?~~ → JSON only, validated and sanitized
3. ~~Plugin crashes?~~ → Exponential backoff + auto-disable after 3 failures
4. Network access for Docker/K8s plugins? → Yes, with manifest declaration
5. Plugin dependencies? → Defer to post-MVP
