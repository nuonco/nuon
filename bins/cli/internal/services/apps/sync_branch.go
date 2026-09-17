package apps

import (
	"context"
	"fmt"
	"time"

	"github.com/nuonco/nuon/sdks/nuon-go"
	"github.com/nuonco/nuon/sdks/nuon-go/models"

	"github.com/nuonco/nuon/bins/cli/internal/ui"
	"github.com/nuonco/nuon/bins/cli/internal/ui/bubbles"
	"github.com/nuonco/nuon/pkg/errs"
)

const (
	// buildsStepName must match the step the branch run workflow generator
	// creates for builds
	// (services/ctl-api/internal/app/apps/workflows/app_branch_run.go). The sync
	// returns once that step finishes; the install group steps that follow keep
	// running server-side.
	buildsStepName = "building components and sandbox"

	defaultBranchRunPoll = time.Second * 5
)

// syncViaBranchRun hands an already-uploaded config to a branch run, which syncs
// it, builds the changed components, and rolls it out to the branch's install
// groups. Only the sync and build phases are waited on.
func (s *Service) syncViaBranchRun(ctx context.Context, appID, branchID, dir string, appConfig *models.AppAppConfig, opts SyncOptions) (*syncResult, error) {
	run, err := s.api.TriggerAppBranchRun(ctx, appID, branchID, &models.ServiceTriggerAppBranchRunRequest{
		AppConfigID:   appConfig.ID,
		SyncAppConfig: true,
		PlanOnly:      opts.Preview,
		AutoApprove:   opts.AutoApprove,
	})
	if err != nil {
		return nil, err
	}

	ui.PrintSuccess(fmt.Sprintf("triggered app branch run %s", run.ID))
	result := &syncResult{AppID: appID, Dir: dir, BranchID: branchID, RunID: run.ID}

	synced, err := s.waitForRunConfigSync(ctx, appID, appConfig.ID, run.WorkflowID, opts.PrintJSON)
	if err != nil {
		return result, err
	}
	s.notifySyncResult(parseSyncState(synced.State).Result)

	if opts.NoWait || run.WorkflowID == "" {
		return result, nil
	}

	outcomes, waited, err := s.waitForBranchRunBuilds(ctx, appID, branchID, run.ID, run.WorkflowID, opts.PrintJSON)
	if waited {
		result.Builds = &syncBuildsResult{
			Scheduled:  len(outcomes),
			Waited:     true,
			Components: outcomes,
		}
	}
	if err != nil {
		return result, err
	}

	ui.PrintSuccess("successfully synced " + dir)
	ui.PrintLn(fmt.Sprintf("app branch run %s is rolling the config out to installs", run.ID))
	return result, nil
}

// branchRunSyncErr keeps the exit codes the standalone sync path documents:
// a failure once the config is synced and the run is building components is a
// build failure (exit 3), anything earlier is a sync failure (exit 1).
func branchRunSyncErr(err error, result *syncResult) error {
	if result == nil || result.Builds == nil {
		return err
	}

	return &ui.ErrExitCode{
		Err:  fmt.Errorf("app config synced, but component builds did not succeed: %w", err),
		Code: "builds_failed",
		Exit: buildsFailedExitCode,
	}
}

// waitForBranchRunBuilds blocks until the run's builds step finishes. The
// returned bool reports whether the step was reached at all, so a run that fails
// earlier does not report an empty build list as "nothing to build".
func (s *Service) waitForBranchRunBuilds(ctx context.Context, appID, branchID, runID, workflowID string, printJSON bool) ([]BuildOutcome, bool, error) {
	spinner := bubbles.NewSpinnerView(printJSON, s.cfg.Interactive)
	spinner.Start("building components")

	pollCtx, cancel := context.WithTimeout(ctx, defaultSyncTimeout)
	defer cancel()

	lastDescription := ""
	for {
		steps, err := s.api.GetWorkflowSteps(pollCtx, workflowID)
		if err != nil && !nuon.IsNotFound(err) && !nuon.IsServerError(err) {
			spinner.Fail(err)
			return nil, false, err
		}

		buildsStep := findStepByName(steps, buildsStepName)
		if buildsStep != nil && buildsStep.Status != nil {
			if desc := buildsStep.Status.StatusHumanDescription; desc != "" && desc != lastDescription {
				lastDescription = desc
				spinner.Update(desc)
			}

			if isTerminalStatus(buildsStep.Status.Status) {
				outcomes := s.branchRunBuildOutcomes(pollCtx, appID, branchID, runID)
				if buildsStep.Status.Status == models.AppStatusSuccess {
					spinner.Success("all component builds completed")
					return outcomes, true, nil
				}
				err := errs.NewUserFacing("component builds did not succeed: %s", stepFailureMessage(buildsStep))
				spinner.Fail(err)
				return outcomes, true, err
			}
		}

		if err := s.checkRunFailed(pollCtx, workflowID); err != nil {
			spinner.Fail(err)
			return nil, false, err
		}

		select {
		case <-pollCtx.Done():
			err := errs.NewUserFacing("timed out after %s waiting for component builds", defaultSyncTimeout)
			spinner.Fail(err)
			return nil, true, err
		case <-time.After(defaultBranchRunPoll):
		}
	}
}

// checkRunFailed reports a branch run that reached a terminal non-success state,
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

func (s *Service) branchRunBuildOutcomes(ctx context.Context, appID, branchID, runID string) []BuildOutcome {
	builds, err := s.api.GetAppBranchRunBuilds(ctx, appID, branchID, runID)
	if err != nil {
		return nil
	}

	outcomes := make([]BuildOutcome, 0, len(builds))
	for _, build := range builds {
		outcomes = append(outcomes, BuildOutcome{
			ComponentID:   build.ComponentID,
			ComponentName: build.ComponentName,
			Status:        buildOutcomeForStatus(build.Status),
		})
	}
	return outcomes
}

func buildOutcomeForStatus(status string) string {
	switch status {
	case componentBuildStatusActive:
		return buildOutcomeBuilt
	case componentBuildStatusError:
		return buildOutcomeError
	case componentBuildStatusPolicyFailed:
		return buildOutcomePolicyFailed
	default:
		return status
	}
}

func findStepByName(steps []*models.AppWorkflowStep, name string) *models.AppWorkflowStep {
	for _, step := range steps {
		if step.Name == name {
			return step
		}
	}
	return nil
}

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

func stepFailureMessage(step *models.AppWorkflowStep) string {
	if step.Status == nil {
		return "unknown"
	}
	if step.Status.StatusHumanDescription != "" {
		return step.Status.StatusHumanDescription
	}
	return string(step.Status.Status)
}
