package flow

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"go.temporal.io/sdk/workflow"
	"go.uber.org/zap"

	"github.com/pkg/errors"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/eventloop"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/log"
	activities "github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/workflows/flow/activities"
	statusactivities "github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/workflows/status/activities"
)

var ErrNotApproved error = fmt.Errorf("not approved")

// executeFlowStep executes a single step in the flow. It handles the execution of the step, updates the status, and waits for approval if necessary.
// It returns true if the step needs to be refetched (in case of approval steps), false otherwise.
func (c *WorkflowConductor[DomainSignal]) executeFlowStep(ctx workflow.Context, req eventloop.EventLoopRequest, idx int, step *app.WorkflowStep, flw *app.Workflow) (bool, error) {
	l, err := log.WorkflowLogger(ctx)
	if err != nil {
		return false, nil
	}

	if step.Status.Status == app.StatusAutoSkipped {
		return false, nil
	}

	defer func() {
		if err := activities.AwaitPkgWorkflowsFlowUpdateFlowStepFinishedAtByID(ctx, step.ID); err != nil {
			l.Error("unable to update finished at", zap.Error(err))
		}
	}()

	if err := statusactivities.AwaitPkgStatusUpdateFlowStatus(ctx, statusactivities.UpdateStatusRequest{
		ID: flw.ID,
		Status: app.CompositeStatus{
			Status:                 app.StatusInProgress,
			StatusHumanDescription: "executing step " + strconv.Itoa(step.Idx+1),
			Metadata:               map[string]any{},
		},
	}); err != nil {
		return false, errors.Wrap(err, "unable to update step")
	}

	// handle the ok status, and just mark success + continue
	stepErr := c.executeStep(ctx, req, step)
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
			return false, errors.Wrap(err, "unable to mark step as error")
		}

		return false, stepErr
	}

	// fetch the step after the signal was executed, to gather any new state such as the step target id on it.
	step, err = activities.AwaitPkgWorkflowsFlowGetFlowsStepByFlowStepID(ctx, step.ID)
	if err != nil {
		return false, errors.Wrap(err, "unable to get step")
	}

	if step.ExecutionType != app.WorkflowStepExecutionTypeApproval {
		if err := statusactivities.AwaitPkgStatusUpdateFlowStepStatus(ctx, statusactivities.UpdateStatusRequest{
			ID: step.ID,
			Status: app.CompositeStatus{
				Status: app.StatusSuccess,
			},
		}); err != nil {
			return false, errors.Wrap(err, "unable to mark step as success")
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
			return false, errors.Wrap(err, "unable to update step to success status")
		}

		return false, nil
	}

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
		return false, errors.Wrap(err, "unable to mark step as success")
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
			return false, errors.Wrap(err, "unable to mark step as error")
		}

		return false, errors.Wrap(err, "failed to check for noop plan")
	}
	// check for plan contents here, if noop then mark auto approved + nex step as skipped since its noop change
	if noopPlan {
		if err := c.handleNoopDeployPlan(ctx, step, flw); err != nil {
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
				return false, errors.Wrap(err, "unable to mark step as error")
			}

			return false, errors.Wrap(err, "failed to handle noop plan")
		}

		if !flw.PlanOnly {
			return true, nil
		}
	}

	// Auto approve if plan-only mode is enabled
	if flw.PlanOnly {
		if err := c.handlePlanOnlyApproval(ctx, step, noopPlan); err != nil {
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
				return false, errors.Wrap(err, "unable to mark step as error")
			}
			return false, errors.Wrap(err, "failed to handle plan-only auto-approval")
		}
		return false, nil
	}

	// update the status to awaiting
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
		return false, errors.Wrap(err, "unable to update step to success status")
	}

	resp, err := c.waitForApprovalResponse(ctx, flw, step, idx)
	if err != nil {
		return false, err
	}

	if resp.Type == app.WorkflowStepApprovalResponseTypeApprove {
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
			return false, errors.Wrap(err, "unable to update step to success status")
		}

		return false, nil
	}

	// approval response retry flow
	if resp.Type == app.WorkflowStepApprovalResponseTypeRetryPlan {
		// cloned step which will be retried next
		err := c.cloneWorkflowStep(ctx, step, flw)
		if err != nil {
			return false, errors.Wrap(err, "unable to clone step for retry plan")
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
			return false, errors.Wrap(err, "unable to update step to retry plan status")
		}

		if err := activities.AwaitPkgWorkflowsFlowUpdateFlowStepTargetStatus(ctx, activities.UpdateFlowStepTargetStatusRequest{
			StepID:            step.ID,
			Status:            app.StatusDiscarded,
			StatusDescription: "Retrying step " + strconv.Itoa(step.Idx),
		}); err != nil {
			return false, errors.Wrap(err, "unable to update step target status")
		}

		return true, nil
	}

	if err := statusactivities.AwaitPkgStatusUpdateFlowStepStatus(ctx, statusactivities.UpdateStatusRequest{
		ID: step.ID,
		Status: app.NewCompositeTemporalStatus(ctx, app.WorkflowStepApprovalStatusApprovalDenied, map[string]any{
			"reason": "approval denied",
		}),
	}); err != nil {
		return false, errors.Wrap(err, "unable to update")
	}
	if err := activities.AwaitPkgWorkflowsFlowUpdateFlowStepTargetStatus(ctx, activities.UpdateFlowStepTargetStatusRequest{
		StepID:            step.ID,
		Status:            app.Status(app.InstallDeployApprovalDenied),
		StatusDescription: "Approval denied",
	}); err != nil {
		return false, errors.Wrap(err, "unable to update step target status")
	}

	return false, ErrNotApproved
}

func (c *WorkflowConductor[DomainSignal]) cloneWorkflowStep(ctx workflow.Context, step *app.WorkflowStep, flw *app.Workflow) error {
	_, err := activities.AwaitPkgWorkflowsFlowCreateFlowStep(ctx, activities.CreateFlowStepRequest{
		FlowID:         flw.ID,
		OwnerID:        flw.OwnerID,
		OwnerType:      flw.OwnerType,
		Name:           getCloneStepName(step.Name),
		Signal:         step.Signal,
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
	})
	return err
}

// getCloneStepName generates a new step name for a cloned step.
// this is quick regex based approach to skip unwanted db call
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

	// No retry suffix found, or unable to parse
	return fmt.Sprintf("%s (retry 1)", name)
}

// removeRetryFromStepName removes the retry suffix from a step name if it exists.
// this is quick regex based approach to skip unwanted db call
func removeRetryFromStepName(name string) string {
	re := regexp.MustCompile(`^(.*)\(retry \d+\)$`)
	matches := re.FindStringSubmatch(name)

	if len(matches) == 2 {
		return strings.TrimSpace(matches[1])
	}

	// No retry suffix found
	return name
}

func (c *WorkflowConductor[DomainSignal]) getStepApprovalPlan(ctx workflow.Context, step *app.WorkflowStep) (*activities.ApprovalPlan, error) {
	// assumption here is that, for approval type steps, there will always be a runPlan
	approvalPlan, err := activities.AwaitGetApprovalPlan(ctx, activities.GetApprovalPlanRequest{
		StepTargetID: step.StepTargetID,
	})
	if err != nil {
		return nil, errors.Wrap(err, "unable to get step approval plan")
	}

	return approvalPlan, nil
}

func (c *WorkflowConductor[DomainSignal]) handleNoopDeployPlan(ctx workflow.Context, step *app.WorkflowStep, flw *app.Workflow) error {
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
	currentStepIndex := c.getStepIndex(step.ID, flw.Steps)
	if currentStepIndex == -1 {
		return errors.Errorf("step index not found for step id: %s", step.ID)
	}

	nextStepIndex := currentStepIndex + 1

	if nextStepIndex >= len(flw.Steps) { // this can happen in plan-only mode where we don't have an apply step.
		return nil // we can let the planonly workflow condition update the status
	}

	nextStep := flw.Steps[nextStepIndex]

	if err := statusactivities.AwaitPkgStatusUpdateFlowStepStatus(ctx, statusactivities.UpdateStatusRequest{
		// this needs to be the step next in line
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

	// this needs to be same as previous value
	if err := activities.AwaitPkgWorkflowsFlowUpdateFlowStepTargetStatus(ctx, activities.UpdateFlowStepTargetStatusRequest{
		StepID:            step.ID,
		Status:            app.StatusAutoSkipped,
		StatusDescription: "No changes found in plan, skipping deployment.",
	}); err != nil {
		return errors.Wrap(err, "unable to update step target status")
	}

	return nil
}

func (c *WorkflowConductor[DomainSignal]) getStepIndex(stepID string, steps []app.WorkflowStep) int {
	for i, s := range steps {
		if s.ID == stepID {
			return i
		}
	}
	return -1
}

func (c *WorkflowConductor[DomainSignal]) handlePlanOnlyApproval(ctx workflow.Context, step *app.WorkflowStep, noopPlan bool) error {
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
