package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"github.com/nuonco/nuon/services/dashboard-ui/server/internal/pkg/cctx"
)

type LogStreamsHandler struct {
	l *zap.Logger
}

func NewLogStreamsHandler(l *zap.Logger) *LogStreamsHandler {
	return &LogStreamsHandler{l: l}
}

func (h *LogStreamsHandler) RegisterRoutes(e *gin.Engine) error {
	e.GET("/api/orgs/:orgId/log-streams/:logStreamId", h.GetLogStream)
	e.GET("/api/orgs/:orgId/log-streams/:logStreamId/logs", h.GetLogStreamLogs)
	// SSE endpoint will be added in Phase 5
	return nil
}

func (h *LogStreamsHandler) GetLogStream(c *gin.Context) {
	client, err := cctx.APIClientFromGinContext(c)
	if err != nil {
		respondError(c, http.StatusInternalServerError, err)
		return
	}
	client.SetOrgID(c.Param("orgId"))

	logStream, err := client.GetLogStream(c.Request.Context(), c.Param("logStreamId"))
	if err != nil {
		respondError(c, http.StatusInternalServerError, err)
		return
	}

	respondJSON(c, http.StatusOK, logStream)
}

func (h *LogStreamsHandler) GetLogStreamLogs(c *gin.Context) {
	client, err := cctx.APIClientFromGinContext(c)
	if err != nil {
		respondError(c, http.StatusInternalServerError, err)
		return
	}
	client.SetOrgID(c.Param("orgId"))

	offset := c.DefaultQuery("offset", "")
	logs, err := client.LogStreamReadLogs(c.Request.Context(), c.Param("logStreamId"), offset)
	if err != nil {
		respondError(c, http.StatusInternalServerError, err)
		return
	}

	respondJSON(c, http.StatusOK, logs)
}
