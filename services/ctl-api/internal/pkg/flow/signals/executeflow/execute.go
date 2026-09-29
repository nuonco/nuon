package executeflow

import (
	"time"

	"go.temporal.io/sdk/workflow"
	"go.uber.org/zap"

	"github.com/pkg/errors"

	tmetrics "github.com/nuonco/nuon/pkg/temporal/metrics"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/compositeerrors"
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

type residentScheduleState string

const (
	residentScheduleComplete         residentScheduleState = "complete"
	residentScheduleBlocked          residentScheduleState = "blocked"
	residentScheduleAwaitingApproval residentScheduleState = "awaiting-approval"
	residentScheduleRunnable         residentScheduleState = "runnable"
)

type residentScheduleDecision struct {
	State    residentScheduleState
	Position int
}

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
	initialScheduleState := residentScheduleRunnable
	if s.Resident {
		if s.resumeRequested {
			// A retry-step update raced ahead of the conductor during re-warm:
			// honor its resume so the retry runs instead of being dropped.
			initialRunType = s.resumeRunType
			initialStartIdx = s.resumeStartIdx
			initialStepID = s.resumeStepID
			s.resumeRequested = false
			s.resumeRunType = ""
			s.resumeStepID = ""
		} else {
			decision := s.residentScheduleDecision(ctx)
			initialScheduleState = decision.State
			if pos, ok := residentInitialGroupPosition(decision); ok {
				initialStartIdx = pos
			}
		}
	}
	run, err := s.createRun(ctx, initialRunType, initialStepID, initialStartIdx)
	if err != nil {
		return err
	}

	for {
		var runErr error
		switch initialScheduleState {
		case residentScheduleBlocked:
			runErr = &flow.AwaitRetryPauseErr{}
		case residentScheduleAwaitingApproval:
			runErr = &flow.ApprovalPauseErr{}
		case residentScheduleRunnable:
			runErr = s.executeRun(ctx, run)
		}
		initialScheduleState = residentScheduleRunnable
		_, awaitingRetry := runErr.(*flow.AwaitRetryPauseErr)
		_, awaitingApproval := runErr.(*flow.ApprovalPauseErr)
		if awaitingRetry || awaitingApproval {
			runErr = nil
		}

		// Only a resume requested while parked below is valid; drop stale ones.
		// The exception is a run that unwound for await-retry: a retry-step
		// update that landed while the group was still live (e.g. retry-plan
		// on a parked approval) already cloned the step and is what caused
		// the unwind, so its resume must survive to dispatch the clone.
		if !awaitingRetry {
			s.resumeRequested = false
		}

		if runErr == nil {
			if s.cancelRequested {
				if workflow.GetVersion(ctx, flowCancelStatusVersion, workflow.DefaultVersion, 1) != workflow.DefaultVersion {
					s.updateRunStatus(ctx, run.ID, app.StatusCancelled)
					if s.Resident {
						if err := s.awaitResidentUpdates(ctx); err != nil {
							return err
						}
					}
					// Re-assert cancellation after the run unwinds: a cancel that
					// lands while executeRun is finishing can be overwritten by
					// its final success status write.
					s.writeFlowCancelled(ctx)
				}
				return nil
			}

			if awaitingRetry {
				s.updateRunStatus(ctx, run.ID, app.StatusFailedPendingRetry)
			} else if awaitingApproval {
				s.updateRunStatus(ctx, run.ID, app.AwaitingApproval)
			} else if s.isWorkflowComplete(ctx) {
				s.updateRunStatus(ctx, run.ID, app.StatusSuccess)
				if !s.Resident || (s.updatesInFlight == 0 && !s.appendRequested && !s.resumeRequested) {
					return nil
				}
			} else if !s.Resident {
				// Paused at approval - update run status and wait for resume.
				// Resident hosts skip this: a non-complete state just means a
				// freshly appended step is pending, which parkResident runs next.
				s.updateRunStatus(ctx, run.ID, app.AwaitingApproval)
			}
		} else {
			if s.cancelRequested {
				if workflow.GetVersion(ctx, flowCancelStatusVersion, workflow.DefaultVersion, 1) != workflow.DefaultVersion {
					s.updateRunStatus(ctx, run.ID, app.StatusCancelled)
					if s.Resident {
						if err := s.awaitResidentUpdates(ctx); err != nil {
							return err
						}
					}
					s.writeFlowCancelled(ctx)
				}
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
				if stoppedErr.StepID != "" {
					metadata["step_name"] = stoppedErr.StepID
				}
				if stoppedErr.Reason != "" {
					metadata["stop_reason"] = stoppedErr.Reason
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
						CompositeError:         stoppedErr.CompositeError,
						Metadata:               metadata,
					},
				})
				// A stop is terminal: the host exits now so the Temporal
				// workflow closes and the queue signal errors like the flow.
				// It only stays up when a step is already pending (e.g. a
				// clone from a retry that raced the stop).
				if exit, err := s.residentShouldExit(ctx); err != nil {
					return err
				} else if exit {
					return runErr
				}
			} else {
				// Actual execution error
				s.updateRunStatus(ctx, run.ID, app.StatusError)

				if !s.checkRetryable(ctx) {
					if exit, err := s.residentShouldExit(ctx); err != nil {
						return err
					} else if exit {
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
				if err := s.awaitResidentUpdates(ctx); err != nil {
					return err
				}
				s.writeFlowCancelled(ctx)
				return runErr
			}
			if !parked {
				// Idle timeout with nothing pending — exit cleanly.
				return nil
			}
			// parkResident set resumeStartIdx to the first pending group.
			resumeRunType := s.resumeRunType
			if resumeRunType == "" {
				resumeRunType = app.WorkflowRunTypeResume
			}
			resumeStepID := s.resumeStepID
			s.resumeRunType = ""
			s.resumeStepID = ""
			run, err = s.createRun(ctx, resumeRunType, resumeStepID, s.resumeStartIdx)
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
			s.writeFlowCancelled(ctx)
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

// residentShouldExit reports whether a host may return after a terminal run:
// legacy hosts always do; a resident host does once no update is in flight,
// none seeded a resume, and no group is pending. A later retry re-warms the
// host through update-with-start.
func (s *Signal) residentShouldExit(ctx workflow.Context) (bool, error) {
	if !s.Resident {
		return true, nil
	}
	if err := s.awaitResidentUpdates(ctx); err != nil {
		return false, err
	}
	if s.resumeRequested || s.appendRequested {
		return false, nil
	}
	groups, err := workflowactivities.AwaitPkgWorkflowsFlowGetFlowStepGroups(ctx, s.WorkflowID)
	if err != nil || len(groups) == 0 {
		// Nothing was generated, so nothing can be retried or skipped;
		// parking would only re-run the same failure.
		return true, nil
	}
	_, pending := s.firstPendingGroupPosition(ctx)
	return !pending, nil
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

		// ApprovalPauseErr means we stopped at an approval or pause. Legacy
		// hosts enter the wait loop; resident hosts surface it so the run is
		// recorded as awaiting-approval before parking.
		if _, ok := err.(*flow.ApprovalPauseErr); ok {
			if s.Resident {
				return err
			}
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

	publishAllStepsOnGeneration := workflow.GetVersion(
		ctx, allStepsLoadedOnGenerationVersion, workflow.DefaultVersion, 1,
	) != workflow.DefaultVersion

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
			missingStepsErr := errors.Errorf("workflow %s has no steps and no generate-steps signal", s.WorkflowID)
			_ = statusactivities.AwaitUpdateFlowStatusMetadata(ctx, statusactivities.UpdateFlowStatusMetadataRequest{
				WorkflowID: s.WorkflowID,
				Metadata: map[string]any{
					"error_message": missingStepsErr.Error(),
				},
			})
			_ = statusactivities.AwaitPkgStatusUpdateFlowStatus(ctx, statusactivities.UpdateStatusRequest{
				ID: s.WorkflowID,
				Status: app.CompositeStatus{
					Status:                 app.StatusError,
					StatusHumanDescription: missingStepsErr.Error(),
				},
			})
			return missingStepsErr
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
				if completeErr == nil && publishAllStepsOnGeneration {
					s.markAllStepsLoaded(gCtx, l)
				}
				completeDone.Send(gCtx, true)
			})
		} else if publishAllStepsOnGeneration {
			s.markAllStepsLoaded(ctx, l)
		}
	} else {
		l.Debug("steps already exist, skipping generation", zap.Int("step_count", len(flw.Steps)))
		if publishAllStepsOnGeneration {
			s.markAllStepsLoaded(ctx, l)
		}
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

		if s.Resident && group.Status.Status == app.StatusFailedPendingRetry && !residentPending[group.GroupIdx] {
			return &flow.AwaitRetryPauseErr{}
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
			stepName, reason := "", ""
			var stepCE *compositeerrors.CompositeErrorData
			if workflow.GetVersion(ctx, groupStopReasonVersion, workflow.DefaultVersion, 1) != workflow.DefaultVersion {
				stepName, reason, stepCE = s.groupStopReason(ctx, group)
			}

			s.markRemainingGroupStepsDiscarded(ctx, l, groups, gi)
			s.markRemainingStepsNotAttempted(ctx, l)
			if err := workflowactivities.AwaitPkgWorkflowsFlowUpdateFlowFinishedAtByID(ctx, s.WorkflowID); err != nil {
				l.Error("unable to update finished at", zap.Error(err))
			}
			stoppedErr := flow.NewFlowStoppedErr(stepName, reason)
			stoppedErr.CompositeError = stepCE
			if reason != "" {
				stoppedErr.StatusHumanDescription = "workflow stopped: " + reason
			}
			stoppedErr.RetriesExhausted = s.checkGroupRetriesExhausted(ctx, group)
			return stoppedErr

		case flowdirective.GroupAwaitApproval:
			return flow.NewApprovalPauseErr("")

		case flowdirective.GroupAwaitRetry:
			if s.Resident {
				return &flow.AwaitRetryPauseErr{}
			}

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
			if s.Resident {
				residentPending = make(map[int]bool)
				for _, st := range flw.Steps {
					if !isStepTerminal(st.Status.Status) {
						residentPending[st.GroupIdx] = true
					}
				}
			}

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

			if !publishAllStepsOnGeneration {
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
			}

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

	// Resident hosts skip groups they consider non-pending; guard against a
	// scan that left a step non-terminal before stamping the flow finished.
	if s.Resident && !s.isWorkflowComplete(ctx) {
		return errors.Errorf("workflow %s is not complete after executing all groups", s.WorkflowID)
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

// markResumeRequested arms the parked Execute loop to resume the run at the
// group containing stepID. Call it last in an update handler: the paused loop
// acts on the flag the instant it flips, and reads the fields written just
// above it. The DB lookups in a handler pause it long enough for Execute to
// run, so setting the flag first means resuming from a stale resumeStartIdx.
func (s *Signal) markResumeRequested(ctx workflow.Context, runType app.WorkflowRunType, stepID string) {
	s.resumeRunType = runType
	s.resumeStepID = stepID
	s.resumeStartIdx = s.findGroupPositionForStep(ctx, stepID)
	s.resumeRequested = true
}

// firstPendingGroupPosition returns the position (index into the ordered group
// slice) of the first group that still has a non-terminal step, and whether one
// exists. Resident hosts use it to (a) start a rewarmed run at the first
// unfinished group instead of replaying completed history, and (b) decide
// whether to run or park after each step. Terminal-but-failed groups (e.g. a
// appended step that errored) are skipped, so one failed step never wedges the
// host.
func (s *Signal) residentScheduleDecision(ctx workflow.Context) residentScheduleDecision {
	groups, err := workflowactivities.AwaitPkgWorkflowsFlowGetFlowStepGroups(ctx, s.WorkflowID)
	if err != nil || len(groups) == 0 {
		return residentScheduleDecision{State: residentScheduleRunnable}
	}
	steps, err := workflowactivities.AwaitPkgWorkflowsFlowGetFlowSteps(ctx, workflowactivities.GetFlowStepsRequest{
		FlowID: s.WorkflowID,
	})
	if err != nil {
		return residentScheduleDecision{State: residentScheduleComplete}
	}

	pending := make(map[int]bool)
	unresolved := make(map[int]bool)
	unanswered := make(map[int]bool)
	for _, st := range steps {
		if !isStepTerminal(st.Status.Status) {
			pending[st.GroupIdx] = true
		}
		if st.Status.Status == app.StatusError && !st.Retried {
			unresolved[st.GroupIdx] = true
		}
		if st.Status.Status == app.AwaitingApproval && (st.Approval == nil || st.Approval.Response == nil) {
			unanswered[st.GroupIdx] = true
		}
	}
	for pos, g := range groups {
		// A group that failed without a directive (no stop sweep ran) leaves
		// downstream groups pending; they must wait for a retry or skip of the
		// errored step rather than run past it.
		if g.Status.Status == app.StatusFailedPendingRetry ||
			(g.Status.Status == app.StatusError && unresolved[g.GroupIdx]) {
			return residentScheduleDecision{State: residentScheduleBlocked}
		}
		// A parked approval holds its group until a response is persisted;
		// the group then re-dispatches the step to apply it.
		if unanswered[g.GroupIdx] {
			return residentScheduleDecision{State: residentScheduleAwaitingApproval}
		}
		if pending[g.GroupIdx] {
			return residentScheduleDecision{State: residentScheduleRunnable, Position: pos}
		}
	}
	return residentScheduleDecision{State: residentScheduleComplete}
}

func (s *Signal) firstPendingGroupPosition(ctx workflow.Context) (int, bool) {
	decision := s.residentScheduleDecision(ctx)
	return decision.Position, decision.State == residentScheduleRunnable
}

func residentInitialGroupPosition(decision residentScheduleDecision) (int, bool) {
	if decision.State == residentScheduleRunnable {
		return decision.Position, true
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
		// A retry/skip update in flight has already superseded the failed
		// step but may not have written its clone yet; scanning inside that
		// window would see no failure and run the downstream groups.
		if err := s.awaitResidentUpdates(ctx); err != nil {
			return false, err
		}
		if pos, ok := s.firstPendingGroupPosition(ctx); ok {
			s.resumeStartIdx = pos
			return true, nil
		}

		s.awaitingResume = true
		woke, err := workflow.AwaitWithTimeout(ctx, s.residentIdleTimeout(), func() bool {
			return s.resumeRequested || s.appendRequested || s.cancelRequested
		})
		s.awaitingResume = false
		if woke && s.resumeRequested {
			s.resumeRequested = false
			s.appendRequested = false
			return true, nil
		}
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

func (s *Signal) residentIdleTimeout() time.Duration {
	if s.ResidentIdleTimeout > 0 {
		return s.ResidentIdleTimeout
	}
	return residentIdleTimeout
}

func (s *Signal) awaitResidentUpdates(ctx workflow.Context) error {
	return workflow.Await(ctx, func() bool { return s.updatesInFlight == 0 })
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
	terminalErrorComplete := workflow.GetVersion(ctx, workflowCompleteTerminalErrorVersion, workflow.DefaultVersion, 1) != workflow.DefaultVersion

	for _, step := range steps {
		// Superseded steps keep their original status (e.g. error) for
		// dashboard display, but a clone has taken their place — they must
		// not block workflow completion.
		if step.Retried {
			continue
		}
		switch step.Status.Status {
		case app.StatusSuccess, app.StatusAutoSkipped, app.StatusUserSkipped,
			app.StatusDiscarded, app.StatusCancelled,
			app.WorkflowStepApprovalStatusApproved,
			app.WorkflowStepNoDrift, app.WorkflowStepDrifted:
			continue
		case app.StatusError:
			if !terminalErrorComplete {
				return false
			}
			// Treat a settled failure (terminal directive) as complete so
			// failures do not leak forever-open workflows: the group already
			// acted on it, so nothing will resume this run. A parked error
			// (await-retry, await-approval, or a legacy empty directive)
			// still waits on a user decision and is not complete.
			if flowdirective.Step(step.ResultDirective).IsTerminal() {
				continue
			}
			return false
		default:
			return false
		}
	}

	return true
}

// writeFlowCancelled update's workflow's status to cancelled
func (s *Signal) writeFlowCancelled(ctx workflow.Context) {
	if workflow.GetVersion(ctx, flowCancelStatusVersion, workflow.DefaultVersion, 1) == workflow.DefaultVersion {
		return
	}
	l, _ := log.WorkflowLogger(ctx)
	if err := statusactivities.AwaitPkgStatusUpdateFlowStatus(ctx, statusactivities.UpdateStatusRequest{
		ID: s.WorkflowID,
		Status: app.CompositeStatus{
			Status:                 app.StatusCancelled,
			StatusHumanDescription: "workflow cancelled",
			Metadata: map[string]any{
				"cancel_requested_at": workflow.Now(ctx).Unix(),
			},
		},
	}); err != nil && l != nil {
		l.Error("unable to re-assert cancelled workflow status", zap.Error(err))
	}
}

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
func (s *Signal) groupStopReason(ctx workflow.Context, group *app.WorkflowStepGroup) (string, string, *compositeerrors.CompositeErrorData) {
	steps, err := workflowactivities.AwaitPkgWorkflowsFlowGetFlowSteps(ctx, workflowactivities.GetFlowStepsRequest{
		FlowID: s.WorkflowID,
	})
	if err != nil {
		return "", "", nil
	}
	for i := range steps {
		step := &steps[i]
		if step.GroupIdx != group.GroupIdx || flowdirective.Step(step.ResultDirective) != flowdirective.StepStop {
			continue
		}
		return step.Name, stepStopReason(step), step.Status.CompositeError
	}
	return "", "", nil
}

func stepStopReason(step *app.WorkflowStep) string {
	if original, ok := step.Status.Metadata["original_error"].(string); ok && original != "" {
		return original
	}
	return step.Status.StatusHumanDescription
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

// stackChangedRunnerGateVersion gates the stack-change deferral on that check.
// Histories that already recorded the activity must keep sending the same input.
const stackChangedRunnerGateVersion = "execute-flow-stack-changed-runner-gate-v1"

// flowCancelStatusVersion gates the cancelled-status writes added on the
// cancel-return paths; in-flight histories never scheduled those activities.
const flowCancelStatusVersion = "execute-flow-cancel-status-v1"

// groupStopReasonVersion gates the GetFlowSteps lookup that derives the stop
// reason; in-flight histories never scheduled it before the sweeps.
const groupStopReasonVersion = "execute-flow-group-stop-reason-v1"

// workflowCompleteTerminalErrorVersion gates terminal errored steps counting as
// complete because in-flight histories previously parked after every error.
const workflowCompleteTerminalErrorVersion = "execute-flow-terminal-error-complete-v1"

const allStepsLoadedOnGenerationVersion = "execute-flow-all-steps-loaded-on-generation-v1"

func (s *Signal) markAllStepsLoaded(ctx workflow.Context, l *zap.Logger) {
	if err := statusactivities.AwaitUpdateFlowStatusMetadata(ctx, statusactivities.UpdateFlowStatusMetadataRequest{
		WorkflowID: s.WorkflowID,
		Metadata: map[string]any{
			"all_steps_loaded": true,
		},
	}); err != nil {
		l.Error("unable to mark all steps loaded", zap.Error(err))
	}
}

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

	req := workflowactivities.CheckFlowRunnerDisabledRequest{FlowID: s.WorkflowID}
	if workflow.GetVersion(ctx, stackChangedRunnerGateVersion, workflow.DefaultVersion, 1) != workflow.DefaultVersion {
		req.HonorStackChanged = true
	}
	disabled, err := workflowactivities.AwaitCheckFlowRunnerDisabled(ctx, req)
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
