package service

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
)

type RestartReleaseReleaseRequest struct{}

//	@BasePath	/v1/releases
//
// Restart an release's event loop
//
//	@Summary	restart an releases event loop
//	@Schemes
//	@Description	restart release event loop
//	@Param			release_id	path	string							true	"release ID"
//	@Param			req			body	RestartReleaseReleaseRequest	true	"Input"
//	@Tags			releases/admin
//	@Accept			json
//	@Produce		json
//	@Success		200	{boolean}	true
//	@Router			/v1/releases/{release_id}/admin-restart [POST]
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

	org, err := s.getOrg(ctx, release.OrgID)
	if err != nil {
		ctx.Error(fmt.Errorf("unable to get release org: %w", err))
		return
	}

	s.hooks.Restart(ctx, release.ID, org.SandboxMode)
	ctx.JSON(http.StatusOK, true)
}
