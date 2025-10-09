---
name: part
description: Close a Claude Code session (like IRC /part #channel)
argument-hint: [session-id-or-name]
auto-approve:
  - Bash(it2 session send-key:*)
  - Bash(it2 session close:*)
  - Bash(it2 session list:*)
---

Close/exit a Claude Code session gracefully.

Arguments: [session-id-or-name] (optional, defaults to current session)

1. Identify the target session (use ITERM_SESSION_ID if no argument provided)
2. Send Ctrl+D to gracefully exit Claude: `it2 session send-key <session-id> "^D"`
3. Wait briefly for Claude to exit
4. Optionally close the session: `it2 session close <session-id>`

Report which session was closed to the user.
