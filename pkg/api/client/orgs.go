package client

import (
	"context"
	"fmt"

	"github.com/powertoolsdev/mono/pkg/api/client/client/operations"
	"github.com/powertoolsdev/mono/pkg/api/client/models"
)

func (c *client) GetOrg(ctx context.Context, orgID string) (*models.AppOrg, error) {
	resp, err := c.genClient.Operations.GetV1OrgsOrgID(&operations.GetV1OrgsOrgIDParams{
		OrgID:   orgID,
		Context: ctx,
	})
	if err != nil {
		return nil, fmt.Errorf("unable to get org: %w", err)
	}

	return resp.Payload, nil
}

func (c *client) GetOrgs(ctx context.Context) ([]*models.AppOrg, error) {
	resp, err := c.genClient.Operations.GetV1Orgs(&operations.GetV1OrgsParams{
		Context: ctx,
	})
	if err != nil {
		return nil, fmt.Errorf("unable to get orgs: %w", err)
	}

	return resp.Payload, nil
}

func (c *client) CreateOrg(ctx context.Context, req *models.ServiceCreateOrgRequest) (*models.AppOrg, error) {
	resp, err := c.genClient.Operations.PostV1Orgs(&operations.PostV1OrgsParams{
		Req:     req,
		Context: ctx,
	})
	if err != nil {
		return nil, fmt.Errorf("unable to create org: %w", err)
	}

	return resp.Payload, nil
}

func (c *client) UpdateOrg(ctx context.Context, orgID string, req *models.ServiceUpdateOrgRequest) (*models.AppOrg, error) {
	resp, err := c.genClient.Operations.PatchV1OrgsOrgID(&operations.PatchV1OrgsOrgIDParams{
		OrgID:   orgID,
		Req:     req,
		Context: ctx,
	})
	if err != nil {
		return nil, fmt.Errorf("unable to update org: %w", err)
	}

	return resp.Payload, nil
}

func (c *client) CreateOrgUser(ctx context.Context, orgID string, req *models.ServiceCreateOrgUserRequest) (*models.AppUserOrg, error) {
	resp, err := c.genClient.Operations.PostV1OrgsOrgIDUser(&operations.PostV1OrgsOrgIDUserParams{
		OrgID:   orgID,
		Req:     req,
		Context: ctx,
	})
	if err != nil {
		return nil, fmt.Errorf("unable to update org: %w", err)
	}

	return resp.Payload, nil
}
