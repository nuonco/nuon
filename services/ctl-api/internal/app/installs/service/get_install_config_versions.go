package service

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/cctx"
)

// @ID						GetInstallConfigVersions
// @Summary				get config versions for an install
// @Description			Returns the install config version history, ordered by most recent first.
// @Param					install_id	path	string	true	"install ID"
// @Tags					installs
// @Accept					json
// @Produce				json
// @Security				APIKey && OrgID
// @Failure				400	{object}	stderr.ErrResponse
// @Failure				401	{object}	stderr.ErrResponse
// @Failure				403	{object}	stderr.ErrResponse
// @Failure				404	{object}	stderr.ErrResponse
// @Failure				500	{object}	stderr.ErrResponse
// @Success				200	{array}		app.InstallConfigVersion
// @Router					/v1/installs/{install_id}/config-versions [GET]
func (s *service) GetInstallConfigVersions(ctx *gin.Context) {
	org, err := cctx.OrgFromContext(ctx)
	if err != nil {
		ctx.Error(err)
		return
	}

	installID := ctx.Param("install_id")

	var versions []app.InstallConfigVersion
	res := s.db.WithContext(ctx).
		Preload("InstallConfigSync").
		Preload("InstallConfigSync.VCSConnectionCommit").
		Where(app.InstallConfigVersion{
			InstallID: installID,
			OrgID:     org.ID,
		}).
		Order("created_at DESC").
		Limit(50).
		Find(&versions)
	if res.Error != nil {
		ctx.Error(fmt.Errorf("unable to get install config versions: %w", res.Error))
		return
	}

	ctx.JSON(http.StatusOK, versions)
}
