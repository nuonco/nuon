package planinstallgroup

import (
	"github.com/pkg/errors"
	"go.temporal.io/sdk/workflow"

	"github.com/nuonco/nuon/services/ctl-api/internal/app/apps/signals/branches/activities"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/signal"
)

var _ signal.SignalWithOnSkip = (*Signal)(nil)

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
	return nil
}
