---
name: claude-plugin-validate-all
description: Discover and validate all Claude Code plugin and marketplace manifests in the repository
---

Find all `.claude-plugin` directories in the repository and validate their manifests.

**Discovery Logic:**

1. Find git repository root
2. Search for all `.claude-plugin` directories
3. For each directory found:
   - Check for `plugin.json` or `marketplace.json`
   - Validate the manifest
   - Report results
4. Provide summary of validation results

**Usage:**

```bash
/claude-plugin-validate-all
```

**Implementation:**

```bash
# Function to validate all discovered manifests
validate_all_discovered() {
    local repo_root
    local total=0
    local passed=0
    local failed=0

    # Find git root
    repo_root=$(git rev-parse --show-toplevel 2>/dev/null || pwd)

    echo "Searching for manifests in: $repo_root"
    echo ""

    # Find all .claude-plugin directories
    while IFS= read -r plugin_dir; do
        local parent_dir=$(dirname "$plugin_dir")
        local manifest=""

        # Determine which manifest exists
        if [ -f "$plugin_dir/plugin.json" ]; then
            manifest="plugin"
            echo "=== Plugin: $parent_dir ==="
        elif [ -f "$plugin_dir/marketplace.json" ]; then
            manifest="marketplace"
            echo "=== Marketplace: $parent_dir ==="
        else
            echo "=== Skipping $plugin_dir (no manifest found) ==="
            continue
        fi

        total=$((total + 1))

        # Validate
        if claude plugin validate "$parent_dir" 2>&1; then
            passed=$((passed + 1))
        else
            failed=$((failed + 1))
        fi
        echo ""
    done < <(find "$repo_root" -type d -name ".claude-plugin" 2>/dev/null)

    # Summary
    echo "========================================"
    echo "Validation Summary:"
    echo "  Total manifests: $total"
    echo "  Passed: $passed ✔"
    echo "  Failed: $failed ✘"
    echo "========================================"

    # Return non-zero if any failed
    [ "$failed" -eq 0 ]
}

validate_all_discovered
```

**Example Output:**

```
Searching for manifests in: /path/to/repo

=== Plugin: integrations/claude-code/plugins/it2-core ===
Validating plugin manifest: /path/to/repo/integrations/claude-code/plugins/it2-core/.claude-plugin/plugin.json
✔ Validation passed

=== Plugin: integrations/claude-code/plugins/it2-claude-automation ===
Validating plugin manifest: /path/to/repo/integrations/claude-code/plugins/it2-claude-automation/.claude-plugin/plugin.json
✔ Validation passed

=== Marketplace: .claude-plugin ===
Validating marketplace manifest: /path/to/repo/.claude-plugin/marketplace.json
✔ Validation passed

========================================
Validation Summary:
  Total manifests: 3
  Passed: 3 ✔
  Failed: 0 ✘
========================================
```

**With Failures:**

```
Searching for manifests in: /path/to/repo

=== Plugin: plugins/broken-plugin ===
Validating plugin manifest: /path/to/repo/plugins/broken-plugin/.claude-plugin/plugin.json
✘ Found 1 error:
  ❯ agents: Invalid input
✘ Validation failed

========================================
Validation Summary:
  Total manifests: 1
  Passed: 0 ✔
  Failed: 1 ✘
========================================
```

**Use Cases:**

- **Pre-commit validation**: Ensure all plugins are valid before committing
- **CI/CD checks**: Validate entire marketplace in pipeline
- **Bulk updates**: After structural changes, validate everything
- **Quality assurance**: Regular checks during development

**Notes:**

- Searches entire repository tree
- Skips `.claude-plugin` directories without manifests
- Provides detailed per-plugin validation output
- Returns non-zero exit code if any validation fails
- Can be used in scripts and automation
