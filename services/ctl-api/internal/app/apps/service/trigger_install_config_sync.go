package service

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"github.com/nuonco/nuon/pkg/config"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/installs/helpers"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/installs/signals/installconfigsync"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/blobstore"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/cctx"
	queueclient "github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/client"
)

type TriggerInstallConfigSyncRequest struct {
	InstallName string `json:"install_name,omitempty"`
}

// @ID						TriggerInstallConfigSync
// @Summary				trigger install config sync from git
// @Description			Triggers a sync of install configs from the installs.toml VCS repo configured in the app config. Optionally specify install_name to sync a single install.
// @Tags					apps
// @Accept					json
// @Param					req				body	TriggerInstallConfigSyncRequest	true	"Input"
// @Param					app_id			path	string							true	"app ID"
// @Param					app_branch_id	path	string							true	"app branch ID"
// @Produce				json
// @Security				APIKey && OrgID
// @Failure				400	{object}	stderr.ErrResponse
// @Failure				401	{object}	stderr.ErrResponse
// @Failure				403	{object}	stderr.ErrResponse
// @Failure				404	{object}	stderr.ErrResponse
// @Failure				500	{object}	stderr.ErrResponse
// @Success				202	{object}	map[string]string
// @Router					/v1/apps/{app_id}/branches/{app_branch_id}/sync-install-configs [post]
func (s *service) TriggerInstallConfigSync(ctx *gin.Context) {
	org, err := cctx.OrgFromContext(ctx)
	if err != nil {
		ctx.Error(err)
		return
	}

	appID := ctx.Param("app_id")
	branchID := ctx.Param("app_branch_id")

	var req TriggerInstallConfigSyncRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.Error(fmt.Errorf("unable to parse request: %w", err))
		return
	}

	var branch app.AppBranch
	if err := s.db.WithContext(ctx).
		Preload("Configs", func(db *gorm.DB) *gorm.DB {
			return db.Order("created_at DESC").Limit(1)
		}).
		Where(app.AppBranch{ID: branchID, AppID: appID, OrgID: org.ID}).
		First(&branch).Error; err != nil {
		ctx.Error(fmt.Errorf("unable to get app branch: %w", err))
		return
	}

	if len(branch.Configs) == 0 {
		ctx.Error(fmt.Errorf("app branch has no config"))
		return
	}

	configID := branch.Configs[0].ID

	var latestAppConfig app.AppConfig
	if err := s.db.WithContext(ctx).
		Where(app.AppConfig{AppID: appID}).
		Order("created_at DESC").
		First(&latestAppConfig).Error; err != nil {
		ctx.Error(fmt.Errorf("unable to get latest app config: %w", err))
		return
	}

	blobCtx := blobstore.WithBlobService(ctx.Request.Context(), s.blobSvc)
	intermediateJSON, err := latestAppConfig.IntermediateConfig.Get(blobCtx)
	if err != nil || intermediateJSON == "" {
		ctx.Error(fmt.Errorf("unable to read intermediate config: %w", err))
		return
	}

	var parsedConfig config.AppConfig
	if err := json.Unmarshal([]byte(intermediateJSON), &parsedConfig); err != nil {
		ctx.Error(fmt.Errorf("unable to parse intermediate config: %w", err))
		return
	}

	if parsedConfig.InstallsConfig == nil {
		ctx.Error(fmt.Errorf("no installs config found in app config - add an installs.toml with a connected_repo or public_repo"))
		return
	}
	if parsedConfig.InstallsConfig.ConnectedRepo == nil && parsedConfig.InstallsConfig.PublicRepo == nil {
		ctx.Error(fmt.Errorf("installs.toml has no connected_repo or public_repo configured"))
		return
	}

	var installs []app.Install
	query := s.db.WithContext(ctx).Where(app.Install{AppID: appID, OrgID: org.ID})
	if req.InstallName != "" {
		query = query.Where(app.Install{Name: req.InstallName})
	}
	if err := query.Find(&installs).Error; err != nil {
		ctx.Error(fmt.Errorf("unable to get installs: %w", err))
		return
	}

	if len(installs) == 0 {
		ctx.Error(fmt.Errorf("no installs found for app"))
		return
	}

	enqueued := 0
	for _, install := range installs {
		queue, err := s.queueClient.GetQueueByOwnerAndName(ctx, install.ID, "installs", helpers.InstallSignalsQueueName)
		if err != nil {
			continue
		}

		_, err = s.queueClient.EnqueueSignal(ctx, &queueclient.EnqueueSignalRequest{
			QueueID:   queue.ID,
			OwnerID:   install.ID,
			OwnerType: "installs",
			Signal: &installconfigsync.Signal{
				InstallID:         install.ID,
				AppBranchID:       branchID,
				AppBranchConfigID: configID,
				TriggeredBy:       "manual",
			},
		})
		if err != nil {
			continue
		}
		enqueued++
	}

	ctx.JSON(http.StatusAccepted, map[string]string{
		"status":   "enqueued",
		"enqueued": fmt.Sprintf("%d", enqueued),
	})
}
