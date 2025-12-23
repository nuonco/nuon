package nuon

import (
	"context"

	"github.com/nuonco/nuon-go/client/operations"
	"github.com/nuonco/nuon-go/models"
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
