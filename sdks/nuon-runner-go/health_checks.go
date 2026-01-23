package nuonrunner

import (
	"context"

	"github.com/nuonco/nuon/sdks/nuon-runner-go/client/operations"
	"github.com/nuonco/nuon/sdks/nuon-runner-go/models"
)

func (c *client) CreateHealthCheck(ctx context.Context, req *models.ServiceCreateRunnerHealthCheckRequest) (*models.AppRunnerHealthCheck, error) {
	resp, err := c.genClient.Operations.CreateRunnerHealthCheck(&operations.CreateRunnerHealthCheckParams{
		RunnerID: c.RunnerID,
		Req:      req,
		Context:  ctx,
	}, c.getAuthInfo())
	if err != nil {
		return nil, err
	}

	return resp.Payload, nil
}
