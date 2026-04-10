package activities

import (
	"context"
)

type UpdateRunnerGroupSettings struct {
	RunnerID           string `json:"runner_id" validate:"required"`
	LocalAWSIAMRoleARN string `json:"runner_iam_role_arn"`
}

// @temporal-gen-v2 activity
func (a *Activities) UpdateRunnerGroupSettings(ctx context.Context, req *UpdateRunnerGroupSettings) error {
	// NOTE(jm): we no longer need this, because we were previously updating the stack to run the runner locally
	// with the runner instance role.
	return nil
}
