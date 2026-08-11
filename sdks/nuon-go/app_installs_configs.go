package nuon

import (
	"context"

	"github.com/nuonco/nuon/sdks/nuon-go/client/operations"
	"github.com/nuonco/nuon/sdks/nuon-go/models"
)

func (c *client) CreateAppInstallsConfig(ctx context.Context, appID string, req *models.ServiceCreateAppInstallsConfigRequest) (*models.AppAppInstallsConfig, error) {
	resp, err := c.genClient.Operations.CreateAppInstallsConfig(&operations.CreateAppInstallsConfigParams{
		Context: ctx,
		AppID:   appID,
		Req:     req,
	}, c.getOrgIDAuthInfo())
	if err != nil {
		return nil, err
	}

	return resp.Payload, nil
}
