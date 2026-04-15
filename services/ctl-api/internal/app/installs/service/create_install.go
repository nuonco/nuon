package service

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"go.uber.org/zap"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/installs/helpers"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/installs/signals"
	installscreated "github.com/nuonco/nuon/services/ctl-api/internal/app/installs/signals/v2/created"
	polldependencies "github.com/nuonco/nuon/services/ctl-api/internal/app/installs/signals/v2/polldependencies"
	"github.com/nuonco/nuon/services/ctl-api/internal/middlewares/stderr"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/cctx"
	executeflow "github.com/nuonco/nuon/services/ctl-api/internal/pkg/flow/signals/executeflow"
	validatorPkg "github.com/nuonco/nuon/services/ctl-api/internal/pkg/validator"
)

type CreateInstallV2Request struct {
	AppID string `json:"app_id" validate:"required"`
	helpers.CreateInstallParams
}

func (c *CreateInstallV2Request) Validate(v *validator.Validate) error {
	if err := v.Struct(c); err != nil {
		return validatorPkg.FormatValidationError(err)
	}

	if c.AWSAccount == nil && c.AzureAccount == nil && c.GCPAccount == nil {
		return stderr.ErrUser{
			Description: "one of AWSAccount, AzureAccount, or GCPAccount must be provided",
			Err:         fmt.Errorf("one of AWSAccount, AzureAccount, or GCPAccount must be provided"),
		}
	}

	if c.AWSAccount != nil {
		if c.AWSAccount.Region == "" {
			return stderr.ErrUser{
				Description: "AWSAccount region is required",
				Err:         fmt.Errorf("AWSAccount region is required"),
			}
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
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.Error(stderr.NewInvalidRequest(err))
		return
	}
	if err := req.Validate(s.v); err != nil {
		ctx.Error(fmt.Errorf("invalid request: %w", err))
		return
	}

	org, err := cctx.OrgFromContext(ctx)
	if err != nil {
		ctx.Error(fmt.Errorf("unable to get org: %w", err))
		return
	}
	req.SandboxMode = org.SandboxMode

	install, err := s.helpers.CreateInstall(ctx, req.AppID, &req.CreateInstallParams)
	if err != nil {
		ctx.Error(fmt.Errorf("unable to create install: %w", err))
		return
	}

	workflow, err := s.helpers.CreateWorkflow(ctx,
		install.ID,
		app.WorkflowTypeProvision,
		map[string]string{},
		false,
	)
	if err != nil {
		ctx.Error(err)
		return
	}

	// Send signals: v2 queues or legacy event loop
	useQueues, err := s.featuresClient.AllFeaturesEnabled(ctx, app.OrgFeatureAppBranches, app.OrgFeatureQueues)
	if err != nil {
		ctx.Error(fmt.Errorf("checking features: %w", err))
		return
	}
	if useQueues {
		signalsQueueID, err := s.getInstallSignalsQueueID(ctx, install.ID)
		if err != nil {
			ctx.Error(err)
			return
		}
		workflowsQueueID, err := s.getInstallWorkflowsQueueID(ctx, install.ID)
		if err != nil {
			ctx.Error(err)
			return
		}
		if err := s.enqueueInstallSignal(ctx, signalsQueueID, &installscreated.Signal{
			InstallID: install.ID,
		}, "", ""); err != nil {
			ctx.Error(fmt.Errorf("enqueue signal: %w", err))
			return
		}
		if err := s.enqueueInstallSignal(ctx, signalsQueueID, &polldependencies.Signal{
			InstallID: install.ID,
		}, "", ""); err != nil {
			ctx.Error(fmt.Errorf("enqueue signal: %w", err))
			return
		}
		if err := s.enqueueInstallSignal(ctx, workflowsQueueID, &executeflow.Signal{
			WorkflowID: workflow.ID,
		}, workflow.ID, "install_workflows"); err != nil {
			ctx.Error(fmt.Errorf("enqueue signal: %w", err))
			return
		}
	} else {
		s.evClient.Send(ctx, install.ID, &signals.Signal{
			Type: signals.OperationCreated,
		})
		s.evClient.Send(ctx, install.ID, &signals.Signal{
			Type: signals.OperationPollDependencies,
		})
		s.evClient.Send(ctx, install.ID, &signals.Signal{
			Type:              signals.OperationExecuteFlow,
			InstallWorkflowID: workflow.ID,
		})
	}
	// SyncActionWorkflowTriggers must stay legacy - it starts a child workflow in the event loop
	s.evClient.Send(ctx, install.ID, &signals.Signal{
		Type: signals.OperationSyncActionWorkflowTriggers,
	})

	ctx.Header(app.HeaderInstallWorkflowID, workflow.ID)

	// Update user journey step for first install creation
	user, err := cctx.AccountFromGinContext(ctx)
	if err == nil {
		if err := s.accountsHelpers.UpdateUserJourneyStepForFirstInstallCreate(ctx, user.ID, install.ID); err != nil {
			s.l.Warn("failed to update user journey for first install create", zap.Error(err))
		}
	}

	ctx.JSON(http.StatusCreated, install)
}

type CreateInstallRequest struct {
	helpers.CreateInstallParams
}

func (c *CreateInstallRequest) Validate(v *validator.Validate) error {
	if err := v.Struct(c); err != nil {
		return validatorPkg.FormatValidationError(err)
	}

	if c.AWSAccount == nil && c.AzureAccount == nil && c.GCPAccount == nil {
		return stderr.ErrUser{
			Description: "one of AWSAccount, AzureAccount, or GCPAccount must be provided",
			Err:         fmt.Errorf("one of AWSAccount, AzureAccount, or GCPAccount must be provided"),
		}
	}

	if c.AWSAccount != nil {
		if c.AWSAccount.Region == "" {
			return stderr.ErrUser{
				Description: "AWSAccount region is required",
				Err:         fmt.Errorf("AWSAccount region is required"),
			}
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
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.Error(stderr.NewInvalidRequest(err))
		return
	}
	if err := req.Validate(s.v); err != nil {
		ctx.Error(fmt.Errorf("invalid request: %w", err))
		return
	}

	org, err := cctx.OrgFromContext(ctx)
	if err != nil {
		ctx.Error(fmt.Errorf("unable to get org: %w", err))
		return
	}
	req.SandboxMode = org.SandboxMode

	install, err := s.helpers.CreateInstall(ctx, appID, &req.CreateInstallParams)
	if err != nil {
		ctx.Error(fmt.Errorf("unable to create install: %w", err))
		return
	}

	workflow, err := s.helpers.CreateWorkflow(ctx,
		install.ID,
		app.WorkflowTypeProvision,
		map[string]string{},
		false,
	)
	if err != nil {
		ctx.Error(err)
		return
	}

	// Send signals: v2 queues or legacy event loop
	useQueues, err := s.featuresClient.AllFeaturesEnabled(ctx, app.OrgFeatureAppBranches, app.OrgFeatureQueues)
	if err != nil {
		ctx.Error(fmt.Errorf("checking features: %w", err))
		return
	}
	if useQueues {
		signalsQueueID, err := s.getInstallSignalsQueueID(ctx, install.ID)
		if err != nil {
			ctx.Error(err)
			return
		}
		workflowsQueueID, err := s.getInstallWorkflowsQueueID(ctx, install.ID)
		if err != nil {
			ctx.Error(err)
			return
		}
		if err := s.enqueueInstallSignal(ctx, signalsQueueID, &installscreated.Signal{
			InstallID: install.ID,
		}, "", ""); err != nil {
			ctx.Error(fmt.Errorf("enqueue signal: %w", err))
			return
		}
		if err := s.enqueueInstallSignal(ctx, signalsQueueID, &polldependencies.Signal{
			InstallID: install.ID,
		}, "", ""); err != nil {
			ctx.Error(fmt.Errorf("enqueue signal: %w", err))
			return
		}
		if err := s.enqueueInstallSignal(ctx, workflowsQueueID, &executeflow.Signal{
			WorkflowID: workflow.ID,
		}, workflow.ID, "install_workflows"); err != nil {
			ctx.Error(fmt.Errorf("enqueue signal: %w", err))
			return
		}
	} else {
		s.evClient.Send(ctx, install.ID, &signals.Signal{
			Type: signals.OperationCreated,
		})
		s.evClient.Send(ctx, install.ID, &signals.Signal{
			Type: signals.OperationPollDependencies,
		})
		s.evClient.Send(ctx, install.ID, &signals.Signal{
			Type:              signals.OperationExecuteFlow,
			InstallWorkflowID: workflow.ID,
		})
	}
	// SyncActionWorkflowTriggers must stay legacy - it starts a child workflow in the event loop
	s.evClient.Send(ctx, install.ID, &signals.Signal{
		Type: signals.OperationSyncActionWorkflowTriggers,
	})

	ctx.Header(app.HeaderInstallWorkflowID, workflow.ID)

	// Update user journey step for first install creation
	user, err := cctx.AccountFromGinContext(ctx)
	if err == nil {
		if err := s.accountsHelpers.UpdateUserJourneyStepForFirstInstallCreate(ctx, user.ID, install.ID); err != nil {
			s.l.Warn("failed to update user journey for first install create", zap.Error(err))
		}
	}

	ctx.JSON(http.StatusCreated, install)
}
