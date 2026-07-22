package service

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"go.uber.org/zap"
	"gorm.io/gorm"

	"github.com/nuonco/nuon/pkg/config"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	appshelpers "github.com/nuonco/nuon/services/ctl-api/internal/app/apps/helpers"
	"github.com/nuonco/nuon/services/ctl-api/internal/middlewares/stderr"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/blobstore"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/cctx"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/config/syncer/triggers"
	validatorPkg "github.com/nuonco/nuon/services/ctl-api/internal/pkg/validator"
)

type UpdateAppConfigRequest struct {
	Status            app.AppConfigStatus `json:"status"`
	StatusDescription string              `json:"status_description"`
	State             string              `json:"state"`
	ComponentIDs      []string            `json:"component_ids"`
}

func (c *UpdateAppConfigRequest) Validate(v *validator.Validate) error {
	if err := v.Struct(c); err != nil {
		return validatorPkg.FormatValidationError(err)
	}

	return nil
}

// @ID						UpdateAppConflgV2
// @Description.markdown	update_app_config.md
// @Tags					apps
// @Accept					json
// @Param					req	body	UpdateAppConfigRequest	true	"Input"
// @Produce				json
// @Param					app_id			path	string	true	"app ID"
// @Param					config_id	path	string	true	"app config ID"
// @Security				APIKey
// @Security				OrgID
// @Failure				400	{object}	stderr.ErrResponse
// @Failure				401	{object}	stderr.ErrResponse
// @Failure				403	{object}	stderr.ErrResponse
// @Failure				404	{object}	stderr.ErrResponse
// @Failure				500	{object}	stderr.ErrResponse
// @Success				201	{object}	app.AppConfig
// @Router					/v1/apps/{app_id}/configs/{config_id} [PATCH]
func (s *service) UpdateAppConfigV2(ctx *gin.Context) {
	appConfigID := ctx.Param("config_id")
	org, err := cctx.OrgFromContext(ctx)
	if err != nil {
		ctx.Error(err)
		return
	}

	var req UpdateAppConfigRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.Error(stderr.NewInvalidRequest(err))
		return
	}
	if err := req.Validate(s.v); err != nil {
		ctx.Error(fmt.Errorf("invalid request: %w", err))
		return
	}

	cfg, err := s.updateAppConfig(ctx, org.ID, ctx.Param("app_id"), appConfigID, &req)
	if err != nil {
		ctx.Error(fmt.Errorf("unable to update app config: %w", err))
		return
	}

	// Update journey step when config becomes active (app sync complete)
	if req.Status == app.AppConfigStatusActive {
		if acct, err := cctx.AccountFromGinContext(ctx); err == nil {
			if err := s.accountsHelpers.UpdateUserJourneyStepForFirstAppSync(ctx, acct.ID, cfg.AppID); err != nil {
				s.l.Warn("failed to update app_synced journey step",
					zap.String("account_id", acct.ID),
					zap.String("app_id", cfg.AppID),
					zap.Error(err),
				)
			}
		}

		// Trigger app branch run if config was synced targeting a branch,
		// otherwise sync non-branch-managed installs to the new config.
		if cfg.AppBranchID.Valid && cfg.AppBranchID.String != "" {
			s.triggerAppBranchRunForConfig(ctx, cfg)
		} else {
			s.emitSyncAppConfigInstallsSignal(ctx, cfg.AppID, cfg.ID)
		}
	}

	ctx.JSON(http.StatusCreated, cfg)
}

// @ID						UpdateAppConfig
// @Description.markdown	update_app_config.md
// @Tags					apps
// @Accept					json
// @Param					req	body	UpdateAppConfigRequest	true	"Input"
// @Produce				json
// @Param					app_id			path	string	true	"app ID"
// @Param					app_config_id	path	string	true	"app config ID"
// @Security				APIKey
// @Security				OrgID
// @Deprecated    true
// @Failure				400	{object}	stderr.ErrResponse
// @Failure				401	{object}	stderr.ErrResponse
// @Failure				403	{object}	stderr.ErrResponse
// @Failure				404	{object}	stderr.ErrResponse
// @Failure				500	{object}	stderr.ErrResponse
// @Success				201	{object}	app.AppConfig
// @Router					/v1/apps/{app_id}/config/{app_config_id} [PATCH]
func (s *service) UpdateAppConfig(ctx *gin.Context) {
	appConfigID := ctx.Param("app_config_id")
	org, err := cctx.OrgFromContext(ctx)
	if err != nil {
		ctx.Error(err)
		return
	}

	var req UpdateAppConfigRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.Error(stderr.NewInvalidRequest(err))
		return
	}
	if err := req.Validate(s.v); err != nil {
		ctx.Error(fmt.Errorf("invalid request: %w", err))
		return
	}

	cfg, err := s.updateAppConfig(ctx, org.ID, ctx.Param("app_id"), appConfigID, &req)
	if err != nil {
		ctx.Error(fmt.Errorf("unable to update app config: %w", err))
		return
	}

	// Update journey step when config becomes active (app sync complete)
	if req.Status == app.AppConfigStatusActive {
		if acct, err := cctx.AccountFromGinContext(ctx); err == nil {
			if err := s.accountsHelpers.UpdateUserJourneyStepForFirstAppSync(ctx, acct.ID, cfg.AppID); err != nil {
				s.l.Warn("failed to update app_synced journey step",
					zap.String("account_id", acct.ID),
					zap.String("app_id", cfg.AppID),
					zap.Error(err),
				)
			}
		}

		if cfg.AppBranchID.Valid && cfg.AppBranchID.String != "" {
			s.triggerAppBranchRunForConfig(ctx, cfg)
		} else {
			s.emitSyncAppConfigInstallsSignal(ctx, cfg.AppID, cfg.ID)
		}
	}

	ctx.JSON(http.StatusCreated, cfg)
}

func (s *service) updateAppConfig(ctx context.Context, orgID, appID, appConfigID string, req *UpdateAppConfigRequest) (*app.AppConfig, error) {
	dbCtx := blobstore.WithBlobService(ctx, s.blobSvc)
	var cfg app.AppConfig
	if err := s.db.WithContext(dbCtx).
		Where(app.AppConfig{ID: appConfigID, OrgID: orgID, AppID: appID}).
		First(&cfg).Error; err != nil {
		return nil, fmt.Errorf("unable to find app config: %w", err)
	}

	updateStatus := app.NewCompositeStatus(ctx, app.Status(req.Status))
	updateStatus.StatusHumanDescription = req.StatusDescription

	appConfig := app.AppConfig{
		Status:            req.Status,
		StatusDescription: req.StatusDescription,
		StatusV2:          updateStatus,
		State:             req.State,
	}

	if len(req.ComponentIDs) > 0 {
		appConfig.ComponentIDs = req.ComponentIDs
	}

	if err := s.db.WithContext(dbCtx).Transaction(func(tx *gorm.DB) error {
		if req.Status == app.AppConfigStatusActive {
			if err := s.syncTriggersForAppConfig(dbCtx, tx, &cfg); err != nil {
				return fmt.Errorf("unable to sync triggers: %w", err)
			}
		}

		res := tx.Model(&app.AppConfig{}).
			Where(app.AppConfig{ID: appConfigID, OrgID: orgID, AppID: appID}).
			Updates(appConfig)
		if res.Error != nil {
			return fmt.Errorf("unable to update app config: %w", res.Error)
		}
		if res.RowsAffected < 1 {
			return fmt.Errorf("app config not found %s %w", appConfigID, gorm.ErrRecordNotFound)
		}
		return nil
	}); err != nil {
		return nil, err
	}

	return &cfg, nil
}

func (s *service) syncTriggersForAppConfig(ctx context.Context, db *gorm.DB, appConfig *app.AppConfig) error {
	if appConfig.IntermediateConfig == nil || !appConfig.IntermediateConfig.IsSet() {
		return nil
	}
	intermediateJSON, err := appConfig.IntermediateConfig.Get(ctx)
	if err != nil {
		return fmt.Errorf("load intermediate config: %w", err)
	}
	if intermediateJSON == "" {
		return nil
	}
	var cfg config.AppConfig
	decoder := json.NewDecoder(strings.NewReader(intermediateJSON))
	decoder.UseNumber()
	if err := decoder.Decode(&cfg); err != nil {
		return fmt.Errorf("decode intermediate config: %w", err)
	}
	return triggers.Sync(ctx, db, &cfg, appConfig.OrgID, appConfig.AppID, appConfig.ID)
}

// triggerAppBranchRunForConfig triggers an app branch run if the config
// has an AppBranchID set (i.e., it was synced targeting a branch).
func (s *service) triggerAppBranchRunForConfig(ctx context.Context, cfg *app.AppConfig) {
	if !cfg.AppBranchID.Valid || cfg.AppBranchID.String == "" {
		return
	}

	branchID := cfg.AppBranchID.String

	// Load the branch with its queue and latest config
	var branch app.AppBranch
	if err := s.db.WithContext(ctx).
		Preload("Queue").
		Preload("Configs", func(db *gorm.DB) *gorm.DB {
			return db.Order("config_number DESC").Limit(1)
		}).
		First(&branch, "id = ?", branchID).Error; err != nil {
		s.l.Warn("unable to load app branch for auto-trigger",
			zap.String("app_branch_id", branchID),
			zap.Error(err))
		return
	}

	if branch.Queue.ID == "" || len(branch.Configs) == 0 {
		s.l.Warn("app branch missing queue or config, skipping auto-trigger",
			zap.String("app_branch_id", branchID))
		return
	}

	triggerResp, err := s.helpers.TriggerAppBranchRun(ctx, &appshelpers.TriggerAppBranchRunRequest{
		Run: appshelpers.CreateAppBranchRunRequest{
			AppBranchID:       branchID,
			AppBranchConfigID: branch.Configs[0].ID,
			EventType:         "sync",
		},
		QueueID: branch.Queue.ID,
		Metadata: map[string]string{
			"config_id": branch.Configs[0].ID,
		},
	})
	if err != nil {
		s.l.Warn("unable to trigger app branch run after sync",
			zap.String("app_branch_id", branchID),
			zap.Error(err))
		return
	}

	s.l.Info("triggered app branch run after sync",
		zap.String("app_branch_id", branchID),
		zap.String("run_id", triggerResp.Run.ID),
		zap.String("workflow_id", triggerResp.Workflow.ID))
}
