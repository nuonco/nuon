package provision

import (
	"context"

	"github.com/powertoolsdev/mono/pkg/aws/credentials"
	"go.temporal.io/sdk/temporal"
)

type CheckIAMRoleRequest struct {
	RoleARN string
}

type CheckIAMRoleResponse struct{}

func (a *Activities) CheckIAMRole(ctx context.Context, req CheckIAMRoleRequest) (CheckIAMRoleResponse, error) {
	var resp CheckIAMRoleResponse
	cfg := &credentials.Config{
		AssumeRole: &credentials.AssumeRoleConfig{
			RoleARN:     req.RoleARN,
			SessionName: "workers-installs-check-iam-role",
		},
	}
	_, err := credentials.Fetch(ctx, cfg)
	if err != nil {
		return resp, temporal.NewNonRetryableApplicationError("unable to access iam role", "error", err)
	}

	return resp, nil
}
