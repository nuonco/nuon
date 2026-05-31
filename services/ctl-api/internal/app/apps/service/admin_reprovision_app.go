package service

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"

	appreprovision "github.com/nuonco/nuon/services/ctl-api/internal/app/apps/signals/reprovision"
	queueclient "github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/client"
)

type ReprovisionAppRequest struct{}

// @ID						AdminReprovisionApp
// @Summary				reprovision an app
// @Description.markdown	reprovision_app.md
// @Param					app_id	path	string	true	"app ID for your current app"
// @Tags					apps/admin
// @Security				AdminEmail
// @Accept					json
// @Param					req	body	ReprovisionAppRequest	true	"Input"
// @Produce				json
// @Success				201	{string}	ok
// @Router					/v1/apps/{app_id}/admin-reprovision [POST]
func (s *service) AdminReprovisionApp(ctx *gin.Context) {
	appID := ctx.Param("app_id")

	_, err := s.getApp(ctx, appID)
	if err != nil {
		ctx.Error(err)
		return
	}

	q, err := s.queueClient.GetQueueByOwner(ctx, appID, "apps")
	if err != nil {
		ctx.Error(fmt.Errorf("unable to get app queue: %w", err))
		return
	}
	if _, err := s.queueClient.EnqueueSignal(ctx, &queueclient.EnqueueSignalRequest{
		QueueID:   q.ID,
		OwnerID:   appID,
		OwnerType: "apps",
		Signal:    &appreprovision.Signal{AppID: appID},
	}); err != nil {
		ctx.Error(fmt.Errorf("unable to enqueue app reprovision signal: %w", err))
		return
	}
	ctx.JSON(http.StatusOK, true)
}
