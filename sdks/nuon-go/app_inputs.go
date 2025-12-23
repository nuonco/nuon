package nuon

import (
	"context"

	"github.com/nuonco/nuon-go/client/operations"
	"github.com/nuonco/nuon-go/models"
)

func (c *client) CreateAppInputConfig(ctx context.Context, appID string, req *models.ServiceCreateAppInputConfigRequest) (*models.AppAppInputConfig, error) {
	resp, err := c.genClient.Operations.CreateAppInputConfig(&operations.CreateAppInputConfigParams{
		Req:     req,
		AppID:   appID,
		Context: ctx,
	}, c.getOrgIDAuthInfo())
	if err != nil {
		return nil, err
	}

	return resp.Payload, nil
}

func (c *client) GetAppInputLatestConfig(ctx context.Context, appID string) (*models.AppAppInputConfig, error) {
	resp, err := c.genClient.Operations.GetAppInputLatestConfig(&operations.GetAppInputLatestConfigParams{
		AppID:   appID,
		Context: ctx,
	}, c.getOrgIDAuthInfo())
	if err != nil {
		return nil, err
	}

	return resp.Payload, nil
}

func (c *client) GetAppInputConfig(ctx context.Context, appID, appInputConfigID string) (*models.AppAppInputConfig, error) {
	resp, err := c.genClient.Operations.GetAppInputConfig(&operations.GetAppInputConfigParams{
		AppID:         appID,
		InputConfigID: appInputConfigID,
		Context:       ctx,
	}, c.getOrgIDAuthInfo())
	if err != nil {
		return nil, err
	}

	return resp.Payload, nil
}

func (c *client) GetAppInputConfigs(ctx context.Context, appID string, query *models.GetPaginatedQuery) ([]*models.AppAppInputConfig, bool, error) {
	params := &operations.GetAppInputConfigsParams{
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

	resp, err := c.genClient.Operations.GetAppInputConfigs(params, c.getOrgIDAuthInfo())
	if err != nil {
		return nil, false, err
	}

	if query != nil {
		items, hasMore := handlePagination(resp.Payload, int64(query.Offset), int64(query.Limit))
		return items, hasMore, nil
	}

	return resp.Payload, false, nil
}
