---
name: join
description: Create or switch to a Claude Code session (like IRC /join #channel)
argument-hint: [session-name]
auto-approve:
  - Bash(it2 session list:*)
  - Bash(it2 session split:*)
  - Bash(it2 session send-text:*)
  - Bash(it2 session set-title:*)
  - Bash(it2 session set-badge:*)
  - Bash(it2 session focus:*)
---

Create a new iTerm2 session and launch Claude Code in it, or switch to an existing Claude session.

Arguments: [session-name] (optional)

If session-name is provided:
1. Check if a session with that name already exists using `it2 session list`
2. If it exists, focus on it
3. If it doesn't exist, create a new vertical split session, set its title to the session-name, and launch Claude

If no session-name is provided:
1. Create a new vertical split session
2. Launch Claude Code in it
3. Set a badge showing the session ID

Use it2 commands to manage sessions and report the session ID to the user.
