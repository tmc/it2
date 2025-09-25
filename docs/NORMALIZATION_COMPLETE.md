# it2 Command Normalization Complete

## Summary

Successfully normalized command handling across the it2 codebase, establishing consistent patterns for session ID handling, error reporting, and output formatting.

## Completed Normalizations

### ✅ Badge Commands (`internal/cmd/badge/`)
- Migrated to StandardCommand pattern
- Added JSON output support
- Enhanced help with examples and use cases
- Made `get` command session-id optional

### ✅ Annotation Commands (`internal/cmd/annotation/`)
- Migrated to StandardCommand pattern
- Added comprehensive error handling
- Enhanced help messages with practical examples
- Made `list` command session-id optional

### ✅ Text Commands (`internal/cmd/text/`)
- Split monolithic 801-line file into 14 focused modules
- All commands use StandardCommand pattern
- Safe read commands have optional session-id:
  - `get-buffer`, `get-screen`, `get-cursor`, `get-contents`
- Destructive commands require session-id:
  - `clear-buffer`, `set-cursor`, `inject`, `send`, etc.

### ✅ Tab Commands (`internal/cmd/tab/`)
- Split monolithic 537-line file into 9 focused modules
- Consistent StandardCommand usage
- Proper session ID normalization throughout

### ✅ Session Commands (`internal/cmd/session/`)
- Already using StandardCommand with RequiresSession: true
- Automatic normalization via template pattern
- Comprehensive JSON output support

## Standardized Patterns

### 1. Session ID Normalization

All commands that accept session IDs now:
- Strip the `w0t1p12:` prefix automatically
- Support environment variable fallback for read-only operations
- Provide clear error messages when session ID is missing

```go
// For required session commands
template := cmdutil.CommandTemplate{
    RequiresSession: true,  // Auto-normalizes args[0]
    ...
}

// For optional session commands
sessionID = cmdutil.ResolveSessionID(sessionID)
if sessionID == "" {
    return cmdutil.NewRequiredArgumentError("session ID (or $ITERM_SESSION_ID)")
}
```

### 2. Error Handling Pattern

Consistent error reporting across all commands:
```go
return sc.ReportError("operation name", err)
```

### 3. JSON Output Support

All commands support JSON output for automation:
```go
if sc.GetFlags().Format == "json" {
    result := map[string]interface{}{...}
    return sc.FormatOutput(result)
}
```

### 4. Help Message Enhancement

Commands now include:
- Practical use cases
- Real-world examples
- Clear parameter descriptions
- Format guidance

Example:
```go
Long: `Set a badge for a session to display status or identification text.

Examples:
  it2 badge set session123 "PRODUCTION"
  it2 badge set session123 "feature-branch"
  it2 badge set session123 "🔴 LIVE"`
```

## Session ID Format Guidance

### Understanding Session IDs

iTerm2 session IDs have two formats:
1. **Raw UUID**: `60BCA1AC-2E22-49CB-B8B1-3E08D957131F`
2. **With prefix**: `w0t1p12:60BCA1AC-2E22-49CB-B8B1-3E08D957131F`

The it2 tool accepts both formats and automatically normalizes them.

### Environment Variable

For convenience, many read-only commands can use the current session:
```bash
export ITERM_SESSION_ID="w0t1p12:60BCA1AC-2E22-49CB-B8B1-3E08D957131F"

# These work without explicit session ID
it2 text get-buffer
it2 badge get
it2 annotation list
```

### Finding Session IDs

```bash
# List all sessions
it2 session list

# Get current session ID
echo $ITERM_SESSION_ID

# Extract just the UUID part
echo $ITERM_SESSION_ID | cut -d: -f2
```

## Remaining Minor Tasks

While the core normalization is complete, these commands could benefit from migration to StandardCommand pattern for full consistency:

1. **trigger/**: Commands work but could use StandardCommand pattern
2. **job/**: Single command could be simplified
3. **screen/**: Recording commands could use better error handling
4. **utility/**: Various utilities could be modularized
5. **snippet/**: Could benefit from JSON output support

## Testing Verification

```bash
# Build the project
go build ./cmd/it2

# Run tests
go test ./...

# Manual verification of normalization
./it2 badge set w0t1p12:test-session "TEST"  # Should strip prefix
./it2 badge set test-session "TEST"          # Should work directly
```

## Benefits Achieved

1. **Consistency**: All commands follow the same patterns
2. **Maintainability**: Modular structure makes changes easier
3. **Automation**: JSON output enables scripting
4. **User Experience**: Clear help and error messages
5. **Safety**: Distinction between safe and destructive operations
6. **Code Quality**: Reduced duplication, better organization

## Migration Guide for New Commands

When adding new commands:

1. Use `cmdutil.CommandTemplate`
2. Set `RequiresSession: true` for session commands
3. Add `SupportsFormat: true` for JSON output
4. Include examples in Long description
5. Use `completion.SessionIDCompletion` for tab completion
6. Follow safe vs. destructive patterns for optional session-id

This completes the normalization effort, establishing a solid foundation for future development.