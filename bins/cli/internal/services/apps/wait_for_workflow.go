package apps

import (
	"context"
	"fmt"
	"time"

	"github.com/nuonco/nuon/sdks/nuon-go/models"

	"github.com/nuonco/nuon/bins/cli/internal/ui/bubbles"
	"github.com/nuonco/nuon/pkg/errs"
)

const defaultBranchRunPoll = time.Second * 5

func isTerminalStatus(status models.AppStatus) bool {
	switch status {
	case models.AppStatusSuccess,
		models.AppStatusError,
		models.AppStatusCancelled,
		models.AppStatusDiscarded,
		models.AppStatusUserDashSkipped,
		models.AppStatusAutoDashSkipped:
		return true
	}
	return false
}

// checkRunFailed reports a workflow that reached a terminal non-success state,
// so a wait on one of its steps stops instead of running out its own timeout.
func (s *Service) checkRunFailed(ctx context.Context, workflowID string) error {
	wf, err := s.api.GetWorkflow(ctx, workflowID)
	if err != nil || wf.Status == nil {
		return nil
	}
	if !isTerminalStatus(wf.Status.Status) || wf.Status.Status == models.AppStatusSuccess {
		return nil
	}

	return errs.NewUserFacing("app branch run did not succeed: %s", wf.Status.StatusHumanDescription)
}

func (s *Service) waitForWorkflowComplete(ctx context.Context, workflowID string, asJSON bool) error {
	if workflowID == "" {
		return nil
	}

	wf, err := s.api.GetWorkflow(ctx, workflowID)
	if err != nil {
		return fmt.Errorf("unable to get workflow: %w", err)
	}

	spinner := bubbles.NewSpinnerView(asJSON, s.cfg.Interactive)
	spinner.Start("waiting for preview workflow to complete")

	pollCtx, cancel := context.WithTimeout(ctx, defaultSyncTimeout)
	defer cancel()

	for {
		if wf.Finished || (wf.Status != nil && isTerminalStatus(wf.Status.Status)) {
			break
		}

		desc := "waiting for preview workflow to complete"
		if wf.Status != nil {
			if wf.Status.StatusHumanDescription != "" {
				desc = wf.Status.StatusHumanDescription
			} else if wf.Status.Status != "" {
				desc = fmt.Sprintf("waiting for preview workflow (status: %s)", wf.Status.Status)
			}
		}
		spinner.Update(desc)

		select {
		case <-pollCtx.Done():
			err := errs.NewUserFacing("timed out after %s waiting for preview workflow", defaultSyncTimeout)
			spinner.Fail(err)
			return err
		case <-time.After(defaultBranchRunPoll):
		}

		current, err := s.api.GetWorkflow(pollCtx, workflowID)
		if err != nil {
			spinner.Fail(fmt.Errorf("failed fetching workflow status: %w", err))
			return err
		}
		wf = current
	}

	switch {
	case wf.Status == nil:
		spinner.Fail(fmt.Errorf("workflow finished with no status"))
		return fmt.Errorf("workflow finished with no status")
	case wf.Status.Status == models.AppStatusSuccess:
		spinner.Success("preview workflow completed")
		return nil
	case wf.Status.Status == models.AppStatusCancelled:
		err := fmt.Errorf("preview workflow was cancelled")
		spinner.Fail(err)
		return err
	default:
		err := errs.NewUserFacing("preview workflow did not succeed: %s", wf.Status.StatusHumanDescription)
		spinner.Fail(err)
		return err
	}
}
