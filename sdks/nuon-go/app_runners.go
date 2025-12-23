package nuon

import (
	"context"

	"github.com/nuonco/nuon-go/client/operations"
	"github.com/nuonco/nuon-go/models"
)

func (c *client) CreateAppRunnerConfig(ctx context.Context, appID string, req *models.ServiceCreateAppRunnerConfigRequest) (*models.AppAppRunnerConfig, error) {
	resp, err := c.genClient.Operations.CreateAppRunnerConfig(&operations.CreateAppRunnerConfigParams{
		Req:     req,
		AppID:   appID,
		Context: ctx,
	}, c.getOrgIDAuthInfo())
	if err != nil {
		return nil, err
	}

	return resp.Payload, nil
}

func (c *client) GetAppRunnerLatestConfig(ctx context.Context, appID string) (*models.AppAppRunnerConfig, error) {
	resp, err := c.genClient.Operations.GetAppRunnerLatestConfig(&operations.GetAppRunnerLatestConfigParams{
		AppID:   appID,
		Context: ctx,
	}, c.getOrgIDAuthInfo())
	if err != nil {
		return nil, err
	}

	return resp.Payload, nil
}

func (c *client) GetAppRunnerConfigs(ctx context.Context, appID string, query *models.GetPaginatedQuery) ([]*models.AppAppRunnerConfig, bool, error) {
	params := &operations.GetAppRunnerConfigsParams{
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

	resp, err := c.genClient.Operations.GetAppRunnerConfigs(params, c.getOrgIDAuthInfo())
	if err != nil {
		return nil, false, err
	}

	if query != nil {
		items, hasMore := handlePagination(resp.Payload, int64(query.Offset), int64(query.Limit))
		return items, hasMore, nil
	}

	return resp.Payload, false, nil
}
