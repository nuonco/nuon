package service

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func (s *service) TemporalWorkflowViewer(c *gin.Context) {
	namespace := c.Query("namespace")
	workflowID := c.Query("workflow_id")

	if namespace == "" || workflowID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "namespace and workflow_id are required"})
		return
	}

	wfInfo := s.getWorkflowInfo(c, namespace, workflowID)
	if wfInfo == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Workflow not found or not accessible"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"namespace":       namespace,
		"workflow_id":     workflowID,
		"temporal_ui_url": s.cfg.TemporalUIURL,
		"workflow_info":   wfInfo,
	})
}
