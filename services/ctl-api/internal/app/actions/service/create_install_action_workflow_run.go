package service

import (
	"context"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/pkg/errors"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app/installs/signals"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/middlewares/stderr"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/cctx"
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

// @ID						CreateInstallActionWorkflowRun
// @Summary				create an action workflow run for an install
// @Description.markdown	create_install_action_workflow_run.md
// @Tags					actions
// @Accept					json
// @Param					install_id	path	string									true	"install ID"
// @Param					req			body	CreateInstallActionWorkflowRunRequest	true	"Input"
// @Produce				json
// @Security				APIKey
// @Security				OrgID
// @Failure				400	{object}	stderr.ErrResponse
// @Failure				401	{object}	stderr.ErrResponse
// @Failure				403	{object}	stderr.ErrResponse
// @Failure				404	{object}	stderr.ErrResponse
// @Failure				500	{object}	stderr.ErrResponse
// @Success				201	{string}	ok
// @Router				/v1/installs/{install_id}/action-workflows/runs [post]
func (s *service) CreateInstallActionWorkflowRun(ctx *gin.Context) {
	installID := ctx.Param("install_id")

	var req CreateInstallActionWorkflowRunRequest
	if err := ctx.BindJSON(&req); err != nil {
		ctx.Error(fmt.Errorf("unable toq parse request: %w", err))
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

	installActionWorkflow, err := s.getInstallActionWorkflow(ctx, installID, awc.ActionWorkflowID)
	if err != nil {
		ctx.Error(errors.Wrap(err, "unable to get install action workflow"))
		return
	}

	prependRunEnvVars := PrependRunEnvPrefix(req.RunEnvVars)
	prependRunEnvVars["install_action_workflow_id"] = installActionWorkflow.ID
	account, err := cctx.AccountFromContext(ctx)
	if err != nil {
		ctx.Error(fmt.Errorf("unable to get account from context: %w", err))
		return
	}
	prependRunEnvVars["triggerred_by_id"] = account.ID

	workflow, err := s.CreateInstallWorkflow(ctx,
		installActionWorkflow.InstallID,
		app.InstallWorkflowTypeActionWorkflowRun,
		prependRunEnvVars,
		app.StepErrorBehaviorAbort,
	)
	if err != nil {
		ctx.Error(err)
		return
	}

	s.evClient.Send(ctx, installActionWorkflow.InstallID, &signals.Signal{
		Type:              signals.OperationExecuteWorkflow,
		InstallWorkflowID: workflow.ID,
	})

	ctx.Header(app.HeaderInstallWorkflowID, workflow.ID)

	ctx.JSON(http.StatusCreated, "ok")
}

// PrependRunEnvPrefix modifies the keys in the provided RunEnvVars map
// by prepending "RUNENV_" to each key.
func PrependRunEnvPrefix(runEnvVars map[string]string) map[string]string {
	result := make(map[string]string, len(runEnvVars))

	for key, value := range runEnvVars {
		newKey := "RUNENV_" + key
		result[newKey] = value
	}

	return result
}

func (s *service) CreateInstallWorkflow(ctx context.Context, installID string, workflowType app.InstallWorkflowType, metadata map[string]string, errBehavior app.StepErrorBehavior) (*app.InstallWorkflow, error) {
	installWorkflow := app.InstallWorkflow{
		Type:              workflowType,
		InstallID:         installID,
		Metadata:          generics.ToHstore(metadata),
		Status:            app.NewCompositeStatus(ctx, app.StatusPending),
		StepErrorBehavior: errBehavior,
	}

	res := s.db.WithContext(ctx).Create(&installWorkflow)
	if res.Error != nil {
		return nil, errors.Wrap(res.Error, "unable to create install workflow")
	}

	return &installWorkflow, nil
}
