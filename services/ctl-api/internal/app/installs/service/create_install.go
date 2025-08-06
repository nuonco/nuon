package service

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app/installs/helpers"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app/installs/signals"
)

type CreateInstallRequest struct {
	helpers.CreateInstallParams
}

func (c *CreateInstallRequest) Validate(v *validator.Validate) error {
	if err := v.Struct(c); err != nil {
		return fmt.Errorf("invalid request: %w", err)
	}

	if c.AWSAccount == nil && c.AzureAccount == nil {
		return fmt.Errorf("either AWSAccount or AzureAccount must be provided")
	}

	if c.AWSAccount != nil {
		if c.AWSAccount.Region == "" {
			return fmt.Errorf("AWSAccount region is required")
		}
	}

	return nil
}

// @ID						CreateInstall
// @Summary				create an app install
// @Description.markdown	create_install.md
// @Param					app_id	path	string					true	"app ID"
// @Param					req		body	CreateInstallRequest	true	"Input"
// @Tags					installs
// @Accept					json
// @Produce				json
// @Security				APIKey
// @Security				OrgID
// @Failure				400	{object}	stderr.ErrResponse
// @Failure				401	{object}	stderr.ErrResponse
// @Failure				403	{object}	stderr.ErrResponse
// @Failure				404	{object}	stderr.ErrResponse
// @Failure				500	{object}	stderr.ErrResponse
// @Success				201	{object}	app.Install
// @Router					/v1/apps/{app_id}/installs [post]
func (s *service) CreateInstall(ctx *gin.Context) {
	appID := ctx.Param("app_id")

	var req CreateInstallRequest
	if err := ctx.BindJSON(&req); err != nil {
		ctx.Error(fmt.Errorf("unable to parse request: %w", err))
		return
	}
	if err := req.Validate(s.v); err != nil {
		ctx.Error(fmt.Errorf("invalid request: %w", err))
		return
	}

	install, err := s.helpers.CreateInstall(ctx, appID, &req.CreateInstallParams)
	if err != nil {
		ctx.Error(fmt.Errorf("unable to create install: %w", err))
		return
	}

	// NOTE(jm): eventually, we may want to move these into the workflow itself, but for now they are really system
	// details so we're not including them in the user facing workflows.
	//
	// Maybe at some point they would be added with a `UserFacing: false` boolean on the step itself.
	s.evClient.Send(ctx, install.ID, &signals.Signal{
		Type: signals.OperationCreated,
	})
	s.evClient.Send(ctx, install.ID, &signals.Signal{
		Type: signals.OperationPollDependencies,
	})
	s.evClient.Send(ctx, install.ID, &signals.Signal{
		Type: signals.OperationSyncActionWorkflowTriggers,
	})

	workflow, err := s.helpers.CreateWorkflow(ctx,
		install.ID,
		app.WorkflowTypeProvision,
		map[string]string{},
		app.StepErrorBehaviorAbort,
		false,
	)
	if err != nil {
		ctx.Error(err)
		return
	}

	s.evClient.Send(ctx, install.ID, &signals.Signal{
		Type:              signals.OperationExecuteFlow,
		InstallWorkflowID: workflow.ID,
	})

	ctx.Header(app.HeaderInstallWorkflowID, workflow.ID)

	// TODO(jm): these will be deprecated after the workflow tooling is created
	ctx.JSON(http.StatusCreated, install)
}
