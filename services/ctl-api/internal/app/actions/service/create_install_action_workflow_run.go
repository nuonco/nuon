package service

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/pkg/errors"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
	installsignals "github.com/powertoolsdev/mono/services/ctl-api/internal/app/installs/signals"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/middlewares/stderr"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/db/generics"
)

type CreateInstallActionWorkflowRunRequest struct {
	ActionWorkFlowConfigID string `json:"action_workflow_config_id" binding:"required"`

	RunEnvVars map[string]string `json:"run_env_vars"`
}

func (c *CreateInstallActionWorkflowRunRequest) Validate(v *validator.Validate) error {
	if err := v.Struct(c); err != nil {
		return fmt.Errorf("invalid request: %w", err)
	}
	return nil
}

//	@ID						CreateInstallActionWorkflowRun
//	@Summary				create an action workflow run for an install
//	@Description.markdown	create_install_action_workflow_run.md
//	@Tags					actions
//	@Accept					json
//	@Param					install_id	path	string									true	"install ID"
//	@Param					req			body	CreateInstallActionWorkflowRunRequest	true	"Input"
//	@Produce				json
//	@Security				APIKey
//	@Security				OrgID
//	@Failure				400	{object}	stderr.ErrResponse
//	@Failure				401	{object}	stderr.ErrResponse
//	@Failure				403	{object}	stderr.ErrResponse
//	@Failure				404	{object}	stderr.ErrResponse
//	@Failure				500	{object}	stderr.ErrResponse
//	@Success				201	{object}	app.InstallActionWorkflowRun
//	@Router					/v1/installs/{install_id}/action-workflows/runs [post]
func (s *service) CreateInstallActionWorkflowRun(ctx *gin.Context) {
	installID := ctx.Param("install_id")

	var req CreateInstallActionWorkflowRunRequest
	if err := ctx.BindJSON(&req); err != nil {
		ctx.Error(fmt.Errorf("unable to parse request: %w", err))
		return
	}

	if err := req.Validate(s.v); err != nil {
		ctx.Error(fmt.Errorf("invalid request: %w", err))
		return
	}

	awc, err := s.findActionWorkflowConfig(ctx, req.ActionWorkFlowConfigID)
	if err != nil {
		ctx.Error(fmt.Errorf("unable to get app: %w", err))
		return
	}

	if !awc.WorkflowConfigCanTriggerManually() {
		ctx.Error(stderr.ErrUser{
			Err:         fmt.Errorf("manual trigger is not allowed"),
			Description: "please update action config to allow manual triggering",
		})
		return
	}

	//
	installActionWorkflow, err := s.getInstallActionWorkflow(ctx, installID, awc.ActionWorkflowID)
	if err != nil {
		ctx.Error(errors.Wrap(err, "unable to get install action workflow"))
		return
	}

	run, err := s.createActionWorkflowRun(ctx, installID, installActionWorkflow.ID, awc, req.RunEnvVars)
	if err != nil {
		ctx.Error(fmt.Errorf("unable to create app: %w", err))
		return
	}

	s.evClient.Send(ctx, installID, &installsignals.Signal{
		ActionWorkflowRunID: run.ID,
		Type:                installsignals.OperationActionWorkflowRun,
	})

	ctx.JSON(http.StatusCreated, run)
}

func (s *service) createActionWorkflowRun(ctx *gin.Context, installID, installActionWorkflowID string, cfg *app.ActionWorkflowConfig, runEnvVars map[string]string) (*app.InstallActionWorkflowRun, error) {
	steps := make([]app.InstallActionWorkflowRunStep, 0)
	for _, step := range cfg.Steps {
		steps = append(steps, app.InstallActionWorkflowRunStep{
			Status: app.InstallActionWorkflowRunStepStatusPending,
			StepID: step.ID,
		})
	}

	trigger := app.InstallActionWorkflowManualTrigger{
		InstallActionWorkflowRun: app.InstallActionWorkflowRun{
			InstallID:               installID,
			InstallActionWorkflowID: installActionWorkflowID,
			ActionWorkflowConfigID:  cfg.ID,
			TriggerType:             app.ActionWorkflowTriggerTypeManual,
			Status:                  app.InstallActionRunStatusQueued,
			StatusDescription:       "Queued",
			Steps:                   steps,
			RunEnvVars:              generics.ToHstore(runEnvVars),
		},
	}

	res := s.db.WithContext(ctx).
		Create(&trigger)
	if res.Error != nil {
		return nil, fmt.Errorf("unable to create action workflow: %w", res.Error)
	}

	return &trigger.InstallActionWorkflowRun, nil
}
