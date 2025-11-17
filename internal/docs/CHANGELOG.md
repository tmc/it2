# Plugin Architecture Design Updates

## Date: 2025-11-17

### Multi-Claude Review Process

Six Claude sessions debated and refined the plugin security architecture:

1. **Minimalist (5B6B399A)** - Identified inconsistencies, critical gaps
2. **Pragmatist (B35C6025)** - Flagged operational concerns
3. **Impartial (20C3F69B)** - Balanced review, prioritized P0/P1/P2
4. **Attribution (96CF7F6E)** - Plugin composition concerns
5. **Codex (21111A58)** - Sandbox orchestration patterns
6. **Gemini (E80A550D)** - Policy engine, manifest validation

### Key Outcomes

**Consensus Reached:**
- ✅ Always sandbox (even verified plugins)
- ✅ 3-level trust (FIRST_PARTY/COMMUNITY/LOCAL)
- ✅ 4-layer execution flow (simplified from 7-8)
- ✅ Policy engine with JSON deny rules
- ✅ Credential filtering mandatory
- ✅ Concrete sandbox profiles (Seatbelt/Landlock)

**Rejected as Over-Engineering:**
- ❌ Message bus (use callbacks)
- ❌ 5-level trust hierarchy
- ❌ 7-8 layer execution flow
- ❌ Heuristic retry detection
- ❌ Automatic retry-without-sandbox

**Deferred to Post-MVP:**
- Two-tier approval cache (optional, for cwd escalation detection)
- Daemon-based caching (start with file cache)
- Transparency log for reproducible builds
- Windows sandboxing (AppContainer)

## Updated Files

### 1. plugin-security.md

**Added:**
- Credential filtering implementation (FilterBuffer, containsCredential)
- Concrete sandbox profiles (Seatbelt strict/relaxed, Landlock strict/relaxed)
- Policy engine implementation (JSON deny rules)
- 3-level trust model (FIRST_PARTY/COMMUNITY/LOCAL)
- Audit log format (TSV with structured fields)
- MVP vs Post-MVP phasing

**Removed:**
- "Retry-without-sandbox" (changed to "retry-with-relaxed-sandbox" in post-MVP)
- Vague "Phase 2" promises

**Changed:**
- Trust levels: 5 → 3
- Defense layers: Added credential filtering, policy engine
- Implementation split into MVP/Post-MVP

### 2. plugin-architecture.md

**Added:**
- Expanded manifest schema (environment, filesystem, network, subprocess permissions)
- Trust determination logic
- Manifest validation (hard vs soft denials)
- Anti-spoofing host matching
- Output sanitization spec
- Error handling (exponential backoff, auto-disable)
- Cache strategy (file-based with TTL)
- Performance targets (<100ms for 10 plugins)

**Removed:**
- Daemon as primary path (now post-MVP optional)
- Undefined "open questions" (answered with concrete decisions)

**Changed:**
- Execution model: Daemon → file-based cache (MVP)
- Communication: Added input spec (JSON stdin), output validation
- Implementation: Added policy.go, trust.go, isolation.go, audit.go

### 3. plugin-execution.md (NEW)

**Created comprehensive execution guide:**
- 4-layer flow diagram (Policy → Manifest → Sandbox → Audit)
- Complete code flow example
- Post-MVP retry flow (with relaxed sandbox, not unsandboxed)
- Post-MVP two-tier cache (cwd escalation detection)
- Test strategy (unit, integration, e2e)
- Performance targets
- Comparison table (Full vs Simplified vs Minimal)

## Security Primitives Validated

All reviewers agreed these are essential:

1. **Policy Engine** - Rule-based allow/deny before execution
2. **Manifest Validation** - Hard/soft denials, schema enforcement
3. **Credential Filtering** - ENV vars + buffer scanning, never exposed
4. **Sandbox Profiles** - Platform-specific (Seatbelt/Landlock) with concrete syscall rules
5. **Trust Levels** - 3 tiers determining approval + sandbox profile
6. **Audit Logging** - Append-only trail with structured data
7. **Output Sanitization** - Escape filtering, size limits, JSON validation
8. **Anti-Spoofing** - Exact segment matching for host wildcards

## Implementation Guidance

### MVP Priority (P0 - Must Have)

1. Policy engine with JSON deny rules
2. Credential filtering (ENV + buffer)
3. Concrete sandbox profiles
4. Manifest validation with hard/soft denials
5. 3-level trust model
6. Basic audit log
7. Output sanitization

### Post-MVP (P1/P2 - Nice to Have)

1. Two-tier approval cache (detect cwd escalation)
2. Retry with relaxed sandbox (on false positives)
3. Daemon-based caching (if file cache too slow)
4. Transparency log (reproducible builds)
5. Windows sandboxing (AppContainer)

### Deferred (P3 - Future)

1. Plugin dependencies/composition
2. Advanced policy engine (Rego/OPA)
3. 5-level trust hierarchy
4. Heuristic denial detection

## Reference Implementations

- **Codex patterns**: `/Volumes/tmc/go/src/github.com/openai/codex/codex-go-patterns/`
  - Combined orchestrator (full 7-layer flow)
  - Two-tier approval cache
  - Sandbox transformation patterns
  - Credential filtering

- **Gemini patterns**: `/Volumes/tmc/go/src/github.com/google-gemini/gemini-cli/internal/plugins/`
  - Policy engine design
  - Trust model implementation
  - Manifest validation
  - Simple integration (618 lines)

## Attribution

This design emerged from collaborative review across 6 Claude sessions:

- Initial docs: Session 5404B77F
- Critical review: Sessions 5B6B, B35C, 20C3, 96CF
- Pattern implementations: Sessions 2111, E80A
- Synthesis: Session 5404B77F

All session metadata captured in git notes for full attribution chain.

## Next Steps

1. Implement MVP security primitives in order (P0 list)
2. Create reference plugin (it2-session-git) with complete manifest
3. Build test suite (unit → integration → e2e)
4. Measure performance against targets
5. Iterate based on real-world false positive data
