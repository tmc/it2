# it2-session-snapshot

Orchestrates session snapshots by discovering and running all `it2-session-snapshot-*` plugins.

## Usage

```bash
# Snapshot current session
it2-session-snapshot

# Snapshot specific session
it2-session-snapshot <session-id>

# Verbose output
it2-session-snapshot -verbose

# Run plugins sequentially (default is parallel)
it2-session-snapshot -parallel=false
```

## How it Works

1. Discovers all executables in PATH matching `it2-session-snapshot-*`
2. Runs each plugin with the session ID as argument
3. Plugins store snapshots in `~/.it2/sessions/$SESSION_ID/snapshots/`

## Built-in Plugins

### it2-session-snapshot-claude
Snapshots the `.claude` directory to a git repository.

Stores in: `~/.it2/sessions/$SESSION_ID/snapshots/claude/.git`

### it2-session-snapshot-filesystem
Captures filesystem state:
- Directory tree (via `tree`)
- File listing (via `ls -laR`)
- Git status and diff (if in git repo)
- Metadata (timestamp, cwd, hostname, user)

Stores in: `~/.it2/sessions/$SESSION_ID/snapshots/filesystem/`

## Creating Custom Snapshot Plugins

Create an executable named `it2-session-snapshot-<name>` that:
1. Accepts session ID as first argument (or uses `$ITERM_SESSION_ID`)
2. Stores snapshots in `~/.it2/sessions/$SESSION_ID/snapshots/<name>/`
3. Exits with 0 on success, non-zero on failure

Example:
```bash
#!/bin/bash
SESSION_ID="${1:-$ITERM_SESSION_ID}"
SNAPSHOT_DIR="$HOME/.it2/sessions/$SESSION_ID/snapshots/mydata"
mkdir -p "$SNAPSHOT_DIR"
# ... capture your data ...
```

Make it executable and place in PATH:
```bash
chmod +x it2-session-snapshot-mydata
cp it2-session-snapshot-mydata ~/bin/
```
