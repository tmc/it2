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
