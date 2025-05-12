package service

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
	"gorm.io/gorm"
)

// @ID						DeleteHelmRelease
// @Summary				delete  helm release
// @Description.markdown	delete_helm_release.md
// @Tags					runners/runner
// @Param					helm_chart_id	path	string					true	"helm chart ID"
// @Param					namespace	path	string					true	"namespace"
// @Param					key		path	string					true	"key"
// @Accept					json
// @Produce				json
// @Security				APIKey
// @Security				OrgID
// @Failure				400	{object}	stderr.ErrResponse
// @Failure				401	{object}	stderr.ErrResponse
// @Failure				403	{object}	stderr.ErrResponse
// @Failure				404	{object}	stderr.ErrResponse
// @Failure				500	{object}	stderr.ErrResponse
// @Success				200	{array}	app.HelmRelease
// @Router					/v1/helm-releases/{helm_chart_id}/releases/{namespace}/{key} [delete]
func (s *service) DeleteHelmRelease(ctx *gin.Context) {
	helmChartID := ctx.Param("helm_chart_id")
	namespace := ctx.Param("namespace")
	key := ctx.Param("key")

	if helmChartID == "" {
		ctx.Error(fmt.Errorf("helm_chart_id was not set"))
		return
	}

	if namespace == "" {
		ctx.Error(fmt.Errorf("namespace was not set"))
		return
	}

	if key == "" {
		ctx.Error(fmt.Errorf("key was not set"))
		return
	}

	err := s.deleteHelmRelease(ctx, helmChartID, namespace, key)
	if err != nil && err == gorm.ErrRecordNotFound {
		ctx.JSON(http.StatusNotFound, nil)
		return
	} else if err != nil {
		ctx.Error(fmt.Errorf("unable to delete helm release: %w", err))
		return
	}

	ctx.JSON(http.StatusOK, nil)
}

func (s *service) deleteHelmRelease(ctx *gin.Context, installID, namespace, key string) error {
	res := s.db.WithContext(ctx).Model(&app.HelmRelease{}).
		Where("helm_chart_id = ? and namespace = ? and key = ?", installID, namespace, key).
		Delete(&app.HelmRelease{})
	if res.Error != nil {
		return res.Error
	}
	return nil

}
