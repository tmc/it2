# cmdutil Refactoring Opportunities

This document outlines potential refactoring opportunities for the `internal/cmdutil` package.

## Completed Refactoring

✅ **URL Parameter Cleanup** (2024-09-28)
- Removed `wsURL` parameter from `cmdutil.ConnectClient` calls
- Updated 98+ command files to use simplified API
- Centralized URL handling in `ConnectClient` function

✅ **Helper Function Addition** (2024-09-28)
- Added `GetTimeout(cmd)` helper function
- Added `GetFormat(cmd)` helper function
- Added `CreateContextFromCommand(cmd)` convenience function

## Future Refactoring Opportunities

### 1. Replace Common GetFlags Patterns

Many commands still use the pattern:
```go
_, timeout, format := cmdutil.GetFlags(cmd)
ctx, cancel := cmdutil.CreateContext(timeout)
```

This could be simplified to:
```go
ctx, cancel := cmdutil.CreateContextFromCommand(cmd)
format := cmdutil.GetFormat(cmd)
```

**Files to update**: ~60 command files
**Estimated effort**: Medium (1-2 hours)
**Benefits**: Cleaner, more readable code

### 2. Session Resolution Helper

Many session commands have the pattern:
```go
c, err := cmdutil.ConnectClient(ctx)
if err != nil {
    return fmt.Errorf("failed to connect: %w", err)
}
defer c.Close()

sessionID, err = c.ResolveSessionID(ctx, sessionID)
if err != nil {
    return fmt.Errorf("failed to resolve session ID: %w", err)
}
```

Could be simplified to:
```go
c, sessionID, err := cmdutil.ConnectAndResolveSession(ctx, sessionID)
if err != nil {
    return err
}
defer c.Close()
```

**Files to update**: ~30 session command files
**Estimated effort**: Medium (2-3 hours)
**Benefits**: Reduced boilerplate, consistent error handling

### 3. StandardCommand Migration

Some files use raw cobra commands instead of the `StandardCommand` pattern:
- More consistent error handling
- Built-in formatting support
- Reduced boilerplate

**Files to update**: Various command files still using raw cobra
**Estimated effort**: Large (4-6 hours)
**Benefits**: Consistency, better error handling, less code duplication

### 4. Global Flag Access in ConnectClient

Currently `ConnectClient` hardcodes `ws://localhost:1912`. It could read the global `--url` flag:

```go
func ConnectClient(ctx context.Context) (*client.Client, error) {
    // Read from global flags or environment
    wsURL := getGlobalURL() // reads --url flag
    c := client.New(wsURL)
    // ...
}
```

**Estimated effort**: Small (30 minutes)
**Benefits**: Enables proxy support without API changes

### 5. Template-Based Command Generation

For commands with similar patterns, we could use more templates:

```go
cmd := cmdutil.NewCommandFromTemplate(cmdutil.SessionCommandTemplate{
    Use: "get-info [<session-id>]",
    Short: "Get session information",
    SessionFunc: func(sc *StandardCommand, sessionID string) error {
        // implementation
    },
})
```

**Files to benefit**: Most command files
**Estimated effort**: Large (6-8 hours)
**Benefits**: Massive reduction in boilerplate, consistency

## Implementation Priority

1. **High Priority**: Replace common GetFlags patterns (easy wins)
2. **Medium Priority**: Session resolution helper (good impact)
3. **Low Priority**: Template generation (large effort, long-term benefit)

## Notes

- Always maintain backward compatibility during refactoring
- Test thoroughly after each change
- Consider batching related changes into single commits
- Update documentation after major refactoring