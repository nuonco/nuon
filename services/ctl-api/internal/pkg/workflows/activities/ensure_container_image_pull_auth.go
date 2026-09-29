package activities

import (
	"go.temporal.io/sdk/workflow"

	"github.com/pkg/errors"

	plantypes "github.com/nuonco/nuon/pkg/plans/types"
)

func EnsureContainerImagePullAuth(ctx workflow.Context, plan *plantypes.BuildPlan) error {
	if plan == nil || plan.ContainerImagePullPlan == nil {
		return nil
	}
	if plan.SandboxMode != nil && plan.SandboxMode.Enabled {
		return nil
	}

	cfg := plan.ContainerImagePullPlan.RepoCfg
	if err := EnsureGARAuth(ctx, cfg); err != nil {
		return errors.Wrap(err, "unable to get GAR access token")
	}
	if err := EnsureACRAuth(ctx, cfg); err != nil {
		return errors.Wrap(err, "unable to get ACR access token")
	}

	return nil
}
