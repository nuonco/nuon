package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"github.com/nuonco/nuon/services/dashboard-ui/server/internal/pkg/cctx"
)

type AppsHandler struct {
	l *zap.Logger
}

func NewAppsHandler(l *zap.Logger) *AppsHandler {
	return &AppsHandler{l: l}
}

func (h *AppsHandler) RegisterRoutes(e *gin.Engine) error {
	e.GET("/api/orgs/:orgId/apps", h.GetApps)
	e.GET("/api/orgs/:orgId/apps/:appId", h.GetApp)
	e.GET("/api/orgs/:orgId/apps/:appId/components", h.GetAppComponents)
	e.GET("/api/orgs/:orgId/apps/:appId/configs", h.GetAppConfigs)
	e.GET("/api/orgs/:orgId/apps/:appId/configs/:configId", h.GetAppConfig)
	e.GET("/api/orgs/:orgId/apps/:appId/installs", h.GetAppInstalls)
	e.GET("/api/orgs/:orgId/apps/:appId/actions", h.GetAppActionWorkflows)
	e.GET("/api/orgs/:orgId/apps/:appId/actions/:actionId", h.GetAppActionWorkflow)
	return nil
}

func (h *AppsHandler) GetApps(c *gin.Context) {
	client, err := cctx.APIClientFromGinContext(c)
	if err != nil {
		respondError(c, http.StatusInternalServerError, err)
		return
	}
	client.SetOrgID(c.Param("orgId"))

	apps, _, err := client.GetApps(c.Request.Context(), paginationFromQuery(c))
	if err != nil {
		respondError(c, http.StatusInternalServerError, err)
		return
	}

	respondJSON(c, http.StatusOK, apps)
}

func (h *AppsHandler) GetApp(c *gin.Context) {
	client, err := cctx.APIClientFromGinContext(c)
	if err != nil {
		respondError(c, http.StatusInternalServerError, err)
		return
	}
	client.SetOrgID(c.Param("orgId"))

	app, err := client.GetApp(c.Request.Context(), c.Param("appId"))
	if err != nil {
		respondError(c, http.StatusInternalServerError, err)
		return
	}

	respondJSON(c, http.StatusOK, app)
}

func (h *AppsHandler) GetAppComponents(c *gin.Context) {
	client, err := cctx.APIClientFromGinContext(c)
	if err != nil {
		respondError(c, http.StatusInternalServerError, err)
		return
	}
	client.SetOrgID(c.Param("orgId"))

	components, _, err := client.GetAppComponents(c.Request.Context(), c.Param("appId"), paginationFromQuery(c))
	if err != nil {
		respondError(c, http.StatusInternalServerError, err)
		return
	}

	respondJSON(c, http.StatusOK, components)
}

func (h *AppsHandler) GetAppConfigs(c *gin.Context) {
	client, err := cctx.APIClientFromGinContext(c)
	if err != nil {
		respondError(c, http.StatusInternalServerError, err)
		return
	}
	client.SetOrgID(c.Param("orgId"))

	configs, _, err := client.GetAppConfigs(c.Request.Context(), c.Param("appId"), paginationFromQuery(c))
	if err != nil {
		respondError(c, http.StatusInternalServerError, err)
		return
	}

	respondJSON(c, http.StatusOK, configs)
}

func (h *AppsHandler) GetAppConfig(c *gin.Context) {
	client, err := cctx.APIClientFromGinContext(c)
	if err != nil {
		respondError(c, http.StatusInternalServerError, err)
		return
	}
	client.SetOrgID(c.Param("orgId"))

	config, err := client.GetAppConfig(c.Request.Context(), c.Param("appId"), c.Param("configId"), nil)
	if err != nil {
		respondError(c, http.StatusInternalServerError, err)
		return
	}

	respondJSON(c, http.StatusOK, config)
}

func (h *AppsHandler) GetAppInstalls(c *gin.Context) {
	client, err := cctx.APIClientFromGinContext(c)
	if err != nil {
		respondError(c, http.StatusInternalServerError, err)
		return
	}
	client.SetOrgID(c.Param("orgId"))

	installs, _, err := client.GetAppInstalls(c.Request.Context(), c.Param("appId"), paginationFromQuery(c))
	if err != nil {
		respondError(c, http.StatusInternalServerError, err)
		return
	}

	respondJSON(c, http.StatusOK, installs)
}

func (h *AppsHandler) GetAppActionWorkflows(c *gin.Context) {
	client, err := cctx.APIClientFromGinContext(c)
	if err != nil {
		respondError(c, http.StatusInternalServerError, err)
		return
	}
	client.SetOrgID(c.Param("orgId"))

	actions, _, err := client.GetActionWorkflows(c.Request.Context(), c.Param("appId"), paginationFromQuery(c))
	if err != nil {
		respondError(c, http.StatusInternalServerError, err)
		return
	}

	respondJSON(c, http.StatusOK, actions)
}

func (h *AppsHandler) GetAppActionWorkflow(c *gin.Context) {
	client, err := cctx.APIClientFromGinContext(c)
	if err != nil {
		respondError(c, http.StatusInternalServerError, err)
		return
	}
	client.SetOrgID(c.Param("orgId"))

	action, err := client.GetAppActionWorkflow(c.Request.Context(), c.Param("appId"), c.Param("actionId"))
	if err != nil {
		respondError(c, http.StatusInternalServerError, err)
		return
	}

	respondJSON(c, http.StatusOK, action)
}
