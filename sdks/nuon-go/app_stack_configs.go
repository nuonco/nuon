package nuon

import (
	"context"

	"github.com/nuonco/nuon-go/client/operations"
	"github.com/nuonco/nuon-go/models"
)

func (c *client) CreateAppStackConfig(ctx context.Context, appID string, req *models.ServiceCreateAppStackConfigRequest) (*models.AppAppStackConfig, error) {
	resp, err := c.genClient.Operations.CreateAppStackConfig(&operations.CreateAppStackConfigParams{
		Req:     req,
		AppID:   appID,
		Context: ctx,
	}, c.getOrgIDAuthInfo())
	if err != nil {
		return nil, err
	}

	return resp.Payload, nil
}

func (c *client) GetAppStackConfig(ctx context.Context, appID, configID string) (*models.AppAppStackConfig, error) {
	params := &operations.GetAppStackConfigParams{
		AppID:    appID,
		ConfigID: configID,
		Context:  ctx,
	}

	resp, err := c.genClient.Operations.GetAppStackConfig(params, c.getOrgIDAuthInfo())
	if err != nil {
		return nil, err
	}

	return resp.Payload, nil
}
