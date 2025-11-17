# Plugin Execution Flow

## Overview

Simplified 4-layer security model for it2 plugin execution.

## MVP Flow (4 Layers)

```
┌─────────────────────────────────────────────┐
│  User Request                               │
│  $ it2 session list                         │
└────────────────┬────────────────────────────┘
                 │
                 ▼
┌─────────────────────────────────────────────┐
│  Layer 1: Policy Check                      │
│  • Check deny rules (JSON config)           │
│  • Check path restrictions                  │
│  • Hard denial → stop                       │
└────────────────┬────────────────────────────┘
                 │ if allowed
                 ▼
┌─────────────────────────────────────────────┐
│  Layer 2: Manifest Validation               │
│  • Parse manifest JSON                      │
│  • Validate schema                          │
│  • Unparseable → hard denial (stop)         │
│  • Network required → soft denial (prompt)  │
└────────────────┬────────────────────────────┘
                 │ if valid
                 ▼
┌─────────────────────────────────────────────┐
│  Layer 3: Sandboxed Execution               │
│  • Determine trust level                    │
│  • Select sandbox profile:                  │
│    - FIRST_PARTY → relaxed                  │
│    - COMMUNITY → relaxed                    │
│    - LOCAL → strict                         │
│  • Filter credentials from buffer           │
│  • Execute in OS sandbox                    │
│    - macOS: Seatbelt                        │
│    - Linux: Landlock                        │
│  • Timeout enforcement (100ms default)      │
└────────────────┬────────────────────────────┘
                 │
                 ▼
┌─────────────────────────────────────────────┐
│  Layer 4: Audit & Cache                     │
│  • Log to audit.log (TSV format)            │
│  • Cache result (file-based, 30s TTL)       │
│  • Return enriched data                     │
└─────────────────────────────────────────────┘
```

## Code Flow

```go
// internal/plugins/executor.go
func (e *Executor) Execute(sessionID string) (*EnrichedSession, error) {
    // Layer 1: Policy Check
    if err := e.policy.Check(plugin.Name, cwd); err != nil {
        e.audit.Log(plugin, "policy-denied", err)
        return nil, err
    }

    // Layer 2: Manifest Validation
    manifest, err := ValidateManifest(plugin.Path)
    if err != nil {
        switch err.(type) {
        case HardDenial:
            e.audit.Log(plugin, "manifest-invalid", err)
            return nil, err
        case SoftDenial:
            // Prompt user for approval
            if !e.approvalFunc(plugin, err.Error()) {
                return nil, ErrUserDenied
            }
        }
    }

    // Layer 3: Sandboxed Execution
    trust := DetermineTrust(plugin)
    sandbox := SelectSandbox(trust)

    // Filter credentials
    buffer := e.getBuffer(sessionID, manifest.Permissions.ReadBufferLines)
    safeBuffer := FilterBuffer(buffer)

    // Execute with timeout
    ctx, cancel := context.WithTimeout(context.Background(),
        time.Duration(manifest.Permissions.TimeoutMs)*time.Millisecond)
    defer cancel()

    result, err := e.runSandboxed(ctx, plugin, sandbox, safeBuffer)

    // Layer 4: Audit & Cache
    e.audit.Log(plugin, "executed", result, err)
    if err == nil {
        e.cache.Set(sessionID, result)
    }

    return result, err
}
```

## Post-MVP: Retry with Relaxed Sandbox

When sandbox denials occur (false positives), offer escalation:

```
┌─────────────────────────────────────────────┐
│  Layer 3: Sandboxed Execution               │
│  • Execute in strict sandbox                │
└────────────────┬────────────────────────────┘
                 │
                 ├──> Success → Layer 4
                 │
                 └──> Sandbox Denial
                       │
                       ▼
┌─────────────────────────────────────────────┐
│  Escalation Flow (Post-MVP)                 │
│  • Detect sandbox violation                 │
│  • Prompt: "Retry with relaxed sandbox?"    │
│  • If approved:                             │
│    - Execute with RELAXED sandbox           │
│    - Still sandboxed (not unsandboxed!)     │
│  • Cache escalation decision                │
└────────────────┬────────────────────────────┘
                 │
                 └──> Layer 4
```

Implementation:

```go
func (e *Executor) executeWithRetry(ctx context.Context, plugin Plugin) (*Result, error) {
    // First attempt: Strict/Relaxed based on trust
    trust := DetermineTrust(plugin)
    sandbox := SelectSandbox(trust)

    result, err := e.runSandboxed(ctx, plugin, sandbox, buffer)
    if err == nil {
        return result, nil
    }

    // Check if it's a sandbox denial
    if !IsSandboxDenial(err) {
        return nil, err
    }

    // Prompt for escalation
    if !e.approvalFunc(plugin, "retry with relaxed sandbox?") {
        return nil, err
    }

    // Retry with relaxed sandbox (still sandboxed!)
    relaxedSandbox := RELAXED
    return e.runSandboxed(ctx, plugin, relaxedSandbox, buffer)
}
```

## Post-MVP: Two-Tier Cache

Detect privilege escalation via directory change:

```go
// internal/plugins/approval.go
type ApprovalCache struct {
    decisions map[string]Decision
}

func (c *ApprovalCache) buildKey(plugin, version, cwd string) string {
    // Tier 1: Plugin identity
    tier1 := fmt.Sprintf("%s@%s", plugin, version)

    // Tier 2: Execution context (includes cwd)
    hash := sha256.Sum256([]byte(plugin + version + cwd))
    tier2 := fmt.Sprintf("%s:%x", tier1, hash[:8])

    return tier2
}

func (c *ApprovalCache) Check(plugin Plugin, cwd string) (Decision, bool) {
    key := c.buildKey(plugin.Name, plugin.Version, cwd)
    decision, ok := c.decisions[key]
    return decision, ok
}
```

Use case:

```bash
# Session 1: User approves plugin in /workspace
$ cd /workspace
$ it2 session list
→ Prompt: "Approve it2-session-git@1.0.0?" [YES]
→ Cached: "it2-session-git@1.0.0:/workspace" → APPROVED

# Later: Plugin runs in different directory
$ cd /etc
$ it2 session list
→ Cache miss (different cwd hash)
→ Prompt: "it2-session-git@1.0.0 accessing /etc. Allow?" [y/N]
```

This prevents silent privilege escalation.

Cost: 20 lines, SHA256 hash computation.

## Test Strategy

### Unit Tests

```go
// internal/plugins/policy_test.go
func TestPolicyDenyList(t *testing.T) {
    policy := &Policy{
        DenyList: []string{"evil-plugin"},
    }

    err := policy.Check("evil-plugin", "/workspace")
    if err != ErrPolicyDenied {
        t.Errorf("expected denial, got: %v", err)
    }
}

// internal/plugins/sandbox_test.go
func TestCredentialFiltering(t *testing.T) {
    lines := []string{
        "normal line",
        "export AWS_SECRET_KEY=abc123",
        "another normal line",
    }

    filtered := FilterBuffer(lines)

    if len(filtered) != 2 {
        t.Errorf("expected 2 lines, got %d", len(filtered))
    }

    for _, line := range filtered {
        if containsCredential(line) {
            t.Errorf("credential leaked: %s", line)
        }
    }
}
```

### Integration Tests

```go
// internal/plugins/executor_test.go
func TestPluginCannotReadSSHKeys(t *testing.T) {
    // Create test plugin that tries to read ~/.ssh/id_rsa
    plugin := &Plugin{
        Path: "./testdata/malicious-plugin",
        Manifest: &Manifest{
            Permissions: Permissions{
                Filesystem: FilesystemPerms{
                    Read: []string{"~/.ssh/**"},
                },
            },
        },
    }

    executor := NewExecutor(STRICT)
    result, err := executor.Execute("test-session", plugin)

    // Should fail with sandbox denial
    if err == nil {
        t.Error("expected sandbox denial, plugin succeeded")
    }

    if !IsSandboxDenial(err) {
        t.Errorf("expected sandbox denial, got: %v", err)
    }
}
```

### E2E Tests

```bash
# Test: Plugin tries to access /etc/passwd
$ ./it2-test-plugin-access-etc &
$ it2 session list
# Expected: Sandbox denial, plugin auto-disabled after 3 failures

# Test: Plugin outputs malicious escape sequence
$ ./it2-test-plugin-terminal-injection &
$ it2 session list
# Expected: Escape sequences stripped, only SGR allowed
```

## Performance Targets

- **Policy check**: <1ms
- **Manifest validation**: <5ms
- **Sandboxed execution**: <100ms (per plugin)
- **Audit log write**: <1ms (async)
- **Cache read**: <1ms (file read)
- **Total (10 plugins)**: <100ms (with cache), <1000ms (cache miss)

## Comparison: Full vs Simplified vs Minimal

| Feature | Full (Codex) | Simplified (Recommended) | Minimal |
|---------|-------------|-------------------------|---------|
| Layers | 7-8 | 4 | 3 |
| Trust levels | 5 | 3 | 1 |
| Cache | Two-tier in-memory | File-based (optional two-tier) | File-based |
| Message bus | Yes | No (callbacks) | No |
| Heuristic retry | Yes | No (explicit) | No |
| Policy engine | Full Rego | JSON deny rules | None |
| Complexity | High | Medium | Low |
| Security | Very High | High | Medium |
| Best for | Multi-agent systems | CLI tools (it2) | Proof of concept |
| LOC | 3200+ | ~600 | ~300 |

## References

- Codex patterns: `/Volumes/tmc/go/src/github.com/openai/codex/codex-go-patterns/`
- Gemini patterns: `/Volumes/tmc/go/src/github.com/google-gemini/gemini-cli/internal/plugins/`
- Attribution system: `attribution-system.md`
- Review synthesis: `review-synthesis.md`
