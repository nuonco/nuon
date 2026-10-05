package service

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/cctx"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/db"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/db/scopes"
)

// @ID GetAppBundles
// @Summary list app bundles
// @Tags app-bundles
// @Produce json
// @Security APIKey
// @Security OrgID
// @Param app_id path string true "app ID"
// @Param app_config_id query string false "exact app config ID"
// @Param status query string false "publish status" Enums(queued,publishing,active,error)
// @Param offset query int false "pagination offset" Default(0)
// @Param limit query int false "page size" Default(10)
// @Success 200 {array} bundleResponse
// @Failure 400 {object} stderr.ErrResponse
// @Failure 500 {object} stderr.ErrResponse
// @Router /v1/apps/{app_id}/bundles [get]
func (s *service) ListBundles(ctx *gin.Context) {
	if !s.requireBundleExport(ctx) {
		return
	}
	org, err := cctx.OrgFromContext(ctx)
	if err != nil {
		ctx.Error(err)
		return
	}
	status := app.AppBundleStatus(ctx.Query("status"))
	switch status {
	case "", app.AppBundleStatusQueued, app.AppBundleStatusPublishing, app.AppBundleStatusActive, app.AppBundleStatusError:
	default:
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "status must be queued, publishing, active, or error"})
		return
	}
	var bundles []app.AppBundle
	res := s.db.WithContext(ctx).
		Scopes(scopes.WithOffsetPagination).
		Where(app.AppBundle{OrgID: org.ID, AppID: ctx.Param("app_id"), AppConfigID: ctx.Query("app_config_id"), Status: status}).
		Order("created_at DESC, id DESC").Find(&bundles)
	if res.Error != nil {
		ctx.Error(fmt.Errorf("unable to list app bundles: %w", res.Error))
		return
	}
	bundles, err = db.HandlePaginatedResponse(ctx, bundles)
	if err != nil {
		ctx.Error(err)
		return
	}
	response := make([]bundleResponse, 0, len(bundles))
	for _, bundle := range bundles {
		response = append(response, responseFromBundle(bundle))
	}
	ctx.JSON(http.StatusOK, response)
}
