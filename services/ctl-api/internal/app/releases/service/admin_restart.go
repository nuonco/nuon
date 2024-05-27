package service

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/app/releases/signals"
)

type RestartReleaseReleaseRequest struct{}

// @ID AdminRestartRelease
// @Summary	restart an releases event loop
// @Description	restart_release.md
// @Param			release_id	path	string							true	"release ID"
// @Param			req			body	RestartReleaseReleaseRequest	true	"Input"
// @Tags			releases/admin
// @Accept			json
// @Produce		json
// @Success		200	{boolean}	true
// @Router			/v1/releases/{release_id}/admin-restart [POST]
func (s *service) RestartRelease(ctx *gin.Context) {
	releaseID := ctx.Param("release_id")

	var req RestartReleaseReleaseRequest
	if err := ctx.BindJSON(&req); err != nil {
		ctx.Error(fmt.Errorf("unable to parse request: %w", err))
		return
	}

	release, err := s.getRelease(ctx, releaseID)
	if err != nil {
		ctx.Error(fmt.Errorf("unable to get release: %w", err))
		return
	}

	s.evClient.Send(ctx, release.ID, &signals.Signal{
		Type: signals.OperationRestart,
	})
	ctx.JSON(http.StatusOK, true)
}
