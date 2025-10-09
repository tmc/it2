---
name: who
description: List all active Claude Code sessions (like IRC /who)
auto-approve:
  - Bash(it2 session list:*)
  - Bash(it2 text get-buffer:*)
  - Bash(it2 text get-screen:*)
---

List all active iTerm2 sessions that appear to be running Claude Code.

1. Use `it2 session list --format json` to get all sessions
2. For each session, check if it's running Claude by examining the buffer or title
3. Display a formatted list showing:
   - Session ID (first 8 chars)
   - Session title
   - Current status (idle/processing/waiting for input)
   - Working directory if available

Present the results in a clean table format.
