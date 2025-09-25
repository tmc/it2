# iTerm2 CLI Normalization Plan

## Executive Summary

This document outlines a comprehensive normalization plan for the iTerm2 CLI codebase. Based on analysis of 96 Go files across 34 command directories, significant inconsistencies exist in command structure, error handling, flag management, and utility usage. This plan provides a priority-ordered roadmap to standardize the codebase following Go best practices.

## Current State Analysis

### Statistics
- **Total Go files**: 96 files across 34 command directories
- **Files using cmdutil utilities**: 32 (34%)
- **Files NOT using cmdutil utilities**: 61 (66%)
- **Commands with > 500 lines**: 5 files (indicating monolithic structures)
- **Command factories found**: 197 `newCommand` functions

### Key Inconsistencies Identified

1. **Command Structure**: Mix of monolithic files (tab.go: 537 lines) and well-modularized commands (session package)
2. **Flag Handling**: Inconsistent usage of `cmdutil.GetFlags()` vs `cmdutil.GetExtendedFlags()` vs manual flag extraction
3. **Error Patterns**: Varied error formatting and context handling
4. **Client Connection**: Some commands manually handle connection logic vs using `cmdutil.ConnectClient()`
5. **Response Formatting**: Mixed usage of `formatting` package patterns

## Priority-Ordered Normalization Tasks

### Phase 1: Foundation (High Priority - 2 weeks)

#### Task 1.1: Expand cmdutil Package
**Files to modify**: `/internal/cmdutil/cmdutil.go`

Add standardized utilities for common patterns:

```go
// StandardCommand provides the foundation for all commands
type StandardCommand struct {
    cmd     *cobra.Command
    client  *client.Client
    ctx     context.Context
    cancel  context.CancelFunc
    flags   *StandardFlags
}

type StandardFlags struct {
    WsURL       string
    Timeout     time.Duration
    Format      string
    Columns     []string
    SortBy      string
    SortReverse bool
}

// NewStandardCommand creates a command with standard setup
func NewStandardCommand(use, short string) *StandardCommand

// ExecuteWithClient runs a command with standard connection handling
func (sc *StandardCommand) ExecuteWithClient(fn func(*client.Client, context.Context) error) error

// AddStandardFlags adds common flags to commands
func (sc *StandardCommand) AddStandardFlags()

// AddExtendedFlags adds sorting/filtering flags
func (sc *StandardCommand) AddExtendedFlags()
```

#### Task 1.2: Create Command Templates
**New files to create**:
- `/internal/cmdutil/templates.go` - Standard command templates
- `/internal/cmdutil/validation.go` - Input validation utilities
- `/internal/cmdutil/errors.go` - Standardized error handling

```go
// CommandTemplate provides standard command structure
type CommandTemplate struct {
    Use              string
    Short            string
    Long             string
    Args             cobra.PositionalArgs
    RequiresClient   bool
    RequiresSession  bool
    SupportsSorting  bool
    SupportsColumns  bool
    RunE             func(*StandardCommand, []string) error
}

// NewCommandFromTemplate creates standardized cobra command
func NewCommandFromTemplate(template CommandTemplate) *cobra.Command
```

#### Task 1.3: Standardize Error Handling
**Pattern to implement across all commands**:

```go
// StandardError provides consistent error formatting
type StandardError struct {
    Operation string
    Cause     error
    Details   map[string]interface{}
}

func (e *StandardError) Error() string {
    return fmt.Sprintf("failed to %s: %w", e.Operation, e.Cause)
}

// Common error constructors
func NewConnectionError(cause error) error
func NewValidationError(field, reason string) error
func NewNotFoundError(resource, id string) error
```

### Phase 2: Monolithic File Refactoring (High Priority - 3 weeks)

#### Task 2.1: Break Up Large Files
**Priority order by file size and complexity**:

1. **`/internal/cmd/utility/utility.go` (1276 lines)**
   - Split into separate files: `cursor.go`, `timestamp.go`, `shell.go`, `password.go`, `ssh.go`, `touchbar.go`
   - Each sub-command gets its own file following session package pattern

2. **`/internal/cmd/text/text.go` (800 lines)**
   - Split into: `text_insert.go`, `text_color.go`, `text_cursor.go`, `text_scroll.go`

3. **`/internal/cmd/app/app.go` (757 lines)**
   - Split into: `app_list.go`, `app_focus.go`, `app_hide.go`, `app_quit.go`

4. **`/internal/cmd/preference/preference.go` (641 lines)**
   - Split into: `preference_get.go`, `preference_set.go`, `preference_list.go`

5. **`/internal/cmd/tab/tab.go` (536 lines)**
   - Already well-structured but could split: `tab_list.go`, `tab_create.go`, `tab_manage.go`

**Refactoring Pattern**:
```go
// Before: monolithic file
func NewCommand() *cobra.Command {
    cmd := &cobra.Command{...}
    cmd.AddCommand(newSubCmd1())
    cmd.AddCommand(newSubCmd2())
    // 500+ lines of implementation
}

// After: modularized approach
// main.go
func NewCommand() *cobra.Command {
    cmd := &cobra.Command{...}
    cmd.AddCommand(newSubCmd1Command())  // in subcmd1.go
    cmd.AddCommand(newSubCmd2Command())  // in subcmd2.go
    return cmd
}
```

### Phase 3: Standardization Implementation (Medium Priority - 4 weeks)

#### Task 3.1: Migrate to StandardCommand Pattern
**Files requiring migration** (61 files not using cmdutil):

Priority order:
1. Session commands (already partially using cmdutil)
2. Window commands
3. Tab commands
4. High-usage utility commands
5. Remaining commands

**Migration template**:
```go
// Before
func newListCommand() *cobra.Command {
    return &cobra.Command{
        Use:   "list",
        Short: "List items",
        RunE: func(cmd *cobra.Command, args []string) error {
            // Manual flag extraction
            wsURL, _ := cmd.Flags().GetString("url")
            timeout, _ := cmd.Flags().GetDuration("timeout")

            // Manual client connection
            c := client.New(wsURL)
            ctx, cancel := context.WithTimeout(context.Background(), timeout)
            defer cancel()

            if err := c.Connect(ctx); err != nil {
                return err
            }
            // Implementation...
        },
    }
}

// After
func newListCommand() *cobra.Command {
    template := cmdutil.CommandTemplate{
        Use:             "list",
        Short:           "List items",
        RequiresClient:  true,
        SupportsSorting: true,
        RunE: func(sc *cmdutil.StandardCommand, args []string) error {
            // Standard client and context already available
            // Implementation using sc.client, sc.ctx...
        },
    }
    return cmdutil.NewCommandFromTemplate(template)
}
```

#### Task 3.2: Standardize Formatting Patterns
**Files to update**: All commands using `formatting` package

Implement consistent formatting interface:
```go
type FormattingOptions struct {
    Format      string
    Columns     []string
    SortBy      string
    SortReverse bool
}

// Standard formatting method for all list commands
func FormatAndOutput(data interface{}, opts FormattingOptions) error
```

### Phase 4: Enhanced Features (Medium Priority - 2 weeks)

#### Task 4.1: Unified Flag Management
Create flag management system:

```go
// Flag registry for consistent behavior
type FlagRegistry struct {
    flags map[string]FlagDefinition
}

type FlagDefinition struct {
    Name        string
    Type        FlagType
    Default     interface{}
    Description string
    Validator   func(interface{}) error
}

// Standard flag sets
var (
    ConnectionFlags = []FlagDefinition{...}
    FormattingFlags = []FlagDefinition{...}
    SessionFlags    = []FlagDefinition{...}
)
```

#### Task 4.2: Plugin System Integration
**Files to update**: Commands using plugins

Standardize plugin integration:
```go
// Standard plugin application
func ApplyPlugins(ctx context.Context, data interface{}, pluginType PluginType) error {
    registry := plugins.NewRegistry()
    if err := registry.DiscoverAndRegister(); err != nil {
        return err
    }
    return registry.ApplyEnrichment(ctx, data, pluginType)
}
```

### Phase 5: Testing and Validation (Low Priority - 2 weeks)

#### Task 5.1: Comprehensive Test Coverage
**Files to create/update**: Test files for all modified commands

Test coverage targets:
- Command construction: 100%
- Flag handling: 100%
- Error paths: 90%
- Happy paths: 95%

#### Task 5.2: Integration Testing
Create end-to-end tests validating:
- Command execution flows
- Flag parsing consistency
- Error message consistency
- Output format consistency

## New Utility Functions to be Created

### `/internal/cmdutil/validation.go`
```go
// Input validation utilities
func ValidateSessionID(id string) error
func ValidateWindowID(id string) error
func ValidateTabID(id string) error
func ValidateFormat(format string) error
func ValidateTimeout(timeout time.Duration) error

// Resource existence validation
func ValidateSessionExists(ctx context.Context, client *client.Client, id string) error
func ValidateWindowExists(ctx context.Context, client *client.Client, id string) error
func ValidateTabExists(ctx context.Context, client *client.Client, id string) error
```

### `/internal/cmdutil/helpers.go`
```go
// Common command helpers
func ParseColumnList(columns string) []string
func BuildSortOptions(sortBy string, reverse bool) SortOptions
func FormatDuration(d time.Duration) string
func FormatTimestamp(t time.Time) string

// Resource ID normalization
func NormalizeResourceID(id, resourceType string) string
func ParseResourceReference(ref string) (resourceType, id string, err error)
```

### `/internal/cmdutil/output.go`
```go
// Standardized output functions
func PrintSuccess(format string, args ...interface{})
func PrintError(err error)
func PrintWarning(format string, args ...interface{})
func PrintInfo(format string, args ...interface{})

// Progress indication
type ProgressReporter interface {
    Start(total int)
    Update(current int, message string)
    Finish(message string)
}
```

## Command Structure Refactoring Plan

### Current Inconsistent Patterns

1. **Mixed file organization**:
   - Some commands: single monolithic file
   - Others: well-modularized (session package)
   - Inconsistent naming: `cmd.go` vs `command.go` vs package name

2. **Inconsistent constructor patterns**:
   - `NewCommand()` vs `newCommand()`
   - Mixed visibility rules
   - Inconsistent parameter passing

### Target Structure

**Standard directory layout**:
```
internal/cmd/[category]/
├── [category].go          # Main command with subcommand registration
├── [category]_list.go     # List operations
├── [category]_create.go   # Create operations
├── [category]_delete.go   # Delete operations
├── [category]_get.go      # Get/read operations
├── [category]_set.go      # Set/update operations
├── [category]_test.go     # Tests
└── doc.go                 # Package documentation
```

**Standard command factory**:
```go
// Each [category].go file
func NewCommand() *cobra.Command {
    cmd := &cobra.Command{
        Use:   "category",
        Short: "Manage iTerm2 [category]",
        Long:  "Commands for managing iTerm2 [category] operations",
    }

    // Register subcommands
    cmd.AddCommand(newListCommand())
    cmd.AddCommand(newCreateCommand())
    cmd.AddCommand(newDeleteCommand())

    return cmd
}
```

### Files Requiring Structure Changes

**High Priority** (monolithic → modular):
1. `/internal/cmd/utility/utility.go` → 7 separate files
2. `/internal/cmd/text/text.go` → 4 separate files
3. `/internal/cmd/app/app.go` → 4 separate files
4. `/internal/cmd/preference/preference.go` → 3 separate files

**Medium Priority** (structure improvements):
1. `/internal/cmd/profile/profile.go` → 5 separate files
2. `/internal/cmd/prompt/prompt.go` → 3 separate files
3. `/internal/cmd/snippet/snippet.go` → 4 separate files

**Low Priority** (minor cleanup):
1. Commands already well-structured but missing consistency
2. Test file organization
3. Documentation standardization

## Standardization Patterns to Apply

### 1. Error Handling Pattern
```go
// Standard error handling across all commands
func (sc *StandardCommand) handleError(operation string, err error) error {
    if err == nil {
        return nil
    }

    // Add context and operation info
    return &StandardError{
        Operation: operation,
        Cause:     err,
        Details:   map[string]interface{}{
            "command": sc.cmd.Name(),
            "args":    sc.cmd.Args,
        },
    }
}
```

### 2. Client Connection Pattern
```go
// Standardized client lifecycle management
func (sc *StandardCommand) withClient(fn func(*client.Client, context.Context) error) error {
    c, err := cmdutil.ConnectClient(sc.ctx, sc.flags.WsURL)
    if err != nil {
        return sc.handleError("connect to iTerm2", err)
    }
    defer c.Close()

    return fn(c, sc.ctx)
}
```

### 3. Resource Validation Pattern
```go
// Consistent resource validation
func validateResourceArgs(cmd *cobra.Command, args []string, resourceType string) error {
    if len(args) == 0 {
        return NewValidationError(resourceType+"_id", "required")
    }

    id := args[0]
    switch resourceType {
    case "session":
        return ValidateSessionID(id)
    case "window":
        return ValidateWindowID(id)
    case "tab":
        return ValidateTabID(id)
    default:
        return NewValidationError("resource_type", "unknown type: "+resourceType)
    }
}
```

### 4. Response Formatting Pattern
```go
// Standard success/failure reporting
func (sc *StandardCommand) reportResult(operation string, success bool, details interface{}) error {
    if success {
        PrintSuccess("Successfully %s", operation)
        if details != nil && sc.flags.Format != "quiet" {
            return FormatAndOutput(details, FormattingOptions{
                Format: sc.flags.Format,
                Columns: sc.flags.Columns,
                SortBy: sc.flags.SortBy,
                SortReverse: sc.flags.SortReverse,
            })
        }
        return nil
    }

    return fmt.Errorf("failed to %s", operation)
}
```

## Testing Strategy

### 1. Unit Test Standards
- **Coverage target**: 90%+ for modified files
- **Test structure**: One test file per command file
- **Mock strategy**: Use interfaces for external dependencies

```go
// Standard test structure
func TestCommandName(t *testing.T) {
    tests := []struct {
        name        string
        args        []string
        flags       map[string]interface{}
        mockSetup   func(*mocks.Client)
        expectError bool
        expectOutput string
    }{
        // Test cases
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            // Test implementation
        })
    }
}
```

### 2. Integration Test Plan
- **Command execution**: Test full command pipelines
- **Flag combinations**: Test all supported flag combinations
- **Error scenarios**: Test network failures, invalid inputs
- **Output formats**: Validate JSON, YAML, table outputs

### 3. Regression Test Suite
- **Baseline establishment**: Capture current command outputs
- **Behavior validation**: Ensure normalization doesn't change user-visible behavior
- **Performance testing**: Ensure no performance regressions

## Implementation Timeline

### Week 1-2: Foundation (Phase 1)
- [ ] Expand cmdutil package with standard patterns
- [ ] Create command templates and error handling
- [ ] Implement validation utilities
- [ ] Create comprehensive test framework

### Week 3-5: Monolithic Refactoring (Phase 2)
- [ ] Split utility.go into 7 files
- [ ] Split text.go into 4 files
- [ ] Split app.go into 4 files
- [ ] Split preference.go into 3 files
- [ ] Update imports and test files

### Week 6-9: Standardization (Phase 3)
- [ ] Migrate 61 non-cmdutil files to StandardCommand pattern
- [ ] Implement consistent formatting patterns
- [ ] Standardize error handling across all commands
- [ ] Update all test files

### Week 10-11: Enhanced Features (Phase 4)
- [ ] Implement unified flag management
- [ ] Standardize plugin integration
- [ ] Add advanced validation features
- [ ] Performance optimizations

### Week 12-13: Testing and Validation (Phase 5)
- [ ] Complete test coverage for all modified files
- [ ] Run integration test suite
- [ ] Performance regression testing
- [ ] Documentation updates

## Success Metrics

### Code Quality Metrics
- **Cyclomatic complexity**: Reduce average complexity by 30%
- **Code duplication**: Eliminate 90%+ of identified duplication
- **File size**: No single file > 500 lines (except generated)
- **Test coverage**: Achieve 90%+ coverage on modified files

### Consistency Metrics
- **Error message format**: 100% consistent error formatting
- **Flag usage**: 100% usage of standard flag utilities
- **Command structure**: 100% compliance with standard patterns
- **Documentation**: 100% commands have consistent doc strings

### Maintainability Metrics
- **New command creation**: Reduce time by 70% with templates
- **Bug fix time**: Improve average time by 40% with better structure
- **Feature addition**: Reduce cross-command feature addition time by 60%

## Risk Mitigation

### Breaking Changes Risk
- **Mitigation**: Preserve all public APIs during refactoring
- **Validation**: Comprehensive regression test suite
- **Rollback plan**: Git branch strategy for easy rollback

### Performance Risk
- **Mitigation**: Benchmark existing command performance before changes
- **Validation**: Performance regression tests in CI
- **Optimization**: Profile and optimize any performance regressions

### Complexity Risk
- **Mitigation**: Incremental rollout, phase-based approach
- **Validation**: Code review requirements for all changes
- **Documentation**: Comprehensive implementation documentation

This normalization plan provides a structured approach to modernizing the iTerm2 CLI codebase while maintaining backward compatibility and improving maintainability. The phased approach allows for incremental progress with validation at each step.