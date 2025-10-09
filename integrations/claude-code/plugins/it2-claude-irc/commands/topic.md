---
name: topic
description: Set or view the "topic" (badge/title) of a Claude session (like IRC /topic)
argument-hint: [session-id] [new-topic]
auto-approve:
  - Bash(it2 session list:*)
  - Bash(it2 session set-badge:*)
  - Bash(it2 session set-title:*)
  - Bash(it2 session get-title:*)
---

Set or view the badge and title for a Claude Code session.

Arguments: [session-id] [new-topic] (both optional)

If no arguments:
- Show the current session's title and badge

If only session-id:
- Show that session's title and badge

If both arguments:
- Set the session's badge and title to the new topic
- Badge format: session ID (first 8 chars) on line 1, topic on line 2
- Title: topic

Example: `/topic ABC123 Backend Development`
