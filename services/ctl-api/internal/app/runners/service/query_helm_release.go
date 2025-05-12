package service

import (
	"fmt"
	"net/http"
	"sort"

	"github.com/gin-gonic/gin"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/helm"
)

// @ID						QueryHelmRelease
// @Summary				query  helm releases
// @Description.markdown	query_helm_releases.md
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
// @Router					/v1/helm-releases/{helm_chart_id}/releases/{namespace}/query [get]

func (s *service) QueryHelmRelease(ctx *gin.Context) {
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

	queryParams := ctx.Request.URL.Query()

	keys := make([]string, 0, len(queryParams))
	for key := range queryParams {
		keys = append(keys, key)
	}
	sort.Strings(keys)

	helmReleases, err := s.queryHelmReleases(ctx, helmChartID, namespace, keys)
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

func (s *service) queryHelmReleases(ctx *gin.Context, helmChartID, namespace string, keys []string) ([]app.HelmRelease, error) {
	var labelMap = map[string]bool{
		"modifiedAt": true,
		"createdAt":  true,
		"version":    true,
		"status":     true,
		"owner":      true,
		"name":       true,
	}

	query := s.db.Model(&app.HelmRelease{}).
		Select("name", "namespace", "body").
		Where("helm_chart_id = ? and namespace = ?", helmChartID, namespace)

	queryParams := ctx.Request.URL.Query()
	for _, key := range keys {
		values := queryParams[key]
		if len(values) > 0 && labelMap[key] {
			query = query.Where(key+" = ?", values[0])
		} else if len(values) > 0 {
			return nil, fmt.Errorf("unknown label %s", key)
		}
	}

	var releases []app.HelmRelease
	if err := query.Find(&releases).Error; err != nil {
		return nil, fmt.Errorf("failed to query helm releases: %w", err)
	}

	return releases, nil
}
