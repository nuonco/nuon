package service

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

// @ID        GetAppSandboxBuild
// @Summary   get app sandbox build
// @Tags      apps
// @Accept    json
// @Produce   json
// @Param     app_id    path  string  true  "app ID"
// @Param     build_id  path  string  true  "sandbox build ID"
// @Security  APIKey
// @Security  OrgID
// @Failure   400  {object}  stderr.ErrResponse
// @Failure   401  {object}  stderr.ErrResponse
// @Failure   404  {object}  stderr.ErrResponse
// @Failure   500  {object}  stderr.ErrResponse
// @Success   200  {object}  app.AppSandboxBuild
// @Router    /v1/apps/{app_id}/sandbox/builds/{build_id} [get]
func (s *service) GetAppSandboxBuild(ctx *gin.Context) {
	appID := ctx.Param("app_id")
	buildID := ctx.Param("build_id")

	_, err := s.getApp(ctx, appID)
	if err != nil {
		ctx.Error(fmt.Errorf("unable to get app: %w", err))
		return
	}

	var build app.AppSandboxBuild
	res := s.db.WithContext(ctx).
		Preload("CreatedBy").
		Preload("LogStream").
		Preload("RunnerJob").
		Preload("VCSConnectionCommit").
		Where("app_id = ?", appID).
		First(&build, "id = ?", buildID)
	if res.Error != nil {
		ctx.Error(fmt.Errorf("unable to get sandbox build: %w", res.Error))
		return
	}

	ctx.JSON(http.StatusOK, build)
}
