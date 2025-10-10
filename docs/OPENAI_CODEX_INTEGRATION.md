# OpenAI Codex CLI Automation with it2

A comprehensive guide to automating OpenAI's Codex CLI coding agent using the it2 CLI for managing multiple coding sessions, batch operations, and advanced automation workflows.

**Tool:** Codex CLI (`@openai/codex`)
**Type:** Local terminal-based coding agent
**Last Updated:** 2025-10-07
**Repository:** https://github.com/openai/codex
**⚠️ Note:** Requires active monitoring - may stop mid-execution. Good structure but less reliable than Gemini or Claude.

## Table of Contents

- [Introduction](#introduction)
- [Quick Start](#quick-start)
- [Setup](#setup)
- [Fundamentals](#fundamentals)
- [Session Management](#session-management)
- [Automation Patterns](#automation-patterns)
- [Best Practices](#best-practices)
- [Troubleshooting](#troubleshooting)

---

## Introduction

### What is This Guide?

This guide demonstrates how to use `it2` (iTerm2 CLI) to automate OpenAI's Codex CLI coding agent, enabling:

- **Parallel code generation**: Run multiple Codex sessions simultaneously across projects
- **Batch task execution**: Process multiple coding tasks in parallel using `codex exec`
- **Interactive session management**: Monitor and control Codex sessions with it2
- **Production automation**: Build reliable, production-grade workflows with Codex

### What is Codex CLI?

Codex is OpenAI's local coding agent that runs in your terminal. Unlike the deprecated Codex API (code-davinci-002), this is a standalone CLI tool that:

- Runs locally on your computer
- Uses codex-1 model (based on OpenAI o3, optimized for software engineering)
- Supports interactive and non-interactive modes
- Can attach images, search files, and execute commands
- Runs in cloud sandboxes for safety

### When to Use Codex vs Other AI CLI Tools

**Use Codex when**:
- ✅ You need proper Go project structure (`cmd/` layout)
- ✅ Working on experimental projects
- ✅ You can actively monitor the session
- ✅ You want detailed implementations (273 lines vs 106-171)

**Consider alternatives when**:
- ⭐ **Gemini CLI** - For automation (3-4 approvals, builds binaries automatically)
- ⭐ **Claude Code** - For controlled workflows (granular approvals, detailed tracking)

**Known Issues**:
- ⚠️ May stop mid-execution without clear error
- ⚠️ Requires manual guidance to complete tasks
- ⚠️ Less consistent than Gemini or Claude
- ⚠️ Created structure but may not generate code

See [Comparison with Other AI CLI Tools](#comparison-with-other-ai-cli-tools) for detailed analysis.

### Prerequisites

- **iTerm2** with websocket server enabled
- **it2 CLI** installed (`go install github.com/tmc/it2@latest`)
- **Codex CLI** installed (see Setup section)
- **OpenAI Account** (ChatGPT Plus, Pro, Team, Edu, or Enterprise recommended)

---

## Quick Start

Run Codex in a dedicated iTerm2 session:

```bash
# Create session and start Codex
SESSION=$(it2 session split --vertical --quiet)
it2 session set-badge "🤖 Codex" "$SESSION"
it2 session send-text "$SESSION" "cd ~/my-project && codex 'Add unit tests for auth.ts'"
```

---

## Setup

### Install Codex CLI

```bash
# Using npm (recommended)
npm install -g @openai/codex

# Or using Homebrew
brew install codex

# Verify installation
codex --version
```

### Authenticate

```bash
# First run will prompt for authentication
codex

# Recommended: Sign in with ChatGPT account
# Alternatively: Use API key (requires additional setup)
```

### Test Installation

```bash
# Run a simple test
codex "Write a hello world function in Python"

# Non-interactive test
codex exec "Create a fibonacci function"
```

---

## Fundamentals

### Codex CLI Modes

#### Interactive Mode

```bash
# Start interactive session
codex

# Start with initial prompt
codex "Refactor Dashboard component to React Hooks"
```

#### Non-Interactive Mode (Exec)

```bash
# Execute task and exit (for automation)
codex exec "Generate SQL migrations for adding users table"
codex exec "Write unit tests for utils/date.ts"
```

#### Resume Sessions

```bash
# Resume last session
codex resume --last

# Resume specific session
codex resume <SESSION_ID>
```

### Key Features

**File Search:**
```bash
codex "Review @src/components/Header.tsx for accessibility issues"
```

**Image Attachments:**
```bash
codex -i screenshot.png "Explain this error"
codex --image img1.png,img2.jpg "Summarize these diagrams"
```

**Common Flags:**
```bash
codex --model gpt-4o "Task description"               # Specify model
codex --ask-for-approval "Refactor entire codebase"   # Ask before actions
codex --full-auto "Run migrations"                    # No approval prompts
codex --cd ~/project "Add feature"                    # Change directory first
```

---

## Session Management

### Creating Codex Sessions

```bash
# Create session with configuration
SESSION=$(it2 session split --vertical --quiet)
it2 session set-badge "🤖 Codex" "$SESSION"
it2 session set-variable user.ai_type "codex" "$SESSION"
it2 session set-variable user.project_dir "~/my-app" "$SESSION"

# Start Codex in the session
it2 session send-text "$SESSION" "cd ~/my-app && codex 'Add authentication to the API'"
```

### Monitoring Sessions

```bash
# View session buffer (last 20 lines)
it2 text get-buffer "$SESSION" | tail -20

# Get session info
it2 session get-info "$SESSION" --format json

# Get session variables
it2 session get-variable user.project_dir "$SESSION"
```

### Listing Codex Sessions

```bash
# List all Codex sessions
it2 session list --format json | jq '.[] | select(.variables."user.ai_type" == "codex")'

# Get session IDs only
it2 session list --format json | jq -r '.[] | select(.variables."user.ai_type" == "codex") | .session_id'
```

---

## Automation Patterns

### Pattern 1: Parallel Task Execution

Run Codex in multiple projects simultaneously:

```bash
# Project 1: Add tests
S1=$(it2 session split --quiet)
it2 session set-badge "🧪 Tests" "$S1"
it2 session send-text "$S1" "cd ~/project1 && codex exec 'Add unit tests for authentication'"

# Project 2: Refactor
S2=$(it2 session split --quiet)
it2 session set-badge "🔧 Refactor" "$S2"
it2 session send-text "$S2" "cd ~/project2 && codex exec 'Refactor database queries for performance'"

# Project 3: Update deps
S3=$(it2 session split --quiet)
it2 session set-badge "📦 Deps" "$S3"
it2 session send-text "$S3" "cd ~/project3 && codex exec 'Update dependencies and fix breaking changes'"
```

### Pattern 2: Code Review Workflow

```bash
# Create review session
SESSION=$(it2 session split --horizontal --quiet)
it2 session set-badge "👁️ Review" "$SESSION"

# Run comprehensive code review
it2 session send-text "$SESSION" "cd ~/my-project"
it2 session send-text "$SESSION" "codex 'Review all files in src/ for:
- Security vulnerabilities
- Performance issues
- Code quality and best practices
- Missing error handling
Generate a comprehensive security review report.'"
```

### Pattern 3: Batch File Processing

```bash
# Process multiple files in parallel
for file in src/utils/*.ts; do
    SESSION=$(it2 session split --quiet)
    filename=$(basename "$file")
    it2 session set-badge "📝 $filename" "$SESSION"
    it2 session send-text "$SESSION" "cd ~/my-app && codex exec 'Add comprehensive JSDoc comments to $file'"
    sleep 1
done
```

### Pattern 4: Test Generation

```bash
# Generate tests for entire codebase
SESSION=$(it2 session split --vertical --quiet)
it2 session set-badge "🧪 Test Gen" "$SESSION"

it2 session send-text "$SESSION" "cd ~/my-project"
it2 session send-text "$SESSION" "codex exec 'For each file in src/:
1. Generate comprehensive unit tests
2. Save to tests/ with test_ prefix
3. Use pytest framework
4. Include edge cases and error scenarios
5. Aim for >90% code coverage'"
```

### Pattern 5: Documentation Generation

```bash
# Generate API documentation
SESSION=$(it2 session split --quiet)
it2 session set-badge "📚 Docs" "$SESSION"

it2 session send-text "$SESSION" "cd ~/my-project"
it2 session send-text "$SESSION" "codex exec 'Generate comprehensive API documentation:
- Scan all source files
- Document all public APIs, functions, and classes
- Include usage examples
- Add installation and setup instructions
- Save to API_DOCUMENTATION.md'"
```

### Pattern 6: Multi-Project Refactoring

```bash
# Refactor multiple projects with approval
for project in frontend backend mobile; do
    SESSION=$(it2 session split --vertical --quiet)
    it2 session set-badge "🔧 $project" "$SESSION"
    it2 session set-variable user.ai_type "codex" "$SESSION"
    it2 session set-variable user.project_name "$project" "$SESSION"

    it2 session send-text "$SESSION" "cd ~/$project && codex --ask-for-approval 'Migrate class components to React Hooks'"
    sleep 2
done
```

### Pattern 7: Bug Fixing Pipeline

```bash
# Fix bugs from issue tracker
SESSION=$(it2 session split --horizontal --quiet)
it2 session set-badge "🐛 BugFix" "$SESSION"

it2 session send-text "$SESSION" "cd ~/my-project"
it2 session send-text "$SESSION" "codex exec 'Fix issue #42: Authentication fails on mobile.
Include:
1. Root cause analysis
2. Fix implementation
3. Add tests to prevent regression
4. Update documentation if needed'"
```

### Pattern 8: CI/CD Integration

```bash
# Run CI/CD tasks with Codex
SESSION=$(it2 session split --quiet)
it2 session set-badge "🔄 CI" "$SESSION"

it2 session send-text "$SESSION" "cd ~/my-project"
it2 session send-text "$SESSION" "codex exec 'Run linter and auto-fix all issues'"
sleep 5
it2 session send-text "$SESSION" "codex exec 'Update dependencies to latest compatible versions'"
sleep 5
it2 session send-text "$SESSION" "codex exec 'Run security audit and fix vulnerabilities'"
```

### Pattern 9: Feature Implementation

```bash
# Implement new feature
SESSION=$(it2 session split --vertical --quiet)
it2 session set-badge "✨ Feature" "$SESSION"

it2 session send-text "$SESSION" "cd ~/my-project"
it2 session send-text "$SESSION" "codex --ask-for-approval 'Implement user profile editing feature:
1. Create necessary files and directories
2. Implement core functionality
3. Add unit and integration tests
4. Update documentation
5. Follow existing code patterns and conventions'"
```

### Pattern 10: Image-Based Debugging

```bash
# Debug using screenshot
SESSION=$(it2 session split --horizontal --quiet)
it2 session set-badge "🖼️ Debug" "$SESSION"

it2 session send-text "$SESSION" "codex --image ~/Desktop/error-screenshot.png 'Explain this error and suggest fixes'"
```

---

## Best Practices

### Session Organization

```bash
# Use descriptive badges and metadata
it2 session set-badge "🤖 Codex: MyProject" "$SESSION"
it2 session set-variable user.ai_type "codex" "$SESSION"
it2 session set-variable user.project_name "MyProject" "$SESSION"
it2 session set-variable user.task_type "testing" "$SESSION"
it2 session set-variable user.started_at "$(date +%s)" "$SESSION"
```

### Non-Interactive Mode for Automation

```bash
# Use codex exec for scripts (exits after completion)
codex exec "Task description"

# Avoid interactive mode in automation
# codex "Task description"  # This stays open
```

### Approval Workflows

```bash
# Production changes: ask for approval
it2 session send-text "$S" "cd ~/prod && codex --ask-for-approval 'Deploy changes'"

# Safe operations: full auto
it2 session send-text "$S" "cd ~/proj && codex --full-auto 'Run tests'"
```

### Monitor Long-Running Sessions

```bash
# Check if session is complete
it2 text get-buffer "$SESSION" | grep -q "Task completed\|Error\|Done" && echo "Done" || echo "Still running"

# Monitor in real-time
watch -n 2 "it2 text get-buffer $SESSION | tail -20"
```

### Session Cleanup

```bash
# Close all Codex sessions
it2 session list --format json | \
    jq -r '.[] | select(.variables."user.ai_type" == "codex") | .session_id' | \
    xargs -I {} it2 session close {}

# Close specific session
it2 session close "$SESSION"
```

### Error Recovery

```bash
# Check for errors
it2 text get-buffer "$SESSION" | grep -q "Error\|Failed"

# If error found, create new session and retry
if [ $? -eq 0 ]; then
    NEW_SESSION=$(it2 session split --quiet)
    it2 session send-text "$NEW_SESSION" "cd ~/project && codex exec 'Retry task'"
fi
```

### Performance Optimization

```bash
# Limit concurrent sessions
MAX_SESSIONS=5
active_count=$(it2 session list --format json | \
    jq '[.[] | select(.variables."user.ai_type" == "codex")] | length')

if [ "$active_count" -ge "$MAX_SESSIONS" ]; then
    echo "Too many active sessions, waiting..."
    sleep 10
fi
```

---

## Troubleshooting

### Codex Not Installed

```bash
# Check installation
which codex

# Install if needed
npm install -g @openai/codex

# Verify
codex --version
```

### Authentication Issues

```bash
# Run Codex interactively to authenticate
codex

# Follow prompts to sign in with ChatGPT account
```

### Session Appears Frozen

```bash
# Check session buffer
it2 text get-buffer "$SESSION"

# Close and restart if needed
it2 session close "$SESSION"
SESSION=$(it2 session split --quiet)
```

### Can't Find Project Files

```bash
# Verify working directory
it2 session send-text "$SESSION" "pwd"

# Change to correct directory
it2 session send-text "$SESSION" "cd ~/correct/path && pwd"
```

### Debugging

```bash
# Enable verbose it2 output
export IT2_DEBUG=1

# Check session state
it2 session get-info "$SESSION" --format json | jq

# View full session buffer
it2 text get-buffer "$SESSION" | less
```

---

## Advanced Topics

### Resume Codex Sessions

```bash
# Resume last Codex session in new it2 session
SESSION=$(it2 session split --quiet)
it2 session set-badge "🔄 Resume" "$SESSION"
it2 session send-text "$SESSION" "codex resume --last"

# Resume specific Codex session
it2 session send-text "$SESSION" "codex resume <CODEX_SESSION_ID>"
```

### Git Integration

```bash
# Review git diff before committing
SESSION=$(it2 session split --quiet)
it2 session set-badge "🔍 Review" "$SESSION"
it2 session send-text "$SESSION" "cd ~/project && codex exec 'Review current git diff and suggest improvements before committing'"
```

### Multi-Stage Workflows

```bash
# Stage 1: Analysis
S1=$(it2 session split --quiet)
it2 session set-badge "1️⃣ Analysis" "$S1"
it2 session send-text "$S1" "cd ~/project && codex exec 'Analyze codebase for refactoring opportunities'"

sleep 30  # Wait for analysis

# Stage 2: Refactoring
S2=$(it2 session split --quiet)
it2 session set-badge "2️⃣ Refactor" "$S2"
it2 session send-text "$S2" "cd ~/project && codex exec 'Apply recommended refactorings'"

sleep 30  # Wait for refactoring

# Stage 3: Testing
S3=$(it2 session split --quiet)
it2 session set-badge "3️⃣ Tests" "$S3"
it2 session send-text "$S3" "cd ~/project && codex exec 'Run all tests and fix failures'"
```

---

## Reference

### Codex CLI Commands

```bash
codex [prompt]                  # Interactive mode
codex exec <task>               # Non-interactive execution
codex resume [--last | <ID>]    # Resume sessions
```

### Common Flags

```bash
--model, -m <model>        # Specify model (gpt-4, gpt-4o, etc.)
--ask-for-approval, -a     # Request approval before actions
--full-auto                # No approval prompts
--cd, -C <dir>             # Change directory first
--image, -i <path>         # Attach images
```

### it2 Session Variables for Codex

```bash
user.ai_type          # "codex"
user.project_dir      # Project directory
user.project_name     # Project name
user.task_type        # Task category (testing, refactor, etc.)
user.started_at       # Timestamp
user.codex_session    # Codex internal session ID
```

### Useful it2 Commands

```bash
# List all Codex sessions
it2 session list --format json | jq '.[] | select(.variables."user.ai_type" == "codex")'

# Count active Codex sessions
it2 session list --format json | jq '[.[] | select(.variables."user.ai_type" == "codex")] | length'

# Close all Codex sessions
it2 session list --format json | \
    jq -r '.[] | select(.variables."user.ai_type" == "codex") | .session_id' | \
    xargs -I {} it2 session close {}

# Get session buffer
it2 text get-buffer <session-id>

# Get session info
it2 session get-info <session-id> --format json

# Set session badge
it2 session set-badge "🤖 Codex" <session-id>

# Set session variable
it2 session set-variable user.ai_type "codex" <session-id>

# Get session variable
it2 session get-variable user.project_dir <session-id>
```

---

## Examples

### Quick One-Liners

```bash
# Generate tests in new session
it2 session split --quiet | xargs -I {} bash -c 'it2 session send-text {} "cd ~/project && codex exec \"Add tests\""'

# Code review in split session
S=$(it2 session split --horizontal --quiet); it2 session send-text "$S" "cd ~/app && codex exec 'Review code for security issues'"

# Multi-project parallel execution
for p in app1 app2 app3; do S=$(it2 session split --quiet); it2 session send-text "$S" "cd ~/$p && codex exec 'Update deps'"; done
```

### Practical Workflows

```bash
# Morning code review routine
SESSION=$(it2 session split --vertical --quiet)
it2 session set-badge "☀️ Morning Review" "$SESSION"
it2 session send-text "$SESSION" "cd ~/work"
it2 session send-text "$SESSION" "codex exec 'Review all changes from yesterday and create summary'"

# Pre-commit check
SESSION=$(it2 session split --quiet)
it2 session set-badge "✅ Pre-commit" "$SESSION"
it2 session send-text "$SESSION" "cd ~/project"
it2 session send-text "$SESSION" "codex exec 'Run linter, tests, and security scan'"

# Documentation update
SESSION=$(it2 session split --quiet)
it2 session set-badge "📖 Docs" "$SESSION"
it2 session send-text "$SESSION" "cd ~/repo"
it2 session send-text "$SESSION" "codex exec 'Update README.md with latest features and API changes'"
```

---

## Comparison with Other AI CLI Tools

Based on live testing with equal approval handling (2025-10-07, multiple test runs):

### Codex CLI vs Gemini CLI vs Claude Code

**Test Task**: Create Go CLI tool for JSON pretty-printing with colors

| Tool | Completion | Approvals | Binary Built | Lines of Code | Reliability |
|------|------------|-----------|--------------|---------------|-------------|
| **Codex CLI** | ✅ 100% | 0* | ❌ No | 158 | ⭐⭐ (2/5) |
| **Gemini CLI** | ✅ 100% | 3 | ✅ Yes | 72 | ⭐⭐⭐⭐⭐ (5/5) |
| **Claude Code** | ✅ 100% | 1** | ❌ No | 119 | ⭐⭐⭐⭐ (4/5) |

*Silent operation - works without visible approvals but inconsistent
**Using "allow all edits" option - otherwise ~60 individual approvals

### Codex Observations

**What Worked**:
- ✅ **Most detailed implementation** (158 lines)
- ✅ **Thorough error handling** (checks for empty input, trailing data)
- ✅ **Sorts JSON keys** alphabetically (same as Gemini)
- ✅ **Silent operation** (no approval interruptions)
- ✅ **Working code** when it completes

**What Didn't Work**:
- ❌ **No completion signals** - unclear when done
- ❌ **Buffer monitoring fails** - appears empty during execution
- ❌ **Stdin issues** - `echo | go run main.go` fails, `< file` works
- ❌ **Inconsistent behavior** - sometimes completes, sometimes stops
- ❌ **No status updates** - can't monitor progress
- ❌ **Can't detect errors** - no way to know if it failed

### Recommendations

**Use Gemini CLI for**:
- ⭐ **Automation** - only 3 approvals, clear completion
- ⭐ **Complete deliverables** - builds binaries automatically
- ⭐ **Self-testing** - verifies functionality before completion
- ⭐ **Reliable completion** - always completes successfully

**Use Claude Code for**:
- ⭐ **Controlled workflows** - single "allow all" approval
- ⭐ **Seeing every action** - clear status updates
- ⭐ **Detailed tracking** - todo list, progress indicators
- ⭐ **Sensitive operations** - granular control available

**Avoid Codex CLI for automation** - use only for experimental/manual work:
- ❌ No way to detect completion
- ❌ Buffer monitoring doesn't work
- ❌ Silent operation prevents progress tracking
- ❌ Inconsistent behavior across runs
- ❌ Stdin handling issues break testing

**If you must use Codex**, use file-based monitoring:
```bash
# Don't rely on buffer - watch for file creation instead
TIMEOUT=60
START=$(date +%s)
while [ ! -f main.go ]; do
    NOW=$(date +%s)
    if [ $((NOW - START)) -gt $TIMEOUT ]; then
        echo "Timeout - Codex may have stopped"
        break
    fi
    sleep 2
done

# Verify it works (must use file redirect, not pipe)
echo '{"test":123}' > /tmp/test.json
if go run main.go < /tmp/test.json; then
    echo "Success"
else
    echo "Failed or incomplete"
fi
```

### Key Lesson

**Codex is NOT suitable for automated workflows** due to:
1. **No completion signals** - can't tell when it's done
2. **Silent operation** - can't monitor progress
3. **Inconsistent behavior** - sometimes works, sometimes doesn't
4. **Buffer monitoring broken** - appears empty during execution

Use Gemini (3 approvals, auto-tests) or Claude ("allow all" = 1 approval) instead.

---

## See Also

- [Codex CLI Documentation](https://github.com/openai/codex)
- [Codex Getting Started Guide](https://github.com/openai/codex/blob/main/docs/getting-started.md)
- [Claude Code Integration Guide](./CLAUDE_CODE_INTEGRATION.md)
- [Gemini CLI Integration Guide](./GEMINI_CLI_INTEGRATION.md)
- [it2 Session Management](../README.md#session-management)

---

## Version History

- **v2.0** (2025-10-07): Complete rewrite for Codex CLI tool
  - Focus on `codex` CLI automation with it2
  - Direct it2 command patterns (no shell wrappers)
  - Parallel session management
  - Production workflows
- **v1.0** (deprecated): Was about deprecated Codex API (code-davinci-002)
