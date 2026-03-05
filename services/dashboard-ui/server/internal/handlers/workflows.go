package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	nuon "github.com/nuonco/nuon/sdks/nuon-go"
	"github.com/nuonco/nuon/services/dashboard-ui/server/internal"
)

type WorkflowsHandler struct {
	cfg *internal.Config
	l   *zap.Logger
}

func NewWorkflowsHandler(cfg *internal.Config, l *zap.Logger) *WorkflowsHandler {
	return &WorkflowsHandler{cfg: cfg, l: l}
}

func (h *WorkflowsHandler) RegisterRoutes(e *gin.Engine) error {
	e.GET("/api/orgs/:orgId/workflows/:workflowId/steps/:stepId/approvals/:approvalId/contents", h.GetApprovalContents)
	return nil
}

func (h *WorkflowsHandler) GetApprovalContents(c *gin.Context) {
	orgID := c.Param("orgId")
	workflowID := c.Param("workflowId")
	stepID := c.Param("stepId")
	approvalID := c.Param("approvalId")

	token, err := c.Cookie(authCookie)
	if err != nil || token == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	client, err := nuon.New(
		nuon.WithURL(h.cfg.APIUrl),
		nuon.WithAuthToken(token),
		nuon.WithOrgID(orgID),
	)
	if err != nil {
		h.l.Error("failed to create nuon client", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create client"})
		return
	}

	contents, err := client.GetWorkflowStepApprovalContents(c.Request.Context(), workflowID, stepID, approvalID)
	if err != nil {
		h.l.Error("failed to get approval contents", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, contents)
}
