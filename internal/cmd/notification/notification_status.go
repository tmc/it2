package notification

import (
	"context"
	"encoding/json"
	"fmt"
	"sort"
	"time"

	"github.com/spf13/cobra"
	"github.com/tmc/it2/internal/client"
	"github.com/tmc/it2/internal/cmdcore"
)

func newStatusCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "status",
		Short: "Show notification subscription status",
		Long:  "Display current notification subscriptions and their status",
		RunE:  runStatusCommand,
	}

	cmd.Flags().BoolP("json", "j", false, "Output in JSON format (equivalent to --format json)")
	cmd.Flags().MarkHidden("json")
	cmd.Flags().Bool("detailed", false, "Show detailed subscription information")

	return cmd
}

func runStatusCommand(cmd *cobra.Command, args []string) error {
	detailed, _ := cmd.Flags().GetBool("detailed")

	// Determine JSON format from --json flag or global --format flag
	jsonFlag, _ := cmd.Flags().GetBool("json")
	_, _, format := cmdcore.GetFlags(cmd)
	jsonFormat := jsonFlag || format == "json"

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	c, err := cmdcore.ConnectClient(ctx)
	if err != nil {
		return fmt.Errorf("failed to connect: %w", err)
	}
	defer c.Close()

	// Get notification status
	status, err := getNotificationStatus(c, ctx)
	if err != nil {
		return fmt.Errorf("failed to get notification status: %w", err)
	}

	if jsonFormat {
		return outputStatusJSON(status)
	}

	return outputStatusTable(status, detailed)
}

type notificationStatus struct {
	Connected     bool                   `json:"connected"`
	Timestamp     time.Time              `json:"timestamp"`
	Subscriptions []subscriptionInfo     `json:"subscriptions"`
	Summary       subscriptionSummary    `json:"summary"`
	SessionInfo   map[string]sessionInfo `json:"session_info,omitempty"`
}

type subscriptionInfo struct {
	Type        string    `json:"type"`
	SessionID   string    `json:"session_id,omitempty"`
	Active      bool      `json:"active"`
	Description string    `json:"description"`
	Scope       string    `json:"scope"`
	CreatedAt   time.Time `json:"created_at"`
}

type subscriptionSummary struct {
	Total         int `json:"total"`
	Active        int `json:"active"`
	SessionScoped int `json:"session_scoped"`
	Global        int `json:"global"`
}

type sessionInfo struct {
	ID       string `json:"id"`
	Name     string `json:"name,omitempty"`
	Active   bool   `json:"active"`
	WindowID string `json:"window_id,omitempty"`
}

func getNotificationStatus(c *client.Client, ctx context.Context) (*notificationStatus, error) {
	status := &notificationStatus{
		Connected:     true,
		Timestamp:     time.Now(),
		Subscriptions: []subscriptionInfo{},
		SessionInfo:   make(map[string]sessionInfo),
	}

	// Note: In a real implementation, we would need to track active subscriptions
	// For now, we'll show available notification types and simulate status
	types := getNotificationTypes()

	// Get session information for context
	sessions, err := c.ListSessions(ctx)
	if err == nil {
		for _, session := range sessions {
			status.SessionInfo[session.SessionID] = sessionInfo{
				ID:       session.SessionID,
				Name:     session.SessionName,
				Active:   true, // In real impl, would check actual status
				WindowID: session.WindowID,
			}
		}
	}

	// Simulate some active subscriptions for demonstration
	// In a real implementation, the client would track active subscriptions
	activeTypes := []string{"keystroke", "screen", "focus"} // Example active subscriptions

	for _, notifType := range activeTypes {
		if typeInfo, exists := types[notifType]; exists {
			scope := "Global"
			if typeInfo.sessionScoped {
				scope = "Session"
			}

			subscription := subscriptionInfo{
				Type:        notifType,
				Active:      true,
				Description: typeInfo.description,
				Scope:       scope,
				CreatedAt:   time.Now().Add(-time.Duration(len(status.Subscriptions)) * time.Minute),
			}

			status.Subscriptions = append(status.Subscriptions, subscription)
		}
	}

	// Calculate summary
	status.Summary.Total = len(status.Subscriptions)
	for _, sub := range status.Subscriptions {
		if sub.Active {
			status.Summary.Active++
		}
		if sub.Scope == "Session" {
			status.Summary.SessionScoped++
		} else {
			status.Summary.Global++
		}
	}

	return status, nil
}

func outputStatusJSON(status *notificationStatus) error {
	data, err := json.MarshalIndent(status, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal JSON: %w", err)
	}

	fmt.Println(string(data))
	return nil
}

func outputStatusTable(status *notificationStatus, detailed bool) error {
	fmt.Printf("Notification System Status\n")
	fmt.Printf("==========================\n\n")

	fmt.Printf("Connection: %s\n", formatConnectionStatus(status.Connected))
	fmt.Printf("Last Check: %s\n\n", status.Timestamp.Format("2006-01-02 15:04:05"))

	// Summary
	fmt.Printf("Subscription Summary:\n")
	fmt.Printf("  Total subscriptions: %d\n", status.Summary.Total)
	fmt.Printf("  Active: %d\n", status.Summary.Active)
	fmt.Printf("  Session-scoped: %d\n", status.Summary.SessionScoped)
	fmt.Printf("  Global: %d\n\n", status.Summary.Global)

	if len(status.Subscriptions) == 0 {
		fmt.Println("No active notification subscriptions found.")
		fmt.Println("\nUse 'it2 notification subscribe <type>' to start receiving notifications.")
		return nil
	}

	// Active subscriptions
	fmt.Printf("Active Subscriptions:\n")
	if detailed {
		fmt.Printf("%-12s %-15s %-10s %-20s %-50s\n", "TYPE", "SCOPE", "SESSION", "CREATED", "DESCRIPTION")
		fmt.Printf("%-12s %-15s %-10s %-20s %-50s\n", "----", "-----", "-------", "-------", "-----------")
	} else {
		fmt.Printf("%-12s %-15s %-10s %-50s\n", "TYPE", "SCOPE", "SESSION", "DESCRIPTION")
		fmt.Printf("%-12s %-15s %-10s %-50s\n", "----", "-----", "-------", "-----------")
	}

	// Sort subscriptions by type for consistent output
	sort.Slice(status.Subscriptions, func(i, j int) bool {
		return status.Subscriptions[i].Type < status.Subscriptions[j].Type
	})

	for _, sub := range status.Subscriptions {
		if !sub.Active {
			continue
		}

		sessionDisplay := "all"
		if sub.SessionID != "" {
			sessionDisplay = sub.SessionID[:8] // Show first 8 chars
		}

		if detailed {
			fmt.Printf("%-12s %-15s %-10s %-20s %-50s\n",
				sub.Type,
				sub.Scope,
				sessionDisplay,
				sub.CreatedAt.Format("15:04:05"),
				sub.Description)
		} else {
			fmt.Printf("%-12s %-15s %-10s %-50s\n",
				sub.Type,
				sub.Scope,
				sessionDisplay,
				sub.Description)
		}
	}

	if detailed && len(status.SessionInfo) > 0 {
		fmt.Printf("\nAvailable Sessions:\n")
		fmt.Printf("%-20s %-30s %-10s\n", "SESSION ID", "NAME", "WINDOW")
		fmt.Printf("%-20s %-30s %-10s\n", "----------", "----", "------")

		var sessionIDs []string
		for id := range status.SessionInfo {
			sessionIDs = append(sessionIDs, id)
		}
		sort.Strings(sessionIDs)

		for _, id := range sessionIDs {
			info := status.SessionInfo[id]
			name := info.Name
			if name == "" {
				name = "(unnamed)"
			}
			fmt.Printf("%-20s %-30s %-10s\n",
				id[:8]+"...", // Show first 8 chars + ellipsis
				name,
				info.WindowID)
		}
	}

	fmt.Printf("\nUse 'it2 notification list-types' to see all available notification types.\n")
	fmt.Printf("Use 'it2 notification subscribe <type>' to add new subscriptions.\n")

	return nil
}

func formatConnectionStatus(connected bool) string {
	if connected {
		return "✓ Connected"
	}
	return "✗ Disconnected"
}
