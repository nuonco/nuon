package nuon

import (
	"context"

	"github.com/nuonco/nuon-go/client/operations"
	"github.com/nuonco/nuon-go/models"
)

func (c *client) CreateInstallConfig(ctx context.Context, installID string, req *models.ServiceCreateInstallConfigRequest) (*models.AppInstallConfig, error) {
	resp, err := c.genClient.Operations.CreateInstallConfig(&operations.CreateInstallConfigParams{
		InstallID: installID,
		Req:       req,
		Context:   ctx,
	}, c.getOrgIDAuthInfo())
	if err != nil {
		return nil, err
	}

	return resp.Payload, nil
}

func (c *client) UpdateInstallConfig(ctx context.Context, installID, configID string, req *models.ServiceUpdateInstallConfigRequest) (*models.AppInstallConfig, error) {
	resp, err := c.genClient.Operations.UpdateInstallConfig(&operations.UpdateInstallConfigParams{
		InstallID: installID,
		ConfigID:  configID,
		Req:       req,
		Context:   ctx,
	}, c.getOrgIDAuthInfo())
	if err != nil {
		return nil, err
	}

	return resp.Payload, nil
}
