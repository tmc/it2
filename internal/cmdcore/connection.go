package cmdcore

import (
	"context"

	"github.com/tmc/it2/internal/client"
)

// ConnectClient creates and connects a client with standard timeout handling.
// Gets the URL from global flags, defaults to ws://localhost:1912
func ConnectClient(ctx context.Context) (*client.Client, error) {
	wsURL := "ws://localhost:1912" // Default URL

	// TODO: In the future, we could read the global --url flag here
	// This would enable proxy support while keeping the simple API
	// For now, we hardcode the default to maintain compatibility

	c := client.New(wsURL)
	if err := c.Connect(ctx); err != nil {
		return nil, err
	}
	return c, nil
}
