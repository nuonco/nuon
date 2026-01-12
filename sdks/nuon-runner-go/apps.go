package nuonrunner

import (
	"context"

	"github.com/nuonco/nuon/sdks/nuon-runner-go/client/operations"
	"github.com/nuonco/nuon/sdks/nuon-runner-go/models"
)

func (c *client) GetAppConfig(ctx context.Context, appID, appConfigID string) (*models.AppAppConfig, error) {
	resp, err := c.genClient.Operations.GetRunnerAppConfig(&operations.GetRunnerAppConfigParams{
		AppID:       appID,
		AppConfigID: appConfigID,
		Context:     ctx,
	}, c.getAuthInfo())
	if err != nil {
		return nil, err
	}

	return resp.Payload, nil
}
