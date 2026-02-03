package nuon

import (
	"context"

	"github.com/nuonco/nuon/sdks/nuon-go/client/operations"
	"github.com/nuonco/nuon/sdks/nuon-go/models"
)

func (c *client) CreateAppPoliciesConfig(ctx context.Context, appID string, req *models.ServiceCreateAppPoliciesConfigRequest) (*models.AppAppPoliciesConfig, error) {
	resp, err := c.genClient.Operations.CreateAppPoliciesConfig(&operations.CreateAppPoliciesConfigParams{
		Req:     req,
		AppID:   appID,
		Context: ctx,
	}, c.getOrgIDAuthInfo())
	if err != nil {
		return nil, err
	}

	return resp.Payload, nil
}

func (c *client) GetLatestAppPoliciesConfig(ctx context.Context, appID string) (*models.AppAppPoliciesConfig, error) {
	resp, err := c.genClient.Operations.GetLatestAppPoliciesConfig(&operations.GetLatestAppPoliciesConfigParams{
		AppID:   appID,
		Context: ctx,
	}, c.getOrgIDAuthInfo())
	if err != nil {
		return nil, err
	}

	return resp.Payload, nil
}

func (c *client) GetAppPoliciesConfig(ctx context.Context, appID, appPoliciesID string) (*models.AppAppPoliciesConfig, error) {
	params := &operations.GetAppPoliciesConfigParams{
		AppID:            appID,
		PoliciesConfigID: appPoliciesID,
		Context:          ctx,
	}

	resp, err := c.genClient.Operations.GetAppPoliciesConfig(params, c.getOrgIDAuthInfo())
	if err != nil {
		return nil, err
	}

	return resp.Payload, nil
}

func (c *client) GetAppPoliciesConfigs(ctx context.Context, appID string, query *models.GetPaginatedQuery) ([]*models.AppAppPoliciesConfig, bool, error) {
	params := &operations.GetAppPoliciesConfigsParams{
		AppID:   appID,
		Context: ctx,
	}

	if query != nil {
		if query.Limit > 0 {
			limit := int64(query.Limit)
			params.Limit = &limit
		}
		if query.Offset > 0 {
			offset := int64(query.Offset)
			params.Offset = &offset
		}
	}

	resp, err := c.genClient.Operations.GetAppPoliciesConfigs(params, c.getOrgIDAuthInfo())
	if err != nil {
		return nil, false, err
	}

	hasMore := len(resp.Payload) > 0 && query != nil && query.Limit > 0 && len(resp.Payload) >= query.Limit
	return resp.Payload, hasMore, nil
}
