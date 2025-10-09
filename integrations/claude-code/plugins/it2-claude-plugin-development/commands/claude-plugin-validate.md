---
name: claude-plugin-validate
description: Validate Claude Code marketplace or plugin by discovering manifest files
---

Intelligently validate Claude Code manifests by discovering them in common locations.

**Discovery Logic:**

1. **Check for marketplace at repository root**: `../../.claude-plugin/marketplace.json` (relative to current directory)
2. **Check for local plugin**: `.claude-plugin/plugin.json` in current directory
3. **Check for marketplace in current directory**: `.claude-plugin/marketplace.json`
4. **If manifest found**: Validate it
5. **If none found**: Report where manifests were searched

**Usage:**

Simply run the command - it will discover and validate manifests automatically:
```bash
/claude-plugin-validate
```

**Implementation:**

```bash
# Function to find and validate manifests
validate_discovered() {
    local repo_root
    local current_dir="$(pwd)"

    # Try to find git root
    repo_root=$(git rev-parse --show-toplevel 2>/dev/null || echo "$current_dir")

    # 1. Check for marketplace at repo root
    if [ -f "$repo_root/.claude-plugin/marketplace.json" ]; then
        echo "Found marketplace at repository root:"
        claude plugin validate "$repo_root/.claude-plugin/marketplace.json"
        return $?
    fi

    # 2. Check for local plugin
    if [ -f "$current_dir/.claude-plugin/plugin.json" ]; then
        echo "Found plugin in current directory:"
        claude plugin validate "$current_dir"
        return $?
    fi

    # 3. Check for local marketplace
    if [ -f "$current_dir/.claude-plugin/marketplace.json" ]; then
        echo "Found marketplace in current directory:"
        claude plugin validate "$current_dir/.claude-plugin/marketplace.json"
        return $?
    fi

    # 4. Nothing found
    echo "No manifest found. Searched:"
    echo "  - $repo_root/.claude-plugin/marketplace.json (repository marketplace)"
    echo "  - $current_dir/.claude-plugin/plugin.json (local plugin)"
    echo "  - $current_dir/.claude-plugin/marketplace.json (local marketplace)"
    return 1
}

validate_discovered
```

**Example Output:**

When marketplace found at root:
```
Found marketplace at repository root:
Validating marketplace manifest: /path/to/repo/.claude-plugin/marketplace.json
✔ Validation passed
```

When plugin found locally:
```
Found plugin in current directory:
Validating plugin manifest: /path/to/plugin/.claude-plugin/plugin.json
✔ Validation passed
```

When nothing found:
```
No manifest found. Searched:
  - /path/to/repo/.claude-plugin/marketplace.json (repository marketplace)
  - /current/dir/.claude-plugin/plugin.json (local plugin)
  - /current/dir/.claude-plugin/marketplace.json (local marketplace)
```

**Use Cases:**

- Quick validation from anywhere in repository
- CI/CD validation scripts
- Pre-commit hooks
- Development workflow integration

**Notes:**

- Searches in priority order (marketplace root → local plugin → local marketplace)
- Uses git to find repository root if available
- Falls back to current directory if not in git repository
- Returns appropriate exit codes for scripting
