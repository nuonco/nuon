package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"github.com/nuonco/nuon/services/dashboard-ui/server/internal/pkg/cctx"
)

type VCSHandler struct {
	l *zap.Logger
}

func NewVCSHandler(l *zap.Logger) *VCSHandler {
	return &VCSHandler{l: l}
}

func (h *VCSHandler) RegisterRoutes(e *gin.Engine) error {
	e.GET("/api/orgs/:orgId/vcs-connections/:connectionId", h.GetVCSConnection)
	// NOTE: GetVCSConnectionRepos and CheckVCSConnectionStatus are not yet
	// in the nuon-go SDK. These routes will be added as SDK methods are implemented.
	return nil
}

func (h *VCSHandler) GetVCSConnection(c *gin.Context) {
	client, err := cctx.APIClientFromGinContext(c)
	if err != nil {
		respondError(c, http.StatusInternalServerError, err)
		return
	}
	client.SetOrgID(c.Param("orgId"))

	conn, err := client.GetVCSConnection(c.Request.Context(), c.Param("connectionId"))
	if err != nil {
		respondError(c, http.StatusInternalServerError, err)
		return
	}

	respondJSON(c, http.StatusOK, conn)
}
