package service

import (
	"context"
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/pkg/errors"

	"github.com/powertoolsdev/mono/pkg/generics"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app/installs/helpers"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app/installs/signals"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/middlewares/stderr"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/db/plugins"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/eventloop"
)

type RetryWorkflowByIDRequest struct {
	// StepID is the ID of the step to start the retry from
	StepID string `json:"step_id" swaggertype:"string"`
	// Retry indicates whether to retry the current step or not
	Operation RetryOperation `json:"operation" swaggertype:"string"`
}

type RetryWorkflowByIDResponse struct {
	WorkflowID string `json:"workflow_id" swaggertype:"string"`
}

func (c *RetryWorkflowByIDRequest) Validate(v *validator.Validate) error {
	if err := v.Struct(c); err != nil {
		return fmt.Errorf("invalid request: %w", err)
	}
	return nil
}

// @ID						RetryOwnerWorkflowByID
// @Summary					rerun the workflow steps starting from input step id, can be used to retry a failed step
// @Description.markdown	retry_workflow_by_id.md
// @Param					workflow_id	path	string					true	"workflow ID"
// @Param					req			body	RetryWorkflowByIDRequest	true	"Input"
// @Tags					installs
// @Accept					json
// @Produce					json
// @Security				APIKey
// @Security				OrgID
// @Failure					400	{object}	stderr.ErrResponse
// @Failure					401	{object}	stderr.ErrResponse
// @Failure					403	{object}	stderr.ErrResponse
// @Failure					404	{object}	stderr.ErrResponse
// @Failure					500	{object}	stderr.ErrResponse
// @Success					201	{object}	RetryWorkflowByIDResponse
// @Router					/v1/workflows/{workflow_id}/retry [post]
func (s *service) RetryOwnerWorkflow(ctx *gin.Context) {
	var req RetryWorkflowByIDRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.Error(stderr.ErrUser{
			Err: err,
		})
		return
	}

	if err := req.Validate(s.v); err != nil {
		ctx.Error(err)
		return
	}

	workflowID := ctx.Param("workflow_id")
	workflow, err := s.getWorkflow(ctx, workflowID)
	if err != nil {
		ctx.Error(stderr.ErrUser{
			Err: fmt.Errorf("install workflow not found: %s", workflow.ID),
		})
		return
	}

	step, err := s.getWorkflowStep(ctx, workflow.ID, req.StepID)
	if err != nil {
		ctx.Error(stderr.ErrUser{
			Err: fmt.Errorf("install workflow step not found: %s", req.StepID),
		})
		return
	}

	var stalePlan bool
	var rePlanStepID string
	switch step.Signal.Type {
	case string(signals.OperationProvisionSandboxApplyPlan),
		string(signals.OperationDeprovisionSandboxApplyPlan),
		string(signals.OperationExecuteDeployComponentApplyPlan),
		string(signals.OperationExecuteTeardownComponentApplyPlan),
		string(signals.OperationReprovisionSandboxApplyPlan):

		var planStepSignal []eventloop.SignalType
		switch step.StepTargetType {
		case plugins.TableName(s.db, app.InstallDeploy{}):
			runnerJob, err := s.getRunnerJob(ctx, step.StepTargetID, app.RunnerJobOperationTypeApplyPlan)
			if err != nil {
				ctx.Error(stderr.ErrUser{
					Err: fmt.Errorf("component runner job not found for owner id %s", step.StepTargetID),
				})
				return
			}

			// in in future we support pulumi, add it here
			if runnerJob.Type == app.RunnerJobTypeTerraformDeploy {
				switch step.Signal.Type {
				case string(signals.OperationExecuteDeployComponentApplyPlan):
					planStepSignal = []eventloop.SignalType{
						signals.OperationExecuteDeployComponentPlanOnly,
						signals.OperationExecuteDeployComponentSyncAndPlan,
					}
				case string(signals.OperationExecuteTeardownComponentApplyPlan):
					planStepSignal = []eventloop.SignalType{
						signals.OperationExecuteTeardownComponentSyncAndPlan,
					}
				}
			}
		case plugins.TableName(s.db, app.InstallSandboxRun{}):
			switch step.Signal.Type {
			case string(signals.OperationProvisionSandboxApplyPlan):
				planStepSignal = []eventloop.SignalType{signals.OperationProvisionSandboxPlan}
			case string(signals.OperationReprovisionSandboxApplyPlan):
				planStepSignal = []eventloop.SignalType{signals.OperationReprovisionSandboxPlan}
			case string(signals.OperationDeprovisionSandboxApplyPlan):
				planStepSignal = []eventloop.SignalType{signals.OperationDeprovisionSandboxPlan}
			}
		default:
			// its a terraform apply step ( sandbox or component deploy )
		}

		// if we have a plan step signal, we need to fetch the plan step for the apply step
		if len(planStepSignal) != 0 {
			rePlanStep, err := s.getPlanStepForApplyStep(ctx, workflow, step, &planStepSignal)
			if err != nil {
				ctx.Error(stderr.ErrUser{
					Err: fmt.Errorf("unable to fetch plan step for apply step %s", step.ID),
				})
				return
			}
			rePlanStepID = rePlanStep.ID
			stalePlan = true
		}
	default:
		// its not apply step
	}

	if step.Status.Status != app.StatusError {
		ctx.Error(stderr.ErrUser{
			Err: fmt.Errorf("install workflow %s can't be retried", workflow.ID),
		})
		return
	}

	switch req.Operation {
	case RetryOperationRetryStep:
		if !step.Retryable {
			ctx.Error(stderr.ErrUser{
				Err: fmt.Errorf("install workflow step %s can't be %s", req.StepID, req.Operation),
			})
			return
		}
	case RetryOperationSkipStep:
		if !step.Skippable {
			ctx.Error(stderr.ErrUser{
				Err: fmt.Errorf("install workflow step %s can't be %s", req.StepID, req.Operation),
			})
			return
		}
	}

	if req.Operation == RetryOperationRetryStep {
		if err = s.helpers.UpdateInstallWorkflowStepRetry(ctx, helpers.UpdateInstallWorkflowStepRetry{
			StepID: req.StepID,
		}); err != nil {
			ctx.Error(stderr.ErrSystem{
				Err: fmt.Errorf("failed to update install workflow step retry: %w", err),
			})
			return
		}
	}

	// TODO: support more than just installs workflow retries
	if workflow.OwnerType != "installs" {
		ctx.Error(stderr.ErrUser{
			Err: fmt.Errorf("workflow %s retry not support for owner type", workflow.ID),
		})
		return
	}

	s.evClient.Send(ctx, workflow.OwnerID, &signals.Signal{
		Type:              signals.OperationRerunFlow,
		InstallWorkflowID: workflow.ID,
		RerunConfiguration: signals.RerunConfiguration{
			StepID:        req.StepID,
			StepOperation: signals.RerunOperation(req.Operation),
			StalePlan:     stalePlan,
			RePlanStepID:  rePlanStepID,
		},
	})

	ctx.JSON(201, RetryWorkflowByIDResponse{
		WorkflowID: workflow.ID,
	})
}

// getRunnerJob returns slice of runner jobs with specific owner id
func (s *service) getRunnerJob(ctx context.Context, ownerID string, operation app.RunnerJobOperationType) (*app.RunnerJob, error) {
	var job app.RunnerJob
	res := s.db.WithContext(ctx).
		Where("owner_id = ? AND operation = ? ", ownerID, operation).
		Order("created_at DESC").
		First(&job)
	if res.Error != nil {
		return nil, errors.Wrap(res.Error, "unable to get workflow step")
	}
	return &job, nil
}

func (s *service) getPlanStepForApplyStep(ctx context.Context, workflow *app.Workflow, applyStep *app.WorkflowStep, signalType *[]eventloop.SignalType) (*app.WorkflowStep, error) {
	var steps []app.WorkflowStep
	res := s.db.WithContext(ctx).
		Where("step_target_id = ? ", applyStep.StepTargetID).
		Order("created_at DESC").
		Find(&steps)
	if res.Error != nil {
		return nil, errors.Wrap(res.Error, "unable to get workflow step")
	}

	var planStep *app.WorkflowStep
	for _, s := range steps {
		if generics.SliceContains(eventloop.SignalType(s.Signal.Type), *signalType) {
			planStep = &s
			break
		}
	}

	return planStep, nil
}
