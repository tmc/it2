package tab

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/spf13/cobra"
	"github.com/tmc/it2/internal/cmdutil"
	"github.com/tmc/it2/internal/completion"
)

// newSetColorCommand creates the set-color subcommand
func newSetColorCommand() *cobra.Command {
	var rgb string
	var hex string

	template := cmdutil.CommandTemplate{
		Use:            "set-color [<tab-id>]",
		Short:          "Set the color of a tab",
		Long: `Set the color of a tab using RGB or hex color formats.

The color can be specified using either:
- --rgb flag: Comma-separated RGB values (0-255), e.g., "255,0,0" for red
- --hex flag: Hexadecimal color code, e.g., "#FF0000" or "FF0000"

If no tab ID is provided, the current tab is used.

The color persists across sessions and is stored in the profile.`,
		Args:           cobra.MaximumNArgs(1),
		RequiresClient: true,
		SupportsFormat: true,
		ValidArgsFunc:  completion.TabIDCompletion,
		PreRunE: func(cmd *cobra.Command, args []string) error {
			// Ensure exactly one color format is specified
			if (rgb == "" && hex == "") || (rgb != "" && hex != "") {
				return fmt.Errorf("exactly one of --rgb or --hex must be specified")
			}
			return nil
		},
		RunE: func(sc *cmdutil.StandardCommand, args []string) error {
			var tabID string
			if len(args) > 0 {
				tabID = args[0]
			} else {
				tabID = ""
			}

			// Resolve tab ID (handles empty string by using current session's tab)
			tabID, err := sc.GetClient().ResolveTabID(sc.GetContext(), tabID)
			if err != nil {
				return sc.ReportError("resolve tab ID", err)
			}

			// Parse color
			var red, green, blue float64
			if rgb != "" {
				red, green, blue, err = parseRGB(rgb)
				if err != nil {
					return sc.ReportError("parse RGB color", err)
				}
			} else {
				red, green, blue, err = parseHex(hex)
				if err != nil {
					return sc.ReportError("parse hex color", err)
				}
			}

			// Set the tab color
			err = sc.GetClient().SetTabColor(sc.GetContext(), tabID, red, green, blue)
			if err != nil {
				return sc.ReportError("set tab color", err)
			}

			// Report success with JSON output support
			if sc.GetFlags().Format == "json" {
				result := map[string]interface{}{
					"tab_id": tabID,
					"red":    red,
					"green":  green,
					"blue":   blue,
					"set":    true,
				}
				return sc.FormatOutput(result)
			}

			sc.ReportSuccess("Successfully set color of tab %s to RGB(%.2f, %.2f, %.2f)", tabID, red, green, blue)
			return nil
		},
	}

	cmd := cmdutil.NewCommandFromTemplate(template)
	cmd.Flags().StringVar(&rgb, "rgb", "", "RGB color (e.g., \"255,0,0\")")
	cmd.Flags().StringVar(&hex, "hex", "", "Hex color (e.g., \"#FF0000\" or \"FF0000\")")

	return cmd
}

// newGetColorCommand creates the get-color subcommand
func newGetColorCommand() *cobra.Command {
	template := cmdutil.CommandTemplate{
		Use:            "get-color [<tab-id>]",
		Short:          "Get the current color of a tab",
		Long: `Get the current color of a tab.

Returns the RGB color components (0.0-1.0) and whether tab color is enabled.
If tab color is not enabled, returns an indication that the default color is used.

If no tab ID is provided, the current tab is used.`,
		Args:           cobra.MaximumNArgs(1),
		RequiresClient: true,
		SupportsFormat: true,
		ValidArgsFunc:  completion.TabIDCompletion,
		RunE: func(sc *cmdutil.StandardCommand, args []string) error {
			var tabID string
			if len(args) > 0 {
				tabID = args[0]
			} else {
				tabID = ""
			}

			// Resolve tab ID (handles empty string by using current session's tab)
			tabID, err := sc.GetClient().ResolveTabID(sc.GetContext(), tabID)
			if err != nil {
				return sc.ReportError("resolve tab ID", err)
			}

			// Get the tab color
			red, green, blue, enabled, err := sc.GetClient().GetTabColor(sc.GetContext(), tabID)
			if err != nil {
				return sc.ReportError("get tab color", err)
			}

			// Report result with JSON output support
			if sc.GetFlags().Format == "json" {
				result := map[string]interface{}{
					"tab_id":  tabID,
					"enabled": enabled,
				}
				if enabled {
					result["red"] = red
					result["green"] = green
					result["blue"] = blue
					result["hex"] = rgbToHex(red, green, blue)
				}
				return sc.FormatOutput(result)
			}

			if !enabled {
				sc.ReportSuccess("Tab %s: color not set (using default)", tabID)
			} else {
				hexColor := rgbToHex(red, green, blue)
				sc.ReportSuccess("Tab %s: RGB(%.2f, %.2f, %.2f) / %s", tabID, red, green, blue, hexColor)
			}
			return nil
		},
	}

	return cmdutil.NewCommandFromTemplate(template)
}

// newClearColorCommand creates the clear-color subcommand
func newClearColorCommand() *cobra.Command {
	template := cmdutil.CommandTemplate{
		Use:            "clear-color [<tab-id>]",
		Short:          "Clear the color of a tab",
		Long: `Clear the color of a tab, reverting to the default color.

This disables the custom tab color and allows the tab to use the default color
from the profile.

If no tab ID is provided, the current tab is used.`,
		Args:           cobra.MaximumNArgs(1),
		RequiresClient: true,
		SupportsFormat: true,
		ValidArgsFunc:  completion.TabIDCompletion,
		RunE: func(sc *cmdutil.StandardCommand, args []string) error {
			var tabID string
			if len(args) > 0 {
				tabID = args[0]
			} else {
				tabID = ""
			}

			// Resolve tab ID (handles empty string by using current session's tab)
			tabID, err := sc.GetClient().ResolveTabID(sc.GetContext(), tabID)
			if err != nil {
				return sc.ReportError("resolve tab ID", err)
			}

			// Clear the tab color
			err = sc.GetClient().ClearTabColor(sc.GetContext(), tabID)
			if err != nil {
				return sc.ReportError("clear tab color", err)
			}

			// Report success with JSON output support
			if sc.GetFlags().Format == "json" {
				result := map[string]interface{}{
					"tab_id":  tabID,
					"cleared": true,
				}
				return sc.FormatOutput(result)
			}

			sc.ReportSuccess("Successfully cleared color of tab %s", tabID)
			return nil
		},
	}

	return cmdutil.NewCommandFromTemplate(template)
}

// parseRGB parses an RGB color string like "255,0,0" into normalized float64 values (0.0-1.0)
func parseRGB(rgb string) (red, green, blue float64, err error) {
	parts := strings.Split(rgb, ",")
	if len(parts) != 3 {
		return 0, 0, 0, fmt.Errorf("RGB must have 3 components separated by commas (e.g., \"255,0,0\")")
	}

	values := make([]float64, 3)
	for i, part := range parts {
		part = strings.TrimSpace(part)
		val, err := strconv.ParseInt(part, 10, 32)
		if err != nil {
			return 0, 0, 0, fmt.Errorf("invalid RGB component %d: %s", i+1, part)
		}
		if val < 0 || val > 255 {
			return 0, 0, 0, fmt.Errorf("RGB component %d out of range (0-255): %d", i+1, val)
		}
		values[i] = float64(val) / 255.0
	}

	return values[0], values[1], values[2], nil
}

// parseHex parses a hex color string like "#FF0000" or "FF0000" into normalized float64 values (0.0-1.0)
func parseHex(hex string) (red, green, blue float64, err error) {
	// Remove # prefix if present
	hex = strings.TrimPrefix(hex, "#")

	// Validate length
	if len(hex) != 6 {
		return 0, 0, 0, fmt.Errorf("hex color must be 6 characters (e.g., \"FF0000\" or \"#FF0000\")")
	}

	// Parse hex components
	values := make([]float64, 3)
	for i := 0; i < 3; i++ {
		component := hex[i*2 : i*2+2]
		val, err := strconv.ParseInt(component, 16, 32)
		if err != nil {
			return 0, 0, 0, fmt.Errorf("invalid hex component at position %d-%d: %s", i*2, i*2+1, component)
		}
		values[i] = float64(val) / 255.0
	}

	return values[0], values[1], values[2], nil
}

// rgbToHex converts normalized RGB values (0.0-1.0) to a hex color string
func rgbToHex(red, green, blue float64) string {
	r := int(red * 255.0)
	g := int(green * 255.0)
	b := int(blue * 255.0)
	return fmt.Sprintf("#%02X%02X%02X", r, g, b)
}
