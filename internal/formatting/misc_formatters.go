package formatting

import (
	"fmt"
	"strings"

	"github.com/tmc/it2/internal/client"
	pb "github.com/tmc/it2/proto"
)

// FormatSelection formats a selection for display.
func (f *Formatter) FormatSelection(selection *pb.Selection) error {
	if selection == nil || len(selection.GetSubSelections()) == 0 {
		fmt.Println("No selection")
		return nil
	}

	switch f.format {
	case "json":
		return f.formatJSON(selection)
	case "yaml":
		return f.formatYAML(selection)
	default:
		fmt.Println("Current selection:")
		for i, sub := range selection.GetSubSelections() {
			fmt.Printf("  Selection %d: %v\n", i+1, sub)
		}
	}
	return nil
}

// FormatStringList formats a list of strings with an optional title.
func (f *Formatter) FormatStringList(title string, items []string) error {
	switch f.format {
	case "json":
		data := map[string][]string{strings.ToLower(strings.ReplaceAll(title, " ", "_")): items}
		return f.formatJSON(data)
	case "yaml":
		data := map[string][]string{strings.ToLower(strings.ReplaceAll(title, " ", "_")): items}
		return f.formatYAML(data)
	default:
		if title != "" {
			fmt.Printf("%s:\n", title)
		}
		if len(items) == 0 {
			fmt.Println("  (none)")
		} else {
			for _, item := range items {
				fmt.Printf("  - %s\n", item)
			}
		}
		return nil
	}
}

// FormatTmuxConnections formats tmux connection information.
func (f *Formatter) FormatTmuxConnections(connections []*client.TmuxConnection) error {
	switch f.format {
	case "json":
		return f.formatJSON(connections)
	case "yaml":
		return f.formatYAML(connections)
	default:
		return f.formatTmuxConnectionsText(connections)
	}
}

func (f *Formatter) formatTmuxConnectionsText(connections []*client.TmuxConnection) error {
	if len(connections) == 0 {
		fmt.Println("No tmux connections found")
		return nil
	}

	fmt.Printf("Tmux Connections (%d):\n", len(connections))
	for i, conn := range connections {
		fmt.Printf("  %d. Connection ID: %s\n", i+1, conn.ConnectionId)
		if conn.OwningSessionId != "" {
			fmt.Printf("     Owning Session: %s\n", conn.OwningSessionId)
		}
		fmt.Println()
	}
	return nil
}

// FormatKeyBindings formats keyboard bindings output.
func (f *Formatter) FormatKeyBindings(bindings []interface{}) error {
	switch f.format {
	case "json":
		return f.formatJSON(bindings)
	case "yaml":
		return f.formatYAML(bindings)
	default:
		return f.formatKeyBindingsText(bindings)
	}
}

func (f *Formatter) formatKeyBindingsText(bindings []interface{}) error {
	if len(bindings) == 0 {
		fmt.Println("No keyboard bindings found")
		return nil
	}

	fmt.Printf("Keyboard Bindings (%d):\n", len(bindings))
	for i, binding := range bindings {
		fmt.Printf("  %d. %v\n", i+1, binding)
	}
	return nil
}

// FormatFindResults formats text search results.
func (f *Formatter) FormatFindResults(results []*client.FindResult, showLineNumbers bool) error {
	switch f.format {
	case "json":
		return f.formatJSON(results)
	case "yaml":
		return f.formatYAML(results)
	default:
		return f.formatFindResultsText(results, showLineNumbers)
	}
}

func (f *Formatter) formatFindResultsText(results []*client.FindResult, showLineNumbers bool) error {
	if len(results) == 0 {
		fmt.Println("No matches found")
		return nil
	}

	fmt.Printf("Found %d matches:\n", len(results))
	for i, result := range results {
		prefix := fmt.Sprintf("  %d. ", i+1)
		if showLineNumbers {
			prefix += fmt.Sprintf("Line %d, Column %d: ", result.Line+1, result.Column+1)
		}
		fmt.Printf("%s%s\n", prefix, result.Context)
	}
	return nil
}

// FormatReplaceResults formats text replacement results.
func (f *Formatter) FormatReplaceResults(results []*client.ReplaceResult) error {
	switch f.format {
	case "json":
		return f.formatJSON(results)
	case "yaml":
		return f.formatYAML(results)
	default:
		return f.formatReplaceResultsText(results)
	}
}

func (f *Formatter) formatReplaceResultsText(results []*client.ReplaceResult) error {
	if len(results) == 0 {
		fmt.Println("No replacements made")
		return nil
	}

	successCount := 0
	for _, result := range results {
		if result.Success {
			successCount++
		}
	}

	fmt.Printf("Replacement Results (%d total, %d successful):\n", len(results), successCount)
	for i, result := range results {
		status := "FAILED"
		if result.Success {
			status = "SUCCESS"
		}
		fmt.Printf("  %d. [%s] Line %d, Column %d: '%s' -> '%s'\n",
			i+1, status, result.Line+1, result.Column+1, result.OriginalText, result.NewText)
	}
	return nil
}

// FormatHighlightResults formats text highlighting results.
func (f *Formatter) FormatHighlightResults(results []*client.HighlightResult) error {
	switch f.format {
	case "json":
		return f.formatJSON(results)
	case "yaml":
		return f.formatYAML(results)
	default:
		return f.formatHighlightResultsText(results)
	}
}

func (f *Formatter) formatHighlightResultsText(results []*client.HighlightResult) error {
	if len(results) == 0 {
		fmt.Println("No highlights created")
		return nil
	}

	fmt.Printf("Highlight Results (%d):\n", len(results))
	for i, result := range results {
		duration := "permanent"
		if result.Duration > 0 {
			duration = fmt.Sprintf("%d seconds", result.Duration)
		}
		fmt.Printf("  %d. Line %d, Column %d: '%s' (%s %s, %s)\n",
			i+1, result.Line+1, result.Column+1, result.Text, result.Color, "highlight", duration)
	}
	return nil
}

// FormatBroadcastDomains formats broadcast domain information.
func (f *Formatter) FormatBroadcastDomains(domains []*client.BroadcastDomain) error {
	switch f.format {
	case "json":
		return f.formatJSON(domains)
	case "yaml":
		return f.formatYAML(domains)
	default:
		return f.formatBroadcastDomainsText(domains)
	}
}

func (f *Formatter) formatBroadcastDomainsText(domains []*client.BroadcastDomain) error {
	if len(domains) == 0 {
		fmt.Println("No broadcast domains found")
		return nil
	}

	fmt.Printf("Broadcast Domains (%d):\n", len(domains))
	for i, domain := range domains {
		if len(domain.SessionIds) == 0 {
			fmt.Printf("  %d. (empty domain)\n", i+1)
			continue
		}

		// Show all sessions in the domain
		fmt.Printf("  %d. Broadcast Domain with %d session(s):\n", i+1, len(domain.SessionIds))
		for j, sessionID := range domain.SessionIds {
			fmt.Printf("       %d. %s\n", j+1, sessionID)
		}
		fmt.Println()
	}
	return nil
}

// FormatAnnotations formats annotation information.
func (f *Formatter) FormatAnnotations(annotations []*Annotation) error {
	switch f.format {
	case "json":
		return f.formatJSON(annotations)
	case "yaml":
		return f.formatYAML(annotations)
	default:
		return f.formatAnnotationsText(annotations)
	}
}

func (f *Formatter) formatAnnotationsText(annotations []*Annotation) error {
	if len(annotations) == 0 {
		fmt.Println("No annotations found")
		return nil
	}

	fmt.Printf("Annotations (%d):\n", len(annotations))
	for i, annotation := range annotations {
		fmt.Printf("  %d. ID: %s\n", i+1, annotation.ID)
		fmt.Printf("     Text: %s\n", annotation.Text)
		if annotation.Type != "" {
			fmt.Printf("     Type: %s\n", annotation.Type)
		}
		if annotation.Line > 0 {
			fmt.Printf("     Location: Line %d", annotation.Line)
			if annotation.Column > 0 {
				fmt.Printf(", Column %d", annotation.Column)
			}
			fmt.Println()
		}
		fmt.Printf("     Created: %s\n", annotation.Timestamp.Format("2006-01-02 15:04:05"))
		fmt.Println()
	}
	return nil
}

// FormatTriggers formats trigger information.
func (f *Formatter) FormatTriggers(triggers []*Trigger) error {
	switch f.format {
	case "json":
		return f.formatJSON(triggers)
	case "yaml":
		return f.formatYAML(triggers)
	default:
		return f.formatTriggersText(triggers)
	}
}

func (f *Formatter) formatTriggersText(triggers []*Trigger) error {
	if len(triggers) == 0 {
		fmt.Println("No triggers found")
		return nil
	}

	fmt.Printf("Triggers (%d):\n", len(triggers))
	for i, trigger := range triggers {
		status := "DISABLED"
		if trigger.Enabled {
			status = "ENABLED"
		}

		fmt.Printf("  %d. ID: %s [%s]\n", i+1, trigger.ID, status)
		fmt.Printf("     Pattern: %s\n", trigger.Pattern)
		fmt.Printf("     Action: %s\n", trigger.Action)
		fmt.Printf("     Created: %s\n", trigger.Created.Format("2006-01-02 15:04:05"))
		fmt.Println()
	}
	return nil
}

// FormatCursor formats cursor position and state information.
func (f *Formatter) FormatCursor(cursor *pb.Coord) error {
	if cursor == nil {
		fmt.Println("(cursor information not available)")
		return nil
	}

	switch f.format {
	case "json":
		cursorInfo := map[string]interface{}{
			"x": cursor.GetX(),
			"y": cursor.GetY(),
		}
		return f.formatJSON(cursorInfo)
	case "yaml":
		cursorInfo := map[string]interface{}{
			"x": cursor.GetX(),
			"y": cursor.GetY(),
		}
		return f.formatYAML(cursorInfo)
	default:
		fmt.Printf("Cursor position: (%d, %d)\n", cursor.GetX(), cursor.GetY())
		return nil
	}
}

// Placeholder formatter methods for commands that aren't fully implemented.
// These prevent compilation errors and provide helpful error messages.

// FormatAlertResponse formats alert responses (placeholder).
func (f *Formatter) FormatAlertResponse(response interface{}) error {
	return fmt.Errorf("Alert functionality not implemented - this is a placeholder")
}

// FormatCustomControlResponse formats custom control responses (placeholder).
func (f *Formatter) FormatCustomControlResponse(response interface{}) error {
	return fmt.Errorf("Custom control functionality not implemented - this is a placeholder")
}

// FormatFilePanelResponse formats file panel responses (placeholder).
func (f *Formatter) FormatFilePanelResponse(response interface{}) error {
	return fmt.Errorf("File panel functionality not implemented - this is a placeholder")
}

// FormatLifecycleResponse formats lifecycle responses (placeholder).
func (f *Formatter) FormatLifecycleResponse(response interface{}) error {
	return fmt.Errorf("Lifecycle functionality not implemented - this is a placeholder")
}
