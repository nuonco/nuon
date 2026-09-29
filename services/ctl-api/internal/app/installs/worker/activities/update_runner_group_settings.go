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
	return nil
}
