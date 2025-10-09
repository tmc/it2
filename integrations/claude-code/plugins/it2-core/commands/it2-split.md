---
name: it2-split
description: Create an iTerm2 split session
auto-approve:
  - Bash(it2 session split:*)
  - Bash(it2 session send-text:*)
---

Create a new iTerm2 split session using `it2 session split` and navigate to the current directory.

**Steps:**

1. Run `it2 session split` to create the split (auto-detects best direction)
2. Send `cd {{cwd}}` to the new session to navigate to the current working directory

**Usage:**

Run the command and it will:
- Create a split session
- Navigate the new session to your current directory ({{cwd}})

**Notes:**
- The split direction (horizontal/vertical) is automatically chosen based on available space
- Use `it2 session split --vertical` or `--horizontal` to force a specific direction
- The new session inherits your shell environment
