package service

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/structured_errors"
)

type ReportCompositeErrorsRequest struct {
	Errors structured_errors.CompositeErrors `json:"errors" validate:"required"`
}

// @ID						ReportCompositeErrors
// @Summary				report composite errors for a runner job
// @Description			Append structured errors to the parent entity (deploy, build, sandbox run, action run) of the specified runner job.
// @Param					runner_job_id	path	string							true	"runner job ID"
// @Param					req				body	ReportCompositeErrorsRequest	true	"Input"
// @Tags					runners/runner
// @Accept					json
// @Produce				json
// @Security				APIKey
// @Security				OrgID
// @Failure				400	{object}	stderr.ErrResponse
// @Failure				401	{object}	stderr.ErrResponse
// @Failure				403	{object}	stderr.ErrResponse
// @Failure				404	{object}	stderr.ErrResponse
// @Failure				500	{object}	stderr.ErrResponse
// @Success				201
// @Router					/v1/runner-jobs/{runner_job_id}/errors [POST]
func (s *service) ReportCompositeErrors(ctx *gin.Context) {
	runnerJobID := ctx.Param("runner_job_id")

	var req ReportCompositeErrorsRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.Error(fmt.Errorf("unable to parse request: %w", err))
		return
	}

	if len(req.Errors) == 0 {
		ctx.Status(http.StatusCreated)
		return
	}

	runnerJob, err := s.getRunnerJob(ctx, runnerJobID)
	if err != nil {
		ctx.Error(fmt.Errorf("unable to get runner job: %w", err))
		return
	}

	// Set owner ID on all errors to the runner job ID
	for i := range req.Errors {
		req.Errors[i].OwnerID = runnerJobID
	}

	// Resolve the parent entity via the polymorphic OwnerType/OwnerID
	model, err := s.resolveOwnerModel(runnerJob)
	if err != nil {
		ctx.Error(fmt.Errorf("unable to resolve owner: %w", err))
		return
	}

	if err := structured_errors.Append(s.db.WithContext(ctx), model, runnerJob.OwnerID, req.Errors); err != nil {
		ctx.Error(fmt.Errorf("unable to append errors: %w", err))
		return
	}

	ctx.Status(http.StatusCreated)
}

func (s *service) resolveOwnerModel(runnerJob *app.RunnerJob) (any, error) {
	switch runnerJob.OwnerType {
	case "install_deploys":
		return &app.InstallDeploy{}, nil
	case "component_builds":
		return &app.ComponentBuild{}, nil
	case "install_sandbox_runs":
		return &app.InstallSandboxRun{}, nil
	case "install_action_workflow_runs":
		return &app.InstallActionWorkflowRun{}, nil
	default:
		return nil, fmt.Errorf("unsupported owner type: %s", runnerJob.OwnerType)
	}
}
