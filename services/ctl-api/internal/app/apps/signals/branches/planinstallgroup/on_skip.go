package planinstallgroup

import (
	"github.com/pkg/errors"
	"go.temporal.io/sdk/workflow"

	"github.com/nuonco/nuon/services/ctl-api/internal/app/apps/signals/branches/activities"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/signal"
)

var _ signal.SignalWithOnSkip = (*Signal)(nil)

// recordSkippedInstallGroupRunVersion gates recording a skipped InstallGroupRun
// row for the group's plan step. Older in-flight histories never scheduled this
// activity and must keep replaying without it.
const recordSkippedInstallGroupRunVersion = "record-skipped-install-group-run-v1"

// OnSkip marks this install group's deploy and post-deploy steps as user-skipped
// when the plan approval is skipped. Those steps live in separate step groups,
// so the generic same-group skip logic never reaches them — without this, the
// workflow would continue straight into deploying the group the user skipped.
func (s *Signal) OnSkip(ctx workflow.Context) error {
	if err := activities.AwaitSkipPendingInstallGroupFollowOnSteps(ctx, &activities.SkipPendingInstallGroupFollowOnStepsInput{
		InstallWorkflowID: s.FlowID,
		InstallGroupID:    s.InstallGroupID,
	}); err != nil {
		return errors.Wrap(err, "unable to skip follow-on steps for skipped install group")
	}

	if workflow.GetVersion(ctx, recordSkippedInstallGroupRunVersion, workflow.DefaultVersion, 1) == workflow.DefaultVersion {
		return nil
	}

	// Preview runs use a synthetic target instead of a real install group, and
	// have no InstallGroup row to resolve or group run to track.
	if s.InstallGroupID == "" {
		return nil
	}

	installIDs, groupName, err := s.resolveInstallIDs(ctx)
	if err != nil {
		return errors.Wrap(err, "unable to resolve install group for skipped group run")
	}

	if err := activities.AwaitCreateOrUpdateSkippedInstallGroupRun(ctx, &activities.CreateOrUpdateSkippedInstallGroupRunInput{
		AppBranchRunID:   s.RunID,
		InstallGroupID:   s.InstallGroupID,
		InstallGroupName: groupName,
		TotalInstalls:    len(installIDs),
	}); err != nil {
		return errors.Wrap(err, "unable to record skipped install group run")
	}

	return nil
}
