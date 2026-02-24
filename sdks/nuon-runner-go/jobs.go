package nuonrunner

import (
	"context"

	"github.com/nuonco/nuon/sdks/nuon-runner-go/client/operations"
	"github.com/nuonco/nuon/sdks/nuon-runner-go/models"
)

func (c *client) GetJobs(ctx context.Context, grp models.AppRunnerJobGroup, status models.AppRunnerJobStatus, limit *int64) ([]*models.AppRunnerJob, error) {
	statusStr := string(status)
	grpStr := string(grp)

	resp, err := c.genClient.Operations.GetRunnerJobs(&operations.GetRunnerJobsParams{
		Limit:    limit,
		RunnerID: c.RunnerID,
		Status:   &statusStr,
		Group:    &grpStr,
		Context:  ctx,
	}, c.getAuthInfo())
	if err != nil {
		return nil, err
	}

	return resp.Payload, nil
}

func (c *client) GetJob(ctx context.Context, jobID string) (*models.AppRunnerJob, error) {
	resp, err := c.genClient.Operations.GetRunnerJob(&operations.GetRunnerJobParams{
		RunnerJobID: jobID,
		Context:     ctx,
	}, c.getAuthInfo())
	if err != nil {
		return nil, err
	}

	return resp.Payload, nil
}

func (c *client) GetJobCompositePlan(ctx context.Context, jobID string) (*models.PlantypesCompositePlan, error) {
	resp, err := c.genClient.Operations.GetRunnerJobCompositePlan(&operations.GetRunnerJobCompositePlanParams{
		RunnerJobID: jobID,
		Context:     ctx,
	}, c.getAuthInfo())
	if err != nil {
		return nil, err
	}

	return resp.Payload, nil
}

func (c *client) GetJobPlanJSON(ctx context.Context, jobID string) (string, error) {
	resp, err := c.genClient.Operations.GetRunnerJobPlan(&operations.GetRunnerJobPlanParams{
		RunnerJobID: jobID,
		Context:     ctx,
	}, c.getAuthInfo())
	if err != nil {
		return "", err
	}

	return resp.Payload, nil
}

func (c *client) UpdateJob(ctx context.Context, jobID string, req *models.ServiceUpdateRunnerJobRequest) (*models.AppRunnerJob, error) {
	resp, err := c.genClient.Operations.UpdateRunnerJob(&operations.UpdateRunnerJobParams{
		Req:         req,
		RunnerJobID: jobID,
		Context:     ctx,
	}, c.getAuthInfo())
	if err != nil {
		return nil, err
	}

	return resp.Payload, nil
}
