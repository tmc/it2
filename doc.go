// Package it2 provides a Go client library and CLI for the iTerm2 API.
//
// it2 enables programmatic control of iTerm2 sessions, tabs, and windows
// over the Unix socket and WebSocket APIs. It handles authentication
// automatically and supports plugins for extensible session enrichment.
//
// # CLI Installation
//
//	go install github.com/tmc/it2/cmd/it2@latest
//
// # Library Usage
//
// The internal/client package provides the Go API (internal; use the CLI
// for external automation):
//
//	c := client.New("ws://localhost:1912")
//	if err := c.Connect(ctx); err != nil {
//		log.Fatal(err)
//	}
//	defer c.Close()
//
//	sessions, err := c.ListSessions(ctx)
//	if err != nil {
//		log.Fatal(err)
//	}
//	for _, s := range sessions {
//		fmt.Printf("%s %s\n", s.SessionID, s.WindowID)
//	}
//
//	err = c.SendText(ctx, sessions[0].SessionID, "echo hello\n")
//
// # Connection
//
// The client prefers the Unix socket at
// ~/Library/Application Support/iTerm2/private/socket
// and falls back to ws://localhost:1912.
//
// # See Also
//
//   - CLI documentation: https://github.com/tmc/it2
//   - iTerm2 API: https://iterm2.com/documentation-api.html
package it2
