# Design Review Synthesis

## Critical Issues (Block Implementation)

### 1. Retry-without-sandbox is a Security Hole

**All 3 reviewers agree**: This undermines the entire security model.

- Minimalist: "Never bypass security for UX"
- Pragmatist: "Security hole"
- Impartial: "⚠️ Security hole"

**Decision needed**: Remove or justify with concrete threat model.

### 2. Permission Model Too Weak

Current: `read_buffer_lines`, `timeout_ms`

Missing:
- Network access control
- Filesystem access control
- Environment variable blocklist
- Memory/process limits (DoS prevention)
- Max output size

**Risk**: Plugin can steal credentials from env vars, exhaust memory, inject terminal escape sequences.

### 3. Output Sanitization Missing

**Impartial graded this F** - critical gap for terminal injection.

Plugins output JSON but:
- No escape sequence filtering
- No size limits
- No validation that it's actually JSON
- Can inject ANSI codes to manipulate terminal

**Attack**: Plugin outputs `{"badge": "\x1b]1337;...malicious..."}`

### 4. Input Delivery Protocol Undefined

How do plugins actually get `read_buffer_lines`?

Options:
- Via stdin (serialized)?
- Via file in temp dir?
- Via env var?

**Can't implement without this**.

## Major Issues (Fix Before v1)

### 5. Daemon is Premature

Both Pragmatist and Impartial flag this:
- Adds complexity before validation
- Start with file-based cache
- Measure if too slow
- THEN add daemon if needed

### 6. Sandbox Profiles Undefined

Docs say "Seatbelt (macOS), seccomp (Linux)" but:
- What syscalls are allowed?
- What filesystem paths?
- What network operations?

**Can't implement sandbox without concrete profiles**.

### 7. plugins.sum vs enabled.json Confusion

Security doc mentions `plugins.sum` (hash registry like go.sum).
Architecture doc shows `enabled.json` (enabled list).

How do they interact?

## Inconsistencies

### Trust vs Approval

- Security doc: "user pre-approved"
- Architecture doc: "user enables"

Are these the same operation? If enabled = trusted, why prompt later?

### Embedded vs External

Architecture doc distinguishes embedded/external plugins.
Security implications differ but not explained.

## Missing Details

### Plugin Lifecycle
- Discovery mechanism (registry?)
- Auto-disable on failures?
- Revocation for compromised plugins?
- Update flow when binary changes?

### Dependency Verification
- If plugin imports malicious lib, how verify?
- Recursive reproducible builds?
- Trust go.sum (can be manipulated)?

### Resource Limits
- Memory cap per plugin?
- Process/thread limits?
- File descriptor limits?
- Timeout on blocking operations?

### Error Handling
- Plugin crashes → what happens?
- Exponential backoff?
- Auto-disable after N failures?

## Recommendations

### Do Immediately

1. **Remove retry-without-sandbox** OR write threat model justifying it
2. **Expand permission model**: network, filesystem, env blocklist, memory limits
3. **Add output sanitization spec**: escape filtering, size limits, JSON validation
4. **Define input delivery**: how plugins receive buffer lines
5. **Write sandbox-profiles.md**: concrete syscall/filesystem/network policies

### Do Before v1

6. **Simplify to file cache**: defer daemon complexity
7. **Clarify plugins.sum + enabled.json**: how they work together
8. **Define plugin lifecycle**: discovery → install → update → revoke
9. **Add dependency verification**: strategy for malicious imports
10. **Specify error handling**: crash recovery, auto-disable logic

### Consider for v2

11. Event-driven cache invalidation (git checkout → invalidate)
12. Batch execution (one plugin run for N sessions)
13. Vulnerability checking (warn about known-bad plugins)
14. Plugin dependency declarations

## Verdict

**Foundation**: Sound ✓
- Always-sandbox philosophy correct
- Defense in depth right approach
- XZ backdoor lesson learned

**Execution**: Gaps ⚠️
- Permission model underspecified
- Output sanitization missing
- Operational details undefined
- Daemon premature

**Grade**: B- (right direction, needs hardening)

## What Changed vs Codex/Gemini Patterns?

Reading their code showed:
- **Codex**: Always sandboxes, has retry flow with approval caching
- **Gemini**: Policy engine, manifest schema, trust levels

Our docs have the philosophy but missing:
- Concrete sandbox profiles (what Codex enforces)
- Permission schema details (what Gemini validates)
- Approval caching strategy (how Codex avoids repeat prompts)

**Next**: Wait for their pattern implementations, integrate those details.
