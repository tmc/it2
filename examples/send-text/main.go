package main

import (
	"context"
	"flag"
	"log"
	"time"

	"github.com/tmc/it2"
)

func main() {
	var (
		sessionID = flag.String("session", "", "Session ID to send text to")
		text      = flag.String("text", "echo 'Hello from it2 library!'\n", "Text to send")
		timeout   = flag.Duration("timeout", 5*time.Second, "Connection timeout")
	)
	flag.Parse()

	ctx, cancel := context.WithTimeout(context.Background(), *timeout)
	defer cancel()

	// Create and connect to iTerm2
	client := it2.New()
	if err := client.Connect(ctx); err != nil {
		log.Fatal("Failed to connect:", err)
	}
	defer client.Close()

	// If no session ID provided, use the first available session
	if *sessionID == "" {
		sessions, err := client.ListSessions(ctx)
		if err != nil {
			log.Fatal("Failed to list sessions:", err)
		}
		if len(sessions) == 0 {
			log.Fatal("No sessions found")
		}
		*sessionID = sessions[0].ID
		log.Printf("Using first session: %s (%s)", sessions[0].Name, sessions[0].ID)
	}

	// Send text to the session
	if err := client.SendText(ctx, *sessionID, *text); err != nil {
		log.Fatal("Failed to send text:", err)
	}

	log.Printf("Successfully sent text to session %s", *sessionID)
}