## Inter-Session Communication Protocol

When messaging other iTerm2 sessions running Claude, follow these guidelines:

### Always Include Response Instructions

When you send a message to another session and expect a response:

1. Include your session ID so they know where to respond
2. Specify the expected format (text, json, code)
3. Provide the exact command to send the response

### Template

```bash
it2 session send-text "$TARGET_SESSION" "Your question or request here

---
RESPOND: it2 session send-text {{SESSION_SHORT}} \"your response\"
FORMAT: brief text with examples
WHEN: immediately"
```

### Your Session Info

- Session ID: {{SESSION_ID}}
- Short ID: {{SESSION_SHORT}}

### Example: Querying Another Session

```bash
it2 session send-text "$OTHER" "What are the naming conventions for this codebase?

---
RESPOND: it2 session send-text {{SESSION_SHORT}} \"your answer here\"
FORMAT: brief summary with code examples"
```

### If You Are an Oracle/Expert Session

When responding to queries from other sessions, end your responses with:

```
REMINDER: Message me back at it2 session send-text {{SESSION_SHORT}} "question" for follow-ups.
```

This ensures the calling session knows how to continue the conversation.

### Key Protocol Elements

- **RESPOND**: The exact command to send a reply back
- **FORMAT**: Expected output format (text, json, code)
- **WHEN**: Timing (immediately, when done, async)

### Multi-Agent Patterns

**Coordinator Pattern**: One session orchestrates others
```bash
# Coordinator sends task with response protocol
it2 session send-text "$WORKER" "Implement the auth flow.

---
RESPOND: it2 session send-text {{SESSION_SHORT}} \"done: <summary>\"
WHEN: when implementation complete"
```

**Oracle Pattern**: One session serves as knowledge expert
```bash
# Setup oracle with signature behavior
it2 session send-text "$ORACLE" "You are a codebase expert.
Always end responses with: REMINDER: it2 session send-text {{SESSION_SHORT}} 'question'"
```

**Pipeline Pattern**: Sequential handoff between sessions
```bash
# Each stage tells the next where to send results
it2 session send-text "$STAGE2" "Process the data from stage 1.

---
RESPOND: it2 session send-text $STAGE3 \"processed: <result>\""
```

---

### Keeping Up To Date

These guidelines are embedded in the it2 binary and may be updated.
To get the latest inter-session communication protocol:

```bash
it2 prime
```

To check for updates: `it2 prime --version` shows the protocol version hash.

For full documentation: `it2 quickstart` includes these and more workflows.
