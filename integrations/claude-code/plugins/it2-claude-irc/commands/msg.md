---
name: msg
description: Send a message to a Claude Code session (like IRC /msg user message)
argument-hint: <session-id-or-name> <message>
auto-approve:
  - Bash(it2 session list:*)
  - Bash(it2 session send-text:*)
---

Send text to a specific Claude Code session.

Arguments: <session-id-or-name> <message>

1. Find the session by ID or name using `it2 session list`
2. Send the message text to that session using `it2 session send-text`
3. Report success to the user

Example: `/msg ABC123 analyze this codebase`
