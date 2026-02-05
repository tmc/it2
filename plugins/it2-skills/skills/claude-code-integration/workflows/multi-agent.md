# Multi-Agent Claude Code Workflow

Coordinating multiple Claude Code instances for complex tasks.

## Use Cases

- Parallel research and implementation
- Code review while coding
- Documentation alongside development
- Testing in parallel with fixes

## Setup Pattern

### Two-Agent Layout (Research + Implement)

```bash
#!/bin/bash

PROJECT_DIR="${1:-.}"
MAIN=$ITERM_SESSION_ID

# Create split for second agent
AGENT2=$(it2 session split --vertical -q)

# Badge sessions
it2 session set-badge "$MAIN" "$(echo $MAIN | cut -c1-8)\nResearch"
it2 session set-badge "$AGENT2" "$(echo $AGENT2 | cut -c1-8)\nImplement"

# Start agents
it2 session send-text "$MAIN" "cd '$PROJECT_DIR' && claude"
it2 session send-text "$AGENT2" "cd '$PROJECT_DIR' && claude"

echo "Two-agent setup complete"
echo "Research: $MAIN"
echo "Implement: $AGENT2"
```

## Coordination Patterns

### Sequential Handoff

Agent 1 completes research, then Agent 2 implements:

```bash
# Wait for first agent to finish
while ! it2 session get-screen "$AGENT1" | grep -qE '^\$|^%'; do
  sleep 2
done

# Start second agent with context
it2 session send-text "$AGENT2" "Implement based on the research in the other pane"
```

### Parallel Independent

Both agents work on different aspects:

```bash
# Research agent
it2 session send-text "$RESEARCH" "Research best practices for error handling in Go"

# Implementation agent
it2 session send-text "$IMPL" "Add logging to the API handlers"
```

## Monitoring Multiple Agents

### Status Dashboard

```bash
#!/bin/bash

while true; do
  clear
  echo "=== Claude Agent Status ==="

  for SID in "$AGENT1" "$AGENT2" "$AGENT3"; do
    LABEL=$(echo $SID | cut -c1-8)
    STATE=$(it2 session get-state "$SID" --format json 2>/dev/null | jq -r '.state // "unknown"')
    MODAL=$(it2 session has-modal "$SID" && echo "NEEDS INPUT" || echo "")

    printf "%-10s: %-10s %s\n" "$LABEL" "$STATE" "$MODAL"
  done

  sleep 2
done
```

### Alert on Modal

```bash
# Watch for any agent needing input
watch_agents() {
  while true; do
    for SID in "$@"; do
      if it2 session has-modal "$SID"; then
        it2 notification "Agent $(echo $SID | cut -c1-8) needs input"
        it2 session focus "$SID"
      fi
    done
    sleep 1
  done
}

watch_agents "$AGENT1" "$AGENT2" &
```

## Inter-Session Communication

When one Claude session messages another and expects a response, **always include explicit response instructions**. The receiving session needs to know:
1. **How to respond** - The exact command to send the response back
2. **What format to use** - Plain text, structured output, etc.
3. **When to respond** - Immediately, after completing work, etc.

### Response Hint Pattern

Always end messages to other sessions with a response hint:

```bash
# BAD - No response instructions
it2 session send-text "$OTHER" "What are the naming conventions for SwiftUI views?"

# GOOD - Explicit response instructions
it2 session send-text "$OTHER" "What are the naming conventions for SwiftUI views?

RESPOND via: it2 session send-text $MY_SESSION_ID \"your answer here\"
Format: Brief summary with examples"
```

### Template for Inter-Session Messages

```bash
MY_SID=$(echo $ITERM_SESSION_ID | cut -c1-8)

# Question to another session with response protocol
it2 session send-text "$TARGET" "$(cat <<EOF
[Your question or request here]

---
RESPOND: it2 session send-text $MY_SID "your response"
FORMAT: [text|json|code]
WHEN: [immediately|when done|async]
EOF
)"
```

### Oracle/Expert Pattern

When setting up a session as an oracle (knowledge source) for other sessions:

```bash
# Oracle setup - includes self-identification for callers
it2 session send-text "$ORACLE" "$(cat <<EOF
You are a codebase expert. When other sessions message you with questions:
1. Answer concisely with examples
2. End responses with: "REMINDER: Message me back at it2 session send-text $(echo $ORACLE | cut -c1-8) 'question' for follow-ups"
EOF
)"
```

## Best Practices

1. **Use distinct badges** - Always include session ID prefix
2. **Define clear roles** - Each agent should have a focused task
3. **Monitor for modals** - Agents may block waiting for permissions
4. **Coordinate directories** - Be careful with file conflicts
5. **Log sessions** - Save artifacts for review
6. **Include response hints** - Always tell the other session how to respond back

## Cleanup

```bash
# Close all agent sessions
for SID in "$AGENT1" "$AGENT2"; do
  # Send interrupt to Claude
  it2 session send-key "$SID" ctrl-c
  sleep 1
  # Exit shell
  it2 session send-text "$SID" "exit"
done
```
