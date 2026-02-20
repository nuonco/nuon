package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"github.com/nuonco/nuon/services/dashboard-ui/server/internal/pkg/cctx"
)

type WorkflowsHandler struct {
	l *zap.Logger
}

func NewWorkflowsHandler(l *zap.Logger) *WorkflowsHandler {
	return &WorkflowsHandler{l: l}
}

func (h *WorkflowsHandler) RegisterRoutes(e *gin.Engine) error {
	e.GET("/api/orgs/:orgId/workflows/:workflowId", h.GetWorkflow)
	e.GET("/api/orgs/:orgId/workflows/:workflowId/steps", h.GetWorkflowSteps)
	e.GET("/api/orgs/:orgId/workflows/:workflowId/steps/:stepId", h.GetWorkflowStep)
	e.GET("/api/orgs/:orgId/workflows/:workflowId/steps/:stepId/approvals/:approvalId/contents", h.GetWorkflowStepApprovalContents)
	return nil
}

func (h *WorkflowsHandler) GetWorkflow(c *gin.Context) {
	client, err := cctx.APIClientFromGinContext(c)
	if err != nil {
		respondError(c, http.StatusInternalServerError, err)
		return
	}
	client.SetOrgID(c.Param("orgId"))

	workflow, err := client.GetWorkflow(c.Request.Context(), c.Param("workflowId"))
	if err != nil {
		respondError(c, http.StatusInternalServerError, err)
		return
	}

	respondJSON(c, http.StatusOK, workflow)
}

func (h *WorkflowsHandler) GetWorkflowSteps(c *gin.Context) {
	client, err := cctx.APIClientFromGinContext(c)
	if err != nil {
		respondError(c, http.StatusInternalServerError, err)
		return
	}
	client.SetOrgID(c.Param("orgId"))

	// GetWorkflowSteps doesn't take pagination in the SDK
	steps, err := client.GetWorkflowSteps(c.Request.Context(), c.Param("workflowId"))
	if err != nil {
		respondError(c, http.StatusInternalServerError, err)
		return
	}

	respondJSON(c, http.StatusOK, steps)
}

func (h *WorkflowsHandler) GetWorkflowStep(c *gin.Context) {
	client, err := cctx.APIClientFromGinContext(c)
	if err != nil {
		respondError(c, http.StatusInternalServerError, err)
		return
	}
	client.SetOrgID(c.Param("orgId"))

	step, err := client.GetWorkflowStep(c.Request.Context(), c.Param("workflowId"), c.Param("stepId"))
	if err != nil {
		respondError(c, http.StatusInternalServerError, err)
		return
	}

	respondJSON(c, http.StatusOK, step)
}

func (h *WorkflowsHandler) GetWorkflowStepApprovalContents(c *gin.Context) {
	client, err := cctx.APIClientFromGinContext(c)
	if err != nil {
		respondError(c, http.StatusInternalServerError, err)
		return
	}
	client.SetOrgID(c.Param("orgId"))

	contents, err := client.GetWorkflowStepApprovalContents(c.Request.Context(), c.Param("workflowId"), c.Param("stepId"), c.Param("approvalId"))
	if err != nil {
		respondError(c, http.StatusInternalServerError, err)
		return
	}

	respondJSON(c, http.StatusOK, contents)
}
