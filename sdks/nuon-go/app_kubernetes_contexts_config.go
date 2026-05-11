package nuon

import (
	"context"

	"github.com/nuonco/nuon/sdks/nuon-go/client/operations"
	"github.com/nuonco/nuon/sdks/nuon-go/models"
)

func (c *client) CreateAppKubernetesContextsConfig(ctx context.Context, appID string, req *models.ServiceCreateAppKubernetesContextsConfigRequest) (*models.AppAppKubernetesContextsConfig, error) {
	resp, err := c.genClient.Operations.CreateAppKubernetesContextsConfig(&operations.CreateAppKubernetesContextsConfigParams{
		Req:     req,
		AppID:   appID,
		Context: ctx,
	}, c.getOrgIDAuthInfo())
	if err != nil {
		return nil, err
	}

	return resp.Payload, nil
}
