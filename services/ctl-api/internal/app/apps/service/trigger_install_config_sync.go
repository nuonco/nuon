package service

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/apps/signals/branches/installconfigsync"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/cctx"
	queueclient "github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/client"
)

type TriggerInstallConfigSyncRequest struct {
	InstallName string `json:"install_name,omitempty"`
}

// @ID						TriggerInstallConfigSync
// @Summary				trigger install config sync from git
// @Description			Triggers a sync of install configs from the configured installs VCS repo. Optionally specify install_name to sync a single install.
// @Tags					apps
// @Accept					json
// @Param					req				body	TriggerInstallConfigSyncRequest	true	"Input"
// @Param					app_id			path	string							true	"app ID"
// @Param					app_branch_id	path	string							true	"app branch ID"
// @Produce				json
// @Security				APIKey
// @Security				OrgID
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
		Preload("Configs.InstallsConnectedGithubVCSConfig").
		Preload("Configs.InstallsPublicGitVCSConfig").
		Preload("Configs.ConnectedGithubVCSConfig").
		Preload("Configs.PublicGitVCSConfig").
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

	queue, err := s.queueClient.GetQueueByOwner(ctx, branchID, "app_branches")
	if err != nil {
		ctx.Error(fmt.Errorf("unable to find queue for app branch: %w", err))
		return
	}

	_, err = s.queueClient.EnqueueSignal(ctx, &queueclient.EnqueueSignalRequest{
		QueueID:   queue.ID,
		OwnerID:   branchID,
		OwnerType: "app_branches",
		Signal: &installconfigsync.Signal{
			AppBranchID:       branchID,
			AppBranchConfigID: configID,
			InstallName:       req.InstallName,
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
