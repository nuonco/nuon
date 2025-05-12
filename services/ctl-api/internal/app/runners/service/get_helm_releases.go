package service

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/helm"
)

// @ID						GetHelmReleases
// @Summary				get  helm releases
// @Description.markdown	get_helm_releases.md
// @Tags					runners/runner
// @Param					helm_chart_id	path	string					true	"helm chart ID"
// @Param					namespace	path	string					true	"namespace"
// @Accept					json
// @Produce				json
// @Security				APIKey
// @Security				OrgID
// @Failure				400	{object}	stderr.ErrResponse
// @Failure				401	{object}	stderr.ErrResponse
// @Failure				403	{object}	stderr.ErrResponse
// @Failure				404	{object}	stderr.ErrResponse
// @Failure				500	{object}	stderr.ErrResponse
// @Success				200	{array}	helm.Release
// @Router					/v1/helm-releases/{helm_chart_id}/releases/{namespace} [get]
func (s *service) GetHelmReleases(ctx *gin.Context) {
	helmChartID := ctx.Param("helm_chart_id")
	namespace := ctx.Param("namespace")

	if helmChartID == "" {
		ctx.Error(fmt.Errorf("helm_chart_id was not set"))
		return
	}

	if namespace == "" {
		ctx.Error(fmt.Errorf("namespace was not set"))
		return
	}

	helmReleases, err := s.listHelmReleases(ctx, helmChartID, namespace)
	if err != nil {
		ctx.Error(fmt.Errorf("unable to list workspaces: %w", err))
		return
	}

	releases := make([]helm.Release, 0, len(helmReleases))
	for _, helmRelease := range helmReleases {
		release, err := helm.DecodeRelease(helmRelease.Body)
		if err != nil {
			ctx.Error(fmt.Errorf("unable to decode helm release: %w", err))
			return
		}

		release.Labels = helmRelease.Labels
		releases = append(releases, *release)
	}

	ctx.JSON(http.StatusOK, releases)
}

func (s *service) listHelmReleases(ctx *gin.Context, helmChartID, namespace string) ([]app.HelmRelease, error) {
	releases := []app.HelmRelease{}
	res := s.db.WithContext(ctx).Model(&app.HelmRelease{}).
		Where("helm_chart_id = ? and namespace = ?", helmChartID, namespace).
		Find(&releases)
	if res.Error != nil {
		return nil, res.Error
	}

	return releases, nil

}
