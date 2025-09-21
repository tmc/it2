package client

import (
	"context"
	"encoding/xml"
	"fmt"
	"os"

	"github.com/tmc/it2/proto"
)

// ColorPreset represents a color preset with its settings
type ColorPreset struct {
	Name     string
	Settings map[string]*ColorSetting
}

// ColorSetting represents a single color setting
type ColorSetting struct {
	Red        float32
	Green      float32
	Blue       float32
	Alpha      float32
	ColorSpace string
}

// ListColorPresets returns a list of available color preset names
func (c *Client) ListColorPresets(ctx context.Context) ([]string, error) {
	msg := &proto.ClientOriginatedMessage{
		Submessage: &proto.ClientOriginatedMessage_ColorPresetRequest{
			ColorPresetRequest: &proto.ColorPresetRequest{
				Request: &proto.ColorPresetRequest_ListPresets_{
					ListPresets: &proto.ColorPresetRequest_ListPresets{},
				},
			},
		},
	}

	response, err := c.SendRequest(ctx, msg)
	if err != nil {
		return nil, err
	}

	resp := response.GetColorPresetResponse()
	if resp == nil {
		return nil, fmt.Errorf("expected ColorPresetResponse")
	}

	if resp.GetStatus() != proto.ColorPresetResponse_OK {
		return nil, fmt.Errorf("color preset request failed with status: %v", resp.GetStatus())
	}

	listResp := resp.GetListPresets()
	if listResp == nil {
		return nil, fmt.Errorf("expected ListPresets response")
	}

	return listResp.GetName(), nil
}

// GetColorPreset retrieves the details of a specific color preset
func (c *Client) GetColorPreset(ctx context.Context, name string) (*ColorPreset, error) {
	msg := &proto.ClientOriginatedMessage{
		Submessage: &proto.ClientOriginatedMessage_ColorPresetRequest{
			ColorPresetRequest: &proto.ColorPresetRequest{
				Request: &proto.ColorPresetRequest_GetPreset_{
					GetPreset: &proto.ColorPresetRequest_GetPreset{
						Name: &name,
					},
				},
			},
		},
	}

	response, err := c.SendRequest(ctx, msg)
	if err != nil {
		return nil, err
	}

	resp := response.GetColorPresetResponse()
	if resp == nil {
		return nil, fmt.Errorf("expected ColorPresetResponse")
	}

	if resp.GetStatus() != proto.ColorPresetResponse_OK {
		return nil, fmt.Errorf("color preset request failed with status: %v", resp.GetStatus())
	}

	getResp := resp.GetGetPreset()
	if getResp == nil {
		return nil, fmt.Errorf("expected GetPreset response")
	}

	preset := &ColorPreset{
		Name:     name,
		Settings: make(map[string]*ColorSetting),
	}

	for _, setting := range getResp.GetColorSettings() {
		preset.Settings[setting.GetKey()] = &ColorSetting{
			Red:        setting.GetRed(),
			Green:      setting.GetGreen(),
			Blue:       setting.GetBlue(),
			Alpha:      setting.GetAlpha(),
			ColorSpace: setting.GetColorSpace(),
		}
	}

	return preset, nil
}

// ITerm2ColorPreset represents the structure of an .itermcolors file
type ITerm2ColorPreset struct {
	XMLName xml.Name `xml:"plist"`
	Dict    Dict     `xml:"dict"`
}

// Dict represents a dictionary element in plist
type Dict struct {
	Keys   []string `xml:"key"`
	Values []Value  `xml:"dict"`
}

// Value represents a dictionary value in plist
type Value struct {
	Keys   []string  `xml:"key"`
	Reals  []float64 `xml:"real"`
	Strings []string `xml:"string"`
}

// ImportColorPreset imports a color preset from an .itermcolors file
func (c *Client) ImportColorPreset(ctx context.Context, name, filePath string) error {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("failed to read file: %w", err)
	}

	var plist ITerm2ColorPreset
	if err := xml.Unmarshal(data, &plist); err != nil {
		return fmt.Errorf("failed to parse .itermcolors file: %w", err)
	}

	// Convert the parsed plist to color settings and apply to a profile
	// For now, this would be a client-side operation to set profile properties
	// since the iTerm2 API doesn't have a direct import preset operation

	// This is a placeholder - the actual implementation would need to:
	// 1. Parse the color values from the plist
	// 2. Get the current profile
	// 3. Update the profile's color properties
	// 4. Or create a new preset by applying colors to the default profile

	return fmt.Errorf("import color preset not implemented - requires profile property updates")
}

// ExportColorPreset exports a color preset to an .itermcolors file
func (c *Client) ExportColorPreset(ctx context.Context, name, filePath string) error {
	preset, err := c.GetColorPreset(ctx, name)
	if err != nil {
		return fmt.Errorf("failed to get color preset: %w", err)
	}

	// Convert ColorPreset to .itermcolors plist format
	plist := ITerm2ColorPreset{
		Dict: Dict{
			Keys: make([]string, 0),
			Values: make([]Value, 0),
		},
	}

	// Add XML header and plist structure
	for key, setting := range preset.Settings {
		plist.Dict.Keys = append(plist.Dict.Keys, key)
		plist.Dict.Values = append(plist.Dict.Values, Value{
			Keys: []string{
				"Alpha Component",
				"Blue Component",
				"Color Space",
				"Green Component",
				"Red Component",
			},
			Reals: []float64{
				float64(setting.Alpha),
				float64(setting.Blue),
				float64(setting.Green),
				float64(setting.Red),
			},
			Strings: []string{setting.ColorSpace},
		})
	}

	output, err := xml.MarshalIndent(plist, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal plist: %w", err)
	}

	// Add XML header
	fullOutput := `<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
` + string(output)

	if err := os.WriteFile(filePath, []byte(fullOutput), 0644); err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}

	return nil
}

// ApplyColorPreset applies a color preset to a specific profile
func (c *Client) ApplyColorPreset(ctx context.Context, presetName, profileName string) error {
	preset, err := c.GetColorPreset(ctx, presetName)
	if err != nil {
		return fmt.Errorf("failed to get color preset: %w", err)
	}

	// Apply each color setting to the profile
	for key, setting := range preset.Settings {
		colorValue := map[string]interface{}{
			"Red Component":   setting.Red,
			"Green Component": setting.Green,
			"Blue Component":  setting.Blue,
			"Alpha Component": setting.Alpha,
			"Color Space":     setting.ColorSpace,
		}

		if err := c.SetProfileProperty(ctx, profileName, key, colorValue); err != nil {
			return fmt.Errorf("failed to set color property %s: %w", key, err)
		}
	}

	return nil
}