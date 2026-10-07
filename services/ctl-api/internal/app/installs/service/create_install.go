package service

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"go.uber.org/zap"
	"gorm.io/gorm"

	"github.com/nuonco/nuon/pkg/lifecyclephase"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/installs/helpers"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/installs/signals/appconfigupdated"
	installscreated "github.com/nuonco/nuon/services/ctl-api/internal/app/installs/signals/created"
	polldependencies "github.com/nuonco/nuon/services/ctl-api/internal/app/installs/signals/polldependencies"
	"github.com/nuonco/nuon/services/ctl-api/internal/middlewares/stderr"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/cctx"
	executeflow "github.com/nuonco/nuon/services/ctl-api/internal/pkg/flow/signals/executeflow"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/request"
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
// @Failure				409	{object}	stderr.ErrResponse
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

	workflowRequest, done := s.returnExistingProvisionInstall(ctx, org.ID, req.AppID, &req.CreateInstallParams)
	if done {
		return
	}

	install, err := s.helpers.CreateInstall(ctx, req.AppID, &req.CreateInstallParams)
	if err != nil {
		ctx.Error(fmt.Errorf("unable to create install: %w", err))
		return
	}
	if workflowRequest != nil {
		workflowRequest.PinnedAppConfigID = install.AppConfigID
	}

	workflow, err := s.helpers.CreateWorkflow(ctx,
		install.ID,
		app.WorkflowTypeProvision,
		provisionWorkflowMetadata(req.StackOnly),
		false,
		workflowRequest,
	)
	if err != nil {
		ctx.Error(err)
		return
	}

	lp := lifecyclephase.New(lifecyclephase.Provisioning, provisionPhaseDescription(req.StackOnly))
	s.db.WithContext(ctx).Model(&app.Install{ID: install.ID}).Updates(map[string]any{
		"lifecycle_phase": lp,
	})

	// Send signals via queues
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
	if err := s.enqueueInstallSignal(ctx, workflowsQueueID, executeflow.NewSignal(workflow.ID), workflow.ID, "install_workflows"); err != nil {
		ctx.Error(fmt.Errorf("enqueue signal: %w", err))
		return
	}
	// reconcile cron/drift emitters from app config triggers
	if err := s.enqueueInstallSignal(ctx, signalsQueueID, &appconfigupdated.Signal{
		InstallID: install.ID,
	}, "", ""); err != nil {
		ctx.Error(fmt.Errorf("enqueue reconcile-emitters signal: %w", err))
		return
	}

	// Update user journey step for first install creation
	user, err := cctx.AccountFromGinContext(ctx)
	if err == nil {
		if err := s.accountsHelpers.UpdateUserJourneyStepForFirstInstallCreate(ctx, user.ID, install.ID); err != nil {
			s.l.Warn("failed to update user journey for first install create", zap.Error(err))
		}
	}

	install.WorkflowID = &workflow.ID
	ctx.JSON(http.StatusCreated, install)
}

type CreateInstallRequest struct {
	helpers.CreateInstallParams
}

func (c *CreateInstallRequest) Validate(v *validator.Validate) error {
	if err := v.Struct(c); err != nil {
		return validatorPkg.FormatValidationError(err)
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
// @Failure				409	{object}	stderr.ErrResponse
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

	workflowRequest, done := s.returnExistingProvisionInstall(ctx, org.ID, appID, &req.CreateInstallParams)
	if done {
		return
	}

	install, err := s.helpers.CreateInstall(ctx, appID, &req.CreateInstallParams)
	if err != nil {
		ctx.Error(fmt.Errorf("unable to create install: %w", err))
		return
	}
	if workflowRequest != nil {
		workflowRequest.PinnedAppConfigID = install.AppConfigID
	}

	workflow, err := s.helpers.CreateWorkflow(ctx,
		install.ID,
		app.WorkflowTypeProvision,
		provisionWorkflowMetadata(req.StackOnly),
		false,
		workflowRequest,
	)
	if err != nil {
		ctx.Error(err)
		return
	}

	lp2 := lifecyclephase.New(lifecyclephase.Provisioning, provisionPhaseDescription(req.StackOnly))
	s.db.WithContext(ctx).Model(&app.Install{ID: install.ID}).Updates(map[string]any{
		"lifecycle_phase": lp2,
	})

	// Send signals via queues
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
	if err := s.enqueueInstallSignal(ctx, workflowsQueueID, executeflow.NewSignal(workflow.ID), workflow.ID, "install_workflows"); err != nil {
		ctx.Error(fmt.Errorf("enqueue signal: %w", err))
		return
	}
	// reconcile cron/drift emitters from app config triggers
	if err := s.enqueueInstallSignal(ctx, signalsQueueID, &appconfigupdated.Signal{
		InstallID: install.ID,
	}, "", ""); err != nil {
		ctx.Error(fmt.Errorf("enqueue reconcile-emitters signal: %w", err))
		return
	}

	// Update user journey step for first install creation
	user, err := cctx.AccountFromGinContext(ctx)
	if err == nil {
		if err := s.accountsHelpers.UpdateUserJourneyStepForFirstInstallCreate(ctx, user.ID, install.ID); err != nil {
			s.l.Warn("failed to update user journey for first install create", zap.Error(err))
		}
	}

	install.WorkflowID = &workflow.ID
	ctx.JSON(http.StatusCreated, install)
}

func (s *service) returnExistingProvisionInstall(ctx *gin.Context, orgID, appID string, params *helpers.CreateInstallParams) (*app.WorkflowRequest, bool) {
	if params.RequestID == "" {
		return nil, false
	}
	hashParams := *params
	hashParams.RequestID = ""
	hashParams.SandboxMode = false
	hash, err := request.Hash(struct {
		AppID  string                      `json:"app_id"`
		Params helpers.CreateInstallParams `json:"params"`
	}{AppID: appID, Params: hashParams})
	if err != nil {
		ctx.Error(err)
		return nil, true
	}
	existing, err := s.lookupProvisionWorkflow(ctx, orgID, params.RequestID)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return &app.WorkflowRequest{
			RequestID:   params.RequestID,
			RequestHash: hash,
		}, false
	}
	if err != nil {
		ctx.Error(err)
		return nil, true
	}
	install, err := s.replayProvisionInstall(ctx, existing, hash)
	if err != nil {
		ctx.Error(err)
		return nil, true
	}
	ctx.JSON(http.StatusCreated, install)
	return nil, true
}

func (s *service) lookupProvisionWorkflow(ctx *gin.Context, orgID, requestID string) (*app.Workflow, error) {
	var existing app.Workflow
	err := s.db.WithContext(ctx).
		Where(app.Workflow{
			OrgID:     orgID,
			OwnerType: "installs",
			Type:      app.WorkflowTypeProvision,
		}).
		Where("request->>'request_id' = ?", requestID).
		First(&existing).Error
	if err != nil {
		return nil, err
	}
	return &existing, nil
}

func (s *service) replayProvisionInstall(ctx *gin.Context, workflow *app.Workflow, hash string) (*app.Install, error) {
	var install app.Install
	if err := s.db.WithContext(ctx).Where(app.Install{ID: workflow.OwnerID}).First(&install).Error; err != nil {
		return nil, fmt.Errorf("unable to get install: %w", err)
	}
	if err := request.Check(workflow.Request, hash, install.AppConfigID); err != nil {
		return nil, err
	}
	install.WorkflowID = &workflow.ID
	return &install, nil
}
