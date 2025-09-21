// Package auth provides commands for iTerm2 API authentication.
//
// Authentication is required to access iTerm2's WebSocket API. This package
// provides commands to:
//
//   - request: Request authentication from iTerm2 (interactive)
//   - check: Verify current authentication status
//
// The authentication process involves iTerm2 generating a cookie and key
// pair that must be provided with API requests. These credentials are
// typically stored in environment variables for subsequent use.
//
// Examples:
//
//	it2 auth request
//	it2 auth check
package auth
