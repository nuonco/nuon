package flow

import (
	"fmt"
	"regexp"
	"slices"
	"strconv"
	"strings"
	"time"

	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"
	"go.uber.org/zap"

	"github.com/pkg/errors"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	policyhelpers "github.com/nuonco/nuon/services/ctl-api/internal/app/policy_reports/helpers"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/eventloop"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/log"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/signal"
	signaldb "github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/signal/db"
	statusactivities "github.com/nuonco/nuon/services/ctl-api/internal/pkg/workflows/status/activities"
	activities "github.com/nuonco/nuon/services/ctl-api/internal/pkg/workflows/workflow/activities"
)

var ErrNotApproved error = fmt.Errorf("not approved")

// Policy evaluation metadata keys
const (
	// DenyViolationsKey is used in metadata to store deny-level violations
	DenyViolationsKey = "deny_violations"
	// WarnViolationsKey is used in metadata to store warning-level violations
	WarnViolationsKey = "warn_violations"
	// PassedPolicyIDsKey is used in metadata to store IDs of policies that passed evaluation
	PassedPolicyIDsKey = "passed_policy_ids"
)

// executeFlowStep executes a single step in the flow. It handles the execution of the step, updates the status, and waits for approval if necessary.
// It returns true if the step needs to be refetched (in case of approval steps), false otherwise.
func (c *WorkflowConductor[DomainSignal]) executeFlowStep(ctx workflow.Context, req eventloop.EventLoopRequest, idx int, step *app.WorkflowStep, flw *app.Workflow) (bool, error) {
	refetchStepsInfo := false

	l, err := log.WorkflowLogger(ctx)
	if err != nil {
		return refetchStepsInfo, nil
	}

	if step.Status.Status != app.StatusPending && step.Status.Status != app.StatusNotAttempted {
		l.Debug("step status not pending or not-attempted, exiting",
			zap.String("step_id", step.ID),
			zap.String("workflow_id", flw.ID))
		return refetchStepsInfo, nil
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
			StatusHumanDescription: "executing step " + step.Name,
			Metadata:               map[string]any{},
		},
	}); err != nil {
		return refetchStepsInfo, errors.Wrap(err, "unable to update step")
	}

	// handle the ok status, and just mark success + continue
	stepErr := c.executeStep(ctx, req, step)
	if stepErr != nil {
		if err := statusactivities.AwaitPkgStatusUpdateFlowStepStatus(ctx, statusactivities.UpdateStatusRequest{
			ID: step.ID,
			Status: app.CompositeStatus{
				Status: app.StatusError,
				Metadata: map[string]any{
					"reason": stepErr.Error(),
				},
				StatusHumanDescription: StepHumanDescription(stepErr),
			},
		}); err != nil {
			return refetchStepsInfo, errors.Wrap(err, "unable to mark step as error")
		}

		return refetchStepsInfo, stepErr
	}

	// fetch the step after the signal was executed, to gather any new state such as the step target id on it.
	step, err = activities.AwaitPkgWorkflowsFlowGetFlowsStepByFlowStepID(ctx, step.ID)
	if err != nil {
		return refetchStepsInfo, errors.Wrap(err, "unable to get step")
	}

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
			return refetchStepsInfo, errors.Wrap(err, "unable to mark step as success")
		}

		if err := statusactivities.AwaitPkgStatusUpdateFlowStatus(ctx, statusactivities.UpdateStatusRequest{
			ID: flw.ID,
			Status: app.CompositeStatus{
				Status:                 app.StatusSuccess,
				StatusHumanDescription: "finished executing step " + step.Name,
				Metadata: map[string]any{
					"step_idx": step.Idx,
					"status":   "ok",
				},
			},
		}); err != nil {
			return refetchStepsInfo, errors.Wrap(err, "unable to update step to success status")
		}

		return refetchStepsInfo, nil
	}

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
		return refetchStepsInfo, errors.Wrap(err, "unable to mark step status as checking plan")
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
			return refetchStepsInfo, errors.Wrap(err, "unable to mark step as error")
		}

		return refetchStepsInfo, errors.Wrap(err, "failed to check for noop plan")
	}

	// check for plan contents here, if noop then mark auto approved + nex step as skipped since its noop change
	if noopPlan {
		l.Debug("approval plan contents empty",
			zap.String("step_id", step.ID),
			zap.String("workflow_id", flw.ID))
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
				return refetchStepsInfo, errors.Wrap(err, "unable to mark step as error")
			}

			return refetchStepsInfo, errors.Wrap(err, "failed to handle noop plan")
		}

		if !flw.PlanOnly {
			refetchStepsInfo = true
			return refetchStepsInfo, nil
		}
	}

	// Check policies before approval
	l.Debug("starting policy check",
		zap.String("step_id", step.ID),
		zap.String("step_target_id", step.StepTargetID),
		zap.String("step_target_type", step.StepTargetType),
		zap.String("workflow_id", flw.ID))

	violations, policyContext, policyErr := c.checkPolicies(ctx, step.StepTargetID, step.StepTargetType)
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
		if err := c.processPolicyViolations(ctx, l, step, flw, violations, passedPolicyIDs); err != nil {
			return refetchStepsInfo, errors.Wrap(err, "unable to process check for policy violation")
		}
	}

	l.Debug("policy check completed successfully",
		zap.String("step_id", step.ID),
		zap.String("step_target_id", step.StepTargetID),
		zap.String("step_target_type", step.StepTargetType),
		zap.String("workflow_id", flw.ID))

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
				return refetchStepsInfo, errors.Wrap(err, "unable to mark step as error")
			}
			return refetchStepsInfo, errors.Wrap(err, "failed to handle plan-only auto-approval")
		}

		return refetchStepsInfo, nil
	}

	// update the status to awaiting
	if err := statusactivities.AwaitPkgStatusUpdateFlowStepStatus(ctx, statusactivities.UpdateStatusRequest{
		ID: step.ID,
		Status: app.CompositeStatus{
			Status:                 app.AwaitingApproval,
			StatusHumanDescription: "awaiting approval for " + step.Name,
			Metadata: map[string]any{
				"step_idx": step.Idx,
				"status":   "ok",
			},
		},
	}); err != nil {
		return refetchStepsInfo, errors.Wrap(err, "unable to update step to success status")
	}

	approvalFunc := c.waitForApprovalResponse
	// Use the v2 approval flow (with continue as new) for workflows created after Nov 26, 2025.
	cutoffDate := time.Date(2025, time.November, 26, 0, 0, 0, 0, time.UTC)
	if flw.CreatedAt.After(cutoffDate) {
		approvalFunc = c.waitForApprovalResponseV2
	}

	resp, err := approvalFunc(ctx, flw, step, idx)
	if err != nil {
		return refetchStepsInfo, err
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
				StatusHumanDescription: "approved " + step.Name,
				Metadata: map[string]any{
					"step_idx": step.Idx,
					"status":   "ok",
				},
			},
		}); err != nil {
			return refetchStepsInfo, errors.Wrap(err, "unable to update step to success status")
		}

		return refetchStepsInfo, nil
	// approval response retry flow
	case app.WorkflowStepApprovalResponseTypeRetryPlan:
		l.Debug("handling approval response type: retry plan",
			zap.String("step_id", step.ID),
			zap.String("workflow_id", flw.ID))

		// cloned step which will be retried next
		err := c.cloneWorkflowStep(ctx, step, flw)
		if err != nil {
			return refetchStepsInfo, errors.Wrap(err, "unable to clone step for retry plan")
		}

		if err := statusactivities.AwaitPkgStatusUpdateFlowStepStatus(ctx, statusactivities.UpdateStatusRequest{
			ID: step.ID,
			Status: app.CompositeStatus{
				Status:                 app.WorkflowStepApprovalStatusApprovalRetryPlan,
				StatusHumanDescription: "retrying " + step.Name,
				Metadata: map[string]any{
					"step_idx": step.Idx,
					"status":   "retrying",
				},
			},
		}); err != nil {
			return refetchStepsInfo, errors.Wrap(err, "unable to update step to retry plan status")
		}

		if err := activities.AwaitPkgWorkflowsFlowUpdateFlowStepTargetStatus(ctx, activities.UpdateFlowStepTargetStatusRequest{
			StepID:            step.ID,
			Status:            app.StatusDiscarded,
			StatusDescription: "Retrying step " + step.Name,
		}); err != nil {
			return refetchStepsInfo, errors.Wrap(err, "unable to update step target status")
		}

		refetchStepsInfo = true
		return refetchStepsInfo, nil
	case app.WorkflowStepApprovalResponseTypeSkipCurrent:
		l.Debug("handling approval response type: skip current and continue",
			zap.String("step_id", step.ID),
			zap.String("workflow_id", flw.ID))

		if err := c.markWorkflowApprovalPlanDenied(ctx, flw, step); err != nil {
			if err := statusactivities.AwaitPkgStatusUpdateFlowStatus(ctx, statusactivities.UpdateStatusRequest{
				ID: flw.ID,
				Status: app.CompositeStatus{
					Status:                 app.StatusError,
					StatusHumanDescription: "failed to deny plan and update step status",
					Metadata:               map[string]any{},
				},
			}); err != nil {
				return refetchStepsInfo, errors.Wrap(err, "unable to mark workflow steps approval denied")
			}
		}

		refetchStepsInfo = true
		return refetchStepsInfo, nil
		// update step status to approval denied and somehow figureout how to skip at the top
		// this is not being used rn dashboardui can't trigger this
	case app.WorkflowStepApprovalResponseTypeSkipCurrentAndDependents:
		l.Debug("handling approval response type: skip current and dependents",
			zap.String("step_id", step.ID),
			zap.String("workflow_id", flw.ID))

		if err := c.markDependentStepsAsSkipped(ctx, flw, step); err != nil {
			if err := statusactivities.AwaitPkgStatusUpdateFlowStatus(ctx, statusactivities.UpdateStatusRequest{
				ID: flw.ID,
				Status: app.CompositeStatus{
					Status:                 app.StatusError,
					StatusHumanDescription: "failed to deny plan and update step status",
					Metadata:               map[string]any{},
				},
			}); err != nil {
				return refetchStepsInfo, errors.Wrap(err, "unable to mark workflow steps approval denied and update step status")
			}
		}

		// find all dependent step groups and mark
		refetchStepsInfo = true
		return refetchStepsInfo, nil

	}

	if err := statusactivities.AwaitPkgStatusUpdateFlowStepStatus(ctx, statusactivities.UpdateStatusRequest{
		ID: step.ID,
		Status: app.NewCompositeTemporalStatus(ctx, app.WorkflowStepApprovalStatusApprovalDenied, map[string]any{
			"reason": "approval denied",
		}),
	}); err != nil {
		return refetchStepsInfo, errors.Wrap(err, "unable to update")
	}
	if err := activities.AwaitPkgWorkflowsFlowUpdateFlowStepTargetStatus(ctx, activities.UpdateFlowStepTargetStatusRequest{
		StepID:            step.ID,
		Status:            app.Status(app.InstallDeployApprovalDenied),
		StatusDescription: "Approval denied",
	}); err != nil {
		return refetchStepsInfo, errors.Wrap(err, "unable to update step target status")
	}

	return refetchStepsInfo, ErrNotApproved
}

func (c *WorkflowConductor[DomainSignal]) cloneWorkflowStep(ctx workflow.Context, step *app.WorkflowStep, flw *app.Workflow) error {
	newRetryIndex := step.RetryIndex + 1

	maxRetries := signal.DefaultMaxRetries
	if step.QueueSignal != nil && step.QueueSignal.Signal != nil {
		if mr, ok := step.QueueSignal.Signal.(signal.SignalWithMaxRetries); ok {
			maxRetries = mr.MaxRetries()
		}
	}
	if newRetryIndex > maxRetries {
		return fmt.Errorf("step %s has exceeded maximum retry count of %d", step.ID, maxRetries)
	}

	// If the signal implements Clone(), use it for retry step creation.
	if step.QueueSignal != nil && step.QueueSignal.Signal != nil {
		if cl, ok := step.QueueSignal.Signal.(signal.SignalWithClone); ok {
			return c.createCloneSteps(ctx, step, flw, cl, newRetryIndex)
		}
	}

	_, err := activities.AwaitPkgWorkflowsFlowCreateFlowSteps(ctx, activities.CreateFlowStepsRequest{
		Steps: []activities.CreateFlowStep{
			{
				FlowID:      flw.ID,
				OwnerID:     flw.OwnerID,
				OwnerType:   flw.OwnerType,
				Name:        getCloneStepName(step.Name),
				Signal:      step.Signal,
				QueueSignal: step.QueueSignal,
				Status: app.NewCompositeTemporalStatus(ctx, app.StatusPending, map[string]any{
					"is_retry":  true,
					"retry_idx": newRetryIndex,
				}),
				Idx:            step.Idx,
				ExecutionType:  step.ExecutionType,
				Metadata:       step.Metadata,
				Retryable:      step.Retryable,
				Skippable:      step.Skippable,
				GroupIdx:       step.GroupIdx,
				GroupRetryIdx:  step.GroupRetryIdx,
				StepTargetType: step.StepTargetType,
				RetryIndex:     newRetryIndex,
				Timeout:        step.Timeout,
				// StepTargetID intentionally omitted — the clone must create
				// a fresh target when it executes, not reuse the original.
			},
		},
	})
	return err
}

// createCloneSteps builds steps from a SignalWithClone implementation.
func (c *WorkflowConductor[DomainSignal]) createCloneSteps(ctx workflow.Context, step *app.WorkflowStep, flw *app.Workflow, cl signal.SignalWithClone, retryIndex int) error {
	defs, err := cl.Clone(ctx, removeRetryFromStepName(step.Name))
	if err != nil {
		return fmt.Errorf("unable to clone signal for retry: %w", err)
	}
	steps := make([]activities.CreateFlowStep, 0, len(defs))
	for i, def := range defs {
		steps = append(steps, activities.CreateFlowStep{
			FlowID:      flw.ID,
			OwnerID:     flw.OwnerID,
			OwnerType:   flw.OwnerType,
			Name:        getCloneStepName(def.Name),
			QueueSignal: &signaldb.SignalData{Signal: def.Signal},
			Status: app.NewCompositeTemporalStatus(ctx, app.StatusPending, map[string]any{
				"is_retry":  true,
				"retry_idx": retryIndex,
			}),
			Idx:            step.Idx + i,
			ExecutionType:  app.WorkflowStepExecutionType(def.ExecutionType),
			Metadata:       step.Metadata,
			Retryable:      step.Retryable,
			Skippable:      step.Skippable,
			GroupIdx:       step.GroupIdx,
			GroupRetryIdx:  step.GroupRetryIdx,
			StepTargetType: step.StepTargetType,
			RetryIndex:     retryIndex,
			Timeout:        signal.DeriveTimeout(def.Signal),
		})
	}
	_, createErr := activities.AwaitPkgWorkflowsFlowCreateFlowSteps(ctx, activities.CreateFlowStepsRequest{Steps: steps})
	return createErr
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

func (c *WorkflowConductor[DomainSignal]) getWorkflowStepGroup(ctx workflow.Context, flw *app.Workflow, groupIdx int) (*[]app.WorkflowStep, error) {
	var steps []app.WorkflowStep
	for _, step := range flw.Steps {
		if step.GroupIdx == groupIdx {
			steps = append(steps, step)
		}
	}

	if len(steps) == 0 {
		return nil, fmt.Errorf("workflow steps for groupIdx %d not found", groupIdx)
	}

	return &steps, nil
}

func (c *WorkflowConductor[DomainSignal]) markDependentStepsAsSkipped(ctx workflow.Context, flw *app.Workflow, step *app.WorkflowStep) error {
	if err := c.markWorkflowApprovalPlanDenied(ctx, flw, step); err != nil {
		return errors.Wrap(err, "unable to mark workflow steps approval denied")
	}

	switch app.WorkflowStepTargetType(step.StepTargetType) {
	case app.WorkflowStepTargetTypeInstallSandboxRun, app.WorkflowStepTargetTypeInstallSandboxRuns:
		// skip all the component deploys
		if err := c.markAllComponentDeployStepsSkipped(ctx, flw); err != nil {
			return errors.Wrap(err, "unable to update step to retry plan status")
		}
	case app.WorkflowStepTargetTypeInstallDeploy, app.WorkflowStepTargetTypeInstallDeploys:
		// installID := generics.FromPtrStr(flw.Metadata["install_id"])
		// install, err := appactivities.AwaitGetByInstallID(ctx, installID)
		// if err != nil {
		// 	return errors.Wrap(err, "unable to get install")
		// }
		// appConfig, err := appactivities.AwaitGetAppConfig(ctx, appactivities.GetAppConfigRequest{
		// 	ID: install.AppConfigID,
		// })
		// if err != nil {
		// 	return errors.Wrapf(err, "unable to get app config for install %s", installID)
		// }
		//
		// // find all dependent components
		// var sig map[string]any
		// if err := json.Unmarshal(step.Signal.SignalJSON, &sig); err != nil {
		// 	return c.handleStepErr(ctx, step.ID, err)
		// }
		// subSignal := sig[]
		// _, err = appactivities.AwaitGetComponentDependents(ctx, &appactivities.GetComponentDependentsRequest{
		// 	AppConfigID: appConfig.ID,
		// 	ComponentID: "",
		// })
		//
		// // skip all dependent components
	}
	return nil
}

func (c *WorkflowConductor[DomainSignal]) markAllComponentDeployStepsSkipped(ctx workflow.Context, flw *app.Workflow) error {
	var groupsToSkip []int
	for _, step := range flw.Steps {
		if app.WorkflowStepTargetType(step.StepTargetType) == app.WorkflowStepTargetTypeInstallDeploy || app.WorkflowStepTargetType(step.StepTargetType) == app.WorkflowStepTargetTypeInstallDeploys {
			groupsToSkip = append(groupsToSkip, step.GroupIdx)
		}
	}

	for _, step := range flw.Steps {
		if slices.Contains(groupsToSkip, step.GroupIdx) {
			if err := statusactivities.AwaitPkgStatusUpdateFlowStepStatus(ctx, statusactivities.UpdateStatusRequest{
				// this needs to be the step next in line
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

func (c *WorkflowConductor[DomainSignal]) markWorkflowApprovalPlanDenied(ctx workflow.Context, flw *app.Workflow, approvalStep *app.WorkflowStep) error {
	groupSteps, err := c.getWorkflowStepGroup(ctx, flw, approvalStep.GroupIdx)
	if err != nil {
		return fmt.Errorf("unable to get workflow step group")
	}

	if err := statusactivities.AwaitPkgStatusUpdateFlowStepStatus(ctx, statusactivities.UpdateStatusRequest{
		// this needs to be the step next in line
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

	for _, step := range *groupSteps {
		if step.ID == approvalStep.ID {
			continue
		}

		// todo(sk): this is probably code smell, need better way to handle this
		if !slices.Contains([]app.Status{
			app.StatusPending,
			app.AwaitingApproval,
			app.StatusNotAttempted,
			app.WorkflowStepApprovalStatusApprovalRetryPlan,
		}, step.Status.Status) {
			continue
		}

		if err := statusactivities.AwaitPkgStatusUpdateFlowStepStatus(ctx, statusactivities.UpdateStatusRequest{
			// this needs to be the step next in line
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
			StatusHumanDescription: "Noop Plan, automatically skipped " + step.Name,
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
			StatusHumanDescription: "Noop Plan, automatically skipped " + nextStep.Name,
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
			StatusHumanDescription: "auto-approved (plan-only mode) " + step.Name,
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

// processPolicyViolations separates violations into deny and warn categories
// and updates step metadata with violations and passed policy IDs.
// Returns early with error if deny violations exist.
func (c *WorkflowConductor[DomainSignal]) processPolicyViolations(ctx workflow.Context, l *zap.Logger, step *app.WorkflowStep, flw *app.Workflow, violations []activities.PolicyViolation, passedPolicyIDs []string) error {
	denyViolations, warnViolations := c.separateViolations(violations)

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

// separateViolations categorizes violations into deny and warn severity levels.
func (c *WorkflowConductor[DomainSignal]) separateViolations(violations []activities.PolicyViolation) ([]activities.PolicyViolation, []activities.PolicyViolation) {
	var denyViolations []activities.PolicyViolation
	var warnViolations []activities.PolicyViolation
	for _, v := range violations {
		if v.Severity == "deny" {
			denyViolations = append(denyViolations, v)
		} else {
			warnViolations = append(warnViolations, v)
		}
	}
	return denyViolations, warnViolations
}

// checkPolicies prepares policy evaluation and then evaluates all applicable policies in parallel.
// It returns all violations found across all policies.
//
// policyEvaluationContext is an alias for the shared PolicyEvaluationContext type.
type policyEvaluationContext = policyhelpers.PolicyEvaluationContext

func (c *WorkflowConductor[DomainSignal]) checkPolicies(ctx workflow.Context, stepTargetID, stepTargetType string) ([]activities.PolicyViolation, *policyEvaluationContext, error) {
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

	// Execute all policy evaluations in parallel
	// TODO: extend temporal-gen to generate an Execute* variant that returns workflow.Future
	// so we can use generated activity options instead of manually specifying them here.
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

	// Collect all violations from parallel evaluations
	var allViolations []activities.PolicyViolation
	for _, fut := range futures {
		var result activities.EvaluateSinglePolicyResult
		if err := fut.Get(ctx, &result); err != nil {
			return nil, nil, errors.Wrap(err, "policy evaluation failed")
		}
		allViolations = append(allViolations, result.Violations...)
	}

	return allViolations, &policyEvaluationContext{
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
