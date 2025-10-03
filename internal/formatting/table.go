package formatting

import (
	"fmt"
	"regexp"
	"sort"
	"strconv"
	"strings"

	"github.com/mattn/go-runewidth"
)

// Regular expression to match ANSI escape codes including OSC 8 hyperlinks
var ansiEscapeRegex = regexp.MustCompile(`\x1b\][^\x1b]*\x1b\\|\x1b\[[^m]*m`)

// visibleLen returns the visible length of a string, excluding ANSI escape codes
func visibleLen(s string) int {
	// Strip all ANSI escape sequences including OSC 8 hyperlinks
	stripped := ansiEscapeRegex.ReplaceAllString(s, "")
	// Use runewidth to get proper display width (handles emoji and wide characters)
	return runewidth.StringWidth(stripped)
}

// TableData represents data to be displayed in a table format
type TableData struct {
	Headers []string
	Rows    [][]string
}

// FormatTable formats data as a table with aligned columns
func (f *Formatter) FormatTable(data *TableData) error {
	// Apply column selection if specified
	if len(f.columns) > 0 {
		data = data.SelectColumns(f.columns)
	}

	// Apply sorting if specified
	if f.sortBy != "" {
		data.SortBy(f.sortBy, !f.sortReverse)
	}

	if len(data.Rows) == 0 && len(data.Headers) == 0 {
		fmt.Println("No data")
		return nil
	}

	// Calculate column widths (using visible length, excluding ANSI escape codes)
	colWidths := make([]int, len(data.Headers))
	for i, header := range data.Headers {
		colWidths[i] = visibleLen(header)
	}

	// Check each row to find maximum width per column
	for _, row := range data.Rows {
		for i, cell := range row {
			if i < len(colWidths) {
				cellWidth := visibleLen(cell)
				if cellWidth > colWidths[i] {
					colWidths[i] = cellWidth
				}
			}
		}
	}

	// Print headers
	if len(data.Headers) > 0 {
		f.printTableRow(data.Headers, colWidths)
		f.printTableSeparator(colWidths)
	}

	// Print rows
	for _, row := range data.Rows {
		f.printTableRow(row, colWidths)
	}

	return nil
}

func (f *Formatter) printTableRow(row []string, colWidths []int) {
	var parts []string
	for i, cell := range row {
		width := colWidths[i]
		if i < len(colWidths)-1 {
			// Left-align all columns except the last one
			// Calculate padding based on visible length to account for escape codes
			visLen := visibleLen(cell)
			padding := width - visLen
			if padding < 0 {
				padding = 0
			}
			parts = append(parts, cell+strings.Repeat(" ", padding))
		} else {
			// Don't pad the last column
			parts = append(parts, cell)
		}
	}
	fmt.Println(strings.Join(parts, "    "))
}

func (f *Formatter) printTableSeparator(colWidths []int) {
	var parts []string
	for _, width := range colWidths {
		separator := strings.Repeat("-", width)
		parts = append(parts, separator)
	}
	fmt.Println(strings.Join(parts, "    "))
}

// NewTableData creates a new TableData with headers
func NewTableData(headers []string) *TableData {
	return &TableData{
		Headers: headers,
		Rows:    make([][]string, 0),
	}
}

// AddRow adds a row to the table data
func (td *TableData) AddRow(row []string) {
	// Ensure row has same length as headers, padding with empty strings if needed
	paddedRow := make([]string, len(td.Headers))
	copy(paddedRow, row)
	for i := len(row); i < len(td.Headers); i++ {
		paddedRow[i] = ""
	}
	td.Rows = append(td.Rows, paddedRow)
}

// SelectColumns filters the table to only show specified columns
func (td *TableData) SelectColumns(columns []string) *TableData {
	if len(columns) == 0 {
		return td
	}

	// Create map of column names to indices
	headerMap := make(map[string]int)
	for i, header := range td.Headers {
		headerMap[strings.ToLower(header)] = i
	}

	// Find indices of requested columns
	var indices []int
	var newHeaders []string
	for _, col := range columns {
		if idx, exists := headerMap[strings.ToLower(col)]; exists {
			indices = append(indices, idx)
			newHeaders = append(newHeaders, td.Headers[idx])
		}
	}

	if len(indices) == 0 {
		return td // Return original if no valid columns found
	}

	// Create new table with selected columns
	newTable := &TableData{
		Headers: newHeaders,
		Rows:    make([][]string, 0, len(td.Rows)),
	}

	// Filter each row to selected columns
	for _, row := range td.Rows {
		newRow := make([]string, len(indices))
		for i, idx := range indices {
			if idx < len(row) {
				newRow[i] = row[idx]
			} else {
				newRow[i] = ""
			}
		}
		newTable.Rows = append(newTable.Rows, newRow)
	}

	return newTable
}

// SortBy sorts the table by the specified column
func (td *TableData) SortBy(column string, ascending bool) {
	if len(td.Rows) == 0 {
		return
	}

	// Find column index
	columnIndex := -1
	for i, header := range td.Headers {
		if strings.EqualFold(header, column) {
			columnIndex = i
			break
		}
	}

	if columnIndex == -1 {
		return // Column not found
	}

	// Sort rows by the specified column
	sort.Slice(td.Rows, func(i, j int) bool {
		var cellI, cellJ string
		if columnIndex < len(td.Rows[i]) {
			cellI = td.Rows[i][columnIndex]
		}
		if columnIndex < len(td.Rows[j]) {
			cellJ = td.Rows[j][columnIndex]
		}

		// Try numeric comparison first
		if numI, errI := strconv.ParseFloat(cellI, 64); errI == nil {
			if numJ, errJ := strconv.ParseFloat(cellJ, 64); errJ == nil {
				if ascending {
					return numI < numJ
				}
				return numI > numJ
			}
		}

		// Fall back to string comparison
		result := strings.Compare(strings.ToLower(cellI), strings.ToLower(cellJ))
		if ascending {
			return result < 0
		}
		return result > 0
	})
}
