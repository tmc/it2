package selection

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/spf13/cobra"
	"github.com/tmc/it2/internal/client"
	"github.com/tmc/it2/internal/cmdutil"
	"github.com/tmc/it2/internal/formatting"
)

// NewCommand creates the selection command with all subcommands.
func NewCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "selection",
		GroupID: "content",
		Short:   "Text selection operations in iTerm2",
		Long:    "Commands for getting, setting, and managing text selection",
	}

	cmd.AddCommand(newGetCommand())
	cmd.AddCommand(newSetCommand())
	cmd.AddCommand(newClearCommand())
	cmd.AddCommand(newCopyCommand())

	return cmd
}

func newGetCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "get [<session-id>]",
		Short: "Get current text selection",
		Long:  "Get the current text selection from the specified session (defaults to active session)",
		Args:  cobra.MaximumNArgs(1),
		RunE:  runGetCommand,
	}

	cmd.Flags().Bool("text", false, "Return the selected text content instead of selection coordinates")

	return cmd
}

func runGetCommand(cmd *cobra.Command, args []string) error {
	sessionID := "active"
	if len(args) > 0 {
		sessionID = client.NormalizeSessionID(args[0])
	}

	getText, _ := cmd.Flags().GetBool("text")
	_, timeout, format := cmdutil.GetFlags(cmd)
	ctx, cancel := cmdutil.CreateContext(timeout)
	defer cancel()

	c, err := cmdutil.ConnectClient(ctx)
	if err != nil {
		return fmt.Errorf("failed to connect: %w", err)
	}
	defer c.Close()

	if getText {
		text, err := c.GetSelectionText(ctx, sessionID)
		if err != nil {
			return fmt.Errorf("failed to get selection text: %w", err)
		}
		fmt.Print(text) // Print without newline in case user wants to pipe it
		return nil
	} else {
		selection, err := c.GetSelection(ctx, sessionID)
		if err != nil {
			return fmt.Errorf("failed to get selection: %w", err)
		}

		formatter := formatting.New(format)
		return formatter.FormatSelection(selection)
	}
}

func newSetCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "set [<session-id>]",
		Short: "Set text selection range",
		Long: `Set text selection using coordinates.

Coordinates can be specified as:
  --start-x, --start-y, --end-x, --end-y (individual coordinates)
  --start "x,y" --end "x,y" (coordinate pairs)
  --range "x1,y1:x2,y2" (full range)`,
		Args: cobra.MaximumNArgs(1),
		RunE: runSetCommand,
	}

	cmd.Flags().Int32("start-x", -1, "Start column (0-based)")
	cmd.Flags().Int64("start-y", -1, "Start row (0-based)")
	cmd.Flags().Int32("end-x", -1, "End column (0-based)")
	cmd.Flags().Int64("end-y", -1, "End row (0-based)")
	cmd.Flags().String("start", "", "Start position as 'x,y'")
	cmd.Flags().String("end", "", "End position as 'x,y'")
	cmd.Flags().String("range", "", "Full range as 'x1,y1:x2,y2'")
	cmd.Flags().String("mode", "character", "Selection mode: character, word, line, smart, box, whole-line")

	return cmd
}

func runSetCommand(cmd *cobra.Command, args []string) error {
	sessionID := "active"
	if len(args) > 0 {
		sessionID = client.NormalizeSessionID(args[0])
	}

	// Parse coordinates from various flag formats
	startX, _ := cmd.Flags().GetInt32("start-x")
	startY, _ := cmd.Flags().GetInt64("start-y")
	endX, _ := cmd.Flags().GetInt32("end-x")
	endY, _ := cmd.Flags().GetInt64("end-y")
	startStr, _ := cmd.Flags().GetString("start")
	endStr, _ := cmd.Flags().GetString("end")
	rangeStr, _ := cmd.Flags().GetString("range")
	modeStr, _ := cmd.Flags().GetString("mode")

	var coords [4]int64 // startX, startY, endX, endY

	if rangeStr != "" {
		// Parse range format: "x1,y1:x2,y2"
		parts := strings.Split(rangeStr, ":")
		if len(parts) != 2 {
			return fmt.Errorf("invalid range format, expected 'x1,y1:x2,y2'")
		}

		start := strings.Split(parts[0], ",")
		end := strings.Split(parts[1], ",")
		if len(start) != 2 || len(end) != 2 {
			return fmt.Errorf("invalid range format, expected 'x1,y1:x2,y2'")
		}

		var err error
		coords[0], err = strconv.ParseInt(start[0], 10, 32)
		if err != nil {
			return fmt.Errorf("invalid start x coordinate: %w", err)
		}
		coords[1], err = strconv.ParseInt(start[1], 10, 64)
		if err != nil {
			return fmt.Errorf("invalid start y coordinate: %w", err)
		}
		coords[2], err = strconv.ParseInt(end[0], 10, 32)
		if err != nil {
			return fmt.Errorf("invalid end x coordinate: %w", err)
		}
		coords[3], err = strconv.ParseInt(end[1], 10, 64)
		if err != nil {
			return fmt.Errorf("invalid end y coordinate: %w", err)
		}
	} else if startStr != "" && endStr != "" {
		// Parse start/end format: "x,y"
		start := strings.Split(startStr, ",")
		end := strings.Split(endStr, ",")
		if len(start) != 2 || len(end) != 2 {
			return fmt.Errorf("invalid coordinate format, expected 'x,y'")
		}

		var err error
		coords[0], err = strconv.ParseInt(start[0], 10, 32)
		if err != nil {
			return fmt.Errorf("invalid start x coordinate: %w", err)
		}
		coords[1], err = strconv.ParseInt(start[1], 10, 64)
		if err != nil {
			return fmt.Errorf("invalid start y coordinate: %w", err)
		}
		coords[2], err = strconv.ParseInt(end[0], 10, 32)
		if err != nil {
			return fmt.Errorf("invalid end x coordinate: %w", err)
		}
		coords[3], err = strconv.ParseInt(end[1], 10, 64)
		if err != nil {
			return fmt.Errorf("invalid end y coordinate: %w", err)
		}
	} else if startX >= 0 && startY >= 0 && endX >= 0 && endY >= 0 {
		// Use individual coordinate flags
		coords[0] = int64(startX)
		coords[1] = startY
		coords[2] = int64(endX)
		coords[3] = endY
	} else {
		return fmt.Errorf("must specify coordinates using --range, --start/--end, or individual coordinate flags")
	}

	_, timeout, _ := cmdutil.GetFlags(cmd)
	ctx, cancel := cmdutil.CreateContext(timeout)
	defer cancel()

	c, err := cmdutil.ConnectClient(ctx)
	if err != nil {
		return fmt.Errorf("failed to connect: %w", err)
	}
	defer c.Close()

	if err := c.SetSelectionRange(ctx, sessionID, int32(coords[0]), coords[1], int32(coords[2]), coords[3], modeStr); err != nil {
		return fmt.Errorf("failed to set selection: %w", err)
	}

	fmt.Printf("Selection set: (%d,%d) to (%d,%d)\n", coords[0], coords[1], coords[2], coords[3])
	return nil
}

func newClearCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "clear [<session-id>]",
		Short: "Clear current text selection",
		Long:  "Clear the current text selection in the specified session (defaults to active session)",
		Args:  cobra.MaximumNArgs(1),
		RunE:  runClearCommand,
	}
}

func runClearCommand(cmd *cobra.Command, args []string) error {
	sessionID := "active"
	if len(args) > 0 {
		sessionID = client.NormalizeSessionID(args[0])
	}

	_, timeout, _ := cmdutil.GetFlags(cmd)
	ctx, cancel := cmdutil.CreateContext(timeout)
	defer cancel()

	c, err := cmdutil.ConnectClient(ctx)
	if err != nil {
		return fmt.Errorf("failed to connect: %w", err)
	}
	defer c.Close()

	if err := c.ClearSelection(ctx, sessionID); err != nil {
		return fmt.Errorf("failed to clear selection: %w", err)
	}

	fmt.Println("Selection cleared")
	return nil
}

func newCopyCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "copy [<session-id>]",
		Short: "Copy current selection to clipboard",
		Long:  "Copy the current text selection to the system clipboard",
		Args:  cobra.MaximumNArgs(1),
		RunE:  runCopyCommand,
	}
}

func runCopyCommand(cmd *cobra.Command, args []string) error {
	sessionID := "active"
	if len(args) > 0 {
		sessionID = client.NormalizeSessionID(args[0])
	}

	_, timeout, _ := cmdutil.GetFlags(cmd)
	ctx, cancel := cmdutil.CreateContext(timeout)
	defer cancel()

	c, err := cmdutil.ConnectClient(ctx)
	if err != nil {
		return fmt.Errorf("failed to connect: %w", err)
	}
	defer c.Close()

	if err := c.CopySelection(ctx, sessionID); err != nil {
		return fmt.Errorf("failed to copy selection: %w", err)
	}

	fmt.Println("Selection copied to clipboard")
	return nil
}
