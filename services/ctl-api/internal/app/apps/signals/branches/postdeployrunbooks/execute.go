package postdeployrunbooks

import (
	"errors"
	"fmt"
	"strconv"
	"strings"

	"go.temporal.io/sdk/workflow"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/apps/signals/branches/activities"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/callback"
	statusactivities "github.com/nuonco/nuon/services/ctl-api/internal/pkg/workflows/status/activities"
)

const (
	statusPending    = "pending"
	statusInProgress = "in-progress"
	statusSuccess    = "success"
	statusError      = "error"
	statusCancelled  = "cancelled"
)

// pendingRunbookRun tracks one in-flight runbook run so its callback can be
// awaited after every install in the wave has been started.
type pendingRunbookRun struct {
	entryIdx   int
	runbookIdx int
	installID  string
	workflowID string
	cb         callback.Ref
}

// Execute runs the branch config's post-deploy runbooks against every install
// that deployed successfully in this group.
//
// Runbooks are sequential over the list and parallel over installs: each runbook
// is started on every still-healthy install before any of them are awaited. The
// alternative — awaiting each install's runbook inside the loop that starts them —
// would serialize runbooks across installs, making the step take the sum of every
// install's runbook duration instead of just the slowest.
//
// A failed runbook drops that install out of the remaining runbooks and fails the
// step, stopping the rollout from advancing to the next group.
//
// The step is retryable, so Execute is written to resume: it re-reads the group
// run row, keeps runbooks that already succeeded, and re-attempts the ones that
// failed. It deliberately never writes runbook outcomes onto the install's
// deploy status — that field is the deploy step's result and the input this
// method uses to decide who is eligible, so overwriting it would make a retry
// skip exactly the installs that need re-running.
func (s *Signal) Execute(ctx workflow.Context) error {
	logger := workflow.GetLogger(ctx)

	run, err := activities.AwaitGetAppBranchRunByIDByRunID(ctx, s.RunID)
	if err != nil {
		return fmt.Errorf("unable to get app branch run: %w", err)
	}

	// RunType is how preview runs are identified now; PlanOnly is the legacy flag
	// nothing in the branch flow sets anymore. Check both so a preview never runs
	// migrations or smoke tests against real installs.
	if run.RunType == app.AppBranchRunTypeGitPreview || run.PlanOnly {
		logger.Info("preview run, skipping post-deploy runbooks",
			"run_type", string(run.RunType),
		)
		return nil
	}

	groupRun, err := activities.AwaitGetInstallGroupRun(ctx, &activities.GetInstallGroupRunInput{
		AppBranchRunID: s.RunID,
		InstallGroupID: s.InstallGroupID,
	})
	if err != nil {
		return fmt.Errorf("unable to get install group run: %w", err)
	}

	installEntries := groupRun.Installs
	deployedIDs := make([]string, 0, len(installEntries))
	for i := range installEntries {
		if installEntries[i].Status == statusSuccess {
			deployedIDs = append(deployedIDs, installEntries[i].InstallID)
		}
	}
	if len(deployedIDs) == 0 {
		logger.Info("no installs deployed successfully, skipping post-deploy runbooks")
		return nil
	}

	resolved, err := activities.AwaitResolveBranchPostDeployRunbooks(ctx, &activities.ResolveBranchPostDeployRunbooksInput{
		AppBranchID:       s.AppBranchID,
		AppBranchConfigID: s.AppBranchConfigID,
		InstallIDs:        deployedIDs,
		NewAppConfigID:    run.AppConfigID,
		CreatedByID:       run.CreatedByID,
		Inputs:            branchRunbookInputs(run),
	})
	if err != nil {
		return fmt.Errorf("unable to resolve post-deploy runbooks: %w", err)
	}
	if len(resolved.Runbooks) == 0 {
		return nil
	}

	logger.Info("running post-deploy runbooks",
		"install_group_id", s.InstallGroupID,
		"runbook_count", len(resolved.Runbooks),
		"install_count", len(deployedIDs),
	)

	seedRunbookPlan(installEntries, resolved.Runbooks)
	s.updateStepMetadata(ctx, planDescription(resolved.Runbooks), installEntries)

	var errs []error
	for runbookIdx, rb := range resolved.Runbooks {
		pending := s.startRunbookWave(ctx, run, rb, runbookIdx, installEntries, &errs)

		s.awaitRunbookWave(ctx, rb, pending, installEntries, &errs)

		s.persistGroupRun(ctx, groupRun.InstallGroupRunID, installEntries, app.StatusInProgress,
			runbookDescription(rb.RunbookName, installEntries))
		s.updateStepMetadata(ctx, runbookDescription(rb.RunbookName, installEntries), installEntries)
	}

	summary := postDeployRunbookSummary(installEntries)
	finalStatus := app.StatusSuccess
	if len(failedRunbookNames(installEntries)) > 0 || len(errs) > 0 {
		finalStatus = app.StatusError
	}

	s.persistGroupRun(ctx, groupRun.InstallGroupRunID, installEntries, finalStatus, summary)

	// Write the terminal metadata and summary onto the step itself. Without this
	// the last in-loop update is the only thing the step ever carried, and its
	// description read as mid-flight after the step had already finished.
	s.updateStepMetadata(ctx, summary, installEntries)

	if len(errs) > 0 {
		return fmt.Errorf("post-deploy runbooks had %d errors: %w", len(errs), errors.Join(errs...))
	}
	if finalStatus == app.StatusError {
		return fmt.Errorf("post-deploy runbooks failed: %s", summary)
	}

	return nil
}

// seedRunbookPlan reconciles each deployed install's recorded runbooks with the
// plan, so the step shows every runbook it intends to run from the first render.
//
// It is resume-aware because the step is retryable: a runbook that already
// succeeded is left alone, and one that failed or was cancelled is reset to
// pending with a bumped attempt so the retry starts a genuinely new run rather
// than adopting the failed one through its idempotency key.
func seedRunbookPlan(installEntries []app.InstallGroupRunInstall, runbooks []activities.ResolvedPostDeployRunbook) {
	for i := range installEntries {
		if installEntries[i].Status != statusSuccess {
			continue
		}

		previous := make(map[string]app.InstallGroupRunRunbook, len(installEntries[i].Runbooks))
		for _, rb := range installEntries[i].Runbooks {
			previous[rb.RunbookID] = rb
		}

		planned := make([]app.InstallGroupRunRunbook, 0, len(runbooks))
		for _, rb := range runbooks {
			entry := app.InstallGroupRunRunbook{
				RunbookID:   rb.RunbookID,
				RunbookName: rb.RunbookName,
				Status:      statusPending,
			}
			if prev, ok := previous[rb.RunbookID]; ok {
				switch prev.Status {
				case statusSuccess:
					entry = prev
				case statusError, statusCancelled:
					entry.Attempt = prev.Attempt + 1
				default:
					entry.Attempt = prev.Attempt
					entry.RunID = prev.RunID
					entry.WorkflowID = prev.WorkflowID
				}
			}
			planned = append(planned, entry)
		}

		installEntries[i].Phase = app.InstallGroupRunPhaseRunbook
		installEntries[i].Runbooks = planned
	}
}

// startRunbookWave starts rb on every eligible install, without awaiting any of
// them. An install is eligible if it deployed cleanly, none of its earlier
// runbooks failed, and this runbook has not already succeeded on a prior attempt.
func (s *Signal) startRunbookWave(
	ctx workflow.Context,
	run *app.AppBranchRun,
	rb activities.ResolvedPostDeployRunbook,
	runbookIdx int,
	installEntries []app.InstallGroupRunInstall,
	errs *[]error,
) []pendingRunbookRun {
	logger := workflow.GetLogger(ctx)

	pending := make([]pendingRunbookRun, 0, len(installEntries))
	for i := range installEntries {
		entry := &installEntries[i]
		if entry.Status != statusSuccess || runbookIdx >= len(entry.Runbooks) {
			continue
		}
		if hasFailedRunbook(entry) {
			continue
		}
		if entry.Runbooks[runbookIdx].Status == statusSuccess {
			continue
		}

		installID := entry.InstallID
		cb := callback.New(ctx, fmt.Sprintf("%s:runbook:%d:%d", installID, runbookIdx, entry.Runbooks[runbookIdx].Attempt))

		result, err := activities.AwaitCreateInstallRunbookRunWorkflow(ctx, &activities.CreateInstallRunbookRunWorkflowInput{
			InstallID:       installID,
			RunbookID:       rb.RunbookID,
			RunbookConfigID: rb.RunbookConfigID,
			TriggeredByID:   run.CreatedByID,
			Inputs:          rb.Inputs,
			Callback:        cb,
			IdempotencyKey:  runbookRunIdempotencyKey(s.RunID, installID, rb.RunbookID, entry.Runbooks[runbookIdx].Attempt),
		})
		if err != nil {
			*errs = append(*errs, fmt.Errorf("install %s runbook %s: unable to start: %w", installID, rb.RunbookName, err))
			entry.Runbooks[runbookIdx].Status = statusError
			cancelRemainingRunbooks(entry, runbookIdx+1)
			continue
		}

		s.childWorkflowIDs = append(s.childWorkflowIDs, result.WorkflowID)
		entry.Runbooks[runbookIdx].RunID = result.InstallRunbookRunID
		entry.Runbooks[runbookIdx].WorkflowID = result.WorkflowID
		entry.Runbooks[runbookIdx].Status = statusInProgress

		logger.Info("enqueued post-deploy runbook run",
			"install_id", installID,
			"runbook_name", rb.RunbookName,
			"workflow_id", result.WorkflowID,
			"install_runbook_run_id", result.InstallRunbookRunID,
		)

		// The run had already reached a terminal state, so no completion signal is
		// coming. Record what that state actually was — a deduped run may have
		// failed, and treating every terminal state as success would advance the
		// rollout past a runbook that never passed.
		if result.TerminalStatus != "" {
			if result.TerminalStatus == statusSuccess {
				entry.Runbooks[runbookIdx].Status = statusSuccess
			} else {
				*errs = append(*errs, fmt.Errorf("install %s runbook %s already finished as %s", installID, rb.RunbookName, result.TerminalStatus))
				entry.Runbooks[runbookIdx].Status = statusError
				cancelRemainingRunbooks(entry, runbookIdx+1)
			}
			continue
		}

		pending = append(pending, pendingRunbookRun{
			entryIdx:   i,
			runbookIdx: runbookIdx,
			installID:  installID,
			workflowID: result.WorkflowID,
			cb:         cb,
		})
	}

	return pending
}

// awaitRunbookWave waits for every run started in the wave.
func (s *Signal) awaitRunbookWave(
	ctx workflow.Context,
	rb activities.ResolvedPostDeployRunbook,
	pending []pendingRunbookRun,
	installEntries []app.InstallGroupRunInstall,
	errs *[]error,
) {
	logger := workflow.GetLogger(ctx)

	for _, p := range pending {
		entry := &installEntries[p.entryIdx]

		if _, err := callback.AwaitWithTimeout(ctx, p.cb, callback.FallbackAwaitTimeout); err != nil {
			*errs = append(*errs, fmt.Errorf("install %s runbook %s workflow %s: %w", p.installID, rb.RunbookName, p.workflowID, err))
			entry.Runbooks[p.runbookIdx].Status = statusError
			cancelRemainingRunbooks(entry, p.runbookIdx+1)
			continue
		}

		entry.Runbooks[p.runbookIdx].Status = statusSuccess

		logger.Info("post-deploy runbook completed",
			"install_id", p.installID,
			"runbook_name", rb.RunbookName,
			"workflow_id", p.workflowID,
		)
	}
}

func (s *Signal) persistGroupRun(
	ctx workflow.Context,
	groupRunID string,
	installEntries []app.InstallGroupRunInstall,
	status app.Status,
	description string,
) {
	// Completed/failed stay the deploy tally: this step never rewrites an
	// install's deploy result, so a runbook failure shows as a failed step and a
	// failed group rather than masquerading as a failed deploy.
	completed, failed := tallyDeployed(installEntries)
	_ = activities.AwaitUpdateInstallGroupRun(ctx, &activities.UpdateInstallGroupRunInput{
		InstallGroupRunID: groupRunID,
		Installs:          installEntries,
		CompletedInstalls: completed,
		FailedInstalls:    failed,
		Status: app.CompositeStatus{
			Status:                 status,
			StatusHumanDescription: description,
		},
	})
}

func runbookRunIdempotencyKey(runID, installID, runbookID string, attempt int) string {
	return fmt.Sprintf("branch-run:%s:%s:%s:%d", runID, installID, runbookID, attempt)
}

func hasFailedRunbook(entry *app.InstallGroupRunInstall) bool {
	for _, rb := range entry.Runbooks {
		if rb.Status == statusError {
			return true
		}
	}
	return false
}

// cancelRemainingRunbooks marks the runbooks an install will never reach, so the
// step shows why they didn't run instead of leaving them pending forever.
func cancelRemainingRunbooks(entry *app.InstallGroupRunInstall, from int) {
	for i := from; i < len(entry.Runbooks); i++ {
		if entry.Runbooks[i].Status == statusPending {
			entry.Runbooks[i].Status = statusCancelled
		}
	}
}

// branchRunbookInputs builds the runbook input overrides sourced from the branch
// run's VCS context. Only inputs the runbook config declares are kept downstream.
func branchRunbookInputs(run *app.AppBranchRun) map[string]string {
	inputs := map[string]string{}
	if run.HeadSHA != "" {
		inputs["commit_sha"] = run.HeadSHA
	}
	if run.BaseBranch != "" {
		inputs["base_branch"] = run.BaseBranch
	}
	if run.PRNumber != nil {
		inputs["pr_number"] = strconv.Itoa(*run.PRNumber)
	}
	return inputs
}

func tallyDeployed(installEntries []app.InstallGroupRunInstall) (int, int) {
	completed := 0
	failed := 0
	for _, e := range installEntries {
		switch e.Status {
		case statusSuccess:
			completed++
		case statusError:
			failed++
		}
	}
	return completed, failed
}

func planDescription(runbooks []activities.ResolvedPostDeployRunbook) string {
	names := make([]string, 0, len(runbooks))
	for _, rb := range runbooks {
		names = append(names, rb.RunbookName)
	}
	return "running " + strings.Join(names, " → ")
}

func runbookDescription(runbookName string, installEntries []app.InstallGroupRunInstall) string {
	done := 0
	total := 0
	for _, e := range installEntries {
		for _, rb := range e.Runbooks {
			if rb.RunbookName != runbookName {
				continue
			}
			total++
			if rb.Status != statusPending && rb.Status != statusInProgress {
				done++
			}
		}
	}
	return fmt.Sprintf("running %s (%d/%d installs)", runbookName, done, total)
}

func failedRunbookNames(installEntries []app.InstallGroupRunInstall) []string {
	var order []string
	seen := map[string]struct{}{}
	for _, e := range installEntries {
		for _, rb := range e.Runbooks {
			if rb.Status != statusError {
				continue
			}
			if _, ok := seen[rb.RunbookName]; !ok {
				seen[rb.RunbookName] = struct{}{}
				order = append(order, rb.RunbookName)
			}
		}
	}
	return order
}

// postDeployRunbookSummary names the runbook that stopped the rollout, so the
// step's final description says more than a bare pass/fail.
func postDeployRunbookSummary(installEntries []app.InstallGroupRunInstall) string {
	ran := map[string]struct{}{}
	for _, e := range installEntries {
		for _, rb := range e.Runbooks {
			ran[rb.RunbookName] = struct{}{}
		}
	}

	if len(ran) == 0 {
		return "no post-deploy runbooks to run"
	}

	if failed := failedRunbookNames(installEntries); len(failed) > 0 {
		return fmt.Sprintf("%s failed", strings.Join(failed, ", "))
	}

	if len(ran) == 1 {
		return "1 post-deploy runbook succeeded"
	}
	return fmt.Sprintf("%d post-deploy runbooks succeeded", len(ran))
}

// updateStepMetadata writes the given description and the current per-install
// runbook state onto the flow step. The description is used verbatim — callers
// pass an already-formatted string.
func (s *Signal) updateStepMetadata(ctx workflow.Context, description string, installEntries []app.InstallGroupRunInstall) {
	if s.StepID == "" {
		return
	}

	installs := make([]any, 0, len(installEntries))
	for _, e := range installEntries {
		runbooks := make([]any, 0, len(e.Runbooks))
		for _, rb := range e.Runbooks {
			runbooks = append(runbooks, map[string]any{
				"runbook_id":   rb.RunbookID,
				"runbook_name": rb.RunbookName,
				"run_id":       rb.RunID,
				"workflow_id":  rb.WorkflowID,
				"status":       rb.Status,
			})
		}

		installs = append(installs, map[string]any{
			"install_id": e.InstallID,
			"status":     e.Status,
			"phase":      e.Phase,
			"runbooks":   runbooks,
		})
	}

	_ = statusactivities.AwaitPkgStatusUpdateFlowStepStatus(ctx, statusactivities.UpdateStatusRequest{
		ID: s.StepID,
		Status: app.CompositeStatus{
			Status:                 app.StatusInProgress,
			StatusHumanDescription: description,
			Metadata: map[string]any{
				"installs": installs,
			},
		},
	})
}
