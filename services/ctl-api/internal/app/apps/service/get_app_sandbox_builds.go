package service

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/db"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/db/scopes"
)

// @ID        GetAppSandboxBuilds
// @Summary   get app sandbox builds
// @Tags      apps
// @Accept    json
// @Produce   json
// @Param     app_id  path   string  true   "app ID"
// @Param     offset  query  int     false  "offset of results to return"  Default(0)
// @Param     limit   query  int     false  "limit of results to return"   Default(10)
// @Security  APIKey
// @Security  OrgID
// @Failure   400  {object}  stderr.ErrResponse
// @Failure   401  {object}  stderr.ErrResponse
// @Failure   404  {object}  stderr.ErrResponse
// @Failure   500  {object}  stderr.ErrResponse
// @Success   200  {array}   app.AppSandboxBuild
// @Router    /v1/apps/{app_id}/sandbox/builds [get]
func (s *service) GetAppSandboxBuilds(ctx *gin.Context) {
	appID := ctx.Param("app_id")

	_, err := s.getApp(ctx, appID)
	if err != nil {
		ctx.Error(fmt.Errorf("unable to get app: %w", err))
		return
	}

	var builds []app.AppSandboxBuild
	res := s.db.WithContext(ctx).
		Scopes(scopes.WithOffsetPagination).
		Preload("CreatedBy").
		Preload("LogStream").
		Preload("RunnerJob").
		Preload("VCSConnectionCommit").
		Where("app_id = ?", appID).
		Order("created_at DESC").
		Find(&builds)
	if res.Error != nil {
		ctx.Error(fmt.Errorf("unable to get sandbox builds: %w", res.Error))
		return
	}

	builds, err = db.HandlePaginatedResponse(ctx, builds)
	if err != nil {
		ctx.Error(fmt.Errorf("unable to handle paginated response: %w", err))
		return
	}

	ctx.JSON(http.StatusOK, builds)
}
