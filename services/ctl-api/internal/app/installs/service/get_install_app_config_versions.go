package service

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/cctx"
)

// @ID						GetInstallAppConfigVersions
// @Summary				get app config versions for an install
// @Description			Returns the app config version history for an install, ordered by most recent first.
// @Param					install_id	path	string	true	"install ID"
// @Tags					installs
// @Accept					json
// @Produce				json
// @Security				APIKey
// @Security				OrgID
// @Failure				400	{object}	stderr.ErrResponse
// @Failure				401	{object}	stderr.ErrResponse
// @Failure				403	{object}	stderr.ErrResponse
// @Failure				404	{object}	stderr.ErrResponse
// @Failure				500	{object}	stderr.ErrResponse
// @Success				200	{array}		app.InstallAppConfigVersion
// @Router					/v1/installs/{install_id}/app-config-versions [GET]
func (s *service) GetInstallAppConfigVersions(ctx *gin.Context) {
	org, err := cctx.OrgFromContext(ctx)
	if err != nil {
		ctx.Error(err)
		return
	}

	installID := ctx.Param("install_id")

	versions, err := s.getInstallAppConfigVersions(ctx, installID, org.ID)
	if err != nil {
		ctx.Error(fmt.Errorf("unable to get install app config versions: %w", err))
		return
	}

	ctx.JSON(http.StatusOK, versions)
}

func (s *service) getInstallAppConfigVersions(ctx *gin.Context, installID, orgID string) ([]app.InstallAppConfigVersion, error) {
	var versions []app.InstallAppConfigVersion
	res := s.db.WithContext(ctx).
		Where(app.InstallAppConfigVersion{
			InstallID: installID,
			OrgID:     orgID,
		}).
		Order("created_at DESC").
		Limit(50).
		Find(&versions)
	if res.Error != nil {
		return nil, fmt.Errorf("unable to get install app config versions: %w", res.Error)
	}

	return versions, nil
}
