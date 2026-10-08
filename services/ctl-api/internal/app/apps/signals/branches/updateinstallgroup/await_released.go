package updateinstallgroup

import (
	"fmt"

	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/apps/signals/branches/activities"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/installgrouprelease"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/callback"
)

func (s *Signal) registerDirectiveHandler(ctx workflow.Context) error {
	s.directiveCh = workflow.NewBufferedChannel(ctx, 256)
	return workflow.SetUpdateHandler(ctx, installgrouprelease.UpdateName, func(ctx workflow.Context, directive installgrouprelease.Directive) error {
		s.directiveCh.Send(ctx, directive)
		return nil
	})
}

func (s *Signal) awaitInstallUpdatesReleased(
	ctx workflow.Context,
	groupName string,
	enqueued []enqueuedInstall,
	groupRunID string,
	installEntries []app.InstallGroupRunInstall,
) (*installUpdateFailure, error) {
	logger := workflow.GetLogger(ctx)
	treatCancelAsCancel := workflow.GetVersion(ctx, installGroupCancelAsCancelVersion, workflow.DefaultVersion, 1) != workflow.DefaultVersion

	byID := make(map[string]int, len(enqueued))
	waiting := make(map[string]struct{}, len(enqueued))
	released := map[string]bool{}
	for i, e := range enqueued {
		byID[e.installID] = i
		waiting[e.installID] = struct{}{}
	}

	var errs []error
	var cancelledErr error
	var failure *installUpdateFailure
	timedOut := false
	noteFailure := func(installID, workflowID, detail string) {
		if failure != nil {
			return
		}
		failure = &installUpdateFailure{installID: installID, workflowID: workflowID, detail: detail}
	}
	refresh := func() {
		completed, failed, cancelled := countInstallOutcomes(installEntries)
		desc := fmt.Sprintf("%d/%d installs deployed", completed, len(enqueued))
		if cancelled > 0 {
			desc += fmt.Sprintf(" (%d cancelled)", cancelled)
		}
		if failed > 0 {
			desc += fmt.Sprintf(" (%d failed)", failed)
		}
		if groupRunID != "" {
			_ = activities.AwaitUpdateInstallGroupRun(ctx, &activities.UpdateInstallGroupRunInput{
				InstallGroupRunID: groupRunID,
				Installs:          installEntries,
				CompletedInstalls: completed,
				FailedInstalls:    failed,
				Status: app.CompositeStatus{
					Status:                 app.StatusInProgress,
					StatusHumanDescription: desc,
				},
			})
		}
		results := map[string]string{}
		for _, entry := range installEntries {
			results[entry.InstallID] = entry.Status
		}
		s.updateInstallMetadata(ctx, groupName, enqueued, results)
	}
	finish := func(i int, installID, workflowID, status, detail string, cause error) {
		delete(waiting, installID)
		installEntries[i].Status = status
		switch status {
		case statusSuccess:
			s.updateInstallAppConfigVersionStatus(ctx, installID, app.StatusSuccess, detail)
			logger.Info("install config update completed", "install_id", installID, "workflow_id", workflowID)
		case statusCancelled:
			s.updateInstallAppConfigVersionStatus(ctx, installID, app.StatusCancelled, detail)
			if cancelledErr == nil {
				if cause != nil {
					cancelledErr = cause
				} else {
					cancelledErr = temporal.NewNonRetryableApplicationError(detail, callback.CancelledErrType, nil)
				}
			}
		case statusError:
			s.updateInstallAppConfigVersionStatus(ctx, installID, app.StatusError, detail)
			errs = append(errs, fmt.Errorf("%s", detail))
			noteFailure(installID, workflowID, detail)
		default:
			s.updateInstallAppConfigVersionStatus(ctx, installID, app.StatusInProgress, detail)
		}
		refresh()
	}
	applyCompletion := func(e enqueuedInstall, res *callback.Result, err error) {
		if released[e.installID] {
			return
		}
		i := byID[e.installID]
		switch {
		case err != nil && treatCancelAsCancel && callback.IsCancelled(err):
			finish(i, e.installID, e.workflowID, statusCancelled, err.Error(), err)
		case err != nil:
			finish(i, e.installID, e.workflowID, statusError, err.Error(), nil)
		case res == nil || res.Status != statusSuccess:
			status := "unknown"
			if res != nil && res.Status != "" {
				status = res.Status
			}
			errMsg := fmt.Sprintf("install %s workflow %s: finished as %s", e.installID, e.workflowID, status)
			if treatCancelAsCancel && status == statusCancelled {
				finish(i, e.installID, e.workflowID, statusCancelled, errMsg, nil)
				return
			}
			finish(i, e.installID, e.workflowID, statusError, errMsg, nil)
		default:
			finish(i, e.installID, e.workflowID, statusSuccess, "install workflow completed", nil)
		}
	}

	sel := workflow.NewSelector(ctx)
	for _, e := range enqueued {
		e := e
		ch := workflow.GetSignalChannel(ctx, e.cb.SignalName)
		sel.AddReceive(ch, func(c workflow.ReceiveChannel, more bool) {
			var result callback.Result
			c.Receive(ctx, &result)
			applyCompletion(e, &result, nil)
		})
	}
	sel.AddReceive(s.directiveCh, func(c workflow.ReceiveChannel, more bool) {
		var directive installgrouprelease.Directive
		c.Receive(ctx, &directive)
		i, ok := byID[directive.InstallID]
		if !ok || released[directive.InstallID] {
			return
		}
		if _, done := waiting[directive.InstallID]; !done {
			return
		}
		if directive.Directive != installgrouprelease.DirectiveRelease {
			return
		}
		var result callback.Result
		if workflow.GetSignalChannel(ctx, callback.SignalName(directive.InstallID)).ReceiveAsync(&result) {
			applyCompletion(enqueued[i], &result, nil)
			return
		}
		released[directive.InstallID] = true
		installEntries[i].ReleaseReason = directive.Reason
		installEntries[i].WaitingOnRunID = directive.WaitingOnRunID
		finish(i, directive.InstallID, enqueued[i].workflowID, installgrouprelease.StatusForRelease(directive.Reason), directive.Reason, nil)
	})

	timerCtx, timerCancel := workflow.WithCancel(ctx)
	defer timerCancel()
	sel.AddFuture(workflow.NewTimer(timerCtx, callback.FallbackAwaitTimeout), func(workflow.Future) {
		timedOut = true
	})

	for len(waiting) > 0 && !timedOut && cancelledErr == nil {
		sel.Select(ctx)
	}
	if timedOut {
		for installID := range waiting {
			i := byID[installID]
			finish(i, installID, enqueued[i].workflowID, statusError, fmt.Sprintf("install %s: timed out waiting for the install workflow", installID), nil)
		}
	}

	if cancelledErr != nil {
		return nil, cancelledErr
	}
	if len(errs) > 0 {
		return failure, fmt.Errorf("update install group had %d errors: %v", len(errs), errs)
	}
	return nil, nil
}
