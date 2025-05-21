package service

import (
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/helm"
)

// @ID						UpdateHelmRelease
// @Summary				update  helm release
// @Description.markdown	update_helm_release.md
// @Tags					runners/runner
// @Param					helm_chart_id	path	string					true	"helm chart ID"
// @Param					namespace	path	string					true	"namespace"
// @Param					key		path	string					true	"key"
// @Param					req	body	UpdateHelmReleaseRequest	true	"Input"
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
// @Router					/v1/helm-releases/{helm_chart_id}/releases/{namespace}/{key} [put]

func (s *service) UpdateHelmRelease(ctx *gin.Context) {
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

	var rls helm.Release
	if err := ctx.ShouldBindJSON(&rls); err != nil {
		ctx.Error(fmt.Errorf("failed to bind request body: %w", err))
		return
	}

	err := s.updateHelmRelease(ctx, helmChartID, namespace, key, &rls)
	if err != nil {
		ctx.Error(fmt.Errorf("unable to update helm release: %w", err))
		return
	}

	ctx.JSON(http.StatusOK, nil)

}

func (s *service) updateHelmRelease(ctx *gin.Context, helmChartID, namespace, key string, rls *helm.Release) error {
	body, err := helm.EncodeRelease(rls)
	if err != nil {
		return fmt.Errorf("unable to encode release: %w", err)
	}

	helmRelease := app.HelmRelease{
		Body:      body,
		Name:      rls.Name,
		Version:   rls.Version,
		Status:    string(rls.Info.Status),
		Owner:     "helm",
		UpdatedAt: time.Now().UTC(),
	}

	res := s.db.WithContext(ctx).Model(&app.HelmRelease{}).
		Where("helm_chart_id = ? and namespace = ? and key = ?", helmChartID, namespace, key).
		Updates(&helmRelease)
	if res.Error != nil {
		return res.Error
	}

	return nil

}
