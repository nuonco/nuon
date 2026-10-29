package v2

import (
	"github.com/pkg/errors"
	"go.temporal.io/sdk/workflow"

	"github.com/nuonco/nuon/pkg/generics"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/installs/worker/activities"
)

// ReconcileAppVersion rolls an install from its applied app config to the
// target app config captured in the workflow's metadata at creation time
// (new_app_config_id). This is the per-install workflow a do_not_deploy branch
// run pin leaves outstanding: the pin updates install.AppConfigID immediately,
// but nothing deploys, so the applied config lags behind. Reconcile generates
// exactly the steps AppBranchConfigUpdate would, minus the install-group
// rollout coordination (classifyAndMaybeWait / group-directive gating), since
// reconcile runs stand alone and are never part of a branch install-group run.
//
// A never-provisioned install (no applied config at all, e.g. first provision
// failed before anything applied) has no stack/runner/sandbox for the
// diff-based steps to target, so it reuses the Provision generator instead,
// against whatever app_config_id the install is currently pinned to -- which,
// for a never-provisioned install, is always safe to take as the freshest
// value, since nothing has been applied yet for an earlier pin to disagree
// with.
func ReconcileAppVersion(ctx workflow.Context, flw *app.Workflow) (*app.GenerateStepsResult, error) {
	installID := generics.FromPtrStr(flw.Metadata["install_id"])
	targetAppConfigID := generics.FromPtrStr(flw.Metadata["new_app_config_id"])
	if targetAppConfigID == "" {
		return nil, errors.New("new_app_config_id not found in workflow metadata")
	}

	install, err := activities.AwaitGetByInstallID(ctx, installID)
	if err != nil {
		return nil, errors.Wrap(err, "unable to get install")
	}

	if install.AppliedAppConfigID() == "" {
		return Provision(ctx, flw)
	}

	diffResult, err := activities.AwaitComputeInstallConfigDiff(ctx, &activities.ComputeInstallConfigDiffInput{
		OldAppConfigID: install.AppliedAppConfigID(),
		NewAppConfigID: targetAppConfigID,
	})
	if err != nil {
		return nil, errors.Wrap(err, "unable to compute config diff")
	}

	return buildConfigUpdateSteps(ctx, configUpdateInput{
		flw:            flw,
		installID:      installID,
		newAppConfigID: targetAppConfigID,
		diff:           diffResult.Diff,
		install:        install,
		triggeredBy:    "reconcile",
		metadataSource: "reconcile",
		checkEmptyDiff: false,
	})
}
