package service

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	sandboxbuildsignal "github.com/nuonco/nuon/services/ctl-api/internal/app/apps/signals/v2/sandboxbuild"
	queueclient "github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/client"
)

// @ID        CreateAppSandboxBuild
// @Summary   create app sandbox build
// @Tags      apps
// @Accept    json
// @Produce   json
// @Param     app_id  path  string  true  "app ID"
// @Security  APIKey
// @Security  OrgID
// @Failure   400  {object}  stderr.ErrResponse
// @Failure   401  {object}  stderr.ErrResponse
// @Failure   404  {object}  stderr.ErrResponse
// @Failure   500  {object}  stderr.ErrResponse
// @Success   201  {object}  app.AppSandboxBuild
// @Router    /v1/apps/{app_id}/sandbox/builds [post]
func (s *service) CreateAppSandboxBuild(ctx *gin.Context) {
	appID := ctx.Param("app_id")

	currentApp, err := s.getApp(ctx, appID)
	if err != nil {
		ctx.Error(fmt.Errorf("unable to get app: %w", err))
		return
	}

	// Get the latest app config
	var latestConfig app.AppConfig
	if res := s.db.WithContext(ctx).
		Where("app_id = ?", currentApp.ID).
		Order("created_at DESC").
		First(&latestConfig); res.Error != nil {
		ctx.Error(fmt.Errorf("no app config found for app %s: %w", appID, res.Error))
		return
	}

	// Get the latest sandbox config
	if len(currentApp.AppSandboxConfigs) == 0 {
		ctx.Error(fmt.Errorf("no sandbox config found for app %s", appID))
		return
	}
	latestSandboxConfig := currentApp.AppSandboxConfigs[0]

	// Create the build record immediately so the caller gets an ID back
	build := app.AppSandboxBuild{
		AppID:              currentApp.ID,
		AppConfigID:        latestConfig.ID,
		AppSandboxConfigID: latestSandboxConfig.ID,
		OrgID:              currentApp.OrgID,
		Status:             app.AppSandboxBuildStatusQueued,
		StatusDescription:  "queued and waiting for runner",
	}
	if res := s.db.WithContext(ctx).Create(&build); res.Error != nil {
		ctx.Error(fmt.Errorf("unable to create sandbox build: %w", res.Error))
		return
	}

	// Get the app's sandbox queue
	q, err := s.queueClient.GetQueueByOwner(ctx, currentApp.ID, "apps")
	if err != nil {
		ctx.Error(fmt.Errorf("unable to get sandbox queue for app %s: %w", appID, err))
		return
	}

	// Enqueue the signal with the pre-created build ID
	_, err = s.queueClient.EnqueueSignal(ctx, &queueclient.EnqueueSignalRequest{
		QueueID: q.ID,
		Signal: &sandboxbuildsignal.Signal{
			AppConfigID:       latestConfig.ID,
			AppSandboxBuildID: build.ID,
		},
	})
	if err != nil {
		ctx.Error(fmt.Errorf("unable to enqueue sandbox build signal: %w", err))
		return
	}

	ctx.JSON(http.StatusCreated, build)
}
