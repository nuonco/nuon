package nuonrunner

import (
	"context"
	"errors"
	"time"

	"github.com/nuonco/nuon/sdks/nuon-runner-go/client/operations"
	"github.com/nuonco/nuon/sdks/nuon-runner-go/models"
)

// ErrTailJobsNotAvailable is returned by TailJobs when the org does not
// have the `runner-job-long-poll` feature flag on. Callers should fall
// back to GetJobs polling for the rest of the runner's lifetime to avoid
// reprobing a known-disabled endpoint every iteration.
var ErrTailJobsNotAvailable = errors.New("tail jobs endpoint not available for this org")

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

// TailJobs long-polls the ctl-api for an available job. wait is rounded
// down to seconds for transport. A nil job list is returned with no error
// when the long-poll times out with the queue still empty; callers should
// reissue immediately in that case.
//
// Returns ErrTailJobsNotAvailable on a 404 from the server, which signals
// the `runner-job-long-poll` org feature flag is off.
func (c *client) TailJobs(ctx context.Context, grp models.AppRunnerJobGroup, wait time.Duration) ([]*models.AppRunnerJob, error) {
	grpStr := string(grp)
	waitStr := wait.String()

	resp, err := c.genClient.Operations.TailRunnerJobs(&operations.TailRunnerJobsParams{
		RunnerID: c.RunnerID,
		Group:    &grpStr,
		Wait:     &waitStr,
		Context:  ctx,
	}, c.getAuthInfo())
	if err != nil {
		var notFound *operations.TailRunnerJobsNotFound
		if errors.As(err, &notFound) {
			return nil, ErrTailJobsNotAvailable
		}
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
