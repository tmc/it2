package client

import (
	"context"
	"fmt"

	pb "github.com/tmc/it2/proto"
)

// CreateTab creates a new tab in an existing window
func (c *Client) CreateTab(ctx context.Context, profileName, windowID string) (*pb.CreateTabResponse, error) {
	return c.CreateTabWithOptions(ctx, profileName, windowID, 0, "")
}

// CreateTabWithOptions creates a new tab with advanced options
func (c *Client) CreateTabWithOptions(ctx context.Context, profileName, windowID string, tabIndex uint32, command string) (*pb.CreateTabResponse, error) {
	req := &pb.CreateTabRequest{
		ProfileName: &profileName,
	}

	if windowID != "" {
		req.WindowId = &windowID
	}
	if tabIndex > 0 {
		req.TabIndex = &tabIndex
	}
	if command != "" {
		req.Command = &command
	}

	msg := &pb.ClientOriginatedMessage{
		Submessage: &pb.ClientOriginatedMessage_CreateTabRequest{
			CreateTabRequest: req,
		},
	}

	response, err := c.SendRequest(ctx, msg)
	if err != nil {
		return nil, err
	}

	if response.GetCreateTabResponse() != nil {
		return response.GetCreateTabResponse(), nil
	}

	return nil, fmt.Errorf("unexpected response type")
}

// CloseTabs closes one or more tabs
func (c *Client) CloseTabs(ctx context.Context, tabIDs []string, force bool) (*pb.CloseResponse, error) {
	msg := &pb.ClientOriginatedMessage{
		Submessage: &pb.ClientOriginatedMessage_CloseRequest{
			CloseRequest: &pb.CloseRequest{
				Target: &pb.CloseRequest_Tabs{
					Tabs: &pb.CloseRequest_CloseTabs{
						TabIds: tabIDs,
					},
				},
				Force: &force,
			},
		},
	}

	response, err := c.SendRequest(ctx, msg)
	if err != nil {
		return nil, err
	}

	if response.GetCloseResponse() != nil {
		return response.GetCloseResponse(), nil
	}

	return nil, fmt.Errorf("unexpected response type")
}

// ActivateTab activates and optionally selects a tab
func (c *Client) ActivateTab(ctx context.Context, tabID string, selectTab bool) (*pb.ActivateResponse, error) {
	msg := &pb.ClientOriginatedMessage{
		Submessage: &pb.ClientOriginatedMessage_ActivateRequest{
			ActivateRequest: &pb.ActivateRequest{
				Identifier: &pb.ActivateRequest_TabId{
					TabId: tabID,
				},
				SelectTab: &selectTab,
			},
		},
	}

	response, err := c.SendRequest(ctx, msg)
	if err != nil {
		return nil, err
	}

	if response.GetActivateResponse() != nil {
		return response.GetActivateResponse(), nil
	}

	return nil, fmt.Errorf("unexpected response type")
}

// ReorderTabs reorders tabs within windows
func (c *Client) ReorderTabs(ctx context.Context, assignments map[string][]string) (*pb.ReorderTabsResponse, error) {
	var reqAssignments []*pb.ReorderTabsRequest_Assignment

	for windowID, tabIDs := range assignments {
		assignment := &pb.ReorderTabsRequest_Assignment{
			WindowId: &windowID,
			TabIds:   tabIDs,
		}
		reqAssignments = append(reqAssignments, assignment)
	}

	msg := &pb.ClientOriginatedMessage{
		Submessage: &pb.ClientOriginatedMessage_ReorderTabsRequest{
			ReorderTabsRequest: &pb.ReorderTabsRequest{
				Assignments: reqAssignments,
			},
		},
	}

	response, err := c.SendRequest(ctx, msg)
	if err != nil {
		return nil, err
	}

	if response.GetReorderTabsResponse() != nil {
		return response.GetReorderTabsResponse(), nil
	}

	return nil, fmt.Errorf("unexpected response type")
}

// SetTabLayout sets the layout of a tab
func (c *Client) SetTabLayout(ctx context.Context, tabID string, root *pb.SplitTreeNode) (*pb.SetTabLayoutResponse, error) {
	msg := &pb.ClientOriginatedMessage{
		Submessage: &pb.ClientOriginatedMessage_SetTabLayoutRequest{
			SetTabLayoutRequest: &pb.SetTabLayoutRequest{
				TabId: &tabID,
				Root:  root,
			},
		},
	}

	response, err := c.SendRequest(ctx, msg)
	if err != nil {
		return nil, err
	}

	if response.GetSetTabLayoutResponse() != nil {
		return response.GetSetTabLayoutResponse(), nil
	}

	return nil, fmt.Errorf("unexpected response type")
}