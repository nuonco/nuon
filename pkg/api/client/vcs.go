package client

import (
	"context"
	"fmt"

	"github.com/powertoolsdev/mono/pkg/api/client/client/operations"
	"github.com/powertoolsdev/mono/pkg/api/client/models"
)

// vcs connections
func (c *client) CreateVCSConnection(ctx context.Context, req *models.ServiceCreateConnectionRequest) (*models.AppVCSConnection, error) {
	resp, err := c.genClient.Operations.PostV1VcsConnections(&operations.PostV1VcsConnectionsParams{
		Req:     req,
		Context: ctx,
	})
	if err != nil {
		return nil, fmt.Errorf("unable to create org vcs connection: %w", err)
	}

	return resp.Payload, nil
}

func (c *client) GetVCSConnections(ctx context.Context) ([]*models.AppVCSConnection, error) {
	resp, err := c.genClient.Operations.GetV1VcsConnections(&operations.GetV1VcsConnectionsParams{
		Context: ctx,
	})
	if err != nil {
		return nil, fmt.Errorf("unable to get vcs connections: %w", err)
	}

	return resp.Payload, nil
}

func (c *client) GetVCSConnection(ctx context.Context, connID string) (*models.AppVCSConnection, error) {
	resp, err := c.genClient.Operations.GetV1VcsConnectionsConnectionID(&operations.GetV1VcsConnectionsConnectionIDParams{
		ConnectionID: connID,
		Context:      ctx,
	})
	if err != nil {
		return nil, fmt.Errorf("unable to create get vcs connection: %w", err)
	}

	return resp.Payload, nil
}

func (c *client) GetAllVCSConnectedRepos(ctx context.Context) ([]*models.ServiceRepository, error) {
	resp, err := c.genClient.Operations.GetV1VcsConnectedRepos(&operations.GetV1VcsConnectedReposParams{
		Context: ctx,
	})
	if err != nil {
		return nil, fmt.Errorf("unable to get org connected repos: %w", err)
	}

	return resp.Payload, nil
}
