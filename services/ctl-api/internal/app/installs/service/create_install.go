package service

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"go.uber.org/zap"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app/installs/helpers"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app/installs/signals"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/cctx"
	validatorPkg "github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/validator"
)

type CreateInstallV2Request struct {
	AppID string `json:"app_id" validate:"required"`
	helpers.CreateInstallParams
}

func (c *CreateInstallV2Request) Validate(v *validator.Validate) error {
	if err := v.Struct(c); err != nil {
		return validatorPkg.FormatValidationError(err)
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

// @ID						CreateInstallV2
// @Summary				create an app install
// @Description.markdown	create_install.md
// @Param					req		body	CreateInstallV2Request	true	"Input"
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
// @Router					/v1/installs [post]
func (s *service) CreateInstallV2(ctx *gin.Context) {
	var req CreateInstallV2Request
	if err := ctx.BindJSON(&req); err != nil {
		ctx.Error(fmt.Errorf("unable to parse request: %w", err))
		return
	}
	if err := req.Validate(s.v); err != nil {
		ctx.Error(fmt.Errorf("invalid request: %w", err))
		return
	}

	install, err := s.helpers.CreateInstall(ctx, req.AppID, &req.CreateInstallParams)
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

	// Update user journey step for first install creation
	user, err := cctx.AccountFromGinContext(ctx)
	if err == nil {
		// Only update if this is the user's first install (install_created step incomplete)
		if err := s.accountsHelpers.UpdateUserJourneyStepForFirstInstallCreate(ctx, user.ID, install.ID); err != nil {
			// Log but don't fail the install creation
			s.l.Warn("failed to update user journey for first install create", zap.Error(err))
		}
	}

	// TODO(jm): these will be deprecated after the workflow tooling is created
	ctx.JSON(http.StatusCreated, install)
}

type CreateInstallRequest struct {
	helpers.CreateInstallParams
}

func (c *CreateInstallRequest) Validate(v *validator.Validate) error {
	if err := v.Struct(c); err != nil {
		return validatorPkg.FormatValidationError(err)
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
// @Deprecated    true
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

	// Update user journey step for first install creation
	user, err := cctx.AccountFromGinContext(ctx)
	if err == nil {
		// Only update if this is the user's first install (install_created step incomplete)
		if err := s.accountsHelpers.UpdateUserJourneyStepForFirstInstallCreate(ctx, user.ID, install.ID); err != nil {
			// Log but don't fail the install creation
			s.l.Warn("failed to update user journey for first install create", zap.Error(err))
		}
	}

	// TODO(jm): these will be deprecated after the workflow tooling is created
	ctx.JSON(http.StatusCreated, install)
}
