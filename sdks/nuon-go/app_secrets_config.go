package nuon

import (
	"context"

	"github.com/nuonco/nuon-go/client/operations"
	"github.com/nuonco/nuon-go/models"
)

func (c *client) CreateAppSecretsConfig(ctx context.Context, appID string, req *models.ServiceCreateAppSecretsConfigRequest) (*models.AppAppSecretsConfig, error) {
	resp, err := c.genClient.Operations.CreateAppSecretsConfig(&operations.CreateAppSecretsConfigParams{
		Req:     req,
		AppID:   appID,
		Context: ctx,
	}, c.getOrgIDAuthInfo())
	if err != nil {
		return nil, err
	}

	return resp.Payload, nil
}

func (c *client) GetLatestAppSecretsConfig(ctx context.Context, appID string) (*models.AppAppSecretsConfig, error) {
	resp, err := c.genClient.Operations.GetLatestAppSecretsConfig(&operations.GetLatestAppSecretsConfigParams{
		AppID:   appID,
		Context: ctx,
	}, c.getOrgIDAuthInfo())
	if err != nil {
		return nil, err
	}

	return resp.Payload, nil
}

func (c *client) GetAppSecretsConfig(ctx context.Context, appID, appSecretsID string) (*models.AppAppSecretsConfig, error) {
	params := &operations.GetAppSecretsConfigParams{
		AppID:              appID,
		AppSecretsConfigID: appSecretsID,
		Context:            ctx,
	}

	resp, err := c.genClient.Operations.GetAppSecretsConfig(params, c.getOrgIDAuthInfo())
	if err != nil {
		return nil, err
	}

	return resp.Payload, nil
}
