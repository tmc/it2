// Package prime provides the inter-session communication guidelines
// for multi-agent workflows. The prime message is embedded and versioned
// to track changes across it2 updates.
package prime

import (
	"crypto/sha256"
	_ "embed"
	"encoding/hex"
	"fmt"
	"os"
	"strings"
)

//go:embed prime_message.md
var PrimeMessage string

// Version returns a short hash of the current prime message content.
// This is used to detect when the priming instructions have changed.
func Version() string {
	hash := sha256.Sum256([]byte(PrimeMessage))
	return hex.EncodeToString(hash[:8])
}

// Render outputs the prime message with session-specific placeholders filled in.
func Render(sessionID string) string {
	sidShort := sessionID
	if len(sidShort) > 8 {
		sidShort = sidShort[:8]
	}

	msg := PrimeMessage
	msg = strings.ReplaceAll(msg, "{{SESSION_ID}}", sessionID)
	msg = strings.ReplaceAll(msg, "{{SESSION_SHORT}}", sidShort)

	return msg
}

// RenderToStdout outputs the prime message to stdout with session info.
func RenderToStdout(sessionID string) {
	fmt.Print(Render(sessionID))
}

// GetSessionID returns the current session ID from environment.
func GetSessionID() string {
	return os.Getenv("ITERM_SESSION_ID")
}
