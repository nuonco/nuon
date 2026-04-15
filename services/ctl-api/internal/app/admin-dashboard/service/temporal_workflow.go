package service

import (
	"net/http"

	"github.com/a-h/templ"
	"github.com/gin-gonic/gin"

	"github.com/nuonco/nuon/services/ctl-api/internal/app/admin-dashboard/service/views"
)

func (s *service) TemporalWorkflowViewer(c *gin.Context) {
	namespace := c.Query("namespace")
	workflowID := c.Query("workflow_id")

	if namespace == "" || workflowID == "" {
		// Show search form
		component := views.TemporalWorkflowSearch(namespace, workflowID)
		templ.Handler(component).ServeHTTP(c.Writer, c.Request)
		return
	}

	wfInfo := s.getWorkflowInfo(c, namespace, workflowID)
	if wfInfo == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Workflow not found or not accessible"})
		return
	}

	component := views.TemporalWorkflowDetail(namespace, workflowID, s.cfg.TemporalUIURL, wfInfo)
	templ.Handler(component).ServeHTTP(c.Writer, c.Request)
}
