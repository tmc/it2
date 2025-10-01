# Internal Documentation

This directory contains **implementation and contributor documentation** for developing the it2 CLI tool. These are not user-facing docs.

---

## What Goes Here

### Implementation Guides
- Technical specifications for features
- Architecture and design documents
- Performance analysis and optimization strategies
- API research and comparison studies
- Code examples for contributors

### Not Here
- User guides, tutorials, quickstarts → `docs/` at root
- Command reference, examples → auto-generated from cobra
- Troubleshooting, FAQ → `docs/` at root

---

## Current Documents

### API Research & Analysis

#### `ITERM2_SOURCE_FINDINGS.md` (40K)
**Purpose**: Comprehensive iTerm2 source code analysis
**Audience**: Contributors implementing new features
**Content**:
- 33 RPC methods cataloged from proto/api.proto
- 80+ variables across 4 scopes
- 28 color keys and preset format
- 100+ profile properties
- Performance optimization strategies
- Deprecation warnings

**Use When**: Implementing any feature that uses iTerm2 API

#### `PYTHON_API_COMPARISON.md` (22K)
**Purpose**: Feature parity analysis between Python API and it2 CLI
**Audience**: Contributors prioritizing new features
**Content**:
- Feature-by-feature comparison matrix
- Coverage assessment (~70% functional parity)
- Missing features identified
- CLI advantages over Python API
- Implementation recommendations

**Use When**: Deciding what features to add, understanding gaps

### Feature Design Specifications

#### `TERMINAL_HYPERLINKS_DESIGN.md` (16K)
**Purpose**: OSC 8 clickable links implementation guide
**Audience**: Contributors implementing hyperlink features
**Content**:
- Complete OSC 8 specification
- URL handler registration
- Code examples for Go implementation
- Security considerations
- Testing strategy

**Use When**: Implementing clickable session/tab/window IDs in list output

#### `session-tail-design.md` (2.7K)
**Purpose**: Notification-based tail architecture
**Audience**: Contributors improving tail command
**Content**:
- Current polling vs event-driven comparison
- Notification protocol details
- Coordinate-based tracking approach
- Implementation tasks

**Use When**: Refactoring tail command to use notifications

#### `session-tail-enhancements.md` (6.4K)
**Purpose**: 10 enhancement ideas for tail command
**Audience**: Contributors adding tail features
**Content**:
- Event-driven architecture
- Advanced filtering options
- Multi-session tailing
- Performance optimizations
- Output formatting enhancements

**Use When**: Planning tail command improvements

### Performance & Optimization

#### `BUFFER_FETCH_TRADEOFFS.md` (8.9K)
**Purpose**: Performance optimization strategies for buffer operations
**Audience**: Contributors optimizing performance
**Content**:
- WindowedCoordRange usage
- Style information tradeoffs
- Metadata-only queries
- Transaction API guidance
- Benchmarking recommendations

**Use When**: Optimizing any feature that fetches buffer content

---

## Document Lifecycle

### Creation
- Written during feature research/design phase
- Before major implementation work begins
- After discovering undocumented APIs or patterns

### Maintenance
- Update when APIs change
- Add lessons learned during implementation
- Document performance findings
- Note deprecations and breaking changes

### Archival
- Move to `internal/docs/archive/` when superseded
- Keep for historical reference
- Link from new docs to archived versions

---

## Writing Guidelines

### Format
- Use Markdown
- Include table of contents for docs > 5K
- Use code blocks with language hints
- Include file/line references to source code

### Structure
```markdown
# Title

**Purpose**: One-line summary
**Audience**: Who should read this
**Date**: When created/last updated

## Overview
Brief introduction

## Detailed Sections
...

## References
- Source code locations
- External documentation
- Related internal docs
```

### Content
- ✅ **Do**: Technical details, implementation strategies, API research
- ✅ **Do**: Code examples, protobuf definitions, performance data
- ✅ **Do**: Decision rationale, tradeoff analysis, gotchas
- ❌ **Don't**: User instructions, command examples, troubleshooting
- ❌ **Don't**: Session-specific notes, WIP status (use tmp/docs/)
- ❌ **Don't**: Duplicate information from external docs (link instead)

---

## Related Documentation

### User-Facing (in root `docs/`)
- README.md - Project overview
- QUICKSTART.md - Getting started
- TROUBLESHOOTING.md - Common issues
- Command examples and guides

### Generated (auto-created)
- Command reference from cobra
- API documentation from godoc

### Temporary (in `tmp/docs/`)
- Session collaboration artifacts
- Work-in-progress status
- Research notes

### Source Code (in `internal/`)
- Package documentation (godoc)
- Code comments
- Test files

---

## Contributing

When adding internal documentation:

1. **Check if it belongs here**
   - Is it for contributors? → internal/docs/
   - Is it for users? → docs/
   - Is it temporary? → tmp/docs/

2. **Choose appropriate name**
   - Feature designs: `<feature>-design.md`
   - API research: `<system>-api-analysis.md`
   - Comparisons: `<tool1>-vs-<tool2>.md`
   - Performance: `<area>-optimization.md`

3. **Link from related files**
   - ROADMAP.md should reference designs
   - Code comments should link to detailed docs
   - Design docs should link to API research

4. **Keep updated**
   - Note breaking changes
   - Add lessons learned
   - Update when implementations change

---

## Index by Topic

### iTerm2 API
- `ITERM2_SOURCE_FINDINGS.md` - Complete API catalog
- `PYTHON_API_COMPARISON.md` - Feature parity analysis

### Features
- `TERMINAL_HYPERLINKS_DESIGN.md` - Clickable links
- `session-tail-design.md` - Tail architecture
- `session-tail-enhancements.md` - Tail improvements

### Performance
- `BUFFER_FETCH_TRADEOFFS.md` - Optimization strategies

### Coming Soon
- Authentication and security design
- Plugin system architecture
- Error handling patterns
- Testing strategies
