package formatting

import (
	"encoding/json"
	"os"

	"gopkg.in/yaml.v3"
)

func (f *Formatter) formatJSON(v interface{}) error {
	encoder := json.NewEncoder(os.Stdout)
	encoder.SetIndent("", "  ")
	return encoder.Encode(v)
}

// PrintJSON is a convenience function for printing JSON output.
func PrintJSON(v interface{}) error {
	encoder := json.NewEncoder(os.Stdout)
	encoder.SetIndent("", "  ")
	return encoder.Encode(v)
}

func (f *Formatter) formatYAML(v interface{}) error {
	return PrintYAML(v)
}

// PrintYAML is a convenience function for printing YAML output.
func PrintYAML(v interface{}) error {
	encoder := yaml.NewEncoder(os.Stdout)
	defer encoder.Close()
	return encoder.Encode(v)
}

// FormatGeneric formats any interface{} as JSON.
func (f *Formatter) FormatGeneric(v interface{}) error {
	switch f.format {
	case "json":
		return f.formatJSON(v)
	case "yaml":
		return f.formatYAML(v)
	default:
		// For text format, just use JSON as well
		return f.formatJSON(v)
	}
}
