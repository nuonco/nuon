package nuon

import (
	"context"

	"github.com/nuonco/nuon/sdks/nuon-go/client/operations"
	"github.com/nuonco/nuon/sdks/nuon-go/models"
)

func (c *client) CreateCloudConnection(ctx context.Context, req *models.ServiceCreateRequest) (*models.ServiceConnectionResponse, error) {
	resp, err := c.genClient.Operations.CreateCloudConnection(&operations.CreateCloudConnectionParams{Context: ctx, Req: req}, c.getOrgIDAuthInfo())
	if err != nil {
		return nil, err
	}
	return resp.Payload, nil
}

func (c *client) ListCloudConnections(ctx context.Context) ([]*models.ServiceConnectionResponse, error) {
	resp, err := c.genClient.Operations.ListCloudConnections(&operations.ListCloudConnectionsParams{Context: ctx}, c.getOrgIDAuthInfo())
	if err != nil {
		return nil, err
	}
	return resp.Payload, nil
}

func (c *client) GetCloudConnection(ctx context.Context, connectionID string) (*models.ServiceConnectionResponse, error) {
	resp, err := c.genClient.Operations.GetCloudConnection(&operations.GetCloudConnectionParams{Context: ctx, ConnectionID: connectionID}, c.getOrgIDAuthInfo())
	if err != nil {
		return nil, err
	}
	return resp.Payload, nil
}

func (c *client) VerifyCloudConnection(ctx context.Context, connectionID string, req *models.ServiceVerifyRequest) (*models.ServiceConnectionResponse, error) {
	resp, err := c.genClient.Operations.VerifyCloudConnection(&operations.VerifyCloudConnectionParams{Context: ctx, ConnectionID: connectionID, Req: req}, c.getOrgIDAuthInfo())
	if err != nil {
		return nil, err
	}
	return resp.Payload, nil
}

func (c *client) DeleteCloudConnection(ctx context.Context, connectionID string) error {
	_, err := c.genClient.Operations.DeleteCloudConnection(&operations.DeleteCloudConnectionParams{Context: ctx, ConnectionID: connectionID}, c.getOrgIDAuthInfo())
	return err
}
