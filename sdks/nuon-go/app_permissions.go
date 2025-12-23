package nuon

import (
	"context"

	"github.com/nuonco/nuon-go/client/operations"
	"github.com/nuonco/nuon-go/models"
)

func (c *client) CreateAppPermissionsConfig(ctx context.Context, appID string, req *models.ServiceCreateAppPermissionsConfigRequest) (*models.AppAppPermissionsConfig, error) {
	resp, err := c.genClient.Operations.CreateAppPermissionsConfig(&operations.CreateAppPermissionsConfigParams{
		Req:     req,
		AppID:   appID,
		Context: ctx,
	}, c.getOrgIDAuthInfo())
	if err != nil {
		return nil, err
	}

	return resp.Payload, nil
}

func (c *client) GetLatestAppPermissionsConfig(ctx context.Context, appID string) (*models.AppAppPermissionsConfig, error) {
	resp, err := c.genClient.Operations.GetLatestAppPermissionsConfig(&operations.GetLatestAppPermissionsConfigParams{
		AppID:   appID,
		Context: ctx,
	}, c.getOrgIDAuthInfo())
	if err != nil {
		return nil, err
	}

	return resp.Payload, nil
}

func (c *client) GetAppPermissionsConfig(ctx context.Context, appID, appPermissionsID string) (*models.AppAppPermissionsConfig, error) {
	params := &operations.GetAppPermissionsConfigParams{
		AppID:               appID,
		PermissionsConfigID: appPermissionsID,
		Context:             ctx,
	}

	resp, err := c.genClient.Operations.GetAppPermissionsConfig(params, c.getOrgIDAuthInfo())
	if err != nil {
		return nil, err
	}

	return resp.Payload, nil
}
