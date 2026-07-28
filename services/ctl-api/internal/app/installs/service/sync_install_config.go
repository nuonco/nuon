package service

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/installs/signals/installconfigsync"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/cctx"
	queueclient "github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/client"
)

// @ID						SyncInstallConfig
// @Summary				trigger install config sync for a single install
// @Description			Triggers a sync of this install's config from git.
// @Tags					installs
// @Accept					json
// @Param					install_id	path	string	true	"install ID"
// @Produce				json
// @Security				APIKey
// @Security				OrgID
// @Failure				400	{object}	stderr.ErrResponse
// @Failure				401	{object}	stderr.ErrResponse
// @Failure				403	{object}	stderr.ErrResponse
// @Failure				404	{object}	stderr.ErrResponse
// @Failure				500	{object}	stderr.ErrResponse
// @Success				202	{object}	map[string]string
// @Router					/v1/installs/{install_id}/sync-config [post]
func (s *service) SyncInstallConfig(ctx *gin.Context) {
	org, err := cctx.OrgFromContext(ctx)
	if err != nil {
		ctx.Error(err)
		return
	}

	installID := ctx.Param("install_id")

	var install app.Install
	if err := s.db.WithContext(ctx).
		Where(app.Install{ID: installID, OrgID: org.ID}).
		First(&install).Error; err != nil {
		ctx.Error(fmt.Errorf("unable to get install: %w", err))
		return
	}

	var branch app.AppBranch
	if err := s.db.WithContext(ctx).
		Preload("Configs", func(db *gorm.DB) *gorm.DB {
			return db.Order("created_at DESC").Limit(1)
		}).
		Where(app.AppBranch{AppID: install.AppID}).
		First(&branch).Error; err != nil {
		ctx.Error(fmt.Errorf("unable to find app branch: %w", err))
		return
	}

	if len(branch.Configs) == 0 {
		ctx.Error(fmt.Errorf("app branch has no config"))
		return
	}

	configID := branch.Configs[0].ID

	queue, err := s.queueClient.GetQueueByOwnerAndName(ctx, installID, "installs", "install-signals")
	if err != nil {
		ctx.Error(fmt.Errorf("unable to find queue for install: %w", err))
		return
	}

	_, err = s.queueClient.EnqueueSignal(ctx, &queueclient.EnqueueSignalRequest{
		QueueID:   queue.ID,
		OwnerID:   installID,
		OwnerType: "installs",
		Signal: &installconfigsync.Signal{
			InstallID:         install.ID,
			AppBranchID:       branch.ID,
			AppBranchConfigID: configID,
			TriggeredBy:       "manual",
		},
	})
	if err != nil {
		ctx.Error(fmt.Errorf("unable to enqueue install config sync: %w", err))
		return
	}

	ctx.JSON(http.StatusAccepted, map[string]string{
		"status": "enqueued",
	})
}
