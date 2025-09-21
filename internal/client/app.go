package client

import (
	"context"
	"fmt"

	pb "github.com/tmc/it2/proto"
)

// Focus brings iTerm2 to the foreground
func (c *Client) Focus(ctx context.Context, raiseAll bool) (*pb.FocusResponse, error) {
	msg := &pb.ClientOriginatedMessage{
		Submessage: &pb.ClientOriginatedMessage_FocusRequest{
			FocusRequest: &pb.FocusRequest{
				// Note: raiseAll parameter not supported by current protobuf API
			},
		},
	}

	response, err := c.SendRequest(ctx, msg)
	if err != nil {
		return nil, err
	}

	if response.GetFocusResponse() != nil {
		return response.GetFocusResponse(), nil
	}

	return nil, fmt.Errorf("unexpected response type")
}

// GetProperty gets an app-level property
func (c *Client) GetProperty(ctx context.Context, propertyName string) (string, error) {
	msg := &pb.ClientOriginatedMessage{
		Submessage: &pb.ClientOriginatedMessage_GetPropertyRequest{
			GetPropertyRequest: &pb.GetPropertyRequest{
				Name: &propertyName,
				// No identifier means app-level property
			},
		},
	}

	response, err := c.SendRequest(ctx, msg)
	if err != nil {
		return "", err
	}

	propResp := response.GetGetPropertyResponse()
	if propResp == nil {
		return "", fmt.Errorf("unexpected response type")
	}

	if propResp.GetStatus() != pb.GetPropertyResponse_OK {
		switch propResp.GetStatus() {
		case pb.GetPropertyResponse_UNRECOGNIZED_NAME:
			return "", fmt.Errorf("unrecognized property name: %s", propertyName)
		case pb.GetPropertyResponse_INVALID_TARGET:
			return "", fmt.Errorf("invalid target for property: %s", propertyName)
		default:
			return "", fmt.Errorf("failed to get property: %s", propResp.GetStatus().String())
		}
	}

	return propResp.GetJsonValue(), nil
}

// SetProperty sets an app-level property
func (c *Client) SetProperty(ctx context.Context, propertyName, jsonValue string) error {
	msg := &pb.ClientOriginatedMessage{
		Submessage: &pb.ClientOriginatedMessage_SetPropertyRequest{
			SetPropertyRequest: &pb.SetPropertyRequest{
				Name:      &propertyName,
				JsonValue: &jsonValue,
				// No identifier means app-level property
			},
		},
	}

	response, err := c.SendRequest(ctx, msg)
	if err != nil {
		return err
	}

	propResp := response.GetSetPropertyResponse()
	if propResp == nil {
		return fmt.Errorf("unexpected response type")
	}

	if propResp.GetStatus() != pb.SetPropertyResponse_OK {
		switch propResp.GetStatus() {
		case pb.SetPropertyResponse_UNRECOGNIZED_NAME:
			return fmt.Errorf("unrecognized property name: %s", propertyName)
		case pb.SetPropertyResponse_INVALID_VALUE:
			return fmt.Errorf("invalid value for property %s: %s", propertyName, jsonValue)
		case pb.SetPropertyResponse_INVALID_TARGET:
			return fmt.Errorf("invalid target for property: %s", propertyName)
		case pb.SetPropertyResponse_DEFERRED:
			return fmt.Errorf("property set deferred (will be tried later): %s", propertyName)
		case pb.SetPropertyResponse_IMPOSSIBLE:
			return fmt.Errorf("property cannot be set: %s", propertyName)
		case pb.SetPropertyResponse_FAILED:
			return fmt.Errorf("failed to set property: %s", propertyName)
		default:
			return fmt.Errorf("failed to set property: %s", propResp.GetStatus().String())
		}
	}

	return nil
}