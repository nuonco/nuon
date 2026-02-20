package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"github.com/nuonco/nuon/services/dashboard-ui/server/internal/pkg/cctx"
)

type InstallsHandler struct {
	l *zap.Logger
}

func NewInstallsHandler(l *zap.Logger) *InstallsHandler {
	return &InstallsHandler{l: l}
}

func (h *InstallsHandler) RegisterRoutes(e *gin.Engine) error {
	e.GET("/api/orgs/:orgId/installs", h.GetInstalls)
	e.GET("/api/orgs/:orgId/installs/:installId", h.GetInstall)
	e.GET("/api/orgs/:orgId/installs/:installId/components", h.GetInstallComponents)
	e.GET("/api/orgs/:orgId/installs/:installId/components/:componentId/deploys", h.GetComponentDeploys)
	e.GET("/api/orgs/:orgId/installs/:installId/deploys/:deployId", h.GetDeploy)
	e.GET("/api/orgs/:orgId/installs/:installId/stack", h.GetInstallStack)
	e.GET("/api/orgs/:orgId/installs/:installId/workflows", h.GetInstallWorkflows)
	e.GET("/api/orgs/:orgId/installs/:installId/sandbox/runs", h.GetSandboxRuns)
	return nil
}

func (h *InstallsHandler) GetInstalls(c *gin.Context) {
	client, err := cctx.APIClientFromGinContext(c)
	if err != nil {
		respondError(c, http.StatusInternalServerError, err)
		return
	}
	client.SetOrgID(c.Param("orgId"))

	installs, _, err := client.GetAllInstalls(c.Request.Context(), paginationFromQuery(c))
	if err != nil {
		respondError(c, http.StatusInternalServerError, err)
		return
	}

	respondJSON(c, http.StatusOK, installs)
}

func (h *InstallsHandler) GetInstall(c *gin.Context) {
	client, err := cctx.APIClientFromGinContext(c)
	if err != nil {
		respondError(c, http.StatusInternalServerError, err)
		return
	}
	client.SetOrgID(c.Param("orgId"))

	install, err := client.GetInstall(c.Request.Context(), c.Param("installId"))
	if err != nil {
		respondError(c, http.StatusInternalServerError, err)
		return
	}

	respondJSON(c, http.StatusOK, install)
}

func (h *InstallsHandler) GetInstallComponents(c *gin.Context) {
	client, err := cctx.APIClientFromGinContext(c)
	if err != nil {
		respondError(c, http.StatusInternalServerError, err)
		return
	}
	client.SetOrgID(c.Param("orgId"))

	components, _, err := client.GetInstallComponents(c.Request.Context(), c.Param("installId"), paginationFromQuery(c))
	if err != nil {
		respondError(c, http.StatusInternalServerError, err)
		return
	}

	respondJSON(c, http.StatusOK, components)
}

func (h *InstallsHandler) GetComponentDeploys(c *gin.Context) {
	client, err := cctx.APIClientFromGinContext(c)
	if err != nil {
		respondError(c, http.StatusInternalServerError, err)
		return
	}
	client.SetOrgID(c.Param("orgId"))

	deploys, _, err := client.GetInstallComponentDeploys(c.Request.Context(), c.Param("installId"), c.Param("componentId"), paginationFromQuery(c))
	if err != nil {
		respondError(c, http.StatusInternalServerError, err)
		return
	}

	respondJSON(c, http.StatusOK, deploys)
}

func (h *InstallsHandler) GetDeploy(c *gin.Context) {
	client, err := cctx.APIClientFromGinContext(c)
	if err != nil {
		respondError(c, http.StatusInternalServerError, err)
		return
	}
	client.SetOrgID(c.Param("orgId"))

	deploy, err := client.GetInstallDeploy(c.Request.Context(), c.Param("installId"), c.Param("deployId"))
	if err != nil {
		respondError(c, http.StatusInternalServerError, err)
		return
	}

	respondJSON(c, http.StatusOK, deploy)
}

func (h *InstallsHandler) GetInstallStack(c *gin.Context) {
	client, err := cctx.APIClientFromGinContext(c)
	if err != nil {
		respondError(c, http.StatusInternalServerError, err)
		return
	}
	client.SetOrgID(c.Param("orgId"))

	stack, err := client.GetInstallStack(c.Request.Context(), c.Param("installId"))
	if err != nil {
		respondError(c, http.StatusInternalServerError, err)
		return
	}

	respondJSON(c, http.StatusOK, stack)
}

func (h *InstallsHandler) GetInstallWorkflows(c *gin.Context) {
	client, err := cctx.APIClientFromGinContext(c)
	if err != nil {
		respondError(c, http.StatusInternalServerError, err)
		return
	}
	client.SetOrgID(c.Param("orgId"))

	workflows, _, err := client.GetWorkflows(c.Request.Context(), c.Param("installId"), paginationFromQuery(c))
	if err != nil {
		respondError(c, http.StatusInternalServerError, err)
		return
	}

	respondJSON(c, http.StatusOK, workflows)
}

func (h *InstallsHandler) GetSandboxRuns(c *gin.Context) {
	client, err := cctx.APIClientFromGinContext(c)
	if err != nil {
		respondError(c, http.StatusInternalServerError, err)
		return
	}
	client.SetOrgID(c.Param("orgId"))

	runs, _, err := client.GetInstallSandboxRuns(c.Request.Context(), c.Param("installId"), paginationFromQuery(c))
	if err != nil {
		respondError(c, http.StatusInternalServerError, err)
		return
	}

	respondJSON(c, http.StatusOK, runs)
}
