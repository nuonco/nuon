package service

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/cctx"
)

// @ID ListInstallReleaseDeployments
// @Summary list immutable release deployment history for an install
// @Tags installs
// @Produce json
// @Security APIKey
// @Security OrgID
// @Param install_id path string true "Install ID"
// @Success 200 {array} app.InstallReleaseDeployment
// @Failure 404 {object} stderr.ErrResponse
// @Failure 500 {object} stderr.ErrResponse
// @Router /v1/installs/{install_id}/release-deployments [get]
func (s *service) ListReleaseDeployments(ctx *gin.Context) {
	if !s.customerManagedInstallsEnabled(ctx) {
		return
	}
	org, err := cctx.OrgFromContext(ctx)
	if err != nil {
		ctx.Error(err)
		return
	}
	var install app.Install
	if err := s.db.WithContext(ctx).Where(app.Install{ID: ctx.Param("install_id"), OrgID: org.ID}).First(&install).Error; err != nil {
		ctx.Error(err)
		return
	}
	var deployments []app.InstallReleaseDeployment
	if err := s.db.WithContext(ctx).Where(app.InstallReleaseDeployment{OrgID: org.ID, InstallID: install.ID}).Order("finished_at DESC, created_at DESC").Find(&deployments).Error; err != nil {
		ctx.Error(err)
		return
	}
	ctx.JSON(http.StatusOK, deployments)
}
