package nuon

import (
	"context"

	"github.com/nuonco/nuon/sdks/nuon-go/client/operations"
	"github.com/nuonco/nuon/sdks/nuon-go/models"
)

func (c *client) CreateAppOperationRoleConfig(ctx context.Context, appID string, req *models.ServiceCreateAppOperationRoleConfigRequest) (*models.AppAppOperationRoleConfig, error) {
	resp, err := c.genClient.Operations.CreateAppOperationRoleConfig(&operations.CreateAppOperationRoleConfigParams{
		Req:     req,
		AppID:   appID,
		Context: ctx,
	}, c.getOrgIDAuthInfo())
	if err != nil {
		return nil, err
	}

	return resp.Payload, nil
}
