package formatting

import (
	"fmt"
	"strings"

	pb "github.com/tmc/it2/proto"
)

// stripTrailingEmptyLines removes empty lines from the end of the buffer contents.
// This is useful for TUI applications where trailing blank lines clutter the output.
func stripTrailingEmptyLines(contents []*pb.LineContents) []*pb.LineContents {
	// Find the last non-empty line
	lastNonEmpty := len(contents) - 1
	for lastNonEmpty >= 0 {
		text := contents[lastNonEmpty].GetText()
		if strings.TrimSpace(text) != "" {
			break
		}
		lastNonEmpty--
	}

	// Return slice up to and including the last non-empty line
	if lastNonEmpty < 0 {
		return nil // All lines were empty
	}
	return contents[:lastNonEmpty+1]
}

// FormatBuffer formats buffer contents.
func (f *Formatter) FormatBuffer(resp *pb.GetBufferResponse) error {
	if resp == nil {
		fmt.Println("(no response)")
		return nil
	}

	if resp.GetStatus() != pb.GetBufferResponse_OK {
		fmt.Printf("Error: %v\n", resp.GetStatus())
		return nil
	}

	contents := resp.GetContents()
	if len(contents) == 0 {
		fmt.Println("(buffer is empty)")
		return nil
	}

	// Strip trailing empty lines unless includeEmptyLines is set
	linesToPrint := contents
	if !f.includeEmptyLines {
		linesToPrint = stripTrailingEmptyLines(contents)
	}

	for _, line := range linesToPrint {
		text := line.GetText()
		fmt.Println(text)
	}
	return nil
}

// FormatBufferWithColors formats buffer contents with ANSI color codes.
func (f *Formatter) FormatBufferWithColors(resp *pb.GetBufferResponse) error {
	if resp == nil {
		fmt.Println("(no response)")
		return nil
	}

	if resp.GetStatus() != pb.GetBufferResponse_OK {
		fmt.Printf("Error: %v\n", resp.GetStatus())
		return nil
	}

	contents := resp.GetContents()
	if len(contents) == 0 {
		fmt.Println("(buffer is empty)")
		return nil
	}

	// Strip trailing empty lines unless includeEmptyLines is set
	linesToPrint := contents
	if !f.includeEmptyLines {
		linesToPrint = stripTrailingEmptyLines(contents)
	}

	for _, line := range linesToPrint {
		text := line.GetText()
		styles := line.GetStyle()

		if len(styles) == 0 {
			// No style information, print plain text
			fmt.Println(text)
			continue
		}

		// Convert text to runes for proper indexing
		runes := []rune(text)
		var output strings.Builder
		styleIndex := 0
		runeIndex := 0

		for styleIndex < len(styles) && runeIndex < len(runes) {
			style := styles[styleIndex]
			repeats := int(style.GetRepeats())
			if repeats == 0 {
				repeats = 1
			}

			// Generate ANSI escape codes for this style
			ansiCode := styleToANSI(style)
			if ansiCode != "" {
				output.WriteString(ansiCode)
			}

			// Write the characters with this style
			for i := 0; i < repeats && runeIndex < len(runes); i++ {
				output.WriteRune(runes[runeIndex])
				runeIndex++
			}

			// Reset style after writing
			if ansiCode != "" {
				output.WriteString("\033[0m")
			}

			styleIndex++
		}

		// Write any remaining characters without style
		for runeIndex < len(runes) {
			output.WriteRune(runes[runeIndex])
			runeIndex++
		}

		fmt.Println(output.String())
	}
	return nil
}

// styleToANSI converts iTerm2 CellStyle to ANSI escape codes.
func styleToANSI(style *pb.CellStyle) string {
	if style == nil {
		return ""
	}

	var codes []string

	// Handle foreground color
	switch fg := style.GetFgColor().(type) {
	case *pb.CellStyle_FgStandard:
		if fg.FgStandard < 16 {
			// Standard 16 colors
			if fg.FgStandard < 8 {
				codes = append(codes, fmt.Sprintf("3%d", fg.FgStandard))
			} else {
				codes = append(codes, fmt.Sprintf("9%d", fg.FgStandard-8))
			}
		} else if fg.FgStandard < 256 {
			// 256 color mode
			codes = append(codes, fmt.Sprintf("38;5;%d", fg.FgStandard))
		}
	case *pb.CellStyle_FgRgb:
		if fg.FgRgb != nil {
			// 24-bit true color
			codes = append(codes, fmt.Sprintf("38;2;%d;%d;%d",
				fg.FgRgb.GetRed()>>8,
				fg.FgRgb.GetGreen()>>8,
				fg.FgRgb.GetBlue()>>8))
		}
	}

	// Handle background color
	switch bg := style.GetBgColor().(type) {
	case *pb.CellStyle_BgStandard:
		if bg.BgStandard < 16 {
			// Standard 16 colors
			if bg.BgStandard < 8 {
				codes = append(codes, fmt.Sprintf("4%d", bg.BgStandard))
			} else {
				codes = append(codes, fmt.Sprintf("10%d", bg.BgStandard-8))
			}
		} else if bg.BgStandard < 256 {
			// 256 color mode
			codes = append(codes, fmt.Sprintf("48;5;%d", bg.BgStandard))
		}
	case *pb.CellStyle_BgRgb:
		if bg.BgRgb != nil {
			// 24-bit true color
			codes = append(codes, fmt.Sprintf("48;2;%d;%d;%d",
				bg.BgRgb.GetRed()>>8,
				bg.BgRgb.GetGreen()>>8,
				bg.BgRgb.GetBlue()>>8))
		}
	}

	// Handle text attributes
	if style.GetBold() {
		codes = append(codes, "1")
	}
	if style.GetFaint() {
		codes = append(codes, "2")
	}
	if style.GetItalic() {
		codes = append(codes, "3")
	}
	if style.GetUnderline() {
		codes = append(codes, "4")
	}
	if style.GetBlink() {
		codes = append(codes, "5")
	}
	if style.GetInverse() {
		codes = append(codes, "7")
	}
	if style.GetInvisible() {
		codes = append(codes, "8")
	}
	if style.GetStrikethrough() {
		codes = append(codes, "9")
	}

	if len(codes) > 0 {
		return fmt.Sprintf("\033[%sm", strings.Join(codes, ";"))
	}

	return ""
}

// FormatBufferEscaped formats buffer contents with escape sequences shown as visible characters (like cat -v).
func (f *Formatter) FormatBufferEscaped(resp *pb.GetBufferResponse, includeColors bool) error {
	if resp == nil {
		fmt.Println("(no response)")
		return nil
	}

	if resp.GetStatus() != pb.GetBufferResponse_OK {
		fmt.Printf("Error: %v\n", resp.GetStatus())
		return nil
	}

	contents := resp.GetContents()
	if len(contents) == 0 {
		fmt.Println("(buffer is empty)")
		return nil
	}

	// Strip trailing empty lines unless includeEmptyLines is set
	linesToPrint := contents
	if !f.includeEmptyLines {
		linesToPrint = stripTrailingEmptyLines(contents)
	}

	for _, line := range linesToPrint {
		text := line.GetText()
		styles := line.GetStyle()

		var output strings.Builder

		if includeColors && len(styles) > 0 {
			// Include color codes, but show them as escaped
			runes := []rune(text)
			styleIndex := 0
			runeIndex := 0

			for styleIndex < len(styles) && runeIndex < len(runes) {
				style := styles[styleIndex]
				repeats := int(style.GetRepeats())
				if repeats == 0 {
					repeats = 1
				}

				// Generate ANSI escape codes for this style
				ansiCode := styleToANSI(style)
				if ansiCode != "" {
					// Show the escape sequence as visible characters
					output.WriteString(escapeString(ansiCode))
				}

				// Write the characters with this style
				for i := 0; i < repeats && runeIndex < len(runes); i++ {
					output.WriteString(escapeChar(runes[runeIndex]))
					runeIndex++
				}

				// Show reset sequence as visible
				if ansiCode != "" {
					output.WriteString("^[[0m")
				}

				styleIndex++
			}

			// Write any remaining characters without style
			for runeIndex < len(runes) {
				output.WriteString(escapeChar(runes[runeIndex]))
				runeIndex++
			}
		} else {
			// Just escape control characters in the text
			for _, r := range text {
				output.WriteString(escapeChar(r))
			}
		}

		fmt.Println(output.String())
	}
	return nil
}

// escapeChar converts a single character to its visible representation (like cat -v).
func escapeChar(r rune) string {
	if r == '\t' {
		return "^I"
	} else if r == '\n' {
		return "$\n"
	} else if r == '\r' {
		return "^M"
	} else if r == '\x1b' {
		return "^["
	} else if r < 32 {
		// Control characters: ^A through ^_
		return fmt.Sprintf("^%c", r+64)
	} else if r == 127 {
		// DEL character
		return "^?"
	} else if r >= 128 && r < 160 {
		// High control characters
		return fmt.Sprintf("M-^%c", (r-128)+64)
	} else if r >= 160 {
		// Normal high-bit characters - could show as M-x notation
		// but for now just pass through
		return string(r)
	}
	return string(r)
}

// escapeString converts an ANSI escape sequence to visible representation.
func escapeString(s string) string {
	var output strings.Builder
	for _, r := range s {
		output.WriteString(escapeChar(r))
	}
	return output.String()
}
