package executeworkflowstep

import (
	"context"
	"fmt"
	"regexp"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/pkg/errors"
	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"
	"go.uber.org/zap"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	policyhelpers "github.com/nuonco/nuon/services/ctl-api/internal/app/policy_reports/helpers"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/log"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/client"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/signal"
	sharedactivities "github.com/nuonco/nuon/services/ctl-api/internal/pkg/workflows/activities"
	statusactivities "github.com/nuonco/nuon/services/ctl-api/internal/pkg/workflows/status/activities"
	workflowsflow "github.com/nuonco/nuon/services/ctl-api/internal/pkg/workflows/workflow"
	activities "github.com/nuonco/nuon/services/ctl-api/internal/pkg/workflows/workflow/activities"
)

const SignalType signal.SignalType = "execute-workflow-step"

// Policy evaluation metadata keys (mirrored from flow package to avoid import cycle)
const (
	DenyViolationsKey  = "deny_violations"
	WarnViolationsKey  = "warn_violations"
	PassedPolicyIDsKey = "passed_policy_ids"
)

// Signal encapsulates the full lifecycle of executing a single workflow step.
// When dispatched to the step queue (e.g. install-workflow-steps), it fetches all
// necessary state from the database and runs the complete step execution flow:
// status updates, inner signal dispatch, approval handling, policy checks, and
// noop plan detection.
type Signal struct {
	StepID     string `json:"step_id"`
	WorkflowID string `json:"workflow_id"`

	// OwnerID is the entity that owns the queues (e.g. install ID).
	OwnerID   string `json:"owner_id"`
	OwnerType string `json:"owner_type"`

	// TargetQueueName is the queue where the inner signal (the actual step signal)
	// gets enqueued for execution (e.g. "install-signals").
	TargetQueueName string `json:"target_queue_name"`

	// innerQueueSignalID tracks the currently executing inner signal so that
	// Cancel() can propagate cancellation to it. Set during executeInnerSignal,
	// read during Cancel. Safe because Temporal workflows are single-threaded.
	innerQueueSignalID string
}

var _ signal.Signal = (*Signal)(nil)
var _ signal.SignalWithCancel = (*Signal)(nil)

func (s *Signal) Type() signal.SignalType {
	return SignalType
}

func (s *Signal) Validate(ctx workflow.Context) error {
	if s.StepID == "" {
		return errors.New("step_id is required")
	}
	if s.WorkflowID == "" {
		return errors.New("workflow_id is required")
	}
	if s.OwnerID == "" {
		return errors.New("owner_id is required")
	}
	if s.OwnerType == "" {
		return errors.New("owner_type is required")
	}
	if s.TargetQueueName == "" {
		return errors.New("target_queue_name is required")
	}
	return nil
}

// Cancel propagates cancellation to the inner signal if one is currently executing.
func (s *Signal) Cancel(ctx workflow.Context) error {
	if s.innerQueueSignalID == "" {
		return nil
	}

	cancelCtx, cancelCtxCancel := workflow.NewDisconnectedContext(ctx)
	defer cancelCtxCancel()

	// Cancel the inner signal (e.g. the actual step signal on install-signals queue)
	_, err := client.AwaitCancelSignal(cancelCtx, s.innerQueueSignalID)
	if err != nil {
		l, _ := log.WorkflowLogger(cancelCtx)
		if l != nil {
			l.Warn("failed to cancel inner signal",
				zap.String("inner_queue_signal_id", s.innerQueueSignalID),
				zap.Error(err))
		}
	}

	// Mark the step as cancelled
	statusactivities.AwaitPkgStatusUpdateFlowStepStatus(cancelCtx, statusactivities.UpdateStatusRequest{
		ID: s.StepID,
		Status: app.CompositeStatus{
			Status: app.StatusCancelled,
		},
	})

	return err
}

// Execute runs the full workflow step lifecycle. This replicates the logic from
// WorkflowConductor.executeFlowStep but as a self-contained signal that fetches
// its own state from the database.
func (s *Signal) Execute(ctx workflow.Context) error {
	// Yield to allow cancellation signals to be processed before starting work.
	if err := workflow.Sleep(ctx, 0); err != nil {
		return err
	}

	l, err := log.WorkflowLogger(ctx)
	if err != nil {
		return errors.Wrap(err, "unable to get workflow logger")
	}

	// Fetch step and workflow from the database
	step, err := activities.AwaitPkgWorkflowsFlowGetFlowsStepByFlowStepID(ctx, s.StepID)
	if err != nil {
		return errors.Wrap(err, "unable to get step")
	}

	flw, err := activities.AwaitPkgWorkflowsFlowGetFlowByID(ctx, s.WorkflowID)
	if err != nil {
		return errors.Wrap(err, "unable to get workflow")
	}

	// Check if step is in executable state
	if step.Status.Status != app.StatusPending && step.Status.Status != app.StatusNotAttempted && step.Status.Status != app.StatusQueued {
		l.Debug("step not in executable state, exiting",
			zap.String("step_id", step.ID),
			zap.String("step_status", string(step.Status.Status)),
			zap.String("workflow_id", flw.ID))
		return nil
	}

	defer func() {
		if err := activities.AwaitPkgWorkflowsFlowUpdateFlowStepFinishedAtByID(ctx, step.ID); err != nil {
			l.Error("unable to update finished at", zap.Error(err))
		}
	}()

	// Update flow status to in-progress
	if err := statusactivities.AwaitPkgStatusUpdateFlowStatus(ctx, statusactivities.UpdateStatusRequest{
		ID: flw.ID,
		Status: app.CompositeStatus{
			Status:                 app.StatusInProgress,
			StatusHumanDescription: "executing step " + strconv.Itoa(step.Idx+1),
			Metadata:               map[string]any{},
		},
	}); err != nil {
		return errors.Wrap(err, "unable to update step")
	}

	// Execute the inner signal (equivalent to executeStep)
	stepErr := s.executeInnerSignal(ctx, step)
	if stepErr != nil {
		if err := statusactivities.AwaitPkgStatusUpdateFlowStepStatus(ctx, statusactivities.UpdateStatusRequest{
			ID: step.ID,
			Status: app.CompositeStatus{
				Status: app.StatusError,
				Metadata: map[string]any{
					"reason": "Step failed, review the error in logs and try again.",
				},
				StatusHumanDescription: "Step failed",
			},
		}); err != nil {
			return errors.Wrap(err, "unable to mark step as error")
		}

		return stepErr
	}

	// Refetch the step after signal execution to gather new state (e.g. step target ID)
	step, err = activities.AwaitPkgWorkflowsFlowGetFlowsStepByFlowStepID(ctx, step.ID)
	if err != nil {
		return errors.Wrap(err, "unable to get step")
	}

	// Non-approval steps: mark success and return
	if step.ExecutionType != app.WorkflowStepExecutionTypeApproval {
		l.Debug("step type non approval, step successful",
			zap.String("step_id", step.ID),
			zap.String("workflow_id", flw.ID))
		if err := statusactivities.AwaitPkgStatusUpdateFlowStepStatus(ctx, statusactivities.UpdateStatusRequest{
			ID: step.ID,
			Status: app.CompositeStatus{
				Status: app.StatusSuccess,
			},
		}); err != nil {
			return errors.Wrap(err, "unable to mark step as success")
		}

		if err := statusactivities.AwaitPkgStatusUpdateFlowStatus(ctx, statusactivities.UpdateStatusRequest{
			ID: flw.ID,
			Status: app.CompositeStatus{
				Status:                 app.StatusSuccess,
				StatusHumanDescription: "finished executing step " + strconv.Itoa(step.Idx+1),
				Metadata: map[string]any{
					"step_idx": step.Idx,
					"status":   "ok",
				},
			},
		}); err != nil {
			return errors.Wrap(err, "unable to update step to success status")
		}

		return nil
	}

	// --- Approval flow ---

	l.Debug("looking up approval contents",
		zap.String("step_id", step.ID),
		zap.String("workflow_id", flw.ID))
	if err := statusactivities.AwaitPkgStatusUpdateFlowStepStatus(ctx, statusactivities.UpdateStatusRequest{
		ID: step.ID,
		Status: app.CompositeStatus{
			Status:                 app.StatusCheckPlan,
			StatusHumanDescription: "checking plan for changes",
			Metadata: map[string]any{
				"status": "checking plan for changes",
			},
		},
	}); err != nil {
		return errors.Wrap(err, "unable to mark step status as checking plan")
	}

	noopPlan, err := activities.AwaitCheckNoopPlan(ctx, &activities.CheckNoopPlanRequest{
		StepTargetID: step.StepTargetID,
	})
	if err != nil {
		if err := statusactivities.AwaitPkgStatusUpdateFlowStepStatus(ctx, statusactivities.UpdateStatusRequest{
			ID: step.ID,
			Status: app.CompositeStatus{
				Status: app.StatusError,
				Metadata: map[string]any{
					"reason": "Step failed, failed to check for noop plan.",
				},
				StatusHumanDescription: "Step failed",
			},
		}); err != nil {
			return errors.Wrap(err, "unable to mark step as error")
		}
		return errors.Wrap(err, "failed to check for noop plan")
	}

	// Noop plan: auto-skip this step and the next apply step
	if noopPlan {
		l.Debug("approval plan contents empty",
			zap.String("step_id", step.ID),
			zap.String("workflow_id", flw.ID))
		if err := s.handleNoopDeployPlan(ctx, step, flw); err != nil {
			if err := statusactivities.AwaitPkgStatusUpdateFlowStepStatus(ctx, statusactivities.UpdateStatusRequest{
				ID: step.ID,
				Status: app.CompositeStatus{
					Status: app.StatusError,
					Metadata: map[string]any{
						"reason": "Step failed, unable to handle noop plan.",
					},
					StatusHumanDescription: "Step failed",
				},
			}); err != nil {
				return errors.Wrap(err, "unable to mark step as error")
			}
			return errors.Wrap(err, "failed to handle noop plan")
		}

		if !flw.PlanOnly {
			return nil
		}
	}

	// Check policies
	l.Debug("starting policy check",
		zap.String("step_id", step.ID),
		zap.String("step_target_id", step.StepTargetID),
		zap.String("step_target_type", step.StepTargetType),
		zap.String("workflow_id", flw.ID))

	violations, policyContext, policyErr := s.checkPolicies(ctx, step.StepTargetID, step.StepTargetType)
	if policyErr != nil {
		l.Warn("failed to check policies",
			zap.String("step_id", step.ID),
			zap.String("step_target_id", step.StepTargetID),
			zap.String("step_target_type", step.StepTargetType),
			zap.String("workflow_id", flw.ID),
			zap.Error(policyErr))
	}

	if policyContext != nil {
		policyInputCounts := make(map[string]int, len(policyContext.PolicyIDs))
		for _, policyID := range policyContext.PolicyIDs {
			policyInputCounts[policyID] = policyContext.InputCount
		}
		var validationID *string
		if step.PolicyValidation != nil {
			validationID = &step.PolicyValidation.ID
		}
		reportResult, err := activities.AwaitPersistPolicyReport(ctx, &activities.PersistPolicyReportRequest{
			OrgID:                          policyContext.OrgID,
			AppID:                          policyContext.AppID,
			InstallID:                      policyContext.InstallID,
			InstallSandboxID:               policyContext.InstallSandboxID,
			ComponentID:                    policyContext.ComponentID,
			ComponentBuildID:               policyContext.ComponentBuildID,
			WorkflowStepPolicyValidationID: validationID,
			OwnerID:                        step.StepTargetID,
			OwnerType:                      step.StepTargetType,
			Violations:                     violations,
			PolicyIDs:                      policyContext.PolicyIDs,
			PolicyInputCounts:              policyInputCounts,
			OrgName:                        policyContext.OrgName,
			AppName:                        policyContext.AppName,
			InstallName:                    policyContext.InstallName,
			ComponentName:                  policyContext.ComponentName,
		})
		if err != nil {
			l.Warn("failed to persist policy report", zap.Error(err))
		}

		var passedPolicyIDs []string
		if reportResult != nil {
			passedPolicyIDs = reportResult.PassedPolicyIDs
		}
		if err := s.processPolicyViolations(ctx, l, step, flw, violations, passedPolicyIDs); err != nil {
			return errors.Wrap(err, "unable to process check for policy violation")
		}
	}

	l.Debug("policy check completed successfully",
		zap.String("step_id", step.ID),
		zap.String("step_target_id", step.StepTargetID),
		zap.String("step_target_type", step.StepTargetType),
		zap.String("workflow_id", flw.ID))

	// Auto approve if plan-only mode
	if flw.PlanOnly {
		if err := s.handlePlanOnlyApproval(ctx, step, noopPlan); err != nil {
			if err := statusactivities.AwaitPkgStatusUpdateFlowStepStatus(ctx, statusactivities.UpdateStatusRequest{
				ID: step.ID,
				Status: app.CompositeStatus{
					Status: app.StatusError,
					Metadata: map[string]any{
						"reason": "Step failed, unable to handle plan-only auto-approval.",
					},
					StatusHumanDescription: "Step failed",
				},
			}); err != nil {
				return errors.Wrap(err, "unable to mark step as error")
			}
			return errors.Wrap(err, "failed to handle plan-only auto-approval")
		}
		return nil
	}

	// Update status to awaiting approval
	if err := statusactivities.AwaitPkgStatusUpdateFlowStepStatus(ctx, statusactivities.UpdateStatusRequest{
		ID: step.ID,
		Status: app.CompositeStatus{
			Status:                 app.AwaitingApproval,
			StatusHumanDescription: "awaiting approval " + strconv.Itoa(step.Idx+1),
			Metadata: map[string]any{
				"step_idx": step.Idx,
				"status":   "ok",
			},
		},
	}); err != nil {
		return errors.Wrap(err, "unable to update step to success status")
	}

	// Wait for approval via V2 child workflow
	resp, err := s.waitForApprovalResponse(ctx, flw, step)
	if err != nil {
		return err
	}

	switch resp.Type {
	case app.WorkflowStepApprovalResponseTypeApprove:
		l.Debug("handling approval response type: approved",
			zap.String("step_id", step.ID),
			zap.String("workflow_id", flw.ID))
		if err := statusactivities.AwaitPkgStatusUpdateFlowStepStatus(ctx, statusactivities.UpdateStatusRequest{
			ID: step.ID,
			Status: app.CompositeStatus{
				Status:                 app.WorkflowStepApprovalStatusApproved,
				StatusHumanDescription: "approved " + strconv.Itoa(step.Idx+1),
				Metadata: map[string]any{
					"step_idx": step.Idx,
					"status":   "ok",
				},
			},
		}); err != nil {
			return errors.Wrap(err, "unable to update step to success status")
		}
		return nil

	case app.WorkflowStepApprovalResponseTypeRetryPlan:
		l.Debug("handling approval response type: retry plan",
			zap.String("step_id", step.ID),
			zap.String("workflow_id", flw.ID))

		err := s.cloneWorkflowStep(ctx, step, flw)
		if err != nil {
			return errors.Wrap(err, "unable to clone step for retry plan")
		}

		if err := statusactivities.AwaitPkgStatusUpdateFlowStepStatus(ctx, statusactivities.UpdateStatusRequest{
			ID: step.ID,
			Status: app.CompositeStatus{
				Status:                 app.WorkflowStepApprovalStatusApprovalRetryPlan,
				StatusHumanDescription: "retrying " + strconv.Itoa(step.Idx),
				Metadata: map[string]any{
					"step_idx": step.Idx,
					"status":   "retrying",
				},
			},
		}); err != nil {
			return errors.Wrap(err, "unable to update step to retry plan status")
		}

		if err := activities.AwaitPkgWorkflowsFlowUpdateFlowStepTargetStatus(ctx, activities.UpdateFlowStepTargetStatusRequest{
			StepID:            step.ID,
			Status:            app.StatusDiscarded,
			StatusDescription: "Retrying step " + strconv.Itoa(step.Idx),
		}); err != nil {
			return errors.Wrap(err, "unable to update step target status")
		}
		return nil

	case app.WorkflowStepApprovalResponseTypeSkipCurrent:
		l.Debug("handling approval response type: skip current and continue",
			zap.String("step_id", step.ID),
			zap.String("workflow_id", flw.ID))

		if err := s.markWorkflowApprovalPlanDenied(ctx, flw, step); err != nil {
			statusactivities.AwaitPkgStatusUpdateFlowStatus(ctx, statusactivities.UpdateStatusRequest{
				ID: flw.ID,
				Status: app.CompositeStatus{
					Status:                 app.StatusError,
					StatusHumanDescription: "failed to deny plan and update step status",
					Metadata:               map[string]any{},
				},
			})
		}
		return nil

	case app.WorkflowStepApprovalResponseTypeSkipCurrentAndDependents:
		l.Debug("handling approval response type: skip current and dependents",
			zap.String("step_id", step.ID),
			zap.String("workflow_id", flw.ID))

		if err := s.markDependentStepsAsSkipped(ctx, flw, step); err != nil {
			statusactivities.AwaitPkgStatusUpdateFlowStatus(ctx, statusactivities.UpdateStatusRequest{
				ID: flw.ID,
				Status: app.CompositeStatus{
					Status:                 app.StatusError,
					StatusHumanDescription: "failed to deny plan and update step status",
					Metadata:               map[string]any{},
				},
			})
		}
		return nil
	}

	// Default: approval denied
	if err := statusactivities.AwaitPkgStatusUpdateFlowStepStatus(ctx, statusactivities.UpdateStatusRequest{
		ID: step.ID,
		Status: app.NewCompositeTemporalStatus(ctx, app.WorkflowStepApprovalStatusApprovalDenied, map[string]any{
			"reason": "approval denied",
		}),
	}); err != nil {
		return errors.Wrap(err, "unable to update")
	}
	if err := activities.AwaitPkgWorkflowsFlowUpdateFlowStepTargetStatus(ctx, activities.UpdateFlowStepTargetStatusRequest{
		StepID:            step.ID,
		Status:            app.Status(app.InstallDeployApprovalDenied),
		StatusDescription: "Approval denied",
	}); err != nil {
		return errors.Wrap(err, "unable to update step target status")
	}

	return fmt.Errorf("not approved")
}

// executeInnerSignal handles the actual signal dispatch for a step.
// It updates step status, then enqueues the inner signal to the install-signals
// queue and awaits completion.
func (s *Signal) executeInnerSignal(ctx workflow.Context, step *app.WorkflowStep) error {
	if err := activities.AwaitPkgWorkflowsFlowUpdateFlowStepStartedAtByID(ctx, step.ID); err != nil {
		return err
	}

	if err := statusactivities.AwaitPkgStatusUpdateFlowStepStatus(ctx, statusactivities.UpdateStatusRequest{
		ID: step.ID,
		Status: app.CompositeStatus{
			Status: app.StatusInProgress,
		},
	}); err != nil {
		return err
	}

	if step.ExecutionType == app.WorkflowStepExecutionTypeSkipped {
		if err := statusactivities.AwaitPkgStatusUpdateFlowStepStatus(ctx, statusactivities.UpdateStatusRequest{
			ID: step.ID,
			Status: app.CompositeStatus{
				Status: app.StatusSuccess,
			},
		}); err != nil {
			return err
		}
		return nil
	}

	if step.QueueSignal == nil {
		return nil
	}

	sig := step.QueueSignal.Signal
	if sig == nil {
		return nil
	}

	logger := workflow.GetLogger(ctx)
	logger.Info("enqueuing signal to target queue",
		"step_name", step.Name,
		"step_id", step.ID,
		"owner_id", s.OwnerID,
		"owner_type", s.OwnerType,
		"target_queue", s.TargetQueueName,
	)

	enqueueResp, err := sharedactivities.AwaitEnqueueSignalToOwner(ctx, &sharedactivities.EnqueueSignalToOwnerRequest{
		OwnerID:         s.OwnerID,
		OwnerType:       s.OwnerType,
		QueueName:       s.TargetQueueName,
		Signal:          sig,
		SignalOwnerID:   step.ID,
		SignalOwnerType: "install_workflow_steps",
	})
	if err != nil {
		return errors.Wrapf(err, "unable to enqueue signal for step %s", step.Name)
	}

	// Track the inner signal ID so Cancel() can propagate cancellation
	s.innerQueueSignalID = enqueueResp.QueueSignalID

	logger.Info("waiting for queue signal to complete",
		"step_name", step.Name,
		"queue_signal_id", enqueueResp.QueueSignalID,
	)

	_, err = client.AwaitAwaitSignal(ctx, enqueueResp.QueueSignalID)
	if err != nil {
		return errors.Wrapf(err, "queue signal execution failed for step %s", step.Name)
	}

	logger.Info("queue signal completed successfully", "step_name", step.Name)
	return nil
}

// waitForApprovalResponse waits for an approval response using the V2 child workflow approach.
func (s *Signal) waitForApprovalResponse(ctx workflow.Context, flw *app.Workflow, step *app.WorkflowStep) (*app.WorkflowStepApprovalResponse, error) {
	resp, err := workflowsflow.AwaitWaitForApprovalResponse(ctx, &workflowsflow.WaitForApprovalResponseRequest{
		WorkflowID: flw.ID,
		StepID:     step.ID,
	})
	if err != nil {
		if errors.Is(err, context.DeadlineExceeded) || temporal.IsTimeoutError(err) {
			statusactivities.AwaitPkgStatusUpdateFlowStepStatus(ctx, statusactivities.UpdateStatusRequest{
				ID: step.ID,
				Status: app.NewCompositeTemporalStatus(ctx, app.WorkflowStepApprovalStatusApprovalExpired, map[string]any{
					"err_message": "approval was not accepted",
				}),
			})
			return nil, fmt.Errorf("approval timed out for step %s", step.ID)
		}
		return nil, fmt.Errorf("error waiting for approval response: %w", err)
	}

	return resp, nil
}

// handleNoopDeployPlan handles the case where the plan has no changes.
func (s *Signal) handleNoopDeployPlan(ctx workflow.Context, step *app.WorkflowStep, flw *app.Workflow) error {
	if err := statusactivities.AwaitPkgStatusUpdateFlowStepStatus(ctx, statusactivities.UpdateStatusRequest{
		ID: step.ID,
		Status: app.CompositeStatus{
			Status:                 app.StatusAutoSkipped,
			StatusHumanDescription: "Noop Plan, automatically skipped " + strconv.Itoa(step.Idx+1),
			Metadata: map[string]any{
				"step_idx": step.Idx,
				"status":   "auto-skipped",
			},
		},
	}); err != nil {
		return errors.Wrap(err, "unable to update step to success status")
	}

	currentStepIndex := getStepIndex(step.ID, flw.Steps)
	if currentStepIndex == -1 {
		return errors.Errorf("step index not found for step id: %s", step.ID)
	}

	nextStepIndex := currentStepIndex + 1
	if nextStepIndex >= len(flw.Steps) {
		return nil
	}

	nextStep := flw.Steps[nextStepIndex]

	if err := statusactivities.AwaitPkgStatusUpdateFlowStepStatus(ctx, statusactivities.UpdateStatusRequest{
		ID: nextStep.ID,
		Status: app.CompositeStatus{
			Status:                 app.StatusAutoSkipped,
			StatusHumanDescription: "Noop Plan, automatically skipped " + strconv.Itoa(nextStep.Idx),
			Metadata: map[string]any{
				"step_idx": nextStep.Idx,
				"status":   "auto-skipped",
			},
		},
	}); err != nil {
		return errors.Wrap(err, "unable to update step to success status")
	}

	if err := activities.AwaitPkgWorkflowsFlowUpdateFlowStepTargetStatus(ctx, activities.UpdateFlowStepTargetStatusRequest{
		StepID:            step.ID,
		Status:            app.StatusAutoSkipped,
		StatusDescription: "No changes found in plan, skipping deployment.",
	}); err != nil {
		return errors.Wrap(err, "unable to update step target status")
	}

	if err := activities.AwaitSyncNoopDeployOutputs(ctx, &activities.SyncNoopDeployOutputsRequest{
		StepID: step.ID,
	}); err != nil {
		l, _ := log.WorkflowLogger(ctx)
		if l != nil {
			l.Warn("unable to sync noop deploy outputs", zap.Error(err))
		}
	}

	return nil
}

func getStepIndex(stepID string, steps []app.WorkflowStep) int {
	for i, s := range steps {
		if s.ID == stepID {
			return i
		}
	}
	return -1
}

// handlePlanOnlyApproval auto-approves steps in plan-only mode.
func (s *Signal) handlePlanOnlyApproval(ctx workflow.Context, step *app.WorkflowStep, noopPlan bool) error {
	statusDescription := "Auto-approved in plan-only mode."
	targetStatus := app.WorkflowStepApprovalStatusApproved

	if noopPlan {
		statusDescription = "No drift detected "
		targetStatus = app.WorkflowStepNoDrift
	} else {
		statusDescription = "Drift detected"
		targetStatus = app.WorkflowStepDrifted
	}

	if err := statusactivities.AwaitPkgStatusUpdateFlowStepStatus(ctx, statusactivities.UpdateStatusRequest{
		ID: step.ID,
		Status: app.CompositeStatus{
			Status:                 app.WorkflowStepApprovalStatusApproved,
			StatusHumanDescription: "auto-approved (plan-only mode) " + strconv.Itoa(step.Idx+1),
			Metadata: map[string]any{
				"step_idx":  step.Idx,
				"status":    "auto-approved",
				"plan_only": true,
				"no_op":     noopPlan,
			},
		},
	}); err != nil {
		return errors.Wrap(err, "unable to update step to auto-approved status")
	}

	if err := activities.AwaitPkgWorkflowsFlowUpdateFlowStepTargetStatus(ctx, activities.UpdateFlowStepTargetStatusRequest{
		StepID:            step.ID,
		Status:            targetStatus,
		StatusDescription: statusDescription,
	}); err != nil {
		return errors.Wrap(err, "unable to update step target status")
	}

	return nil
}

// checkPolicies evaluates all applicable policies for the step target.
func (s *Signal) checkPolicies(ctx workflow.Context, stepTargetID, stepTargetType string) ([]activities.PolicyViolation, *policyhelpers.PolicyEvaluationContext, error) {
	prepResult, err := activities.AwaitPrepPolicyEvaluation(ctx, &activities.PrepPolicyEvaluationRequest{
		StepTargetID:   stepTargetID,
		StepTargetType: stepTargetType,
	})
	if err != nil {
		return nil, nil, errors.Wrap(err, "unable to prepare policy evaluation")
	}

	if !prepResult.HasPolicies {
		return nil, nil, nil
	}

	ao := workflow.ActivityOptions{
		StartToCloseTimeout:    1*time.Minute + 30*time.Second,
		ScheduleToCloseTimeout: 2 * time.Minute,
		RetryPolicy:            &temporal.RetryPolicy{MaximumAttempts: 1},
	}
	policyCtx := workflow.WithActivityOptions(ctx, ao)

	var futures []workflow.Future
	for _, policy := range prepResult.Policies {
		fut := workflow.ExecuteActivity(policyCtx, (&activities.Activities{}).EvaluateSinglePolicy, &activities.EvaluateSinglePolicyRequest{
			PolicyID:      policy.PolicyID,
			PolicyName:    policy.PolicyName,
			Contents:      policy.Contents,
			InputJSON:     policy.InputJSON,
			InputIndex:    policy.InputIndex,
			InputIdentity: policy.InputIdentity,
		})
		futures = append(futures, fut)
	}

	var allViolations []activities.PolicyViolation
	for _, fut := range futures {
		var result activities.EvaluateSinglePolicyResult
		if err := fut.Get(ctx, &result); err != nil {
			return nil, nil, errors.Wrap(err, "policy evaluation failed")
		}
		allViolations = append(allViolations, result.Violations...)
	}

	return allViolations, &policyhelpers.PolicyEvaluationContext{
		OrgID:            prepResult.OrgID,
		AppID:            prepResult.AppID,
		InstallID:        prepResult.InstallID,
		InstallSandboxID: prepResult.InstallSandboxID,
		ComponentID:      prepResult.ComponentID,
		ComponentBuildID: prepResult.ComponentBuildID,
		PolicyIDs:        prepResult.PolicyIDs,
		InputCount:       prepResult.InputCount,
		OrgName:          prepResult.OrgName,
		AppName:          prepResult.AppName,
		InstallName:      prepResult.InstallName,
		ComponentName:    prepResult.ComponentName,
	}, nil
}

// processPolicyViolations handles policy evaluation results.
func (s *Signal) processPolicyViolations(ctx workflow.Context, l *zap.Logger, step *app.WorkflowStep, flw *app.Workflow, violations []activities.PolicyViolation, passedPolicyIDs []string) error {
	var denyViolations, warnViolations []activities.PolicyViolation
	for _, v := range violations {
		if v.Severity == "deny" {
			denyViolations = append(denyViolations, v)
		} else {
			warnViolations = append(warnViolations, v)
		}
	}

	l.Info("policy evaluation complete",
		zap.String("step_id", step.ID),
		zap.String("step_target_id", step.StepTargetID),
		zap.String("step_target_type", step.StepTargetType),
		zap.String("workflow_id", flw.ID),
		zap.Int("deny_count", len(denyViolations)),
		zap.Int("warn_count", len(warnViolations)),
		zap.Int("passed_count", len(passedPolicyIDs)))

	if len(denyViolations) > 0 {
		if updateErr := statusactivities.AwaitPkgStatusUpdateFlowStepStatus(ctx, statusactivities.UpdateStatusRequest{
			ID: step.ID,
			Status: app.CompositeStatus{
				Status: app.StatusError,
				Metadata: map[string]any{
					"reason":           "Policy violations found",
					DenyViolationsKey:  denyViolations,
					WarnViolationsKey:  warnViolations,
					PassedPolicyIDsKey: passedPolicyIDs,
				},
				StatusHumanDescription: "Policy check failed",
			},
		}); updateErr != nil {
			return errors.Wrap(updateErr, "unable to mark step as error")
		}
		return fmt.Errorf("policy violations found: %d deny violations", len(denyViolations))
	}

	if updateErr := statusactivities.AwaitPkgStatusUpdateFlowStepStatus(ctx, statusactivities.UpdateStatusRequest{
		ID: step.ID,
		Status: app.CompositeStatus{
			Status: step.Status.Status,
			Metadata: map[string]any{
				DenyViolationsKey:  denyViolations,
				WarnViolationsKey:  warnViolations,
				PassedPolicyIDsKey: passedPolicyIDs,
			},
		},
	}); updateErr != nil {
		l.Warn("failed to update step with policy metadata", zap.Error(updateErr))
	}

	return nil
}

// cloneWorkflowStep creates a clone of the step for retry.
func (s *Signal) cloneWorkflowStep(ctx workflow.Context, step *app.WorkflowStep, flw *app.Workflow) error {
	_, err := activities.AwaitPkgWorkflowsFlowCreateFlowSteps(ctx, activities.CreateFlowStepsRequest{
		Steps: []activities.CreateFlowStep{
			{
				FlowID:         flw.ID,
				OwnerID:        flw.OwnerID,
				OwnerType:      flw.OwnerType,
				Name:           getCloneStepName(step.Name),
				Signal:         step.Signal,
				QueueSignal:    step.QueueSignal,
				Status:         app.NewCompositeTemporalStatus(ctx, app.StatusPending),
				Idx:            step.Idx,
				ExecutionType:  step.ExecutionType,
				Metadata:       step.Metadata,
				Retryable:      step.Retryable,
				Skippable:      step.Skippable,
				GroupIdx:       step.GroupIdx,
				GroupRetryIdx:  step.GroupRetryIdx,
				StepTargetType: step.StepTargetType,
				StepTargetID:   step.StepTargetID,
			},
		},
	})
	return err
}

func getCloneStepName(name string) string {
	re := regexp.MustCompile(`^(.*)\(retry (\d+)\)$`)
	matches := re.FindStringSubmatch(name)

	if len(matches) == 3 {
		base := strings.TrimSpace(matches[1])
		retryCount, err := strconv.Atoi(matches[2])
		if err == nil {
			return fmt.Sprintf("%s (retry %d)", base, retryCount+1)
		}
	}

	return fmt.Sprintf("%s (retry 1)", name)
}

// markWorkflowApprovalPlanDenied marks the approval step and its group siblings as denied/skipped.
func (s *Signal) markWorkflowApprovalPlanDenied(ctx workflow.Context, flw *app.Workflow, approvalStep *app.WorkflowStep) error {
	var groupSteps []app.WorkflowStep
	for _, step := range flw.Steps {
		if step.GroupIdx == approvalStep.GroupIdx {
			groupSteps = append(groupSteps, step)
		}
	}
	if len(groupSteps) == 0 {
		return fmt.Errorf("workflow steps for groupIdx %d not found", approvalStep.GroupIdx)
	}

	if err := statusactivities.AwaitPkgStatusUpdateFlowStepStatus(ctx, statusactivities.UpdateStatusRequest{
		ID: approvalStep.ID,
		Status: app.CompositeStatus{
			Status:                 app.WorkflowStepApprovalStatusApprovalDenied,
			StatusHumanDescription: "Plan changes denied, skipping current step group",
		},
	}); err != nil {
		return errors.Wrap(err, "unable to update step to success status")
	}

	if err := activities.AwaitPkgWorkflowsFlowUpdateFlowStepTargetStatus(ctx, activities.UpdateFlowStepTargetStatusRequest{
		StepID:            approvalStep.ID,
		Status:            app.Status(app.InstallDeployApprovalDenied),
		StatusDescription: "Approval denied",
	}); err != nil {
		return errors.Wrap(err, "unable to update step target status")
	}

	for _, step := range groupSteps {
		if step.ID == approvalStep.ID {
			continue
		}

		if !slices.Contains([]app.Status{
			app.StatusPending,
			app.AwaitingApproval,
			app.StatusNotAttempted,
			app.WorkflowStepApprovalStatusApprovalRetryPlan,
		}, step.Status.Status) {
			continue
		}

		if err := statusactivities.AwaitPkgStatusUpdateFlowStepStatus(ctx, statusactivities.UpdateStatusRequest{
			ID: step.ID,
			Status: app.CompositeStatus{
				Status:                 app.StatusUserSkipped,
				StatusHumanDescription: "Plan denied and skipped by the user.",
			},
		}); err != nil {
			return errors.Wrap(err, "unable to update step to success status")
		}
	}

	return nil
}

// markDependentStepsAsSkipped marks the approval step as denied and skips dependent steps.
func (s *Signal) markDependentStepsAsSkipped(ctx workflow.Context, flw *app.Workflow, step *app.WorkflowStep) error {
	if err := s.markWorkflowApprovalPlanDenied(ctx, flw, step); err != nil {
		return errors.Wrap(err, "unable to mark workflow steps approval denied")
	}

	switch app.WorkflowStepTargetType(step.StepTargetType) {
	case app.WorkflowStepTargetTypeInstallSandboxRun, app.WorkflowStepTargetTypeInstallSandboxRuns:
		if err := s.markAllComponentDeployStepsSkipped(ctx, flw); err != nil {
			return errors.Wrap(err, "unable to update step to retry plan status")
		}
	case app.WorkflowStepTargetTypeInstallDeploy, app.WorkflowStepTargetTypeInstallDeploys:
		// Future: skip dependent components
	}
	return nil
}

// markAllComponentDeployStepsSkipped marks all component deploy steps as skipped.
func (s *Signal) markAllComponentDeployStepsSkipped(ctx workflow.Context, flw *app.Workflow) error {
	var groupsToSkip []int
	for _, step := range flw.Steps {
		if app.WorkflowStepTargetType(step.StepTargetType) == app.WorkflowStepTargetTypeInstallDeploy || app.WorkflowStepTargetType(step.StepTargetType) == app.WorkflowStepTargetTypeInstallDeploys {
			groupsToSkip = append(groupsToSkip, step.GroupIdx)
		}
	}

	for _, step := range flw.Steps {
		if slices.Contains(groupsToSkip, step.GroupIdx) {
			if err := statusactivities.AwaitPkgStatusUpdateFlowStepStatus(ctx, statusactivities.UpdateStatusRequest{
				ID: step.ID,
				Status: app.CompositeStatus{
					Status:                 app.StatusUserSkipped,
					StatusHumanDescription: "Plan denied and skipped by the user.",
				},
			}); err != nil {
				return errors.Wrap(err, "unable to update step to success status")
			}
		}
	}

	return nil
}
