package service

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/cctx"
)

// @ID						GetInstallAppConfigVersionDiff
// @Summary				get the diff for an install app config version
// @Description			Returns the component diff for a specific app config version transition.
// @Param					install_id	path	string	true	"install ID"
// @Param					version_id	path	string	true	"app config version ID"
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
// @Success				200	{object}	app.InstallConfigDiff
// @Router					/v1/installs/{install_id}/app-config-versions/{version_id}/diff [GET]
func (s *service) GetInstallAppConfigVersionDiff(ctx *gin.Context) {
	org, err := cctx.OrgFromContext(ctx)
	if err != nil {
		ctx.Error(err)
		return
	}

	installID := ctx.Param("install_id")
	versionID := ctx.Param("version_id")

	var version app.InstallAppConfigVersion
	if err := s.db.WithContext(ctx.Request.Context()).
		Where(app.InstallAppConfigVersion{
			ID:        versionID,
			InstallID: installID,
			OrgID:     org.ID,
		}).
		First(&version).Error; err != nil {
		ctx.Error(fmt.Errorf("unable to get install app config version: %w", err))
		return
	}

	if version.Diff == nil {
		ctx.JSON(http.StatusOK, &app.InstallConfigDiff{
			Added:     []app.ComponentDiffEntry{},
			Removed:   []app.ComponentDiffEntry{},
			Changed:   []app.ComponentDiffEntry{},
			Unchanged: []app.ComponentDiffEntry{},
		})
		return
	}

	diffJSON, err := version.Diff.Get(ctx.Request.Context())
	if err != nil {
		ctx.Error(fmt.Errorf("unable to load diff blob: %w", err))
		return
	}

	var diff app.InstallConfigDiff
	if err := json.Unmarshal([]byte(diffJSON), &diff); err != nil {
		ctx.Error(fmt.Errorf("unable to parse diff: %w", err))
		return
	}

	ctx.JSON(http.StatusOK, &diff)
}
