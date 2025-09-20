package main

import (
	"context"
	"fmt"
	"log"

	"github.com/tmc/it2"
)

func main() {
	// Create and connect to iTerm2
	client, err := it2.QuickConnect(context.Background())
	if err != nil {
		log.Fatal("Failed to connect:", err)
	}
	defer client.Close()

	// List all sessions
	sessions, err := client.ListSessions(context.Background())
	if err != nil {
		log.Fatal("Failed to list sessions:", err)
	}

	fmt.Printf("Found %d sessions:\n", len(sessions))
	for i, session := range sessions {
		fmt.Printf("%d. %s (ID: %s)\n", i+1, session.Name, session.ID)
	}
}