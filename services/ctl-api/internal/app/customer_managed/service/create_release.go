package service

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"

	customermanaged "github.com/nuonco/nuon/pkg/customer_managed"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/releases"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/cctx"
)

type createReleaseRequest struct {
	AppConfigID string                            `json:"app_config_id" binding:"required"`
	Runbooks    []customermanaged.RunbookTemplate `json:"runbooks,omitempty"`
}

// @ID CreateAppRelease
// @Summary create an immutable application release
// @Tags releases
// @Accept json
// @Produce json
// @Security APIKey
// @Security OrgID
// @Param app_id path string true "app ID"
// @Param request body createReleaseRequest true "release request"
// @Success 200 {object} app.AppRelease
// @Success 201 {object} app.AppRelease
// @Failure 400 {object} stderr.ErrResponse
// @Failure 403 {object} stderr.ErrResponse
// @Failure 500 {object} stderr.ErrResponse
// @Router /v1/apps/{app_id}/releases [post]
func (s *service) CreateRelease(ctx *gin.Context) {
	if !s.customerManagedInstallsEnabled(ctx) {
		return
	}
	var req createReleaseRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.Error(fmt.Errorf("invalid request: %w", err))
		return
	}
	org, err := cctx.OrgFromContext(ctx)
	if err != nil {
		ctx.Error(err)
		return
	}
	release, created, err := releases.CreateAppRelease(ctx, s.db, s.appsHelpers, s.blobSvc, org.ID, ctx.Param("app_id"), req.AppConfigID, req.Runbooks)
	if err != nil {
		ctx.Error(fmt.Errorf("unable to create app release: %w", err))
		return
	}
	status := http.StatusOK
	if created {
		status = http.StatusCreated
	}
	ctx.JSON(status, release)
}
