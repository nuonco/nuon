package nuon

import (
	"context"

	"github.com/nuonco/nuon/sdks/nuon-go/client/operations"
	"github.com/nuonco/nuon/sdks/nuon-go/models"
)

// internal methods
func (c *client) GetApp(ctx context.Context, appID string) (*models.AppApp, error) {
	resp, err := c.genClient.Operations.GetApp(&operations.GetAppParams{
		Context: ctx,
		AppID:   appID,
	}, c.getOrgIDAuthInfo())
	if err != nil {
		return nil, err
	}

	return resp.Payload, nil
}

func (c *client) GetApps(ctx context.Context, query *models.GetPaginatedQuery) ([]*models.AppApp, bool, error) {
	params := &operations.GetAppsParams{
		Context: ctx,
	}

	params.Offset, params.Limit = applyPaginationQuery(query)

	hr := newResponseHeaderReader(&operations.GetAppsReader{})
	resp, err := c.genClient.Operations.GetApps(params, nil, hr.ClientOption())
	if err != nil {
		return nil, false, err
	}

	return resp.Payload, hasNextPage(hr), nil
}

func (c *client) CreateApp(ctx context.Context, req *models.ServiceCreateAppRequest) (*models.AppApp, error) {
	resp, err := c.genClient.Operations.CreateApp(&operations.CreateAppParams{
		Req:     req,
		Context: ctx,
	}, c.getOrgIDAuthInfo())
	if err != nil {
		return nil, err
	}

	return resp.Payload, nil
}

func (c *client) UpdateApp(ctx context.Context, appID string, req *models.ServiceUpdateAppRequest) (*models.AppApp, error) {
	resp, err := c.genClient.Operations.UpdateApp(&operations.UpdateAppParams{
		AppID:   appID,
		Req:     req,
		Context: ctx,
	}, c.getOrgIDAuthInfo())
	if err != nil {
		return nil, err
	}

	return resp.Payload, nil
}

func (c *client) DeleteApp(ctx context.Context, appID string) (bool, error) {
	_, err := c.genClient.Operations.DeleteApp(&operations.DeleteAppParams{
		AppID:   appID,
		Context: ctx,
	}, c.getOrgIDAuthInfo())
	if err != nil {
		return false, err
	}

	return true, nil
}
