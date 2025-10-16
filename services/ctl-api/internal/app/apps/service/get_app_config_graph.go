package service

import (
	"bytes"
	"net/http"

	"github.com/dominikbraun/graph/draw"
	"github.com/gin-gonic/gin"
)

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
// @Failure				400	{object}	stderr.ErrResponse
// @Failure				401	{object}	stderr.ErrResponse
// @Failure				403	{object}	stderr.ErrResponse
// @Failure				404	{object}	stderr.ErrResponse
// @Failure				500	{object}	stderr.ErrResponse
// @Success				200	{object}	string
// @Router					/v1/apps/{app_id}/config/{app_config_id}/graph [get]
func (s *service) GetAppConfigGraph(ctx *gin.Context) {
	appConfigID := ctx.Param("app_config_id")

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
