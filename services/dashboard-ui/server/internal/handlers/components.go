package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"github.com/nuonco/nuon/services/dashboard-ui/server/internal/pkg/cctx"
)

type ComponentsHandler struct {
	l *zap.Logger
}

func NewComponentsHandler(l *zap.Logger) *ComponentsHandler {
	return &ComponentsHandler{l: l}
}

func (h *ComponentsHandler) RegisterRoutes(e *gin.Engine) error {
	e.GET("/api/orgs/:orgId/components/:componentId/builds", h.GetComponentBuilds)
	e.GET("/api/orgs/:orgId/components/:componentId/builds/:buildId", h.GetComponentBuild)
	return nil
}

func (h *ComponentsHandler) GetComponentBuilds(c *gin.Context) {
	client, err := cctx.APIClientFromGinContext(c)
	if err != nil {
		respondError(c, http.StatusInternalServerError, err)
		return
	}
	client.SetOrgID(c.Param("orgId"))

	// GetComponentBuilds requires (componentID, appID, query)
	// appID is passed as empty string since the route doesn't include it;
	// the SDK uses componentID as the primary lookup.
	builds, _, err := client.GetComponentBuilds(c.Request.Context(), c.Param("componentId"), "", paginationFromQuery(c))
	if err != nil {
		respondError(c, http.StatusInternalServerError, err)
		return
	}

	respondJSON(c, http.StatusOK, builds)
}

func (h *ComponentsHandler) GetComponentBuild(c *gin.Context) {
	client, err := cctx.APIClientFromGinContext(c)
	if err != nil {
		respondError(c, http.StatusInternalServerError, err)
		return
	}
	client.SetOrgID(c.Param("orgId"))

	build, err := client.GetComponentBuild(c.Request.Context(), c.Param("componentId"), c.Param("buildId"))
	if err != nil {
		respondError(c, http.StatusInternalServerError, err)
		return
	}

	respondJSON(c, http.StatusOK, build)
}
