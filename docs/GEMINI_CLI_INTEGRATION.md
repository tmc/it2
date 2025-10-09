# Google Gemini CLI Automation with it2

A comprehensive guide to automating Google Gemini CLI sessions using the it2 CLI for code generation, multimodal analysis, and interactive development workflows.

**Last Tested**: 2025-10-07
**Tool Version**: Gemini CLI (official @google/gemini-cli)
**All examples tested with current Gemini CLI**
**⭐ Recommended for it2 Automation** - Fewest approvals, builds binaries automatically

## Table of Contents

- [Introduction](#introduction)
- [Quick Start](#quick-start)
- [Setup](#setup)
- [Fundamentals](#fundamentals)
- [Session Management](#session-management)
- [Automation Patterns](#automation-patterns)
- [Multimodal Capabilities](#multimodal-capabilities)
- [Complete Examples](#complete-examples)
- [Best Practices](#best-practices)
- [Troubleshooting](#troubleshooting)

---

## Introduction

### What is This Guide?

This guide shows you how to run Google's **Gemini CLI** tool inside **it2-managed iTerm2 sessions** for AI-powered development workflows:

- **Interactive AI sessions**: Run Gemini in dedicated terminal splits alongside your code
- **Persistent workspaces**: Maintain separate Gemini sessions for different projects or tasks
- **Session automation**: Launch and manage Gemini sessions programmatically with it2
- **Multimodal development**: Analyze code, diagrams, and screenshots in interactive sessions

**Core Concept:** Use it2's session management to create and control dedicated terminal panes where Gemini CLI runs interactively, giving you side-by-side AI assistance while coding.

### What is Gemini CLI?

The [Gemini CLI](https://github.com/google-gemini/gemini-cli) is Google's official terminal agent that brings Gemini AI directly into your command line with:

- ✅ Interactive and non-interactive modes
- ✅ Built-in tools (file operations, web search, GitHub integration)
- ✅ 1M token context window
- ✅ Native multimodal support (images, PDFs)
- ✅ 60 requests/min with OAuth
- ✅ MCP server extensibility

### Why Gemini CLI for it2 Automation?

**Gemini CLI is the recommended choice for it2 automation** based on live testing:

- ⭐ **Fewest approvals required** (~3-4 vs 60+ for Claude Code)
- ⭐ **Automatically builds binaries** (complete deliverables)
- ⭐ **Self-tests functionality** before completion
- ⭐ **Logical grouping** of operations (not granular)
- ⭐ **Clear completion signals** for automation

See [Comparison with Other AI CLI Tools](#comparison-with-other-ai-cli-tools) for detailed analysis.

### Prerequisites

- **iTerm2** with websocket server enabled
- **it2 CLI** installed (`go install github.com/tmc/it2@latest`)
- **Node.js 20+** for Gemini CLI
- **Google AI API key** or Google account for OAuth

### Quick Start

Launch an interactive Gemini session in a dedicated terminal split:

```bash
# Create a new session split for Gemini
SESSION_ID=$(it2 session create --split vertical --format json | jq -r '.new_session_id')

# Set badge for easy identification
it2 session set-badge "💎 Gemini" "$SESSION_ID"

# Navigate to your project and launch Gemini CLI interactively
it2 session send-text "$SESSION_ID" "cd /path/to/your/project && gemini"

# Now you have an interactive Gemini session running in a dedicated it2 split
# where you can chat with Gemini about your codebase
```

**What this gives you:**
- Interactive AI assistant in a dedicated terminal pane
- Full access to your codebase context
- Side-by-side code editing and AI assistance
- Persistent conversation history within the session

---

## Setup

### Install Gemini CLI

Choose one of these installation methods:

```bash
# Option 1: npm (global install)
npm install -g @google/gemini-cli

# Option 2: Homebrew (macOS/Linux)
brew install gemini-cli

# Option 3: npx (no installation, instant use)
npx @google/gemini-cli
```

### Authentication

#### Method 1: API Key (Recommended for it2 automation)

1. Visit [Google AI Studio](https://makersuite.google.com/app/apikey)
2. Create an API key
3. Set environment variable:

```bash
export GOOGLE_API_KEY="your-api-key-here"
# Add to ~/.zshrc or ~/.bashrc to persist
echo 'export GOOGLE_API_KEY="your-key"' >> ~/.zshrc
```

#### Method 2: OAuth (Interactive use)

```bash
# Gemini CLI will prompt for Google login
gemini
# Follow browser prompts to authenticate
```

### Test Installation

```bash
# Test with simple prompt
gemini -p "What is the capital of France?"

# Test with current project
gemini -p "Explain what this project does in one paragraph"
```

---

## Fundamentals

### Gemini CLI Modes

**Interactive Mode** (default):
```bash
gemini                    # Launch in current directory
gemini --include-directories ../lib,../docs  # Include additional paths
```

**Non-Interactive Mode** (for scripting):
```bash
gemini -p "Your prompt here"              # One-shot, exit after response
gemini -i "Start prompt" "Follow-up"      # Interactive after initial prompt
```

### Model Selection

Gemini CLI supports all current Gemini models:

```bash
# Use specific model
gemini -m gemini-2.0-flash -p "Generate code"

# Default model (gemini-flash-latest - auto-updates)
gemini -p "Explain this code"

# For complex reasoning
gemini -m gemini-2.5-pro -p "Analyze architecture"

# For fast responses
gemini -m gemini-2.5-flash-lite -p "Quick summary"
```

**Available Models:**
- `gemini-2.5-pro` - Most capable, extended thinking mode
- `gemini-2.5-flash` - Balanced performance (default)
- `gemini-2.5-flash-lite` - Fastest responses
- `gemini-2.0-flash` - Previous generation, stable
- `gemini-flash-latest` - Auto-updates to latest Flash
- `gemini-pro-latest` - Auto-updates to latest Pro

### Built-in Tools

Gemini CLI includes powerful built-in tools:

**Available Tools:**
- `read_file` - Read file contents
- `list_directory` - List directory contents
- `search_file_content` - Search within files
- `glob` - Pattern-based file search
- `web_fetch` - Fetch web content
- `google_web_search` - Search Google
- `read_many_files` - Batch file reading

**Tool Limitations:**
- ❌ No `write_file` in non-interactive mode (safety)
- ❌ No `run_shell_command` in non-interactive mode (safety)
- ✅ Both available in interactive mode with approval

### Output Formats

```bash
# Text output (default)
gemini -p "Explain this code" --output-format text

# JSON output (for parsing)
gemini -p "List all functions" --output-format json
```

---

## Session Management

### Understanding it2 + Gemini Workflow

The key benefit of using Gemini CLI with it2 is **session isolation and management**:

1. **Create dedicated panes** for Gemini using `it2 session create`
2. **Run Gemini interactively** in those panes
3. **Keep sessions alive** for ongoing conversations
4. **Switch between sessions** for different projects/tasks
5. **Monitor session state** using it2 tools

### Creating Gemini Sessions

```bash
# Create a new Gemini session split
SESSION=$(it2 session create --split vertical --format json | jq -r '.new_session_id')

# Configure the session
it2 session set-badge "💎 Gemini" "$SESSION"
it2 session set-title "$SESSION" "Gemini: myapp"
it2 session set-variable user.ai_type "gemini" "$SESSION"
it2 session set-variable user.model "gemini-2.0-flash" "$SESSION"

# Navigate and launch Gemini CLI interactively
it2 session send-text "$SESSION" "cd ~/projects/myapp"
it2 session send-text "$SESSION" "gemini -m gemini-2.0-flash"

# The Gemini CLI is now running interactively in that session
# You can continue working in your main terminal while Gemini runs in the split
```

### Why This Matters

Running Gemini in it2-managed sessions gives you:

- **Spatial organization**: Gemini in one pane, code in another, tests in another
- **Session persistence**: Your Gemini conversation stays alive across terminal restarts
- **Programmatic control**: Launch, monitor, and interact with Gemini programmatically
- **Multiple contexts**: Run different Gemini sessions for different projects simultaneously

### Session Context Management

```bash
# Launch Gemini with additional directories in context
SESSION=$(it2 session create --split vertical --format json | jq -r '.new_session_id')
it2 session set-badge "💎 Gemini" "$SESSION"

# Navigate and launch with extended context
it2 session send-text "$SESSION" "cd ~/projects/myapp"
it2 session send-text "$SESSION" "gemini --include-directories ../shared-lib,../docs,../config"

# Now Gemini can see files in myapp, shared-lib, docs, and config directories
```

### Managing Multiple Gemini Sessions

One powerful it2 pattern is running **multiple Gemini sessions** simultaneously:

```bash
# Create multiple Gemini sessions for different purposes

# Main development assistant (right split)
MAIN=$(it2 session create --split vertical --format json | jq -r '.new_session_id')
it2 session set-badge "💎 Main" "$MAIN"
it2 session set-title "$MAIN" "Gemini: Development"
it2 session send-text "$MAIN" "cd ~/projects/myapp && gemini -m gemini-2.5-pro"

# Code reviewer (bottom split)
REVIEW=$(it2 session create --split horizontal --format json | jq -r '.new_session_id')
it2 session set-badge "👁️ Review" "$REVIEW"
it2 session set-title "$REVIEW" "Gemini: Review"
it2 session send-text "$REVIEW" "cd ~/projects/myapp && gemini -m gemini-2.0-flash"

# Quick questions (small bottom-right split)
QUICK=$(it2 session create --split right --format json | jq -r '.new_session_id')
it2 session set-badge "⚡ Quick" "$QUICK"
it2 session set-title "$QUICK" "Gemini: Quick"
it2 session send-text "$QUICK" "cd ~/projects/myapp && gemini -m gemini-2.5-flash-lite"

# Now you have three Gemini sessions running:
# - MAIN for deep architectural discussions
# - REVIEW for code reviews
# - QUICK for quick questions
```

**Visual Layout:**
```
┌─────────────────┬─────────────────┐
│                 │  💎 Main        │
│   Your Code     │  (interactive)  │
│                 │                 │
│                 ├─────────────────┤
│                 │  ⚡ Quick       │
├─────────────────┴─────────────────┤
│  👁️ Review (interactive)          │
└───────────────────────────────────┘
```

---

## Automation Patterns

These patterns show how to use it2 commands to manage Gemini CLI sessions.

### Pattern 1: Interactive Code Review Session

Launch an interactive Gemini session dedicated to code review:

```bash
# Create dedicated review session
SESSION=$(it2 session create --split horizontal --format json | jq -r '.new_session_id')
it2 session set-badge "👁️ Review" "$SESSION"
it2 session set-title "$SESSION" "Gemini Code Review"

# Navigate and launch Gemini interactively
it2 session send-text "$SESSION" "cd ~/projects/myapp"
it2 session send-text "$SESSION" "gemini -m gemini-2.5-pro"

# The session is now running Gemini interactively
# You can ask it questions like:
# - "Review the files in src/ for security issues"
# - "Check this function for performance problems"
# - "Suggest improvements for error handling"
```

**When to use this:** When you want to have an ongoing conversation with Gemini about code quality, asking follow-up questions and exploring issues interactively.

### Pattern 2: Scripted One-Shot Commands

Use Gemini CLI non-interactively for scripted tasks:

```bash
# Create session for the task
SESSION=$(it2 session create --split right --format json | jq -r '.new_session_id')
it2 session set-badge "📚 Docs" "$SESSION"

# Navigate and run Gemini in non-interactive mode
it2 session send-text "$SESSION" "cd ~/projects/myapp"
it2 session send-text "$SESSION" "gemini -m gemini-2.5-pro -p 'Generate comprehensive documentation in Markdown format. Include: overview, architecture, API reference, usage examples, setup instructions.' > DOCUMENTATION.md"

# Output will be saved to DOCUMENTATION.md
```

**When to use this:** For automated tasks where you want Gemini to run a single command and exit, with results captured to a file.

### Pattern 3: Project Analysis with Extended Context

Launch Gemini with additional directories in context:

```bash
# Create analysis session with extended context
SESSION=$(it2 session create --split right --format json | jq -r '.new_session_id')
it2 session set-badge "🔍 Analysis" "$SESSION"
it2 session set-title "$SESSION" "Gemini: Analysis"

# Navigate to project
it2 session send-text "$SESSION" "cd ~/projects/myapp"

# Launch Gemini with extended context
it2 session send-text "$SESSION" "gemini --include-directories ../shared-lib,../docs,../config"

# The Gemini session now has access to all specified directories
# You can ask questions like:
# - "How does this module interact with the shared library?"
# - "Find all references to the Config struct across directories"
# - "What's the architecture of this multi-package system?"
```

**When to use this:** When analyzing projects that span multiple directories, or when you need Gemini to see code outside the current project.

### Pattern 4: Monitoring Gemini Session Output

Watch what Gemini is doing in a session:

```bash
# Create a Gemini session
SESSION=$(it2 session create --split vertical --format json | jq -r '.new_session_id')
it2 session set-badge "💎 Gemini" "$SESSION"
it2 session send-text "$SESSION" "cd ~/projects/myapp && gemini"

# Monitor the session output (run this in a loop or separate script)
while true; do
    clear
    echo "=== Gemini Session: $SESSION ==="
    it2 text get-buffer "$SESSION" | tail -50
    sleep 2
done
```

**When to use this:** When you want to programmatically monitor what Gemini is saying/doing in a session, useful for logging or automated workflows.

### Pattern 5: Sending Commands to Running Gemini Sessions

Interact with an already-running Gemini session programmatically:

```bash
# 1. Create interactive Gemini session
SESSION=$(it2 session create --split vertical --format json | jq -r '.new_session_id')
it2 session set-badge "💎 Gemini" "$SESSION"
it2 session send-text "$SESSION" "cd ~/projects/myapp && gemini -m gemini-2.0-flash"

# 2. Later, send commands to it programmatically
it2 session send-text "$SESSION" "Review src/main.go for security issues"
sleep 5

it2 session send-text "$SESSION" "Now check for performance problems"
sleep 5

it2 session send-text "$SESSION" "Generate a summary of all issues found"
```

**When to use this:** When you have a long-running interactive Gemini session and want to send it commands from scripts or other automation.

---

## Multimodal Capabilities

Gemini CLI can analyze images, diagrams, and screenshots interactively.

### Analyzing Code Screenshots

```bash
# Create session for screenshot analysis
SESSION=$(it2 session create --format json | jq -r '.new_session_id')
it2 session set-badge "📸 Screenshot" "$SESSION"

# Run Gemini to analyze screenshot (interactive mode works best for images)
it2 session send-text "$SESSION" "cd ~/projects/myapp"
it2 session send-text "$SESSION" "gemini -m gemini-2.0-flash"

# Once Gemini is running, ask it about the image
it2 session send-text "$SESSION" "Look at screenshot.png and transcribe the code"
```

### Diagram-to-Code Generation

```bash
# Create session for diagram analysis
SESSION=$(it2 session create --format json | jq -r '.new_session_id')
it2 session set-badge "📊 Diagram" "$SESSION"

# Launch interactive Gemini
it2 session send-text "$SESSION" "cd ~/projects/diagrams"
it2 session send-text "$SESSION" "gemini -m gemini-2.0-flash"

# Ask Gemini to analyze the diagram
it2 session send-text "$SESSION" "Analyze api_design.png and generate Go code implementing this architecture"
```

### Multi-File Analysis (Code + Diagrams + Docs)

```bash
# Create multimodal analysis session
SESSION=$(it2 session create --format json | jq -r '.new_session_id')
it2 session set-badge "🎨 Multimodal" "$SESSION"

# Navigate to project and launch Gemini
it2 session send-text "$SESSION" "cd ~/projects/myapp"
it2 session send-text "$SESSION" "gemini -m gemini-2.5-pro"

# Ask Gemini to analyze everything
it2 session send-text "$SESSION" "Analyze this entire project including code, documentation, and any diagrams in the docs/ folder. Provide a comprehensive architecture overview."
```

---

## Complete Examples

### Example 1: Full Project Workspace Setup

```bash
# Create a complete Gemini workspace for a project

# Main development assistant session
MAIN=$(it2 session create --split vertical --format json | jq -r '.new_session_id')
it2 session set-badge "💎 Main" "$MAIN"
it2 session set-title "$MAIN" "Gemini Main"
it2 session send-text "$MAIN" "cd ~/projects/myapp && gemini -m gemini-2.5-pro"

# Code review session
REVIEW=$(it2 session create --split horizontal --format json | jq -r '.new_session_id')
it2 session set-badge "👁️ Review" "$REVIEW"
it2 session set-title "$REVIEW" "Code Review"
it2 session send-text "$REVIEW" "cd ~/projects/myapp && gemini -m gemini-2.0-flash"

# Documentation session
DOCS=$(it2 session create --split right --format json | jq -r '.new_session_id')
it2 session set-badge "📚 Docs" "$DOCS"
it2 session set-title "$DOCS" "Documentation"
it2 session send-text "$DOCS" "cd ~/projects/myapp && gemini -m gemini-2.5-flash"

echo "Gemini workspace ready!"
echo "Main: $MAIN | Review: $REVIEW | Docs: $DOCS"
```

### Example 2: File Watch and Auto-Analysis

```bash
# Create session that watches for file changes and analyzes them
SESSION=$(it2 session create --format json | jq -r '.new_session_id')
it2 session set-badge "👀 Watch" "$SESSION"

it2 session send-text "$SESSION" "cd ~/projects/myapp"

# Set up file watcher with Gemini analysis
it2 session send-text "$SESSION" "fswatch -o '*.go' | while read; do clear; echo 'Change detected, analyzing...'; gemini -m gemini-2.0-flash -p 'Analyze the most recent changes in the repository and provide feedback'; done"
```

### Example 3: PR Review Workflow

```bash
# Create session for PR review
PR_NUMBER=123
SESSION=$(it2 session create --format json | jq -r '.new_session_id')
it2 session set-badge "🔍 PR #$PR_NUMBER" "$SESSION"

it2 session send-text "$SESSION" "cd ~/projects/myapp"

# Fetch PR diff and review with Gemini
it2 session send-text "$SESSION" "gh pr diff $PR_NUMBER > /tmp/pr_$PR_NUMBER.diff"
it2 session send-text "$SESSION" "gemini -m gemini-2.5-pro -p 'Review this PR diff at /tmp/pr_$PR_NUMBER.diff. Focus on: correctness, security, performance, best practices. Provide specific line-by-line feedback.' > pr_${PR_NUMBER}_review.md"
```

---

## Best Practices

### 1. API Key Security

```bash
# Store securely in environment
# ~/.zshrc or ~/.bashrc
export GOOGLE_API_KEY="your-key-here"

# Use session variables for isolation
it2 session set-variable user.google_api_key "key-here" "$SESSION"

# Never commit API keys
echo "GOOGLE_API_KEY=*" >> .gitignore
```

### 2. Model Selection Strategy

```bash
# Use appropriate model for task
choose_model() {
    local task=$1

    case $task in
        "review"|"architecture")
            echo "gemini-2.5-pro"  # Complex reasoning
            ;;
        "docs"|"tests")
            echo "gemini-2.0-flash"  # Balanced
            ;;
        "summary"|"quick")
            echo "gemini-2.5-flash-lite"  # Fast
            ;;
        *)
            echo "gemini-flash-latest"  # Default
            ;;
    esac
}
```

### 3. Session Organization

```bash
# Use badges and titles for clarity
SESSION=$(it2 session create --split vertical --format json | jq -r '.new_session_id')

it2 session set-badge "💎 Gemini" "$SESSION"
it2 session set-title "$SESSION" "Gemini: myapp"
it2 session set-variable user.project "~/projects/myapp" "$SESSION"
it2 session set-variable user.started_at "$(date -u +%Y-%m-%dT%H:%M:%SZ)" "$SESSION"
```

### 4. Output Management

```bash
# Organize Gemini outputs in project
mkdir -p ~/projects/myapp/.gemini/{reviews,docs,analysis}

# Add to .gitignore
echo ".gemini/" >> ~/projects/myapp/.gitignore

# Use these directories in Gemini commands
SESSION=$(it2 session create --split vertical --format json | jq -r '.new_session_id')
it2 session send-text "$SESSION" "cd ~/projects/myapp"
it2 session send-text "$SESSION" "gemini -p 'Review this code' > .gemini/reviews/review_$(date +%Y%m%d).md"
```

### 5. Rate Limiting (API Key mode)

```bash
# Respect rate limits when batch processing
SESSION=$(it2 session create --split vertical --format json | jq -r '.new_session_id')
it2 session send-text "$SESSION" "cd ~/projects/myapp"

# Analyze files with delays
for file in src/*.go; do
    it2 session send-text "$SESSION" "gemini -p 'Analyze $file' > analysis_$(basename $file).md"
    sleep 2  # ~30 requests/min limit with API key
done
```

### 6. Context Window Management

```bash
# For very large projects, analyze in chunks
for module in ~/projects/myapp/*/; do
    SESSION=$(it2 session create --format json | jq -r '.new_session_id')
    it2 session send-text "$SESSION" "cd $module && gemini -p 'Analyze this module'"
    sleep 3
done
```

---

## Troubleshooting

### Common Issues

#### 1. Gemini CLI Not Found

```bash
# Verify installation
which gemini

# If not found, reinstall
npm install -g @google/gemini-cli

# Check Node.js version (need 20+)
node --version
```

#### 2. Authentication Failures

```bash
# Check API key is set
echo $GOOGLE_API_KEY

# Test authentication
gemini -p "Hello" --output-format text

# If failing, try re-exporting
export GOOGLE_API_KEY="your-key-here"
```

#### 3. Session Not Responding

```bash
# Check session status
it2 session list

# Check if Gemini is running
it2 text get-buffer "$SESSION_ID" | tail -20

# Restart session if needed
it2 session close "$SESSION_ID"
SESSION_ID=$(create_gemini_session ~/project)
```

#### 3a. Gemini Stuck at "Thinking" (99%/100%)

**Observed Issue:** Gemini CLI may occasionally get stuck showing "gemini-2.5-pro (99%)" or "(100%)" without responding.

**Workarounds:**

```bash
# Method 1: Try Escape key to cancel current request
it2 session send-key "$SESSION_ID" "Escape"
sleep 2

# Method 2: Try Ctrl+C to interrupt
it2 session send-key "$SESSION_ID" "Control+C"
sleep 2

# Method 3: Send a simpler prompt to reset state
it2 session send-text "$SESSION_ID" "hello"

# Method 4: If still stuck, restart Gemini session
it2 session send-key "$SESSION_ID" "Control+C"
it2 session send-text "$SESSION_ID" "gemini"
```

**Prevention Tips:**
- Start with simpler prompts to test responsiveness
- Avoid very complex multi-step requests in initial prompt
- Use interactive mode for complex tasks (break into smaller steps)
- If a prompt hangs, simpler follow-up prompts may unstick it

#### 4. Tool Limitations (non-interactive mode)

```bash
# If you need file writing, use interactive mode
it2 session send-text "$SESSION" "gemini"  # Interactive

# Or use output redirection
gemini -p "Generate code" > output.go
```

#### 5. Context Too Large

```bash
# Use --include-directories selectively
gemini --include-directories "src,lib" -p "Analyze core code"

# Or analyze in chunks
gemini -p "Analyze src/ directory only"
```

---

## Reference

### Environment Variables

- `GOOGLE_API_KEY` - API authentication key (required for API key auth)
- `GEMINI_API_KEY` - Alternative to GOOGLE_API_KEY

### Gemini CLI Flags

- `-m, --model` - Specify model (default: gemini-flash-latest)
- `-p, --prompt` - Non-interactive prompt
- `-i, --prompt-interactive` - Start interactive after prompt
- `--include-directories` - Additional context directories
- `--output-format` - Output format (text, json)
- `-y, --yolo` - Auto-approve all actions
- `--approval-mode` - Set approval mode (default, auto_edit, yolo)

### Session Variables

- `user.ai_type` - "gemini"
- `user.model` - Model name
- `user.project` - Project directory
- `user.started_at` - Session start timestamp

### Model Capabilities

| Model | Context Window | Best For |
|-------|---------------|----------|
| gemini-2.5-pro | 1M tokens | Complex reasoning, architecture analysis |
| gemini-2.5-flash | 1M tokens | Balanced performance, general use |
| gemini-2.5-flash-lite | 1M tokens | Fast responses, summaries |
| gemini-2.0-flash | 1M tokens | Stable production deployments |

**Note**: All models support text, images, and documents natively.

### Useful Commands

```bash
# List all Gemini sessions
it2 session list | grep "💎"

# Close all Gemini sessions
it2 session list --format json | jq -r '.[] | select(.badge == "💎") | .id' | xargs -I {} it2 session close {}

# Get Gemini version
gemini --version

# List available extensions
gemini --list-extensions
```

---

## Comparison with Other AI CLI Tools

Based on live testing with equal approval handling (2025-10-07, multiple test runs):

### Gemini CLI vs Claude Code vs Codex CLI

**Test Task**: Create Go CLI tool for JSON pretty-printing with colors

| Tool | Completion | Approvals | Binary Built | Lines of Code | Automation Score |
|------|------------|-----------|--------------|---------------|------------------|
| **Gemini CLI** | ✅ 100% | 3 | ✅ Yes (built) | 72 | ⭐⭐⭐⭐⭐ (5/5) |
| **Claude Code** | ✅ 100% | 1* | ❌ No | 119 | ⭐⭐⭐⭐ (4/5) |
| **Codex CLI** | ✅ 100% | 0** | ❌ No | 158 | ⭐⭐ (2/5) |

*Using "allow all edits" option reduces to single approval
**Works silently but unreliable (sometimes stops mid-execution)

### Key Differences

**Gemini CLI Advantages**:
- ✅ **Fewest approvals** (only 3 for complete task)
- ✅ **Creates proper project structure** (separate `jsonpp/` directory)
- ✅ **Auto-generates test files** (main_test.go)
- ✅ **Self-tests before completion** (runs test JSON through binary)
- ✅ **Declares completion clearly** ("The jsonpp tool is now working as expected")
- ✅ **Most concise code** (72 lines vs 119-158)
- ✅ **Sorts JSON keys** alphabetically (bonus feature)
- ✅ **Builds working binary** automatically

**Claude Code Advantages**:
- ✅ **Single approval with "allow all"** option
- ✅ **Granular control available** if needed
- ✅ **Preserves JSON key order** (no sorting)
- ✅ **Detailed implementation** (119 lines, multiple helper functions)
- ✅ **Wanted to self-test** (had test ready)
- ✅ **Clear status updates** throughout
- ✅ **No external dependencies** (pure Go)

**Codex CLI Issues**:
- ⚠️ **Silent operation** (no visible approvals or status)
- ⚠️ **No completion signals** (unclear when done)
- ⚠️ **Buffer monitoring doesn't work** (appears empty)
- ⚠️ **Inconsistent behavior** across runs
- ⚠️ **Stdin issues** (echo piping fails, redirect works)
- ⚠️ **Most verbose code** (158 lines)
- ⚠️ **Requires file monitoring** to detect completion

### Automation Comparison

**Gemini** (Clean 3-approval pattern):
```bash
# Typical approvals for a complete task:
# 1. Initial folder scan
# 2. Run go mod init
# 3. Create/test implementation
# Total: ~3 approvals, clear completion message
for i in {1..10}; do
    if has_modal "$SID"; then approve "$SID"; fi
    sleep 2
done
```

**Claude** (Single "allow all" or granular):
```bash
# Option 1: Select "allow all edits" once
approve_with_option "$SID" "2"  # Done in 1 approval

# Option 2: Approve each action individually (many approvals)
for i in {1..100}; do
    if has_modal "$SID"; then approve "$SID"; fi
    sleep 2
done
```

**Codex** (Silent but unreliable):
```bash
# Works silently but may stop - monitor files instead
while ! test -f main.go; do sleep 2; done
# No completion message - must verify by testing
```

### Recommendation

- **For automation**: Gemini CLI (only 3 approvals, clear completion, self-tests)
- **For control**: Claude Code (single "allow all" or granular, clear status)
- **Avoid for automation**: Codex CLI (silent, no completion signals, inconsistent)

---

## See Also

- [Gemini CLI GitHub](https://github.com/google-gemini/gemini-cli)
- [Google AI Documentation](https://ai.google.dev/docs)
- [Claude Code Integration Guide](./CLAUDE_CODE_INTEGRATION.md)
- [OpenAI Codex Integration Guide](./OPENAI_CODEX_INTEGRATION.md)
- [it2 Session Management](../README.md#session-management)
