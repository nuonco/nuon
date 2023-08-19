package client

import (
	"context"
	"fmt"

	"github.com/powertoolsdev/mono/pkg/api/client/client/operations"
	"github.com/powertoolsdev/mono/pkg/api/client/models"
)

func (c *client) GetOrg(ctx context.Context) (*models.AppOrg, error) {
	resp, err := c.genClient.Operations.GetV1OrgsCurrent(&operations.GetV1OrgsCurrentParams{
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

func (c *client) DeleteOrg(ctx context.Context) (bool, error) {
	resp, err := c.genClient.Operations.DeleteV1OrgsCurrent(&operations.DeleteV1OrgsCurrentParams{
		Context: ctx,
	})
	if err != nil {
		return false, fmt.Errorf("unable to create org: %w", err)
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

func (c *client) UpdateOrg(ctx context.Context, req *models.ServiceUpdateOrgRequest) (*models.AppOrg, error) {
	resp, err := c.genClient.Operations.PatchV1OrgsCurrent(&operations.PatchV1OrgsCurrentParams{
		Req:     req,
		Context: ctx,
	})
	if err != nil {
		return nil, fmt.Errorf("unable to update org: %w", err)
	}

	return resp.Payload, nil
}

func (c *client) CreateOrgUser(ctx context.Context, req *models.ServiceCreateOrgUserRequest) (*models.AppUserOrg, error) {
	resp, err := c.genClient.Operations.PostV1OrgsCurrentUser(&operations.PostV1OrgsCurrentUserParams{
		Req:     req,
		Context: ctx,
	})
	if err != nil {
		return nil, fmt.Errorf("unable to create org user: %w", err)
	}

	return resp.Payload, nil
}
