package service

import (
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/helm"
)

type CreateHelmReleaseRequest any

// @ID						CreateHelmRelease
// @Summary				create  helm release
// @Description.markdown	create_helm_release.md
// @Tags					runners/runner
// @Param					helm_chart_id	path	string					true	"helm chart ID"
// @Param					namespace	path	string					true	"namespace"
// @Param					key		path	string					true	"key"
// @Param					req	body	CreateHelmReleaseRequest	true	"Input"
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
// @Router					/v1/helm-releases/{helm_chart_id}/releases/{namespace}/{key} [post]
func (s *service) CreateHelmRelease(ctx *gin.Context) {
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

	helmRelease, err := s.createHelmRelease(ctx, helmChartID, namespace, key, &rls)
	if err != nil {
		ctx.Error(fmt.Errorf("unable to create helm release: %w", err))
		return
	}

	ctx.JSON(http.StatusCreated, helmRelease)
}

func (s *service) createHelmRelease(ctx *gin.Context, helmChartID, namespace, key string, rls *helm.Release) (any, error) {
	body, err := helm.EncodeRelease(rls)
	if err != nil {
		return nil, fmt.Errorf("unable to encode release: %w", err)
	}

	helmRelease := app.HelmRelease{
		HelmChartID: helmChartID,
		Key:         key,
		Type:        "helm.sh/release.v1",
		Body:        body,
		Name:        rls.Name,
		Namespace:   namespace,
		Version:     int(rls.Version),
		Status:      rls.Info.Status.String(),
		Owner:       "helm",
		CreatedAt:   time.Now().UTC(),
		Labels:      rls.Labels,
	}

	res := s.db.WithContext(ctx).Model(&app.HelmRelease{}).Create(&helmRelease)
	if res.Error != nil {
		return nil, fmt.Errorf("unable to create helm release: %w", res.Error)
	}

	return helmRelease, nil
}
