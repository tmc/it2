// Package cmdcore provides core command infrastructure for iTerm2 CLI operations.
//
// This package contains essential scaffolding used by most commands:
//   - Flag parsing and extraction (GetFlags, GetTimeout, GetFormat)
//   - Context creation with timeout handling
//   - Client connection management
//
// These are the most fundamental building blocks that commands depend on to:
//   1. Extract configuration from command flags
//   2. Create properly configured contexts
//   3. Establish connections to iTerm2
//
// This package is intentionally minimal and focused on core command infrastructure.
// Higher-level abstractions like command templates and validation live in separate packages.
package cmdcore
