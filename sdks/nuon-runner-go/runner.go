package nuonrunner

import (
	"context"

	"github.com/nuonco/nuon/sdks/nuon-runner-go/client/operations"
	"github.com/nuonco/nuon/sdks/nuon-runner-go/models"
)

func (c *client) GetRunner(ctx context.Context) (*models.AppRunner, error) {
	resp, err := c.genClient.Operations.GetRunner(&operations.GetRunnerParams{
		RunnerID: c.RunnerID,
		Context:  ctx,
	}, c.getAuthInfo())
	if err != nil {
		return nil, err
	}

	return resp.Payload, nil
}
