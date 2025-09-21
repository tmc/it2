package client

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	pb "github.com/tmc/it2/proto"
)

// ListProfilesSimple returns just profile names
func (c *Client) ListProfilesSimple(ctx context.Context) ([]string, error) {
	msg := &pb.ClientOriginatedMessage{
		Submessage: &pb.ClientOriginatedMessage_ListProfilesRequest{
			ListProfilesRequest: &pb.ListProfilesRequest{
				Properties: []string{"Name"}, // Only request names
			},
		},
	}

	response, err := c.SendRequest(ctx, msg)
	if err != nil {
		return nil, fmt.Errorf("failed to list profiles: %w", err)
	}

	if resp := response.GetListProfilesResponse(); resp != nil {
		profiles := resp.GetProfiles()
		names := make([]string, len(profiles))
		for i, profile := range profiles {
			for _, prop := range profile.GetProperties() {
				if prop.GetKey() == "Name" {
					names[i] = strings.Trim(prop.GetJsonValue(), "\"")
					break
				}
			}
		}
		return names, nil
	}

	return nil, fmt.Errorf("unexpected response type")
}

// ListProfilesDetailed returns detailed profile information
func (c *Client) ListProfilesDetailed(ctx context.Context) ([]*pb.ListProfilesResponse_Profile, error) {
	msg := &pb.ClientOriginatedMessage{
		Submessage: &pb.ClientOriginatedMessage_ListProfilesRequest{
			ListProfilesRequest: &pb.ListProfilesRequest{},
		},
	}

	response, err := c.SendRequest(ctx, msg)
	if err != nil {
		return nil, fmt.Errorf("failed to list profiles: %w", err)
	}

	if resp := response.GetListProfilesResponse(); resp != nil {
		return resp.GetProfiles(), nil
	}

	return nil, fmt.Errorf("unexpected response type")
}

// ListProfiles returns profile names, optionally with detailed information
// This is the main method that should be used
func (c *Client) ListProfiles(ctx context.Context, detailed bool) ([]string, error) {
	if !detailed {
		return c.ListProfilesSimple(ctx)
	}

	profiles, err := c.ListProfilesDetailed(ctx)
	if err != nil {
		return nil, err
	}

	names := make([]string, len(profiles))
	for i, profile := range profiles {
		// Extract name from properties
		for _, prop := range profile.GetProperties() {
			if prop.GetKey() == "Name" {
				names[i] = strings.Trim(prop.GetJsonValue(), "\"")
				break
			}
		}
	}
	return names, nil
}

// GetProfile gets all properties for a specific profile by name
func (c *Client) GetProfile(ctx context.Context, profileName string) (map[string]interface{}, error) {
	// First list all profiles to find the GUID for the given name
	profiles, err := c.ListProfilesDetailed(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to list profiles: %w", err)
	}

	var profileGUID string
	var targetProfile *pb.ListProfilesResponse_Profile

	for _, profile := range profiles {
		for _, prop := range profile.GetProperties() {
			if prop.GetKey() == "Name" && strings.Trim(prop.GetJsonValue(), "\"") == profileName {
				targetProfile = profile
				// Find the GUID property for this profile
				for _, guidProp := range profile.GetProperties() {
					if guidProp.GetKey() == "Guid" {
						profileGUID = strings.Trim(guidProp.GetJsonValue(), "\"")
						break
					}
				}
				break
			}
		}
		if profileGUID != "" {
			break
		}
	}

	if targetProfile == nil {
		return nil, fmt.Errorf("profile not found: %s", profileName)
	}

	// Convert all properties to a map
	properties := make(map[string]interface{})
	for _, prop := range targetProfile.GetProperties() {
		// Parse JSON value
		var value interface{}
		jsonValue := prop.GetJsonValue()
		if err := json.Unmarshal([]byte(jsonValue), &value); err != nil {
			// If JSON parsing fails, store as string
			properties[prop.GetKey()] = jsonValue
		} else {
			properties[prop.GetKey()] = value
		}
	}

	return properties, nil
}

// GetProfileProperty gets a specific property for a profile by name
// Deprecated: Use GetProfile to get all properties at once
func (c *Client) GetProfileProperty(ctx context.Context, profileName, key string) (interface{}, error) {
	props, err := c.GetProfile(ctx, profileName)
	if err != nil {
		return nil, err
	}

	value, ok := props[key]
	if !ok {
		return nil, fmt.Errorf("property not found: %s", key)
	}

	return value, nil
}

// SetProfileProperty sets a specific property for a profile by name
func (c *Client) SetProfileProperty(ctx context.Context, profileName, key string, value interface{}) error {
	// First get the profile to find its GUID
	props, err := c.GetProfile(ctx, profileName)
	if err != nil {
		return err
	}

	profileGUID, ok := props["Guid"].(string)
	if !ok {
		return fmt.Errorf("profile GUID not found for: %s", profileName)
	}

	// Convert value to JSON
	var jsonValue string
	if valueStr, ok := value.(string); ok {
		// For string values, check if it's already JSON formatted
		if strings.HasPrefix(valueStr, "\"") && strings.HasSuffix(valueStr, "\"") {
			jsonValue = valueStr
		} else {
			jsonValue = fmt.Sprintf("\"%s\"", valueStr)
		}
	} else {
		jsonBytes, err := json.Marshal(value)
		if err != nil {
			return fmt.Errorf("failed to marshal value: %w", err)
		}
		jsonValue = string(jsonBytes)
	}

	msg := &pb.ClientOriginatedMessage{
		Submessage: &pb.ClientOriginatedMessage_SetProfilePropertyRequest{
			SetProfilePropertyRequest: &pb.SetProfilePropertyRequest{
				Target: &pb.SetProfilePropertyRequest_GuidList_{
					GuidList: &pb.SetProfilePropertyRequest_GuidList{
						Guids: []string{profileGUID},
					},
				},
				Key:       stringPtr(key),
				JsonValue: stringPtr(jsonValue),
			},
		},
	}

	response, err := c.SendRequest(ctx, msg)
	if err != nil {
		return fmt.Errorf("failed to set profile property: %w", err)
	}

	if resp := response.GetSetProfilePropertyResponse(); resp != nil {
		if resp.GetStatus() != pb.SetProfilePropertyResponse_OK {
			return fmt.Errorf("failed to set profile property: %v", resp.GetStatus())
		}
	}

	return nil
}

// SetProfileProperties sets multiple properties for a profile efficiently
// This avoids repeated GUID lookups by doing it once upfront
func (c *Client) SetProfileProperties(ctx context.Context, profileName string, properties map[string]interface{}) error {
	if len(properties) == 0 {
		return nil
	}

	// Get the profile GUID once
	props, err := c.GetProfile(ctx, profileName)
	if err != nil {
		return err
	}

	profileGUID, ok := props["Guid"].(string)
	if !ok {
		return fmt.Errorf("profile GUID not found for: %s", profileName)
	}

	// Set properties one by one using the GUID we already have
	successCount := 0
	errorCount := 0

	for key, value := range properties {
		// Skip certain properties that shouldn't be imported
		if key == "Guid" {
			continue
		}

		// Convert value to JSON
		var jsonValue string
		if valueStr, ok := value.(string); ok {
			// For string values, check if it's already JSON formatted
			if strings.HasPrefix(valueStr, "\"") && strings.HasSuffix(valueStr, "\"") {
				jsonValue = valueStr
			} else {
				jsonValue = fmt.Sprintf("\"%s\"", valueStr)
			}
		} else {
			jsonBytes, err := json.Marshal(value)
			if err != nil {
				fmt.Printf("  Failed to marshal %s: %v\n", key, err)
				errorCount++
				continue
			}
			jsonValue = string(jsonBytes)
		}

		msg := &pb.ClientOriginatedMessage{
			Submessage: &pb.ClientOriginatedMessage_SetProfilePropertyRequest{
				SetProfilePropertyRequest: &pb.SetProfilePropertyRequest{
					Target: &pb.SetProfilePropertyRequest_GuidList_{
						GuidList: &pb.SetProfilePropertyRequest_GuidList{
							Guids: []string{profileGUID},
						},
					},
					Key:       stringPtr(key),
					JsonValue: stringPtr(jsonValue),
				},
			},
		}

		response, err := c.SendRequest(ctx, msg)
		if err != nil {
			fmt.Printf("  Failed to set %s: %v\n", key, err)
			errorCount++
			continue
		}

		if resp := response.GetSetProfilePropertyResponse(); resp != nil {
			if resp.GetStatus() != pb.SetProfilePropertyResponse_OK {
				fmt.Printf("  Failed to set %s: %v\n", key, resp.GetStatus())
				errorCount++
				continue
			}
		}

		fmt.Printf("  Set %s\n", key)
		successCount++
	}

	fmt.Printf("Bulk update completed: %d properties set successfully, %d failed\n", successCount, errorCount)
	return nil
}