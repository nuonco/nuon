package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"github.com/nuonco/nuon/services/dashboard-ui/server/internal/pkg/cctx"
)

type RunnersHandler struct {
	l *zap.Logger
}

func NewRunnersHandler(l *zap.Logger) *RunnersHandler {
	return &RunnersHandler{l: l}
}

func (h *RunnersHandler) RegisterRoutes(e *gin.Engine) error {
	// NOTE: Many runner SDK methods (GetRunner, GetRunnerJobs, GetRunnerSettings,
	// GetRunnerLatestHeartbeat, GetRunnerRecentHealthChecks) are not yet in the
	// nuon-go SDK. Only GetRunnerJobPlan exists. These routes will be added
	// as SDK methods are implemented.
	e.GET("/api/orgs/:orgId/runners/jobs/:runnerJobId/plan", h.GetRunnerJobPlan)
	return nil
}

func (h *RunnersHandler) GetRunnerJobPlan(c *gin.Context) {
	client, err := cctx.APIClientFromGinContext(c)
	if err != nil {
		respondError(c, http.StatusInternalServerError, err)
		return
	}
	client.SetOrgID(c.Param("orgId"))

	plan, err := client.GetRunnerJobPlan(c.Request.Context(), c.Param("runnerJobId"))
	if err != nil {
		respondError(c, http.StatusInternalServerError, err)
		return
	}

	respondJSON(c, http.StatusOK, plan)
}
