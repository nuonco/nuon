package activities

import (
	"go.temporal.io/sdk/workflow"

	"github.com/pkg/errors"

	plantypes "github.com/nuonco/nuon/pkg/plans/types"
)

// EnsureContainerImagePullAuth embeds source-registry credentials into a
// container image build plan so the runner can pull an image it has no identity
// for.
//
// A sandboxed build is skipped entirely: the control-plane executor short
// circuits before the copy (see pkg/runner/controlplane.isSandboxableBuildHandler),
// so minting a real vendor token is work that cannot succeed. Sandbox-mode orgs
// have no access to the vendor's registry, which turns a job that was supposed
// to be faked into a hard failure at plan time.
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
