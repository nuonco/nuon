package service

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/blobstore"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/cctx"
)

// @ID						GetInstallConfigVersionDiff
// @Summary				get the diff for an install config version
// @Description			Returns the config diff for a specific install config version.
// @Param					install_id	path	string	true	"install ID"
// @Param					version_id	path	string	true	"config version ID"
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
// @Success				200	{object}	map[string]interface{}
// @Router					/v1/installs/{install_id}/config-versions/{version_id}/diff [GET]
func (s *service) GetInstallConfigVersionDiff(ctx *gin.Context) {
	org, err := cctx.OrgFromContext(ctx)
	if err != nil {
		ctx.Error(err)
		return
	}

	installID := ctx.Param("install_id")
	versionID := ctx.Param("version_id")

	var version app.InstallConfigVersion
	if err := s.db.WithContext(ctx.Request.Context()).
		Where(app.InstallConfigVersion{
			ID:        versionID,
			InstallID: installID,
			OrgID:     org.ID,
		}).
		First(&version).Error; err != nil {
		ctx.Error(fmt.Errorf("unable to get install config version: %w", err))
		return
	}

	if version.Diff == nil || !version.Diff.IsSet() {
		ctx.JSON(http.StatusOK, map[string]any{})
		return
	}

	blobCtx := blobstore.WithBlobService(ctx.Request.Context(), s.blobSvc)
	diffJSON, err := version.Diff.Get(blobCtx)
	if err != nil {
		ctx.Error(fmt.Errorf("unable to load diff blob: %w", err))
		return
	}

	ctx.Data(http.StatusOK, "application/json", []byte(diffJSON))
}
