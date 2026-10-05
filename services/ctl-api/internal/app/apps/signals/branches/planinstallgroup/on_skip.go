package planinstallgroup

import (
	"time"

	"github.com/pkg/errors"
	"go.temporal.io/sdk/workflow"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/apps/signals/branches/activities"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/signal"
	statusactivities "github.com/nuonco/nuon/services/ctl-api/internal/pkg/workflows/status/activities"
)

var _ signal.SignalWithOnSkip = (*Signal)(nil)

// OnSkip marks this install group's deploy and post-deploy steps as user-skipped
// when the plan approval is skipped. Those steps live in separate step groups,
// so the generic same-group skip logic never reaches them — without this, the
// workflow would continue straight into deploying the group the user skipped.
func (s *Signal) OnSkip(ctx workflow.Context) error {
	out, err := activities.AwaitGetPendingInstallGroupDeployStep(ctx, &activities.GetPendingInstallGroupDeployStepInput{
		InstallWorkflowID: s.FlowID,
		InstallGroupID:    s.InstallGroupID,
	}, &workflow.ActivityOptions{ScheduleToCloseTimeout: time.Minute})
	if err != nil {
		return errors.Wrap(err, "unable to find deploy step for skipped install group")
	}

	for _, stepID := range out.StepIDs {
		if err := statusactivities.AwaitPkgStatusUpdateFlowStepStatus(ctx, statusactivities.UpdateStatusRequest{
			ID: stepID,
			Status: app.CompositeStatus{
				Status:                 app.StatusUserSkipped,
				StatusHumanDescription: "install group plan skipped, deploy skipped",
			},
		}); err != nil {
			return err
		}
	}
	return nil
}
