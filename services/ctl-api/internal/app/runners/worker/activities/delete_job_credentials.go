package activities

import (
	"context"
)

type DeleteJobCredentialsRequest struct {
	JobID string `validate:"required"`
}

// @temporal-gen-v2 activity
// @start-to-close-timeout 1m
// @max-retries 3
func (a *Activities) DeleteJobCredentials(ctx context.Context, req DeleteJobCredentialsRequest) error {
	return a.acctClient.DeleteRunnerJobCredentials(ctx, req.JobID)
}
