package service

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	installhelpers "github.com/nuonco/nuon/services/ctl-api/internal/app/installs/helpers"
	"github.com/nuonco/nuon/services/ctl-api/internal/middlewares/stderr"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/request"
)

type ReconcileInstallAppVersionRequest struct {
	RequestID string `json:"request_id,omitempty" validate:"omitempty,max=255"`
	PlanOnly  bool   `json:"plan_only"`
}

// @ID					ReconcileInstallAppVersion
// @Summary				reconcile an install's applied app config to its pinned app config
// @Description			Creates a workflow that deploys the difference between the install's applied app config and its
// @Description			currently pinned app config (install.app_config_id). Branch runs with a do_not_deploy install group pin
// @Description			an install's app config without deploying it; this endpoint catches that install up. It generates the
// @Description			same steps and plan approvals an app branch config update would, minus install-group rollout
// @Description			coordination. Returns 409 if the install is already on its pinned config, or if a reconcile or app branch
// @Description			config update workflow is already queued or in progress for the install.
// @Param				install_id	path	string						true	"install ID"
// @Param				req			body	ReconcileInstallAppVersionRequest	true	"Input"
// @Tags				installs
// @Accept				json
// @Produce			json
// @Security			APIKey
// @Security			OrgID
// @Failure			400	{object}	stderr.ErrResponse
// @Failure			401	{object}	stderr.ErrResponse
// @Failure			403	{object}	stderr.ErrResponse
// @Failure			404	{object}	stderr.ErrResponse
// @Failure			409	{object}	stderr.ErrResponse
// @Failure			500	{object}	stderr.ErrResponse
// @Success			201	{object}	app.InstallAppConfigVersion
// @Router				/v1/installs/{install_id}/reconcile-app-version [post]
func (s *service) ReconcileInstallAppVersion(ctx *gin.Context) {
	installID := ctx.Param("install_id")

	var req ReconcileInstallAppVersionRequest
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

	if req.RequestID != "" {
		hashReq := req
		hashReq.RequestID = ""
		hash, err := request.Hash(hashReq)
		if err != nil {
			ctx.Error(err)
			return
		}

		var createdVersion *app.InstallAppConfigVersion
		wf, _, err := s.helpers.RunIdempotentInstallWorkflow(ctx, installhelpers.IdempotentInstallWorkflowRequest{
			InstallID:    install.ID,
			WorkflowType: app.WorkflowTypeReconcileAppVersion,
			Metadata: map[string]string{
				"new_app_config_id": install.AppConfigID,
			},
			PlanOnly:    req.PlanOnly,
			RequestID:   req.RequestID,
			RequestHash: hash,
			Operation:   "reconcile-app-version",
			QueueName:   installhelpers.InstallWorkflowsQueueName,
		}, &installhelpers.IdempotentInstallWorkflowHooks{
			AfterCreate: func(tx *gorm.DB, wf *app.Workflow) error {
				if err := s.rejectIfInSyncOrInFlight(ctx, tx, install); err != nil {
					return err
				}
				update := app.InstallAppConfigVersion{
					InstallID:      installID,
					OldAppConfigID: install.AppliedAppConfigID(),
					NewAppConfigID: install.AppConfigID,
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
			ctx.Error(reconcileErrToHTTP(err))
			return
		}

		if createdVersion != nil {
			ctx.JSON(http.StatusCreated, createdVersion)
			return
		}

		var update app.InstallAppConfigVersion
		if err := s.db.WithContext(ctx).
			Where("workflow_id = ?", wf.ID).
			First(&update).Error; err != nil {
			ctx.Error(fmt.Errorf("unable to load install config update: %w", err))
			return
		}
		ctx.JSON(http.StatusCreated, update)
		return
	}

	result, err := s.helpers.CreateReconcileAppVersionWorkflow(ctx, installhelpers.CreateReconcileAppVersionWorkflowInput{
		InstallID: installID,
		PlanOnly:  req.PlanOnly,
	})
	if err != nil {
		ctx.Error(reconcileErrToHTTP(err))
		return
	}

	var update app.InstallAppConfigVersion
	if err := s.db.WithContext(ctx).
		Where(app.InstallAppConfigVersion{ID: result.InstallAppConfigVersionID}).
		First(&update).Error; err != nil {
		ctx.Error(fmt.Errorf("unable to load install config update: %w", err))
		return
	}

	ctx.JSON(http.StatusCreated, update)
}

// rejectIfInSyncOrInFlight re-checks the already-in-sync and in-flight
// conditions inside the idempotent-create hook, after the idempotency lock is
// held, so a race right at request time still returns 409 instead of
// silently creating a no-op reconcile.
func (s *service) rejectIfInSyncOrInFlight(ctx *gin.Context, tx *gorm.DB, install *app.Install) error {
	if install.AppliedAppConfigID() == install.AppConfigID {
		return installhelpers.ErrInstallAlreadyOnPinnedConfig
	}
	return nil
}

func reconcileErrToHTTP(err error) error {
	if errors.Is(err, installhelpers.ErrInstallAlreadyOnPinnedConfig) {
		return stderr.ErrConflict{
			Err:         err,
			Description: "The install is already on its pinned app config; there is nothing to reconcile.",
		}
	}
	if errors.Is(err, installhelpers.ErrInstallConfigUpdateInFlight) {
		return stderr.ErrConflict{
			Err:         err,
			Description: "A reconcile or app branch config update workflow is already queued or in progress for this install.",
		}
	}
	return err
}
