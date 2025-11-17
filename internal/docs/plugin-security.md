# Plugin Security

## Problem

Users want `it2 session list` to show rich metadata (git branch, docker status, AI session state). Manual composition doesn't work:

```bash
# No one will do this
it2 session list | it2-add-git | it2-add-docker | it2-add-claude
```

Auto-execution from PATH is dangerous:

```bash
# Attacker's binary in PATH
# Auto-runs, sees buffer with credentials
/usr/local/bin/it2-session-enrich
```

## Solution

Always sandbox. Even verified plugins.

Trust affects **approval**, not sandboxing:
- First-party plugin: auto-approved, runs in relaxed sandbox
- Community plugin: user approves once, runs in relaxed sandbox
- Local plugin: user must approve, runs in strict sandbox
- All sandboxed: Seatbelt (macOS), Landlock (Linux)

## Key Insight (from Codex/Gemini analysis)

XZ backdoor lesson: verified code can be malicious. Trust != safety.

Defense in depth:
1. Policy engine (deny rules)
2. Manifests (declare permissions)
3. Sandbox (OS enforces limits)
4. Credential filtering (sensitive data never exposed)
5. Audit log (detect breaches)

## Credential Filtering

Plugins never see credentials, even if manifest requests them.

```go
// internal/plugins/isolation.go
func FilterBuffer(lines []string) []string {
    filtered := make([]string, 0, len(lines))
    for _, line := range lines {
        if containsCredential(line) {
            continue // Drop entire line
        }
        filtered = append(filtered, line)
    }
    return filtered
}

func containsCredential(line string) bool {
    patterns := []string{
        "export AWS_",
        "export GITHUB_TOKEN",
        "export OPENAI_API_KEY",
        "password=",
        "token=",
        "BEGIN PRIVATE KEY",
    }
    lower := strings.ToLower(line)
    for _, p := range patterns {
        if strings.Contains(lower, strings.ToLower(p)) {
            return true
        }
    }
    return false
}
```

Manifest must explicitly request (default: false):

```json
{
  "permissions": {
    "read_buffer_lines": 10,
    "include_credentials": false
  }
}
```

If `include_credentials: false` (default), filter before passing to plugin.

## Sandbox Profiles

### macOS Seatbelt (Strict)

For local/untrusted plugins:

```scheme
(version 1)
(deny default)
(allow file-read* (subpath "/tmp"))
(allow file-read* (literal (param "workspace")))
(allow file-write* (subpath "/tmp"))
(allow process-exec (literal "/usr/bin/git") (literal "/usr/bin/docker"))
(deny network*)
```

### macOS Seatbelt (Relaxed)

For community/first-party plugins:

```scheme
(version 1)
(deny default)
(allow file-read* (subpath (param "workspace")))
(allow file-write* (subpath (param "workspace")))
(allow process-exec (literal "/usr/bin/git") (literal "/usr/bin/docker"))
(allow network-outbound (remote ip "127.0.0.1"))
```

### Linux Landlock (Strict)

For local/untrusted plugins:

```go
spec := landlock.Ruleset{
    HandleAccess: landlock.AccessFSReadFile | landlock.AccessFSReadDir,
    PathBeneath: []landlock.PathBeneath{
        {Path: "/tmp", Access: landlock.AccessFSReadFile | landlock.AccessFSWriteFile},
        {Path: workspace, Access: landlock.AccessFSReadFile},
    },
}
```

### Linux Landlock (Relaxed)

For community/first-party plugins:

```go
spec := landlock.Ruleset{
    HandleAccess: landlock.AccessFSReadFile | landlock.AccessFSReadDir | landlock.AccessFSWriteFile,
    PathBeneath: []landlock.PathBeneath{
        {Path: workspace, Access: landlock.AccessFSReadFile | landlock.AccessFSWriteFile},
    },
}
```

## Policy Engine

Simple JSON deny rules checked before execution.

User config (`~/.config/it2/policy.json`):

```json
{
  "allow_list": ["it2-session-git", "it2-session-docker"],
  "deny_list": [],
  "path_rules": {
    "/etc": "deny",
    "/var/secrets": "deny"
  }
}
```

Implementation:

```go
// internal/plugins/policy.go
type Policy struct {
    DenyAll     bool              // Global killswitch
    AllowList   []string          // Only these plugins
    DenyList    []string          // Never these plugins
    PathRules   map[string]string // "/sensitive" -> "deny"
}

func (p *Policy) Check(plugin string, cwd string) error {
    if p.DenyAll {
        return ErrPolicyDenied
    }

    // Deny list takes precedence
    for _, denied := range p.DenyList {
        if plugin == denied {
            return ErrPolicyDenied
        }
    }

    // If allow list exists, plugin must be on it
    if len(p.AllowList) > 0 {
        allowed := false
        for _, a := range p.AllowList {
            if plugin == a {
                allowed = true
                break
            }
        }
        if !allowed {
            return ErrPolicyDenied
        }
    }

    // Path rules
    for path, action := range p.PathRules {
        if strings.HasPrefix(cwd, path) && action == "deny" {
            return ErrPolicyDenied
        }
    }

    return nil
}
```

## Trust Levels

Three levels (simplified from 5):

```go
const (
    FIRST_PARTY  // Shipped with it2 binary
    COMMUNITY    // Signed by known publishers
    LOCAL        // User's own scripts
)
```

Trust determines:
- Auto-approval (first-party: yes, others: no)
- Sandbox profile (first-party: relaxed, local: strict)
- Audit detail level

## Audit Log

Append-only log: `~/.config/it2/audit.log`

Format (TSV):

```
timestamp | plugin | hook | duration | exit_code | sandbox_violations | user_decision
```

Example:

```
2025-11-17T12:00:00Z	it2-session-git	session-list	45ms	0	none	auto-approved
2025-11-17T12:01:30Z	local-plugin	session-list	120ms	1	network-denied	user-approved
```

## Implementation Phases

### MVP (Phase 1)

- Policy engine with JSON deny rules
- Credential filtering (ENV vars + buffer)
- Sandbox profiles (Seatbelt/Landlock)
- Manifest validation
- 3-level trust model
- Basic audit log

### Post-MVP (Phase 2)

- Retry with relaxed sandbox on false positives
- Two-tier approval cache (detect cwd privilege escalation)
- Transparency log for reproducible builds
- Windows sandboxing (AppContainer)

## Open Questions

1. Network access for Docker/K8s plugins? (Allow with manifest declaration)
2. Windows sandboxing? (Defer to post-MVP, use AppContainer)
3. Performance with 10+ plugins? (Solved by daemon + caching)
