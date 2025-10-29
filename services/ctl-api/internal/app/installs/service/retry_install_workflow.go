package service

import (
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app/installs/helpers"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app/installs/signals"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/middlewares/stderr"
)

type RetryOperation string

const (
	RetryOperationSkipStep  = "skip-step"
	RetryOperationRetryStep = "retry-step"
)

type RetryWorkflowRequest struct {
	WorkflowID string `json:"workflow_id" swaggertype:"string"`
	// StepID is the ID of the step to start the retry from
	StepID string `json:"step_id" swaggertype:"string"`
	// Retry indicates whether to retry the current step or not
	Operation RetryOperation `json:"operation" swaggertype:"string"`
}

type RetryWorkflowResponse struct {
	WorkflowID string `json:"workflow_id" swaggertype:"string"`
}

func (c *RetryWorkflowRequest) Validate(v *validator.Validate) error {
	if err := v.Struct(c); err != nil {
		return fmt.Errorf("invalid request: %w", err)
	}
	return nil
}

// @ID						RetryWorkflow
// @Summary					rerun the workflow steps starting from input step id, can be used to retry a failed step
// @Description.markdown	retry_workflow.md
// @Param					install_id	path	string					true	"install ID"
// @Param					req			body	RetryWorkflowRequest	true	"Input"
// @Tags					installs
// @Accept					json
// @Produce					json
// @Security				APIKey
// @Security				OrgID
// @Deprecated
// @Failure					400	{object}	stderr.ErrResponse
// @Failure					401	{object}	stderr.ErrResponse
// @Failure					403	{object}	stderr.ErrResponse
// @Failure					404	{object}	stderr.ErrResponse
// @Failure					500	{object}	stderr.ErrResponse
// @Success					201	{object}	RetryWorkflowResponse
// @Router					/v1/installs/{install_id}/retry-workflow [post]
func (s *service) RetryWorkflow(ctx *gin.Context) {
	fmt.Println("rb - RetryWorkflow called")
	install_id := ctx.Param("install_id")

	var req RetryWorkflowRequest
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

	workflow, err := s.getWorkflow(ctx, req.WorkflowID)
	if err != nil {
		ctx.Error(stderr.ErrUser{
			Err: fmt.Errorf("install workflow not found: %s", req.WorkflowID),
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

	if step.Status.Status != app.StatusError {
		ctx.Error(stderr.ErrUser{
			Err: fmt.Errorf("install workflow %s can't be retried", workflow.ID),
		})
		return
	}

	// this feels like code smell since its not explicit
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

	s.evClient.Send(ctx, install_id, &signals.Signal{
		Type:              signals.OperationRerunFlow,
		InstallWorkflowID: workflow.ID,
		RerunConfiguration: signals.RerunConfiguration{
			StepID:        req.StepID,
			StepOperation: signals.RerunOperation(req.Operation),
		},
	})

	ctx.JSON(201, RetryWorkflowResponse{
		WorkflowID: workflow.ID,
	})
}
