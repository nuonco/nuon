package service

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/db/scopes"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/helm"
)

// @ID						GetHelmRelease
// @Summary				get  helm release
// @Description.markdown	get_helm_release.md
// @Tags					runners/runner
// @Param					helm_chart_id	path	string					true	"helm chart ID"
// @Param					namespace	path	string					true	"namespace"
// @Param					key	path	string					true	"key"
// @Param					offset						query	int		false	"offset of results to return"	Default(0)
// @Param					limit						query	int		false	"limit of results to return"	Default(10)
// @Param					page						query	int		false	"page number of results to return"	Default(0)
// @Accept					json
// @Produce				json
// @Security				APIKey
// @Security				OrgID
// @Failure				400	{object}	stderr.ErrResponse
// @Failure				401	{object}	stderr.ErrResponse
// @Failure				403	{object}	stderr.ErrResponse
// @Failure				404	{object}	stderr.ErrResponse
// @Failure				500	{object}	stderr.ErrResponse
// @Success				200	{object}	helm.Release
// @Router					/v1/helm-releases/{helm_chart_id}/releases/{namespace}/{key} [get]
func (s *service) GetHelmRelease(ctx *gin.Context) {
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

	helmRelease, err := s.getHelmRelease(ctx, helmChartID, namespace, key)
	if err != nil {
		ctx.Error(fmt.Errorf("unable to list workspaces: %w", err))
		return
	}

	release, err := helm.DecodeRelease(helmRelease.Body)
	if err != nil {
		ctx.Error(fmt.Errorf("unable to decode helm release: %w", err))
		return
	}

	release.Labels = helmRelease.Labels

	ctx.JSON(http.StatusOK, release)
}

func (s *service) getHelmRelease(ctx *gin.Context, helmChartID, namespace, key string) (*app.HelmRelease, error) {
	release := app.HelmRelease{}
	res := s.db.WithContext(ctx).Model(&app.HelmRelease{}).
		Scopes(scopes.WithOffsetPagination).Where("helm_chart_id = ? and namespace = ? and key = ?", helmChartID, namespace, key).
		Find(&release)
	if res.Error != nil {
		return nil, res.Error
	}

	return &release, nil
}
