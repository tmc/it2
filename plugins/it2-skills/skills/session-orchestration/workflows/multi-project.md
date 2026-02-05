# Multi-Project Development Workflow

Setting up a terminal environment for working on multiple related projects.

## Scenario

You're working on:
- A frontend React app
- A backend API
- A shared library

## Setup Script

```bash
#!/bin/bash

# Define projects
FRONTEND="/path/to/frontend"
BACKEND="/path/to/backend"
SHARED="/path/to/shared-lib"

# Start from clean tab
MAIN=$ITERM_SESSION_ID

# Create layout:
# ┌─────────────┬─────────────┐
# │  Frontend   │   Backend   │
# ├─────────────┴─────────────┤
# │         Shared            │
# └───────────────────────────┘

BACKEND_SID=$(it2 session split --vertical -q)
SHARED_SID=$(it2 session split --horizontal -q)

# Set up Frontend
it2 session set-badge "$MAIN" "$(echo $MAIN | cut -c1-8)\nFrontend"
it2 session send-text "$MAIN" "cd $FRONTEND && npm run dev"

# Set up Backend
it2 session set-badge "$BACKEND_SID" "$(echo $BACKEND_SID | cut -c1-8)\nBackend"
it2 session send-text "$BACKEND_SID" "cd $BACKEND && go run ."

# Set up Shared
it2 session set-badge "$SHARED_SID" "$(echo $SHARED_SID | cut -c1-8)\nShared"
it2 session send-text "$SHARED_SID" "cd $SHARED"

# Focus on shared for development
it2 session focus "$SHARED_SID"

echo "Development environment ready"
echo "Frontend: $MAIN"
echo "Backend:  $BACKEND_SID"
echo "Shared:   $SHARED_SID"
```

## Monitoring All Projects

```bash
# Watch all sessions for issues
it2 session watch --all

# Or tail logs from specific session
it2 session tail -f "$BACKEND_SID"
```

## Coordinated Commands

Send the same command to all project sessions:

```bash
for SID in "$MAIN" "$BACKEND_SID" "$SHARED_SID"; do
  it2 session send-text "$SID" "git pull"
done
```
