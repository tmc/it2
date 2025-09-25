package formatting

import (
	"reflect"
	"testing"
)

func TestNewTableData(t *testing.T) {
	headers := []string{"Name", "Age", "City"}
	table := NewTableData(headers)

	if table == nil {
		t.Fatal("Expected non-nil table data")
	}

	if !reflect.DeepEqual(table.Headers, headers) {
		t.Errorf("Expected headers %v, got %v", headers, table.Headers)
	}

	if len(table.Rows) != 0 {
		t.Errorf("Expected 0 rows, got %d", len(table.Rows))
	}
}

func TestTableData_AddRow(t *testing.T) {
	headers := []string{"Name", "Age", "City"}
	table := NewTableData(headers)

	// Add a complete row
	row1 := []string{"Alice", "25", "New York"}
	table.AddRow(row1)

	if len(table.Rows) != 1 {
		t.Fatalf("Expected 1 row, got %d", len(table.Rows))
	}

	if !reflect.DeepEqual(table.Rows[0], row1) {
		t.Errorf("Expected row %v, got %v", row1, table.Rows[0])
	}

	// Add a partial row (should be padded)
	row2 := []string{"Bob", "30"}
	table.AddRow(row2)

	if len(table.Rows) != 2 {
		t.Fatalf("Expected 2 rows, got %d", len(table.Rows))
	}

	expectedRow2 := []string{"Bob", "30", ""}
	if !reflect.DeepEqual(table.Rows[1], expectedRow2) {
		t.Errorf("Expected padded row %v, got %v", expectedRow2, table.Rows[1])
	}
}

func TestTableData_SelectColumns(t *testing.T) {
	headers := []string{"Name", "Age", "City", "Country"}
	table := NewTableData(headers)
	table.AddRow([]string{"Alice", "25", "New York", "USA"})
	table.AddRow([]string{"Bob", "30", "London", "UK"})

	// Select specific columns
	selectedCols := []string{"Name", "City"}
	filteredTable := table.SelectColumns(selectedCols)

	expectedHeaders := []string{"Name", "City"}
	if !reflect.DeepEqual(filteredTable.Headers, expectedHeaders) {
		t.Errorf("Expected headers %v, got %v", expectedHeaders, filteredTable.Headers)
	}

	expectedRows := [][]string{
		{"Alice", "New York"},
		{"Bob", "London"},
	}
	if !reflect.DeepEqual(filteredTable.Rows, expectedRows) {
		t.Errorf("Expected rows %v, got %v", expectedRows, filteredTable.Rows)
	}
}

func TestTableData_SelectColumns_CaseInsensitive(t *testing.T) {
	headers := []string{"Name", "Age", "City"}
	table := NewTableData(headers)
	table.AddRow([]string{"Alice", "25", "New York"})

	// Test case insensitive column selection
	selectedCols := []string{"name", "CITY"}
	filteredTable := table.SelectColumns(selectedCols)

	expectedHeaders := []string{"Name", "City"}
	if !reflect.DeepEqual(filteredTable.Headers, expectedHeaders) {
		t.Errorf("Expected headers %v, got %v", expectedHeaders, filteredTable.Headers)
	}

	expectedRows := [][]string{
		{"Alice", "New York"},
	}
	if !reflect.DeepEqual(filteredTable.Rows, expectedRows) {
		t.Errorf("Expected rows %v, got %v", expectedRows, filteredTable.Rows)
	}
}

func TestTableData_SelectColumns_EmptyColumns(t *testing.T) {
	headers := []string{"Name", "Age", "City"}
	table := NewTableData(headers)
	table.AddRow([]string{"Alice", "25", "New York"})

	// Empty column selection should return original table
	filteredTable := table.SelectColumns([]string{})

	if filteredTable != table {
		t.Error("Expected same table instance for empty column selection")
	}
}

func TestTableData_SelectColumns_InvalidColumns(t *testing.T) {
	headers := []string{"Name", "Age", "City"}
	table := NewTableData(headers)
	table.AddRow([]string{"Alice", "25", "New York"})

	// Invalid column selection should return original table
	filteredTable := table.SelectColumns([]string{"NonExistent"})

	if filteredTable != table {
		t.Error("Expected same table instance for invalid column selection")
	}
}

func TestTableData_SortBy_StringAscending(t *testing.T) {
	headers := []string{"Name", "Age", "City"}
	table := NewTableData(headers)
	table.AddRow([]string{"Charlie", "35", "Boston"})
	table.AddRow([]string{"Alice", "25", "New York"})
	table.AddRow([]string{"Bob", "30", "London"})

	table.SortBy("Name", true)

	expectedRows := [][]string{
		{"Alice", "25", "New York"},
		{"Bob", "30", "London"},
		{"Charlie", "35", "Boston"},
	}
	if !reflect.DeepEqual(table.Rows, expectedRows) {
		t.Errorf("Expected sorted rows %v, got %v", expectedRows, table.Rows)
	}
}

func TestTableData_SortBy_StringDescending(t *testing.T) {
	headers := []string{"Name", "Age", "City"}
	table := NewTableData(headers)
	table.AddRow([]string{"Alice", "25", "New York"})
	table.AddRow([]string{"Charlie", "35", "Boston"})
	table.AddRow([]string{"Bob", "30", "London"})

	table.SortBy("Name", false)

	expectedRows := [][]string{
		{"Charlie", "35", "Boston"},
		{"Bob", "30", "London"},
		{"Alice", "25", "New York"},
	}
	if !reflect.DeepEqual(table.Rows, expectedRows) {
		t.Errorf("Expected sorted rows %v, got %v", expectedRows, table.Rows)
	}
}

func TestTableData_SortBy_NumericAscending(t *testing.T) {
	headers := []string{"Name", "Age", "City"}
	table := NewTableData(headers)
	table.AddRow([]string{"Charlie", "35", "Boston"})
	table.AddRow([]string{"Alice", "25", "New York"})
	table.AddRow([]string{"Bob", "30", "London"})

	table.SortBy("Age", true)

	expectedRows := [][]string{
		{"Alice", "25", "New York"},
		{"Bob", "30", "London"},
		{"Charlie", "35", "Boston"},
	}
	if !reflect.DeepEqual(table.Rows, expectedRows) {
		t.Errorf("Expected numerically sorted rows %v, got %v", expectedRows, table.Rows)
	}
}

func TestTableData_SortBy_NumericDescending(t *testing.T) {
	headers := []string{"Name", "Age", "City"}
	table := NewTableData(headers)
	table.AddRow([]string{"Alice", "25", "New York"})
	table.AddRow([]string{"Bob", "30", "London"})
	table.AddRow([]string{"Charlie", "35", "Boston"})

	table.SortBy("Age", false)

	expectedRows := [][]string{
		{"Charlie", "35", "Boston"},
		{"Bob", "30", "London"},
		{"Alice", "25", "New York"},
	}
	if !reflect.DeepEqual(table.Rows, expectedRows) {
		t.Errorf("Expected numerically sorted rows %v, got %v", expectedRows, table.Rows)
	}
}

func TestTableData_SortBy_CaseInsensitive(t *testing.T) {
	headers := []string{"Name", "Age", "City"}
	table := NewTableData(headers)
	table.AddRow([]string{"alice", "25", "New York"})
	table.AddRow([]string{"Bob", "30", "London"})
	table.AddRow([]string{"CHARLIE", "35", "Boston"})

	table.SortBy("name", true) // Case insensitive column name

	expectedRows := [][]string{
		{"alice", "25", "New York"},
		{"Bob", "30", "London"},
		{"CHARLIE", "35", "Boston"},
	}
	if !reflect.DeepEqual(table.Rows, expectedRows) {
		t.Errorf("Expected case-insensitive sorted rows %v, got %v", expectedRows, table.Rows)
	}
}

func TestTableData_SortBy_InvalidColumn(t *testing.T) {
	headers := []string{"Name", "Age", "City"}
	table := NewTableData(headers)
	originalRow := []string{"Alice", "25", "New York"}
	table.AddRow(originalRow)

	originalRows := make([][]string, len(table.Rows))
	copy(originalRows, table.Rows)

	table.SortBy("NonExistent", true)

	// Table should remain unchanged
	if !reflect.DeepEqual(table.Rows, originalRows) {
		t.Error("Expected table to remain unchanged when sorting by invalid column")
	}
}

func TestTableData_SortBy_EmptyTable(t *testing.T) {
	headers := []string{"Name", "Age", "City"}
	table := NewTableData(headers)

	// Should not panic on empty table
	table.SortBy("Name", true)

	if len(table.Rows) != 0 {
		t.Error("Expected empty table to remain empty after sorting")
	}
}

func TestTableData_SortBy_MixedNumericString(t *testing.T) {
	headers := []string{"Value"}
	table := NewTableData(headers)
	table.AddRow([]string{"100"})
	table.AddRow([]string{"text"})
	table.AddRow([]string{"50"})
	table.AddRow([]string{"another"})

	table.SortBy("Value", true)

	// Numeric values should come first (sorted numerically), then strings (sorted alphabetically)
	expectedRows := [][]string{
		{"50"},
		{"100"},
		{"another"},
		{"text"},
	}
	if !reflect.DeepEqual(table.Rows, expectedRows) {
		t.Errorf("Expected mixed sort result %v, got %v", expectedRows, table.Rows)
	}
}

func TestTableData_SortBy_FloatingPoint(t *testing.T) {
	headers := []string{"Score"}
	table := NewTableData(headers)
	table.AddRow([]string{"95.5"})
	table.AddRow([]string{"100.0"})
	table.AddRow([]string{"87.25"})

	table.SortBy("Score", true)

	expectedRows := [][]string{
		{"87.25"},
		{"95.5"},
		{"100.0"},
	}
	if !reflect.DeepEqual(table.Rows, expectedRows) {
		t.Errorf("Expected floating point sorted rows %v, got %v", expectedRows, table.Rows)
	}
}