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

func (c *client) ListCloudConnections(ctx context.Context, query *models.GetPaginatedQuery) ([]*models.ServiceConnectionListResponse, bool, error) {
	params := &operations.ListCloudConnectionsParams{Context: ctx}
	params.Offset, params.Limit = applyPaginationQuery(query)
	hr := newResponseHeaderReader(&operations.ListCloudConnectionsReader{})
	resp, err := c.genClient.Operations.ListCloudConnections(params, c.getOrgIDAuthInfo(), hr.ClientOption())
	if err != nil {
		return nil, false, err
	}
	return resp.Payload, hasNextPage(hr), nil
}

func (c *client) GetCloudConnection(ctx context.Context, connectionID string) (*models.ServiceConnectionResponse, error) {
	resp, err := c.genClient.Operations.GetCloudConnection(&operations.GetCloudConnectionParams{Context: ctx, ConnectionID: connectionID}, c.getOrgIDAuthInfo())
	if err != nil {
		return nil, err
	}
	return resp.Payload, nil
}

func (c *client) VerifyCloudConnection(ctx context.Context, connectionID string) (*models.ServiceConnectionResponse, error) {
	resp, err := c.genClient.Operations.VerifyCloudConnection(&operations.VerifyCloudConnectionParams{Context: ctx, ConnectionID: connectionID}, c.getOrgIDAuthInfo())
	if err != nil {
		return nil, err
	}
	return resp.Payload, nil
}

func (c *client) DeleteCloudConnection(ctx context.Context, connectionID string) error {
	_, err := c.genClient.Operations.DeleteCloudConnection(&operations.DeleteCloudConnectionParams{Context: ctx, ConnectionID: connectionID}, c.getOrgIDAuthInfo())
	return err
}
