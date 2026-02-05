# Session Hierarchy Navigation

iTerm2 organizes sessions in a hierarchy: **Application → Windows → Tabs → Sessions (Panes)**

## Hierarchy Commands

### Direct Lookups

```bash
# Get parent tab
it2 session lookup tab "$SID"

# Get parent window
it2 session lookup window "$SID"

# Get split root (top of split tree)
it2 session lookup split-root "$SID"
```

### Relationship Queries

```bash
# Siblings (same parent in split tree)
it2 session lookup siblings "$SID"

# All sessions in same tab
it2 session lookup descendants "$(it2 session lookup tab $SID)"
```

### Spatial Navigation

```bash
# Pane above
it2 session lookup above "$SID"

# Pane below
it2 session lookup below "$SID"

# Pane to the left
it2 session lookup before "$SID"

# Pane to the right
it2 session lookup after "$SID"
```

## Scope-Based Listing

```bash
# All sessions in current tab
it2 session list --scope tab

# All sessions in current window
it2 session list --scope window

# All sessions globally
it2 session list --scope all
```

## JSON Processing

```bash
# Get full hierarchy as JSON
it2 session list --format json | jq 'group_by(.tab_id) | .[] | {
  tab: .[0].tab_id,
  sessions: [.[] | .session_id]
}'
```

## Lineage Traversal

```bash
# Get full lineage (ancestors + descendants)
it2 session lookup lineage "$SID"

# Just ancestors
it2 session lookup ancestors "$SID"

# Just descendants
it2 session lookup descendants "$SID"
```
