package service

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/middlewares/stderr"
	executeflow "github.com/nuonco/nuon/services/ctl-api/internal/pkg/flow/signals/executeflow"
)

type CreateInstallAppConfigUpdateRequest struct {
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
// @Failure				500	{object}	stderr.ErrResponse
// @Success				201	{object}	app.InstallConfigUpdate
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
	)
	if err != nil {
		ctx.Error(fmt.Errorf("unable to create workflow: %w", err))
		return
	}

	// Create the InstallConfigUpdate tracking record
	update := app.InstallConfigUpdate{
		InstallID:      installID,
		OldAppConfigID: install.AppConfigID,
		NewAppConfigID: req.AppConfigID,
		WorkflowID:     &wf.ID,
		Status:         app.NewCompositeStatus(ctx, app.StatusPending),
	}
	if err := s.db.WithContext(ctx).Create(&update).Error; err != nil {
		ctx.Error(fmt.Errorf("unable to create install config update: %w", err))
		return
	}

	// Enqueue the workflow on the install's queue
	queueID, err := s.getInstallWorkflowsQueueID(ctx, installID)
	if err != nil {
		ctx.Error(err)
		return
	}
	if err := s.enqueueInstallSignal(ctx, queueID, &executeflow.Signal{
		WorkflowID: wf.ID,
	}, wf.ID, "install_workflows"); err != nil {
		ctx.Error(fmt.Errorf("enqueue signal: %w", err))
		return
	}

	ctx.JSON(http.StatusCreated, update)
}
