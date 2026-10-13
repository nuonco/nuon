package service

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	installhelpers "github.com/nuonco/nuon/services/ctl-api/internal/app/installs/helpers"
	"github.com/nuonco/nuon/services/ctl-api/internal/middlewares/stderr"
	executeflow "github.com/nuonco/nuon/services/ctl-api/internal/pkg/flow/signals/executeflow"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/request"
)

type CreateInstallAppConfigUpdateRequest struct {
	RequestID   string `json:"request_id,omitempty" validate:"omitempty,max=255"`
	AppConfigID string `json:"app_config_id" validate:"required"`
	PlanOnly    bool   `json:"plan_only"`
}

// @ID						CreateInstallAppConfigUpdate
// @Summary				trigger an app config update for an install
// @Description			Creates a workflow to diff and deploy a new app config to an install.
// @Param					install_id	path	string									true	"install ID"
// @Param					req			body	CreateInstallAppConfigUpdateRequest		true	"Input"
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
// @Success				201	{object}	app.InstallAppConfigVersion
// @Router					/v1/installs/{install_id}/app-config-updates [post]
func (s *service) CreateInstallAppConfigUpdate(ctx *gin.Context) {
	installID := ctx.Param("install_id")

	var req CreateInstallAppConfigUpdateRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.Error(stderr.NewInvalidRequest(err))
		return
	}
	if err := s.v.Struct(req); err != nil {
		ctx.Error(stderr.NewInvalidRequest(err))
		return
	}

	install, err := s.getInstall(ctx, installID)
	if err != nil {
		ctx.Error(err)
		return
	}

	// Verify the app config exists
	var appConfig app.AppConfig
	if err := s.db.WithContext(ctx).First(&appConfig, "id = ?", req.AppConfigID).Error; err != nil {
		ctx.Error(fmt.Errorf("unable to find app config: %w", err))
		return
	}

	if req.RequestID != "" {
		hashReq := req
		hashReq.RequestID = ""
		hash, err := request.Hash(hashReq)
		if err != nil {
			ctx.Error(err)
			return
		}

		var createdVersion *app.InstallAppConfigVersion
		workflow, _, err := s.helpers.RunIdempotentInstallWorkflow(ctx, installhelpers.IdempotentInstallWorkflowRequest{
			InstallID:    install.ID,
			WorkflowType: app.WorkflowTypeAppBranchConfigUpdate,
			Metadata: map[string]string{
				"new_app_config_id": req.AppConfigID,
			},
			PlanOnly:    req.PlanOnly,
			RequestID:   req.RequestID,
			RequestHash: hash,
			Operation:   "app-config-update",
			QueueName:   installhelpers.InstallWorkflowsQueueName,
		}, &installhelpers.IdempotentInstallWorkflowHooks{
			AfterCreate: func(tx *gorm.DB, wf *app.Workflow) error {
				update := app.InstallAppConfigVersion{
					InstallID:      installID,
					OldAppConfigID: install.AppConfigID,
					NewAppConfigID: req.AppConfigID,
					WorkflowID:     &wf.ID,
					Status:         app.NewCompositeStatus(ctx, app.StatusPending),
				}
				if err := installhelpers.CreateInstallAppConfigVersionRow(ctx, tx, &update); err != nil {
					return fmt.Errorf("unable to create install config update: %w", err)
				}
				createdVersion = &update
				return nil
			},
		})
		if err != nil {
			ctx.Error(err)
			return
		}

		if createdVersion != nil {
			ctx.JSON(http.StatusCreated, createdVersion)
			return
		}

		var update app.InstallAppConfigVersion
		if err := s.db.WithContext(ctx).
			Where("workflow_id = ?", workflow.ID).
			First(&update).Error; err != nil {
			ctx.Error(fmt.Errorf("unable to load install config update: %w", err))
			return
		}
		ctx.JSON(http.StatusCreated, update)
		return
	}

	// Create the install workflow
	metadata := map[string]string{
		"new_app_config_id": req.AppConfigID,
	}

	wf, err := s.helpers.CreateWorkflow(
		ctx,
		installID,
		app.WorkflowTypeAppBranchConfigUpdate,
		metadata,
		req.PlanOnly,
		nil,
	)
	if err != nil {
		ctx.Error(fmt.Errorf("unable to create workflow: %w", err))
		return
	}

	// Create the InstallAppConfigVersion tracking record
	update := app.InstallAppConfigVersion{
		InstallID:      installID,
		OldAppConfigID: install.AppConfigID,
		NewAppConfigID: req.AppConfigID,
		WorkflowID:     &wf.ID,
		Status:         app.NewCompositeStatus(ctx, app.StatusPending),
	}
	if err := installhelpers.CreateInstallAppConfigVersionRow(ctx, s.db, &update); err != nil {
		ctx.Error(fmt.Errorf("unable to create install config update: %w", err))
		return
	}

	// Enqueue the workflow on the install's queue
	queueID, err := s.getInstallWorkflowsQueueID(ctx, installID)
	if err != nil {
		ctx.Error(err)
		return
	}
	if err := s.enqueueInstallSignal(ctx, queueID, executeflow.NewSignal(wf.ID), wf.ID, "install_workflows"); err != nil {
		ctx.Error(fmt.Errorf("enqueue signal: %w", err))
		return
	}

	ctx.JSON(http.StatusCreated, update)
}
