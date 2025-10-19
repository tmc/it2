// Package jqutil provides jq filtering utilities for JSON output.
package jqutil

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/itchyny/gojq"
)

// ApplyJQ applies a jq expression to JSON data and returns the filtered/transformed result.
// The input should be valid JSON (typically an array of objects).
// The output is formatted based on what the jq expression returns.
func ApplyJQ(jsonData []byte, jqExpr string) (string, error) {
	// Parse the jq expression
	query, err := gojq.Parse(jqExpr)
	if err != nil {
		return "", fmt.Errorf("invalid jq expression: %w", err)
	}

	// Parse the JSON input
	var input interface{}
	if err := json.Unmarshal(jsonData, &input); err != nil {
		return "", fmt.Errorf("failed to parse JSON: %w", err)
	}

	// Run the jq query
	iter := query.Run(input)
	var results []interface{}

	for {
		v, ok := iter.Next()
		if !ok {
			break
		}
		if err, ok := v.(error); ok {
			return "", fmt.Errorf("jq execution error: %w", err)
		}
		results = append(results, v)
	}

	// Format the output
	// If there's only one result, output it directly (not in an array)
	// If there are multiple results, output them line by line
	var output strings.Builder
	for i, result := range results {
		if i > 0 {
			output.WriteString("\n")
		}

		// Check if the result is a string (common for formatted output)
		if str, ok := result.(string); ok {
			output.WriteString(str)
		} else {
			// Otherwise, marshal to JSON
			resultJSON, err := json.Marshal(result)
			if err != nil {
				return "", fmt.Errorf("failed to marshal result: %w", err)
			}
			output.Write(resultJSON)
		}
	}

	return output.String(), nil
}

// FilterJSONFields filters a JSON array to only include specified fields.
// fields is a comma-separated list of field names.
// If fields is empty, returns the original JSON unchanged.
func FilterJSONFields(jsonData []byte, fields string) ([]byte, error) {
	if fields == "" {
		// No filtering, return original
		return jsonData, nil
	}

	// Parse the JSON input
	var input []interface{}
	if err := json.Unmarshal(jsonData, &input); err != nil {
		return nil, fmt.Errorf("failed to parse JSON: %w", err)
	}

	// Split the fields list
	fieldList := strings.Split(fields, ",")
	for i, f := range fieldList {
		fieldList[i] = strings.TrimSpace(f)
	}

	// Build a jq expression to select only the specified fields
	// For each object in the array, select only the requested fields
	jqFields := make([]string, len(fieldList))
	for i, field := range fieldList {
		// Map common aliases to actual field names
		field = mapFieldAlias(field)
		jqFields[i] = fmt.Sprintf("%s: .%s", field, field)
	}
	jqExpr := fmt.Sprintf(".[] | {%s}", strings.Join(jqFields, ", "))

	// Apply the jq expression
	result, err := ApplyJQ(jsonData, jqExpr)
	if err != nil {
		return nil, err
	}

	// Parse the result back to ensure it's valid JSON array
	// Split by newlines and wrap in array
	lines := strings.Split(strings.TrimSpace(result), "\n")
	var objects []interface{}
	for _, line := range lines {
		if line == "" {
			continue
		}
		var obj interface{}
		if err := json.Unmarshal([]byte(line), &obj); err != nil {
			return nil, fmt.Errorf("failed to parse filtered result: %w", err)
		}
		objects = append(objects, obj)
	}

	// Marshal back to JSON array
	return json.Marshal(objects)
}

// mapFieldAlias maps common field aliases to their actual JSON field names.
func mapFieldAlias(field string) string {
	aliases := map[string]string{
		"id":      "SessionID",
		"name":    "SessionName",
		"title":   "TabTitle",
		"wd":      "WorkingDirectory",
		"cwd":     "WorkingDirectory",
		"command": "CurrentCommand",
		"pid":     "ShellPID",
	}

	if actual, ok := aliases[strings.ToLower(field)]; ok {
		return actual
	}
	return field
}
