package activities

import (
	"context"

	runnershelpers "github.com/nuonco/nuon/services/ctl-api/internal/app/runners/helpers"
)

type UpdateRunnerCloudIdentityRole struct {
	RunnerID         string `json:"runner_id" validate:"required"`
	RunnerIAMRoleARN string `json:"runner_iam_role_arn"`
}

// @temporal-gen-v2 activity
func (a *Activities) UpdateRunnerCloudIdentityRole(ctx context.Context, req *UpdateRunnerCloudIdentityRole) (*runnershelpers.RunnerCloudIdentityRoleUpdate, error) {
	return a.runnersHelpers.SetRunnerCloudIdentityRole(ctx, req.RunnerID, req.RunnerIAMRoleARN)
}
