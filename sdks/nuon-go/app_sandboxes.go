package nuon

import (
	"context"

	"github.com/nuonco/nuon-go/client/operations"
	"github.com/nuonco/nuon-go/models"
)

func (c *client) CreateAppSandboxConfig(ctx context.Context, appID string, req *models.ServiceCreateAppSandboxConfigRequest) (*models.AppAppSandboxConfig, error) {
	resp, err := c.genClient.Operations.CreateAppSandboxConfig(&operations.CreateAppSandboxConfigParams{
		Req:     req,
		AppID:   appID,
		Context: ctx,
	}, c.getOrgIDAuthInfo())
	if err != nil {
		return nil, err
	}

	return resp.Payload, nil
}

func (c *client) GetAppSandboxLatestConfig(ctx context.Context, appID string) (*models.AppAppSandboxConfig, error) {
	resp, err := c.genClient.Operations.GetAppSandboxLatestConfig(&operations.GetAppSandboxLatestConfigParams{
		AppID:   appID,
		Context: ctx,
	}, c.getOrgIDAuthInfo())
	if err != nil {
		return nil, err
	}

	return resp.Payload, nil
}

func (c *client) GetAppSandboxConfigs(ctx context.Context, appID string, query *models.GetPaginatedQuery) ([]*models.AppAppSandboxConfig, bool, error) {
	params := &operations.GetAppSandboxConfigsParams{
		AppID:   appID,
		Context: ctx,
	}

	query = handlePaginationQuery(query)

	if query != nil {
		offset := int64(query.Offset)
		limit := int64(query.Limit)

		params.Offset = &offset
		params.Limit = &limit
	}

	resp, err := c.genClient.Operations.GetAppSandboxConfigs(params, c.getOrgIDAuthInfo())
	if err != nil {
		return nil, false, err
	}

	if query != nil {
		items, hasMore := handlePagination(resp.Payload, int64(query.Offset), int64(query.Limit))
		return items, hasMore, nil
	}

	return resp.Payload, false, nil
}
