package nuon

import (
	"context"

	"github.com/nuonco/nuon-go/client/operations"
	"github.com/nuonco/nuon-go/models"
)

func (c *client) CreateAppBreakGlassConfig(ctx context.Context, appID string, req *models.ServiceCreateAppBreakGlassConfigRequest) (*models.AppAppBreakGlassConfig, error) {
	resp, err := c.genClient.Operations.CreateAppBreakGlassConfig(&operations.CreateAppBreakGlassConfigParams{
		Req:     req,
		AppID:   appID,
		Context: ctx,
	}, c.getOrgIDAuthInfo())
	if err != nil {
		return nil, err
	}

	return resp.Payload, nil
}

func (c *client) GetLatestAppBreakGlassConfig(ctx context.Context, appID string) (*models.AppAppBreakGlassConfig, error) {
	resp, err := c.genClient.Operations.GetLatestAppBreakGlassConfig(&operations.GetLatestAppBreakGlassConfigParams{
		AppID:   appID,
		Context: ctx,
	}, c.getOrgIDAuthInfo())
	if err != nil {
		return nil, err
	}

	return resp.Payload, nil
}

func (c *client) GetAppBreakGlassConfig(ctx context.Context, appID, appBreakGlassID string) (*models.AppAppBreakGlassConfig, error) {
	params := &operations.GetAppBreakGlassConfigParams{
		AppID:              appID,
		BreakGlassConfigID: appBreakGlassID,
		Context:            ctx,
	}

	resp, err := c.genClient.Operations.GetAppBreakGlassConfig(params, c.getOrgIDAuthInfo())
	if err != nil {
		return nil, err
	}

	return resp.Payload, nil
}
