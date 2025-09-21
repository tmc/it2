package client

import (
	"context"
	"fmt"

	pb "github.com/tmc/it2/proto"
)

// BroadcastDomain represents a broadcast domain containing session IDs
type BroadcastDomain struct {
	SessionIds []string `json:"session_ids"`
}

// GetBroadcastDomains returns the current broadcast domains
func (c *Client) GetBroadcastDomains(ctx context.Context) ([]*BroadcastDomain, error) {
	request := &pb.ClientOriginatedMessage{
		Submessage: &pb.ClientOriginatedMessage_GetBroadcastDomainsRequest{
			GetBroadcastDomainsRequest: &pb.GetBroadcastDomainsRequest{},
		},
	}

	response, err := c.SendRequest(ctx, request)
	if err != nil {
		return nil, err
	}

	resp := response.GetGetBroadcastDomainsResponse()
	if resp == nil {
		return nil, fmt.Errorf("no broadcast domains response received")
	}

	domains := make([]*BroadcastDomain, len(resp.BroadcastDomains))
	for i, domain := range resp.BroadcastDomains {
		domains[i] = &BroadcastDomain{
			SessionIds: domain.SessionIds,
		}
	}

	return domains, nil
}

// SetBroadcastDomains sets the broadcast domains
func (c *Client) SetBroadcastDomains(ctx context.Context, domains []*BroadcastDomain) error {
	pbDomains := make([]*pb.BroadcastDomain, len(domains))
	for i, domain := range domains {
		pbDomains[i] = &pb.BroadcastDomain{
			SessionIds: domain.SessionIds,
		}
	}

	request := &pb.ClientOriginatedMessage{
		Submessage: &pb.ClientOriginatedMessage_SetBroadcastDomainsRequest{
			SetBroadcastDomainsRequest: &pb.SetBroadcastDomainsRequest{
				BroadcastDomains: pbDomains,
			},
		},
	}

	response, err := c.SendRequest(ctx, request)
	if err != nil {
		return err
	}

	resp := response.GetSetBroadcastDomainsResponse()
	if resp == nil {
		return fmt.Errorf("no set broadcast domains response received")
	}

	if resp.GetStatus() != pb.SetBroadcastDomainsResponse_OK {
		return fmt.Errorf("set broadcast domains failed: %v", resp.GetStatus())
	}

	return nil
}

// AddToBroadcastGroup adds session IDs to a broadcast group
func (c *Client) AddToBroadcastGroup(ctx context.Context, groupName string, sessionIDs []string) error {
	// Get current domains
	domains, err := c.GetBroadcastDomains(ctx)
	if err != nil {
		return fmt.Errorf("failed to get current broadcast domains: %w", err)
	}

	// Find or create the named group
	var targetDomain *BroadcastDomain
	for _, domain := range domains {
		// For simplicity, we'll use the first session ID as a group identifier
		// In a real implementation, you might want to store group names separately
		if len(domain.SessionIds) > 0 && domain.SessionIds[0] == groupName {
			targetDomain = domain
			break
		}
	}

	if targetDomain == nil {
		// Create new domain
		targetDomain = &BroadcastDomain{
			SessionIds: []string{groupName}, // Group name as identifier
		}
		domains = append(domains, targetDomain)
	}

	// Add session IDs to the domain
	for _, sessionID := range sessionIDs {
		// Check if session is already in the domain
		found := false
		for _, existing := range targetDomain.SessionIds {
			if existing == sessionID {
				found = true
				break
			}
		}
		if !found {
			targetDomain.SessionIds = append(targetDomain.SessionIds, sessionID)
		}
	}

	return c.SetBroadcastDomains(ctx, domains)
}

// RemoveFromBroadcastGroup removes session IDs from a broadcast group
func (c *Client) RemoveFromBroadcastGroup(ctx context.Context, groupName string, sessionIDs []string) error {
	// Get current domains
	domains, err := c.GetBroadcastDomains(ctx)
	if err != nil {
		return fmt.Errorf("failed to get current broadcast domains: %w", err)
	}

	// Find the named group
	var targetDomain *BroadcastDomain
	for _, domain := range domains {
		if len(domain.SessionIds) > 0 && domain.SessionIds[0] == groupName {
			targetDomain = domain
			break
		}
	}

	if targetDomain == nil {
		return fmt.Errorf("broadcast group '%s' not found", groupName)
	}

	// Remove session IDs from the domain
	for _, sessionID := range sessionIDs {
		for i, existing := range targetDomain.SessionIds {
			if existing == sessionID {
				targetDomain.SessionIds = append(targetDomain.SessionIds[:i], targetDomain.SessionIds[i+1:]...)
				break
			}
		}
	}

	return c.SetBroadcastDomains(ctx, domains)
}

// SendTextToBroadcastGroup sends text to all sessions in a broadcast group
func (c *Client) SendTextToBroadcastGroup(ctx context.Context, groupName, text string) error {
	// Get current domains
	domains, err := c.GetBroadcastDomains(ctx)
	if err != nil {
		return fmt.Errorf("failed to get current broadcast domains: %w", err)
	}

	// Find the named group
	var targetDomain *BroadcastDomain
	for _, domain := range domains {
		if len(domain.SessionIds) > 0 && domain.SessionIds[0] == groupName {
			targetDomain = domain
			break
		}
	}

	if targetDomain == nil {
		return fmt.Errorf("broadcast group '%s' not found", groupName)
	}

	// Send text to all sessions in the group (except the first which is the group name)
	for i := 1; i < len(targetDomain.SessionIds); i++ {
		sessionID := targetDomain.SessionIds[i]
		err := c.SendText(ctx, sessionID, text)
		if err != nil {
			return fmt.Errorf("failed to send text to session %s: %w", sessionID, err)
		}
	}

	return nil
}