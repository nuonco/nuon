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

func (s *Signal) executeFlow(ctx workflow.Context) (retErr error) {
	s.executeStarted = true

	if s.mw != nil && s.v != nil {
		tmw, err := tmetrics.New(s.v, tmetrics.WithMetricsWriter(s.mw))
		if err == nil {
			s.tmw = tmw
		}
	}

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

	initialRunType := app.WorkflowRunTypeInitial
	initialStartIdx := 0
	initialStepID := ""
	initialScheduleState := residentScheduleRunnable
	if s.Resident {
		if s.resumeRequested {
			// why: A retry-step update raced ahead of the conductor during re-warm:
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

		// why: Only a resume requested while parked below is valid; drop stale ones.
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
				if exit, err := s.residentShouldExit(ctx); err != nil {
					return err
				} else if exit {
					return runErr
				}
			} else {
				s.updateRunStatus(ctx, run.ID, app.StatusError)

				if !s.checkRetryable(ctx) {
					if exit, err := s.residentShouldExit(ctx); err != nil {
						return err
					} else if exit {
						return runErr
					}
				} else {
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
				return nil
			}
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

		s.resumeRequested = false
		run, err = s.createRun(ctx, s.resumeRunType, s.resumeStepID, s.resumeStartIdx)
		if err != nil {
			return err
		}
	}
}

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
		return true, nil
	}
	_, pending := s.firstPendingGroupPosition(ctx)
	return !pending, nil
}

func (s *Signal) executeRun(ctx workflow.Context, run *app.WorkflowRun) error {
	cfg := s.stepConfig()
	startIdx := run.StartFromIdx

	for {
		err := s.handle(ctx, startIdx)
		if err == nil {
			return nil
		}

		if cerr, ok := err.(*flow.ContinueAsNewErr); ok && cerr != nil {
			startIdx = cerr.StartFromStepIdx
			continue
		}

		if _, ok := err.(*flow.ApprovalPauseErr); ok {
			if s.Resident {
				return err
			}
			return nil
		}

		if _, ok := err.(*flow.FlowStoppedErr); ok {
			return err
		}

		_ = cfg
		return err
	}
}

func (s *Signal) handle(ctx workflow.Context, startFromGroupIdx int) error {
	l, err := log.WorkflowLogger(ctx)
	if err != nil {
		return nil
	}

	flw, err := workflowactivities.AwaitPkgWorkflowsFlowGetFlowByID(ctx, s.WorkflowID)
	if err != nil {
		return errors.Wrap(err, "unable to get workflow object")
	}

	wfTags := []string{"workflow_type", s.WorkflowType, "org_id", s.OrgID}

	if flw.Status.Status == app.StatusCancelled {
		return errors.New("workflow already cancelled")
	}
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

		if s.tmw != nil {
			s.tmw.Incr(ctx, "workflow.started", wfTags...)
			s.tmw.Timing(ctx, "workflow.start_latency", workflow.Now(ctx).Sub(flw.CreatedAt), wfTags...)
		}
	}

	cfg := s.stepConfig()

	var eagerQueueSignalID string
	var eagerGroupCount int

	var completeDone workflow.Channel
	var completedFlw *app.Workflow
	var completeErr error

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

	stepGroups, _ := workflowactivities.AwaitPkgWorkflowsFlowGetFlowStepGroups(ctx, s.WorkflowID)

	var groups []app.WorkflowStepGroup
	if len(stepGroups) > 0 {
		groups = stepGroups
	} else {
		groupIdxs := collectGroupIndices(flw.Steps)
		for _, gIdx := range groupIdxs {
			groups = append(groups, app.WorkflowStepGroup{
				GroupIdx: gIdx,
				Parallel: isGroupParallel(flw.Steps, gIdx),
			})
		}
	}

	l.Debug("executing groups for workflow", zap.Int("group_count", len(groups)))

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

			// why: If cancellation was requested, preserve the cancelled status
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
			if s.pauseRequested {
				return &flow.ApprovalPauseErr{StepID: "paused"}
			}

		case flowdirective.GroupStop:
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
			if err := s.cloneGroupForRetry(ctx, group.GroupIdx); err != nil {
				s.markRemainingGroupStepsDiscarded(ctx, l, groups, gi)
				s.markRemainingStepsNotAttempted(ctx, l)
				if finErr := workflowactivities.AwaitPkgWorkflowsFlowUpdateFlowFinishedAtByID(ctx, s.WorkflowID); finErr != nil {
					l.Error("unable to update finished at", zap.Error(finErr))
				}
				stoppedErr := flow.NewFlowStoppedErr("", err.Error())
				stoppedErr.RetriesExhausted = true
				return stoppedErr
			}
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
			gi--
			continue

		case flowdirective.GroupSkipGroup:
			continue
		}

		if gi+1 == eagerGroupCount && completeDone != nil {
			l.Debug("waiting for parallel step generation to complete", zap.Int("eager_group_count", eagerGroupCount))
			completeDone.Receive(ctx, nil)
			completeDone = nil

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

			next := len(groups)
			for i := range groups {
				if !eagerExecuted[groups[i].GroupIdx] {
					next = i
					break
				}
			}
			gi = next - 1
		}

		if (gi+1-startFromGroupIdx) > 0 && (gi+1-startFromGroupIdx)%5 == 0 {
			return &flow.ContinueAsNewErr{StartFromStepIdx: gi + 1}
		}
	}

	// why: Resident hosts skip groups they consider non-pending; guard against a
	// scan that left a step non-terminal before stamping the flow finished.
	if s.Resident && !s.isWorkflowComplete(ctx) {
		return errors.Errorf("workflow %s is not complete after executing all groups", s.WorkflowID)
	}

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

func isGroupParallel(steps []app.WorkflowStep, groupIdx int) bool {
	for _, step := range steps {
		if step.GroupIdx == groupIdx && step.GroupParallel {
			return true
		}
	}
	return false
}

func collectGroupIndices(steps []app.WorkflowStep) []int {
	seen := make(map[int]bool)
	var groups []int
	for _, step := range steps {
		if !seen[step.GroupIdx] {
			seen[step.GroupIdx] = true
			groups = append(groups, step.GroupIdx)
		}
	}
	return groups
}

func (s *Signal) findGroupPositionForStep(ctx workflow.Context, stepID string) int {
	steps, err := workflowactivities.AwaitPkgWorkflowsFlowGetFlowSteps(ctx, workflowactivities.GetFlowStepsRequest{
		FlowID: s.WorkflowID,
	})
	if err != nil {
		return 0
	}

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

	groupIdxs := collectGroupIndices(steps)
	for i, gIdx := range groupIdxs {
		if gIdx == stepGroupIdx {
			return i
		}
	}
	return 0
}

func (s *Signal) markResumeRequested(ctx workflow.Context, runType app.WorkflowRunType, stepID string) {
	s.resumeRunType = runType
	s.resumeStepID = stepID
	s.resumeStartIdx = s.findGroupPositionForStep(ctx, stepID)
	s.resumeRequested = true
}

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
		if g.Status.Status == app.StatusFailedPendingRetry ||
			(g.Status.Status == app.StatusError && unresolved[g.GroupIdx]) {
			return residentScheduleDecision{State: residentScheduleBlocked}
		}
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

func (s *Signal) parkResident(ctx workflow.Context) (bool, error) {
	for {
		if s.cancelRequested {
			return false, nil
		}
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
			if s.updatesInFlight > 0 {
				continue
			}
			if pos, ok := s.firstPendingGroupPosition(ctx); ok {
				s.resumeStartIdx = pos
				return true, nil
			}
			if s.updatesInFlight > 0 || s.resumeRequested || s.appendRequested {
				continue
			}
			return false, nil
		}
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

	futureGroupIdxs := make(map[int]bool)
	for i := currentGroupPosition + 1; i < len(groups); i++ {
		futureGroupIdxs[groups[i].GroupIdx] = true

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

func (s *Signal) createRun(ctx workflow.Context, runType app.WorkflowRunType, triggerStepID string, startFromIdx int) (*app.WorkflowRun, error) {
	return workflowactivities.AwaitPkgWorkflowsFlowCreateWorkflowRun(ctx, workflowactivities.CreateWorkflowRunRequest{
		WorkflowID:    s.WorkflowID,
		Type:          runType,
		TriggerStepID: triggerStepID,
		StartFromIdx:  startFromIdx,
	})
}

func (s *Signal) updateRunStatus(ctx workflow.Context, runID string, status app.Status) {
	workflowactivities.AwaitPkgWorkflowsFlowUpdateWorkflowRunStatus(ctx, workflowactivities.UpdateWorkflowRunStatusRequest{
		RunID: runID,
		Status: app.CompositeStatus{
			Status: status,
		},
	})
}

func (s *Signal) isWorkflowComplete(ctx workflow.Context) bool {
	steps, err := workflowactivities.AwaitPkgWorkflowsFlowGetFlowStepsByFlowID(ctx, s.WorkflowID)
	if err != nil {
		return false
	}
	terminalErrorComplete := workflow.GetVersion(ctx, workflowCompleteTerminalErrorVersion, workflow.DefaultVersion, 1) != workflow.DefaultVersion

	for _, step := range steps {
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

// why: runnerDisabledCheckVersion gates the pre-group runner check so in-flight
// histories, which never scheduled the activity, still replay deterministically.
const runnerDisabledCheckVersion = "execute-flow-runner-disabled-check-v1"

// why: stackChangedRunnerGateVersion gates the stack-change deferral on that check.
// Histories that already recorded the activity must keep sending the same input.
const stackChangedRunnerGateVersion = "execute-flow-stack-changed-runner-gate-v1"

// why: flowCancelStatusVersion gates the cancelled-status writes added on the
// cancel-return paths; in-flight histories never scheduled those activities.
const flowCancelStatusVersion = "execute-flow-cancel-status-v1"

// why: groupStopReasonVersion gates the GetFlowSteps lookup that derives the stop
// reason; in-flight histories never scheduled it before the sweeps.
const groupStopReasonVersion = "execute-flow-group-stop-reason-v1"

// why: workflowCompleteTerminalErrorVersion gates terminal errored steps counting as
// complete because in-flight histories previously parked after every error.
const workflowCompleteTerminalErrorVersion = "execute-flow-terminal-error-complete-v1"

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
