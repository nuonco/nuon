package provision

import (
	"context"
	"fmt"

	"go.temporal.io/sdk/temporal"

	assumerole "github.com/powertoolsdev/mono/pkg/aws/assume-role"
	"github.com/powertoolsdev/mono/pkg/aws/credentials"
)

type CheckIAMRoleRequest struct {
	RoleARN string `validate:"required"`

	TwoStepConfig *assumerole.TwoStepConfig `validate:"required"`
}

type CheckIAMRoleResponse struct{}

func (a *Activities) CheckIAMRole(ctx context.Context, req CheckIAMRoleRequest) (CheckIAMRoleResponse, error) {
	if err := a.v.Struct(req); err != nil {
		return CheckIAMRoleResponse{}, fmt.Errorf("invalid request: %w", err)
	}

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
