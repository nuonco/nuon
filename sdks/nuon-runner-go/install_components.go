package nuonrunner

import (
	"context"

	"github.com/nuonco/nuon/sdks/nuon-runner-go/client/operations"
	"github.com/nuonco/nuon/sdks/nuon-runner-go/models"
)

func (c *client) GetRunnerInstallComponents(ctx context.Context) (*models.ServiceRunnerInstallComponentsResponse, error) {
	resp, err := c.genClient.Operations.GetRunnerInstallComponents(&operations.GetRunnerInstallComponentsParams{
		RunnerID: c.RunnerID,
		Context:  ctx,
	}, c.getAuthInfo())
	if err != nil {
		return nil, err
	}

	return resp.Payload, nil
}
