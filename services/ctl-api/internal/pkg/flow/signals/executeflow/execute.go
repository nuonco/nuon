package executeflow

import (
	"time"

	"go.temporal.io/sdk/workflow"
	"go.uber.org/zap"

	"github.com/pkg/errors"

	tmetrics "github.com/nuonco/nuon/pkg/temporal/metrics"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/flow"
	flowdirective "github.com/nuonco/nuon/services/ctl-api/internal/pkg/flow/directive"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/log"
	statusactivities "github.com/nuonco/nuon/services/ctl-api/internal/pkg/workflows/status/activities"
	workflowactivities "github.com/nuonco/nuon/services/ctl-api/internal/pkg/workflows/workflow/activities"
)

// residentIdleTimeout bounds how long a Resident workflow parks waiting for the
// next step (e.g. an appended step) after running 0->end. On idle the execute
// loop returns cleanly so the queue signal completes and history stays finite;
// the workflow re-warms lazily on the next dispatch.
const residentIdleTimeout = 15 * time.Minute

// executeFlow runs the workflow conductor with run-based execution.
// Each execution segment (initial, retry, skip, resume) is tracked as a WorkflowRun.
// The flow pauses at approval points and errors, waiting for update handlers to resume.
func (s *Signal) executeFlow(ctx workflow.Context) (retErr error) {
	// Mark the conductor as started so a retry-step update that lands during a
	// re-warm (before the loop reaches its parked state) is still cloned+queued.
	s.executeStarted = true

	// Initialize temporal metrics writer if the underlying metrics writer was injected.
	if s.mw != nil && s.v != nil {
		tmw, err := tmetrics.New(s.v, tmetrics.WithMetricsWriter(s.mw))
		if err == nil {
			s.tmw = tmw
		}
	}

	// Emit workflow completion metrics on exit.
	flowStart := workflow.Now(ctx)
	defer func() {
		if s.tmw == nil {
			return
		}
		wfTags := []string{"workflow_type", s.WorkflowType, "org_id", s.OrgID}
		status := "success"
		if retErr != nil {
			status = "error"
		}
		s.tmw.Incr(ctx, "workflow.completed", append(wfTags, "status", status)...)
		s.tmw.Timing(ctx, "workflow.latency", workflow.Now(ctx).Sub(flowStart), append(wfTags, "status", status)...)
	}()

	// Create and execute the initial run. Resident hosts may rewarm with
	// historical (terminal) steps from earlier steps; start the initial run at
	// the first pending group so completed groups are never replayed. If none
	// are pending the run is a no-op and the loop parks for the next append.
	initialRunType := app.WorkflowRunTypeInitial
	initialStartIdx := 0
	initialStepID := ""
	if s.Resident {
		if s.resumeRequested {
			// A retry-step update raced ahead of the conductor during re-warm:
			// honor its resume so the retry runs instead of being dropped.
			initialRunType = s.resumeRunType
			initialStartIdx = s.resumeStartIdx
			initialStepID = s.resumeStepID
			s.resumeRequested = false
		} else if pos, ok := s.firstPendingGroupPosition(ctx); ok {
			initialStartIdx = pos
		}
	}
	run, err := s.createRun(ctx, initialRunType, initialStepID, initialStartIdx)
	if err != nil {
		return err
	}

	for {
		runErr := s.executeRun(ctx, run)

		// Only a resume requested while parked below is valid; drop stale ones.
		s.resumeRequested = false

		if runErr == nil {
			if s.cancelRequested {
				return nil
			}

			// Run completed without error. Check if workflow is fully done
			// or if we paused at an approval/directive point.
			if s.isWorkflowComplete(ctx) {
				s.updateRunStatus(ctx, run.ID, app.StatusSuccess)
				if !s.Resident {
					return nil
				}
				// Resident: stay alive to accept the next step (e.g. a
				// appended step) instead of completing. Fall through to the
				// shared park below, which honors appendRequested and an idle
				// timeout in resident mode.
			} else if !s.Resident {
				// Paused at approval - update run status and wait for resume.
				// Resident hosts skip this: a non-complete state just means a
				// freshly appended step is pending, which parkResident runs next.
				s.updateRunStatus(ctx, run.ID, app.AwaitingApproval)
			}
		} else {
			if s.cancelRequested {
				return nil
			}

			// FlowStoppedErr is a terminal state — not retryable
			if stoppedErr, ok := runErr.(*flow.FlowStoppedErr); ok {
				s.updateRunStatus(ctx, run.ID, app.StatusError)
				metadata := map[string]any{
					"error_message": runErr.Error(),
					"stopped":       true,
				}
				if stoppedErr.RetriesExhausted {
					metadata["retries_exhausted"] = true
				}
				humanDesc := stoppedErr.StatusHumanDescription
				if humanDesc == "" {
					humanDesc = "workflow stopped"
				}
				_ = statusactivities.AwaitPkgStatusUpdateFlowStatus(ctx, statusactivities.UpdateStatusRequest{
					ID: s.WorkflowID,
					Status: app.CompositeStatus{
						Status:                 app.StatusError,
						StatusHumanDescription: humanDesc,
						Metadata:               metadata,
					},
				})
				// Resident hosts must survive a stopped/failed step: one bad
				// appended step becomes a terminal (errored) group and the host
				// stays up to run later steps. The step's own status is mirrored
				// onto its target row by the inner signal.
				if !s.Resident {
					return runErr
				}
			} else {
				// Actual execution error
				s.updateRunStatus(ctx, run.ID, app.StatusError)

				if !s.checkRetryable(ctx) {
					// Same as above: keep a resident host alive past a step error.
					if !s.Resident {
						return runErr
					}
				} else {
					// Mark workflow as failed, awaiting retry
					_ = statusactivities.AwaitPkgStatusUpdateFlowStatus(ctx, statusactivities.UpdateStatusRequest{
						ID: s.WorkflowID,
						Status: app.CompositeStatus{
							Status:                 app.StatusError,
							StatusHumanDescription: "workflow failed, awaiting retry",
							Metadata: map[string]any{
								"error_message":  runErr.Error(),
								"awaiting_retry": true,
							},
						},
					})
				}
			}
		}

		// Wait for the next thing to run, cancel, or idle out. Resident hosts
		// park until a pending group appears (re-scanning on every wake so a
		// step appended mid-run is never missed); other workflows wait for an
		// explicit resume/cancel.
		if s.Resident {
			parked, perr := s.parkResident(ctx)
			if perr != nil {
				return perr
			}
			if s.cancelRequested {
				s.updateRunStatus(ctx, run.ID, app.StatusCancelled)
				return runErr
			}
			if !parked {
				// Idle timeout with nothing pending — exit cleanly.
				return nil
			}
			// parkResident set resumeStartIdx to the first pending group.
			run, err = s.createRun(ctx, app.WorkflowRunTypeResume, "", s.resumeStartIdx)
			if err != nil {
				return err
			}
			continue
		}

		s.awaitingResume = true
		err = workflow.Await(ctx, func() bool {
			return s.resumeRequested || s.cancelRequested
		})
		s.awaitingResume = false
		if err != nil {
			return err
		}

		if s.cancelRequested {
			s.updateRunStatus(ctx, run.ID, app.StatusCancelled)
			return runErr
		}

		// Create a new run for the resume.
		s.resumeRequested = false
		run, err = s.createRun(ctx, s.resumeRunType, s.resumeStepID, s.resumeStartIdx)
		if err != nil {
			return err
		}
	}
}

// executeRun executes a single workflow run, directly managing step generation
// and execution without going through the WorkflowConductor.
func (s *Signal) executeRun(ctx workflow.Context, run *app.WorkflowRun) error {
	cfg := s.stepConfig()
	startIdx := run.StartFromIdx

	for {
		err := s.handle(ctx, startIdx)
		if err == nil {
			return nil
		}

		// Handle ContinueAsNew (batch size limit)
		if cerr, ok := err.(*flow.ContinueAsNewErr); ok && cerr != nil {
			startIdx = cerr.StartFromStepIdx
			continue
		}

		// ApprovalPauseErr means we stopped at an approval or pause - return nil to enter wait loop
		if _, ok := err.(*flow.ApprovalPauseErr); ok {
			return nil
		}

		// FlowStoppedErr means the workflow was stopped (denied/skipped) — not a retryable error
		if _, ok := err.(*flow.FlowStoppedErr); ok {
			return err
		}

		// Actual failure
		_ = cfg // suppress unused warning in case of early return refactors
		return err
	}
}

// handle manages the full lifecycle of a flow execution: generate steps, then
// dispatch groups sequentially. Each group is dispatched as an execute-workflow-step-group
// signal. After each group, the flow checks the workflow's ResultDirective and the
// pause state.
func (s *Signal) handle(ctx workflow.Context, startFromGroupIdx int) error {
	l, err := log.WorkflowLogger(ctx)
	if err != nil {
		return nil
	}

	flw, err := workflowactivities.AwaitPkgWorkflowsFlowGetFlowByID(ctx, s.WorkflowID)
	if err != nil {
		return errors.Wrap(err, "unable to get workflow object")
	}

	// Build metric tags from the workflow for lifecycle metrics.
	wfTags := []string{"workflow_type", s.WorkflowType, "org_id", s.OrgID}

	if flw.Status.Status == app.StatusCancelled {
		return errors.New("workflow already cancelled")
	}
	// Restore cancel flag from persisted metadata. The in-memory
	// cancelRequested flag is lost across ContinueAsNew boundaries, but
	// cancel_requested_at in the DB survives.
	if flw.Status.Metadata != nil {
		if _, ok := flw.Status.Metadata["cancel_requested_at"]; ok {
			s.cancelRequested = true
			return nil
		}
	}

	defer func() {
		if errors.Is(ctx.Err(), workflow.ErrCanceled) {
			cancelCtx, cancelCtxCancel := workflow.NewDisconnectedContext(ctx)
			defer cancelCtxCancel()

			if err := statusactivities.AwaitPkgStatusUpdateFlowStatus(cancelCtx, statusactivities.UpdateStatusRequest{
				ID: s.WorkflowID,
				Status: app.CompositeStatus{
					Status: app.StatusCancelled,
				},
			}); err != nil {
				l.Error("unable to update status on cancellation", zap.Error(err))
			}
		}
	}()

	if startFromGroupIdx == 0 {
		if err := workflowactivities.AwaitPkgWorkflowsFlowUpdateFlowStartedAtByID(ctx, s.WorkflowID); err != nil {
			return err
		}

		// Emit workflow lifecycle metrics on first execution.
		if s.tmw != nil {
			s.tmw.Incr(ctx, "workflow.started", wfTags...)
			s.tmw.Timing(ctx, "workflow.start_latency", workflow.Now(ctx).Sub(flw.CreatedAt), wfTags...)
		}
	}

	cfg := s.stepConfig()

	// eagerQueueSignalID tracks whether we used eager step group generation.
	// If non-empty, we must call CompleteStepGeneration before executing
	// groups beyond the eager set.
	var eagerQueueSignalID string
	var eagerGroupCount int

	// completeDone, completedFlw, and completeErr are used to run
	// CompleteStepGeneration in a background goroutine so remaining step
	// groups are persisted to the DB while eager groups execute.
	var completeDone workflow.Channel
	var completedFlw *app.Workflow
	var completeErr error

	// Generate steps if the workflow doesn't already have them.
	// Steps may be pre-created (e.g. by tests or by a previous run that was
	// ContinueAsNew'd) — in that case, skip generation.
	if len(flw.Steps) == 0 {
		// Resident host workflows (e.g. interactive append-driven hosts) start with no steps and no
		// generate-steps signal — they exist only to accept append-step updates.
		// Skip generation and return so the execute loop parks for the first
		// step. Once a step is appended, handle() resumes with len(Steps) > 0 and
		// runs only the appended group.
		if s.Resident && (flw.GenerateStepsSignal == nil || flw.GenerateStepsSignal.Signal == nil) {
			l.Debug("resident workflow has no steps; parking for append-step")
			return nil
		}

		l.Debug("generating steps for workflow")
		if err := statusactivities.AwaitPkgStatusUpdateFlowStatus(ctx, statusactivities.UpdateStatusRequest{
			ID: s.WorkflowID,
			Status: app.CompositeStatus{
				Status:                 app.StatusInProgress,
				StatusHumanDescription: "generating steps for workflow",
			},
		}); err != nil {
			return err
		}

		if flw.GenerateStepsSignal == nil || flw.GenerateStepsSignal.Signal == nil {
			return errors.Errorf("workflow %s has no steps and no generate-steps signal", s.WorkflowID)
		}

		// Use eager step groups: fetch and persist the eager groups so we can
		// begin executing them while the remaining groups may still be generating.
		earlyResult, err := flow.GenerateEagerStepGroups(ctx, cfg, flw)
		if err != nil {
			_ = statusactivities.AwaitPkgStatusUpdateFlowStatus(ctx, statusactivities.UpdateStatusRequest{
				ID: s.WorkflowID,
				Status: app.CompositeStatus{
					Status:                 app.StatusError,
					StatusHumanDescription: "error while generating steps",
					Metadata: map[string]any{
						"error_message": err.Error(),
					},
				},
			})

			return errors.Wrap(err, "unable to generate workflow steps")
		}

		flw = earlyResult.Workflow
		eagerQueueSignalID = earlyResult.QueueSignalID
		eagerGroupCount = len(flw.StepGroups)

		if err := statusactivities.AwaitPkgStatusUpdateFlowStatus(ctx, statusactivities.UpdateStatusRequest{
			ID: s.WorkflowID,
			Status: app.CompositeStatus{
				Status:                 app.StatusInProgress,
				StatusHumanDescription: "generated initial step groups, executing",
				Metadata: map[string]any{
					"eager_steps_loaded": true,
				},
			},
		}); err != nil {
			return err
		}

		// Start completing step generation in the background so remaining
		// groups are persisted to the DB (and visible in the UI) while
		// the eager groups execute.
		if eagerQueueSignalID != "" {
			completeDone = workflow.NewChannel(ctx)
			workflow.Go(ctx, func(gCtx workflow.Context) {
				completedFlw, completeErr = flow.CompleteStepGeneration(gCtx, cfg, flw, eagerQueueSignalID)
				completeDone.Send(gCtx, true)
			})
		}
	} else {
		l.Debug("steps already exist, skipping generation", zap.Int("step_count", len(flw.Steps)))
	}

	// Load step groups for the workflow.
	// If groups exist (new path), iterate over them. Otherwise fall back to
	// collecting group indices from steps (backward compat for in-flight workflows).
	stepGroups, _ := workflowactivities.AwaitPkgWorkflowsFlowGetFlowStepGroups(ctx, s.WorkflowID)

	var groups []app.WorkflowStepGroup
	if len(stepGroups) > 0 {
		groups = stepGroups
	} else {
		// Backward compat: build synthetic group objects from step GroupIdx values.
		groupIdxs := collectGroupIndices(flw.Steps)
		for _, gIdx := range groupIdxs {
			groups = append(groups, app.WorkflowStepGroup{
				GroupIdx: gIdx,
				Parallel: isGroupParallel(flw.Steps, gIdx),
			})
		}
	}

	// Execute groups
	l.Debug("executing groups for workflow", zap.Int("group_count", len(groups)))

	// Resident hosts replay nothing: precompute which groups still have a
	// non-terminal step so already-finished groups (earlier appended steps) are
	// skipped even on a cold rewarm that starts at group 0.
	var residentPending map[int]bool
	if s.Resident {
		residentPending = make(map[int]bool)
		for _, st := range flw.Steps {
			if !isStepTerminal(st.Status.Status) {
				residentPending[st.GroupIdx] = true
			}
		}
	}

	var eagerExecuted map[int]bool

	for gi := startFromGroupIdx; gi < len(groups); gi++ {
		if s.cancelRequested {
			s.markRemainingGroupStepsDiscarded(ctx, l, groups, gi-1)
			s.markRemainingStepsNotAttempted(ctx, l)
			return nil
		}

		group := &groups[gi]

		if stopped, err := s.stopIfRunnerDisabled(ctx, l, flw, groups, gi); err != nil {
			return err
		} else if stopped {
			return flow.NewFlowStoppedErr("", "the install runner is disabled")
		}

		if s.Resident && !residentPending[group.GroupIdx] {
			continue
		}

		if eagerExecuted[group.GroupIdx] {
			continue
		}

		l.Debug("dispatching group", zap.Int("group_idx", group.GroupIdx), zap.Int("group_position", gi), zap.String("step_group_id", group.ID), zap.Bool("parallel", group.Parallel))

		directive, err := s.executeGroup(ctx, group, flw)
		if err != nil {
			if errors.Is(ctx.Err(), workflow.ErrCanceled) {
				return err
			}

			if err := workflowactivities.AwaitPkgWorkflowsFlowUpdateFlowFinishedAtByID(ctx, s.WorkflowID); err != nil {
				l.Error("unable to update finished at", zap.Error(err))
			}

			// If cancellation was requested, preserve the cancelled status
			// that the cancel handler already set — don't overwrite it with error.
			if s.cancelRequested {
				_ = statusactivities.AwaitPkgStatusUpdateFlowStatus(ctx, statusactivities.UpdateStatusRequest{
					ID: s.WorkflowID,
					Status: app.CompositeStatus{
						Status:                 app.StatusCancelled,
						StatusHumanDescription: "workflow cancelled",
					},
				})
				s.markRemainingGroupStepsDiscarded(ctx, l, groups, gi)

				return nil
			} else {
				_ = statusactivities.AwaitPkgStatusUpdateFlowStatus(ctx, statusactivities.UpdateStatusRequest{
					ID: s.WorkflowID,
					Status: app.CompositeStatus{
						Status:                 app.StatusError,
						StatusHumanDescription: "error while executing group",
						Metadata: map[string]any{
							"error_message": err.Error(),
							"group_idx":     group.GroupIdx,
						},
					},
				})
			}

			return errors.Wrapf(err, "group %d failed", group.GroupIdx)
		}

		l.Debug("group completed", zap.Int("group_idx", group.GroupIdx), zap.String("directive", directive))

		switch flowdirective.Group(directive) {
		case flowdirective.GroupContinue, "":
			if s.cancelRequested {
				s.markRemainingGroupStepsDiscarded(ctx, l, groups, gi)
				s.markRemainingStepsNotAttempted(ctx, l)
				return nil
			}
			// Check if pause was requested
			if s.pauseRequested {
				return &flow.ApprovalPauseErr{StepID: "paused"}
			}

		case flowdirective.GroupStop:
			// Derive the reason before the sweeps overwrite step statuses.
			stepName, reason := s.groupStopReason(ctx, group)

			s.markRemainingGroupStepsDiscarded(ctx, l, groups, gi)
			s.markRemainingStepsNotAttempted(ctx, l)
			if err := workflowactivities.AwaitPkgWorkflowsFlowUpdateFlowFinishedAtByID(ctx, s.WorkflowID); err != nil {
				l.Error("unable to update finished at", zap.Error(err))
			}
			stoppedErr := flow.NewFlowStoppedErr(stepName, reason)
			if reason != "" {
				stoppedErr.StatusHumanDescription = "workflow stopped: " + reason
			}
			stoppedErr.RetriesExhausted = s.checkGroupRetriesExhausted(ctx, group)
			return stoppedErr

		case flowdirective.GroupAwaitApproval:
			return flow.NewApprovalPauseErr("")

		case flowdirective.GroupRetryGroup:
			// Clone the group and re-dispatch the same group position.
			if err := s.cloneGroupForRetry(ctx, group.GroupIdx); err != nil {
				// Retry limit exceeded: treat as a stop directive.
				s.markRemainingGroupStepsDiscarded(ctx, l, groups, gi)
				s.markRemainingStepsNotAttempted(ctx, l)
				if finErr := workflowactivities.AwaitPkgWorkflowsFlowUpdateFlowFinishedAtByID(ctx, s.WorkflowID); finErr != nil {
					l.Error("unable to update finished at", zap.Error(finErr))
				}
				stoppedErr := flow.NewFlowStoppedErr("", err.Error())
				stoppedErr.RetriesExhausted = true
				return stoppedErr
			}
			// Re-fetch groups
			stepGroups, _ = workflowactivities.AwaitPkgWorkflowsFlowGetFlowStepGroups(ctx, s.WorkflowID)
			if len(stepGroups) > 0 {
				groups = stepGroups
			} else {
				flw, err = workflowactivities.AwaitPkgWorkflowsFlowGetFlowByID(ctx, s.WorkflowID)
				if err != nil {
					return errors.Wrap(err, "unable to re-fetch workflow after retry-group")
				}
				groupIdxs := collectGroupIndices(flw.Steps)
				groups = groups[:0]
				for _, gIdx := range groupIdxs {
					groups = append(groups, app.WorkflowStepGroup{
						GroupIdx: gIdx,
						Parallel: isGroupParallel(flw.Steps, gIdx),
					})
				}
			}
			gi-- // Retry the same group position
			continue

		case flowdirective.GroupSkipGroup:
			continue
		}

		// After the last eager group finishes, wait for the background
		// CompleteStepGeneration goroutine and reload groups.
		if gi+1 == eagerGroupCount && completeDone != nil {
			l.Debug("waiting for parallel step generation to complete", zap.Int("eager_group_count", eagerGroupCount))
			completeDone.Receive(ctx, nil)
			completeDone = nil // only complete once

			if completeErr != nil {
				return errors.Wrap(completeErr, "unable to complete step generation")
			}
			flw = completedFlw

			eagerExecuted = make(map[int]bool, gi+1)
			for i := 0; i <= gi && i < len(groups); i++ {
				eagerExecuted[groups[i].GroupIdx] = true
			}

			// Check for cancellation before overwriting status. The cancel
			// handler may have set StatusCancelled while we were waiting for
			// step generation to complete.
			if s.cancelRequested {
				s.markRemainingGroupStepsDiscarded(ctx, l, groups, gi)
				s.markRemainingStepsNotAttempted(ctx, l)
				return nil
			}

			_ = statusactivities.AwaitPkgStatusUpdateFlowStatus(ctx, statusactivities.UpdateStatusRequest{
				ID: s.WorkflowID,
				Status: app.CompositeStatus{
					Status:                 app.StatusInProgress,
					StatusHumanDescription: "all steps generated",
					Metadata: map[string]any{
						"all_steps_loaded": true,
					},
				},
			})

			// Reload groups from DB now that all are persisted.
			stepGroups, _ = workflowactivities.AwaitPkgWorkflowsFlowGetFlowStepGroups(ctx, s.WorkflowID)
			if len(stepGroups) > 0 {
				groups = stepGroups
			} else {
				groupIdxs := collectGroupIndices(flw.Steps)
				groups = groups[:0]
				for _, gIdx := range groupIdxs {
					groups = append(groups, app.WorkflowStepGroup{
						GroupIdx: gIdx,
						Parallel: isGroupParallel(flw.Steps, gIdx),
					})
				}
			}

			// Eager groups can sit mid-list in group_idx order; rewind so groups ordered before them are not skipped.
			next := len(groups)
			for i := range groups {
				if !eagerExecuted[groups[i].GroupIdx] {
					next = i
					break
				}
			}
			gi = next - 1
		}

		// ContinueAsNew every 5 groups to bound workflow history
		if (gi+1-startFromGroupIdx) > 0 && (gi+1-startFromGroupIdx)%5 == 0 {
			return &flow.ContinueAsNewErr{StartFromStepIdx: gi + 1}
		}
	}

	// All groups done
	if err := workflowactivities.AwaitPkgWorkflowsFlowUpdateFlowFinishedAtByID(ctx, s.WorkflowID); err != nil {
		l.Error("unable to update finished at", zap.Error(err))
	}

	if err := statusactivities.AwaitPkgStatusUpdateFlowStatus(ctx, statusactivities.UpdateStatusRequest{
		ID: s.WorkflowID,
		Status: app.CompositeStatus{
			Status:                 app.StatusSuccess,
			StatusHumanDescription: "successfully executed workflow",
		},
	}); err != nil {
		return err
	}

	return nil
}

// isGroupParallel returns true if any step in the group has GroupParallel=true.
func isGroupParallel(steps []app.WorkflowStep, groupIdx int) bool {
	for _, step := range steps {
		if step.GroupIdx == groupIdx && step.GroupParallel {
			return true
		}
	}
	return false
}

// collectGroupIndices extracts sorted unique GroupIdx values from steps.
func collectGroupIndices(steps []app.WorkflowStep) []int {
	seen := make(map[int]bool)
	var groups []int
	for _, step := range steps {
		if !seen[step.GroupIdx] {
			seen[step.GroupIdx] = true
			groups = append(groups, step.GroupIdx)
		}
	}
	// Steps are already ordered by Idx, so groups come out in order
	return groups
}

// findGroupPositionForStep returns the position (index into groupIdxs) of the
// group that contains the given step. Returns 0 if the step is not found.
func (s *Signal) findGroupPositionForStep(ctx workflow.Context, stepID string) int {
	steps, err := workflowactivities.AwaitPkgWorkflowsFlowGetFlowSteps(ctx, workflowactivities.GetFlowStepsRequest{
		FlowID: s.WorkflowID,
	})
	if err != nil {
		return 0
	}

	// Find the step's GroupIdx
	stepGroupIdx := -1
	for _, step := range steps {
		if step.ID == stepID {
			stepGroupIdx = step.GroupIdx
			break
		}
	}
	if stepGroupIdx == -1 {
		return 0
	}

	// Find the position of that GroupIdx in the ordered group list
	groupIdxs := collectGroupIndices(steps)
	for i, gIdx := range groupIdxs {
		if gIdx == stepGroupIdx {
			return i
		}
	}
	return 0
}

// firstPendingGroupPosition returns the position (index into the ordered group
// slice) of the first group that still has a non-terminal step, and whether one
// exists. Resident hosts use it to (a) start a rewarmed run at the first
// unfinished group instead of replaying completed history, and (b) decide
// whether to run or park after each step. Terminal-but-failed groups (e.g. a
// appended step that errored) are skipped, so one failed step never wedges the
// host.
func (s *Signal) firstPendingGroupPosition(ctx workflow.Context) (int, bool) {
	groups, err := workflowactivities.AwaitPkgWorkflowsFlowGetFlowStepGroups(ctx, s.WorkflowID)
	if err != nil || len(groups) == 0 {
		return 0, false
	}
	steps, err := workflowactivities.AwaitPkgWorkflowsFlowGetFlowSteps(ctx, workflowactivities.GetFlowStepsRequest{
		FlowID: s.WorkflowID,
	})
	if err != nil {
		return 0, false
	}

	pending := make(map[int]bool)
	for _, st := range steps {
		if !isStepTerminal(st.Status.Status) {
			pending[st.GroupIdx] = true
		}
	}
	for pos, g := range groups {
		if pending[g.GroupIdx] {
			return pos, true
		}
	}
	return 0, false
}

// parkResident blocks a resident host until there is a pending group to run
// (e.g. an appended appended step), cancellation is requested, or the idle
// timeout elapses. It returns (true, nil) when a pending group was found
// (resumeStartIdx is set to its position), (false, nil) on idle timeout or
// cancel, and a non-nil error only on context failure. It re-scans on every
// wake so appends that arrived while a run was still executing — i.e. before
// the loop parked — are never missed.
func (s *Signal) parkResident(ctx workflow.Context) (bool, error) {
	for {
		if s.cancelRequested {
			return false, nil
		}
		if pos, ok := s.firstPendingGroupPosition(ctx); ok {
			s.resumeStartIdx = pos
			return true, nil
		}

		s.awaitingResume = true
		woke, err := workflow.AwaitWithTimeout(ctx, residentIdleTimeout, func() bool {
			return s.resumeRequested || s.appendRequested || s.cancelRequested
		})
		s.awaitingResume = false
		s.resumeRequested = false
		s.appendRequested = false
		if err != nil {
			return false, err
		}
		if !woke {
			// The idle timer fired. Before closing, make sure no append-step or
			// retry-step update handler is still running and that no step became
			// pending. Those handlers persist their step rows before they set
			// their wake flags, so closing on the timer alone could drop a step
			// that is still being written. If an update is in flight or a step
			// is pending, loop back to re-scan and run it.
			if s.updatesInFlight > 0 {
				continue
			}
			if pos, ok := s.firstPendingGroupPosition(ctx); ok {
				s.resumeStartIdx = pos
				return true, nil
			}
			// firstPendingGroupPosition yields on activities; re-check the wake
			// flags in case an update handler started during that window.
			if s.updatesInFlight > 0 || s.resumeRequested || s.appendRequested {
				continue
			}
			// Nothing pending: return cleanly so the queue signal completes and
			// history stays bounded. The host re-warms on the next dispatch.
			return false, nil
		}
		// Woke — loop back to re-scan for a pending group.
	}
}

// markRemainingGroupStepsDiscarded marks all remaining groups and their
// non-terminal steps as discarded. This is called when a group returns a
// "stop" directive (e.g. plan denied) so that future groups and their steps
// reflect that they were discarded due to an earlier stop.
func (s *Signal) markRemainingGroupStepsDiscarded(ctx workflow.Context, l *zap.Logger, groups []app.WorkflowStepGroup, currentGroupPosition int) {
	if currentGroupPosition+1 >= len(groups) {
		return
	}

	steps, err := workflowactivities.AwaitPkgWorkflowsFlowGetFlowSteps(ctx, workflowactivities.GetFlowStepsRequest{
		FlowID: s.WorkflowID,
	})
	if err != nil {
		l.Warn("unable to fetch steps to mark as not-attempted", zap.Error(err))
		return
	}

	// Build set of group indices that come after the current group.
	futureGroupIdxs := make(map[int]bool)
	for i := currentGroupPosition + 1; i < len(groups); i++ {
		futureGroupIdxs[groups[i].GroupIdx] = true

		// Mark the group object itself as discarded.
		if groups[i].ID != "" {
			if err := statusactivities.AwaitPkgStatusUpdateFlowStepGroupStatus(ctx, statusactivities.UpdateStatusRequest{
				ID: groups[i].ID,
				Status: app.CompositeStatus{
					Status: app.StatusDiscarded,
					Metadata: map[string]any{
						"reason": "discarded: workflow stopped before group was reached",
					},
				},
			}); err != nil {
				l.Warn("failed to mark group as discarded",
					zap.String("step_group_id", groups[i].ID),
					zap.Error(err))
			}
		}
	}

	for _, step := range steps {
		if !futureGroupIdxs[step.GroupIdx] {
			continue
		}
		if isStepTerminal(step.Status.Status) {
			continue
		}
		if err := statusactivities.AwaitPkgStatusUpdateFlowStepStatus(ctx, statusactivities.UpdateStatusRequest{
			ID: step.ID,
			Status: app.CompositeStatus{
				Status: app.StatusNotAttempted,
				Metadata: map[string]any{
					"reason": "workflow stopped before group was reached",
				},
			},
		}); err != nil {
			l.Warn("failed to mark step as not-attempted", zap.String("step_id", step.ID), zap.Error(err))
		}
	}
}

// markRemainingStepsNotAttempted marks all non-terminal steps in the workflow
// as not-attempted. Called when the workflow is stopped (e.g. retries exhausted)
// so the dashboard clearly shows which steps were never reached.
func (s *Signal) markRemainingStepsNotAttempted(ctx workflow.Context, l *zap.Logger) {
	steps, err := workflowactivities.AwaitPkgWorkflowsFlowGetFlowSteps(ctx, workflowactivities.GetFlowStepsRequest{
		FlowID: s.WorkflowID,
	})
	if err != nil {
		l.Warn("unable to fetch steps to mark as not-attempted", zap.Error(err))
		return
	}

	for _, step := range steps {
		if isStepTerminal(step.Status.Status) {
			continue
		}
		if err := statusactivities.AwaitPkgStatusUpdateFlowStepStatus(ctx, statusactivities.UpdateStatusRequest{
			ID: step.ID,
			Status: app.CompositeStatus{
				Status: app.StatusNotAttempted,
				Metadata: map[string]any{
					"reason": "workflow stopped before step was reached",
				},
			},
		}); err != nil {
			l.Warn("failed to mark step as not-attempted", zap.String("step_id", step.ID), zap.Error(err))
		}
	}
}

// isStepTerminal returns true if the step status is a terminal state.
func isStepTerminal(status app.Status) bool {
	switch status {
	case app.StatusSuccess, app.StatusAutoSkipped, app.StatusUserSkipped,
		app.StatusDiscarded, app.StatusCancelled, app.StatusError,
		app.StatusNotAttempted,
		app.WorkflowStepApprovalStatusApproved, app.WorkflowStepApprovalStatusApprovalDenied,
		app.WorkflowStepApprovalStatusApprovalExpired,
		app.WorkflowStepNoDrift, app.WorkflowStepDrifted:
		return true
	}
	return false
}

// stepConfig returns the StepConfig for this signal.
func (s *Signal) stepConfig() flow.StepConfig {
	return flow.StepConfig{
		GroupQueueName:         s.StepGroupQueueName,
		QueueName:              s.StepQueueName,
		TargetQueueName:        s.StepTargetQueueName,
		GenerateStepsQueueName: s.GenerateStepsQueueName,
		OwnerID:                s.OwnerID,
		OwnerType:              s.OwnerType,
		MW:                     s.tmw,
	}
}

// createRun creates a WorkflowRun record to track this execution segment.
func (s *Signal) createRun(ctx workflow.Context, runType app.WorkflowRunType, triggerStepID string, startFromIdx int) (*app.WorkflowRun, error) {
	return workflowactivities.AwaitPkgWorkflowsFlowCreateWorkflowRun(ctx, workflowactivities.CreateWorkflowRunRequest{
		WorkflowID:    s.WorkflowID,
		Type:          runType,
		TriggerStepID: triggerStepID,
		StartFromIdx:  startFromIdx,
	})
}

// updateRunStatus updates the status of a workflow run.
func (s *Signal) updateRunStatus(ctx workflow.Context, runID string, status app.Status) {
	workflowactivities.AwaitPkgWorkflowsFlowUpdateWorkflowRunStatus(ctx, workflowactivities.UpdateWorkflowRunStatusRequest{
		RunID: runID,
		Status: app.CompositeStatus{
			Status: status,
		},
	})
}

// isWorkflowComplete checks if all steps in the workflow have terminal statuses.
func (s *Signal) isWorkflowComplete(ctx workflow.Context) bool {
	steps, err := workflowactivities.AwaitPkgWorkflowsFlowGetFlowStepsByFlowID(ctx, s.WorkflowID)
	if err != nil {
		return false
	}

	for _, step := range steps {
		switch step.Status.Status {
		case app.StatusSuccess, app.StatusAutoSkipped, app.StatusUserSkipped,
			app.StatusDiscarded, app.StatusCancelled,
			app.WorkflowStepApprovalStatusApproved,
			app.WorkflowStepNoDrift, app.WorkflowStepDrifted:
			continue
		default:
			return false
		}
	}

	return true
}

// checkRetryable checks if the workflow is still eligible for retry.
func (s *Signal) checkRetryable(ctx workflow.Context) bool {
	resp, err := workflowactivities.AwaitCheckWorkflowRetryable(ctx, workflowactivities.CheckWorkflowRetryableRequest{
		WorkflowID: s.WorkflowID,
	})
	if err != nil {
		return false
	}
	return resp.Retryable
}

// groupStopReason returns the name and status text of the step that caused
// the group to stop. The step that writes the StepStop directive owns the
// user-facing phrasing; this is only a lookup.
func (s *Signal) groupStopReason(ctx workflow.Context, group *app.WorkflowStepGroup) (string, string) {
	steps, err := workflowactivities.AwaitPkgWorkflowsFlowGetFlowSteps(ctx, workflowactivities.GetFlowStepsRequest{
		FlowID: s.WorkflowID,
	})
	if err != nil {
		return "", ""
	}
	for i := range steps {
		step := &steps[i]
		if step.GroupIdx != group.GroupIdx || flowdirective.Step(step.ResultDirective) != flowdirective.StepStop {
			continue
		}
		return step.Name, step.Status.StatusHumanDescription
	}
	return "", ""
}

// checkGroupRetriesExhausted checks if any step in the group has retries_exhausted
// metadata, indicating the stop was caused by retry exhaustion.
func (s *Signal) checkGroupRetriesExhausted(ctx workflow.Context, group *app.WorkflowStepGroup) bool {
	steps, err := workflowactivities.AwaitPkgWorkflowsFlowGetFlowSteps(ctx, workflowactivities.GetFlowStepsRequest{
		FlowID: s.WorkflowID,
	})
	if err != nil {
		return false
	}
	for _, step := range steps {
		if step.GroupIdx == group.GroupIdx && step.Status.Status == app.StatusError {
			if v, ok := step.Status.Metadata["retries_exhausted"]; ok {
				if b, ok := v.(bool); ok && b {
					return true
				}
			}
		}
	}
	return false
}

// runnerDisabledCheckVersion gates the pre-group runner check so in-flight
// histories, which never scheduled the activity, still replay deterministically.
const runnerDisabledCheckVersion = "execute-flow-runner-disabled-check-v1"

// stopIfRunnerDisabled halts a workflow whose install runner was disabled after
// it started. Creation already rejects these, so without this the workflow would
// fail one runner-dependent group at a time and report each as its own error.
func (s *Signal) stopIfRunnerDisabled(
	ctx workflow.Context,
	l *zap.Logger,
	flw *app.Workflow,
	groups []app.WorkflowStepGroup,
	groupPosition int,
) (bool, error) {
	if workflow.GetVersion(ctx, runnerDisabledCheckVersion, workflow.DefaultVersion, 1) == workflow.DefaultVersion {
		return false, nil
	}

	if flw.OwnerType != "installs" || !flw.Type.RequiresInstallRunner() {
		return false, nil
	}

	disabled, err := workflowactivities.AwaitCheckFlowRunnerDisabled(ctx, workflowactivities.CheckFlowRunnerDisabledRequest{
		FlowID: s.WorkflowID,
	})
	if err != nil {
		// A failed check must not take down a workflow that would otherwise run.
		l.Warn("unable to check whether the install runner is disabled", zap.Error(err))
		return false, nil
	}
	if !disabled {
		return false, nil
	}

	l.Warn("install runner is disabled, stopping workflow",
		zap.String("workflow_id", s.WorkflowID),
		zap.Int("group_position", groupPosition))

	s.markRemainingGroupStepsDiscarded(ctx, l, groups, groupPosition-1)
	s.markRemainingStepsNotAttempted(ctx, l)

	if err := workflowactivities.AwaitPkgWorkflowsFlowUpdateFlowFinishedAtByID(ctx, s.WorkflowID); err != nil {
		l.Error("unable to update finished at", zap.Error(err))
	}

	return true, nil
}
