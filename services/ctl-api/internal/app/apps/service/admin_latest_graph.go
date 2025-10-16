package service

import (
	"bytes"
	"fmt"
	"net/http"

	"github.com/dominikbraun/graph/draw"

	"github.com/gin-gonic/gin"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/cctx"
)

type AdminAppConfigGraphRequest struct {
	ConfigID string
}

// @ID						AdminAppConfigGraphApp
// @Summary				get an app's graph
// @Description.markdown	app_config_graph.md
// @Tags					apps/admin
// @Security				AdminEmail
// @Accept					json
// @Param			req		body	AdminAppConfigGraphRequest	true	"Input"
// @Param				app_id	path	string					true	"app id"
// @Produce				json
// @Success				201	{string}	ok
// @Router					/v1/apps/{app_id}/admin-config-graph [POST]
func (s *service) AdminConfigGraph(ctx *gin.Context) {
	appID := ctx.Param("app_id")
	orgID, err := cctx.OrgIDFromContext(ctx)
	if err != nil {
		ctx.Error(err)
		return
	}

	var req AdminAppConfigGraphRequest
	if err := ctx.BindJSON(&req); err != nil {
		ctx.Error(fmt.Errorf("unable to parse request: %w", err))
		return
	}

	cfgID := req.ConfigID
	if cfgID == "" || cfgID == "string" {
		cfgs, err := s.getAppConfigs(ctx, orgID, appID)
		if err != nil {
			ctx.Error(err)
			return
		}
		cfgID = cfgs[0].ID
	}

	appCfg, err := s.helpers.GetFullAppConfig(ctx, cfgID, true)
	if err != nil {
		ctx.Error(err)
		return
	}

	graph, err := s.helpers.GetConfigGraph(ctx, appCfg)
	if err != nil {
		ctx.Error(err)
		return
	}

	// Create a buffer to store the DOT graph
	var buf bytes.Buffer
	if err := draw.DOT(graph, &buf, draw.GraphAttribute("name", "name")); err != nil {
		ctx.Error(err)
		return
	}

	ctx.String(http.StatusOK, buf.String())
}
