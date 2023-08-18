package client

import (
	"context"
	"fmt"

	"github.com/powertoolsdev/mono/pkg/api/client/client/operations"
	"github.com/powertoolsdev/mono/pkg/api/client/models"
)

// vcs connections
func (c *client) CreateOrgVCSConnection(ctx context.Context, orgID string, req *models.ServiceCreateOrgConnectionRequest) (*models.AppVCSConnection, error) {
	resp, err := c.genClient.Operations.PostV1VcsOrgIDConnection(&operations.PostV1VcsOrgIDConnectionParams{
		OrgID:   orgID,
		Context: ctx,
	})
	if err != nil {
		return nil, fmt.Errorf("unable to create org vcs connection: %w", err)
	}

	return resp.Payload, nil
}

func (c *client) GetOrgVCSConnectedRepos(ctx context.Context, orgID string) ([]*models.ServiceRepository, error) {
	resp, err := c.genClient.Operations.GetV1VcsOrgIDConnectedRepos(&operations.GetV1VcsOrgIDConnectedReposParams{
		OrgID:   orgID,
		Context: ctx,
	})
	if err != nil {
		return nil, fmt.Errorf("unable to get org connected repos: %w", err)
	}

	return resp.Payload, nil
}
