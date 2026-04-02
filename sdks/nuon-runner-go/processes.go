package nuonrunner

import (
	"context"

	"github.com/nuonco/nuon/sdks/nuon-runner-go/client/operations"
	"github.com/nuonco/nuon/sdks/nuon-runner-go/models"
)

func (c *client) CreateProcess(ctx context.Context, req *models.ServiceCreateRunnerProcessRequest) (*models.AppRunnerProcess, error) {
	resp, err := c.genClient.Operations.CreateRunnerProcess(&operations.CreateRunnerProcessParams{
		RunnerID: c.RunnerID,
		Req:      req,
		Context:  ctx,
	}, c.getAuthInfo())
	if err != nil {
		return nil, err
	}

	return resp.Payload, nil
}

func (c *client) GetProcess(ctx context.Context, processID string) (*models.AppRunnerProcess, error) {
	resp, err := c.genClient.Operations.GetRunnerProcess(&operations.GetRunnerProcessParams{
		RunnerID:  c.RunnerID,
		ProcessID: processID,
		Context:   ctx,
	}, c.getAuthInfo())
	if err != nil {
		return nil, err
	}

	return resp.Payload, nil
}

func (c *client) UpdateProcess(ctx context.Context, processID string, req *models.ServiceUpdateRunnerProcessRequest) (*models.AppRunnerProcess, error) {
	resp, err := c.genClient.Operations.UpdateRunnerProcess(&operations.UpdateRunnerProcessParams{
		RunnerID:  c.RunnerID,
		ProcessID: processID,
		Req:       req,
		Context:   ctx,
	}, c.getAuthInfo())
	if err != nil {
		return nil, err
	}

	return resp.Payload, nil
}
