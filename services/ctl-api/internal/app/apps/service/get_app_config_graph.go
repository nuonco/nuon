package service

import (
	"bytes"
	"fmt"
	"net/http"

	"github.com/dominikbraun/graph/draw"
	"github.com/gin-gonic/gin"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/middlewares/stderr"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/cctx"
)

// @ID						GetAppConfigGraphV2
// @Summary				get an app config graph
// @Description.markdown	get_app_config_graph.md
// @Param					app_id			path	string	true	"app ID"
// @Param					config_id	path	string	true	"app config ID"
// @Tags					apps
// @Accept					json
// @Produce				json
// @Security				APIKey
// @Security				OrgID
// @Failure				400	{object}	stderr.ErrResponse
// @Failure				401	{object}	stderr.ErrResponse
// @Failure				403	{object}	stderr.ErrResponse
// @Failure				404	{object}	stderr.ErrResponse
// @Failure				500	{object}	stderr.ErrResponse
// @Success				200	{object}	string
// @Router					/v1/apps/{app_id}/configs/{config_id}/graph [get]
func (s *service) GetAppConfigGraphV2(ctx *gin.Context) {
	appConfigID := ctx.Param("config_id")

	if err := s.requireAppConfigOwner(ctx, ctx.Param("app_id"), appConfigID); err != nil {
		ctx.Error(err)
		return
	}

	appConfig, err := s.helpers.GetFullAppConfig(ctx, appConfigID, true)
	if err != nil {
		ctx.Error(err)
		return
	}

	graph, err := s.helpers.GetConfigGraph(ctx, appConfig)
	if err != nil {
		ctx.Error(err)
		return
	}

	var buf bytes.Buffer
	if err := draw.DOT(graph, &buf, draw.GraphAttribute("name", "name")); err != nil {
		ctx.Error(err)
		return
	}

	ctx.JSON(http.StatusOK, buf.String())
}

// @ID						GetAppConfigGraph
// @Summary				get an app config graph
// @Description.markdown	get_app_config_graph.md
// @Param					app_id			path	string	true	"app ID"
// @Param					app_config_id	path	string	true	"app config ID"
// @Tags					apps
// @Accept					json
// @Produce				json
// @Security				APIKey
// @Security				OrgID
// @Deprecated    true
// @Failure				400	{object}	stderr.ErrResponse
// @Failure				401	{object}	stderr.ErrResponse
// @Failure				403	{object}	stderr.ErrResponse
// @Failure				404	{object}	stderr.ErrResponse
// @Failure				500	{object}	stderr.ErrResponse
// @Success				200	{object}	string
// @Router					/v1/apps/{app_id}/config/{app_config_id}/graph [get]
func (s *service) GetAppConfigGraph(ctx *gin.Context) {
	appConfigID := ctx.Param("app_config_id")

	if err := s.requireAppConfigOwner(ctx, ctx.Param("app_id"), appConfigID); err != nil {
		ctx.Error(err)
		return
	}

	appConfig, err := s.helpers.GetFullAppConfig(ctx, appConfigID, true)
	if err != nil {
		ctx.Error(err)
		return
	}

	graph, err := s.helpers.GetConfigGraph(ctx, appConfig)
	if err != nil {
		ctx.Error(err)
		return
	}

	var buf bytes.Buffer
	if err := draw.DOT(graph, &buf, draw.GraphAttribute("name", "name")); err != nil {
		ctx.Error(err)
		return
	}

	ctx.JSON(http.StatusOK, buf.String())
}

// requireAppConfigOwner confirms an app config id taken from the URL belongs to
// the app and org also named in the URL. GetFullAppConfig looks a config up by
// id alone, which is safe for callers passing an id read off an already-scoped
// install, but not for one supplied by the client.
func (s *service) requireAppConfigOwner(ctx *gin.Context, appID, appConfigID string) error {
	orgID, err := cctx.OrgIDFromContext(ctx)
	if err != nil {
		return err
	}

	var count int64
	res := s.db.WithContext(ctx).
		Model(&app.AppConfig{}).
		Where(app.AppConfig{
			ID:    appConfigID,
			AppID: appID,
			OrgID: orgID,
		}).
		Count(&count)
	if res.Error != nil {
		return fmt.Errorf("unable to resolve app config owner: %w", res.Error)
	}
	if count == 0 {
		return stderr.ErrNotFound{
			Err:         fmt.Errorf("app config %q not found for app %q", appConfigID, appID),
			Description: "app config not found",
		}
	}

	return nil
}
