package service

import (
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"go.temporal.io/api/workflowservice/v1"
	"go.uber.org/zap"
)

type WorkflowStats struct {
	HistoryLength    int64  `json:"history_length"`
	HistorySizeBytes int64  `json:"history_size_bytes"`
	CANCount         int64  `json:"can_count"`
	StartTime        string `json:"start_time"`
	Status           string `json:"status"`
}

// TemporalWorkflowStats returns history length, size, CAN count, and age for a single workflow.
func (s *service) TemporalWorkflowStats(c *gin.Context) {
	namespace := c.Query("namespace")
	workflowID := c.Query("workflow_id")

	if namespace == "" || workflowID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "namespace and workflow_id are required"})
		return
	}

	nsClient, err := s.temporalClient.GetNamespaceClient(namespace)
	if err != nil {
		s.l.Error("failed to get namespace client", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to connect to namespace"})
		return
	}

	ctx := c.Request.Context()
	stats := &WorkflowStats{}

	desc, err := nsClient.DescribeWorkflowExecution(ctx, workflowID, "")
	if err != nil {
		s.l.Warn("failed to describe workflow for stats", zap.String("workflow_id", workflowID), zap.Error(err))
		c.JSON(http.StatusNotFound, gin.H{"error": "Workflow not found"})
		return
	}

	if info := desc.WorkflowExecutionInfo; info != nil {
		stats.HistoryLength = info.HistoryLength
		stats.HistorySizeBytes = info.HistorySizeBytes
		stats.Status = cleanStatus(info.Status.String())
		if info.StartTime != nil {
			stats.StartTime = info.StartTime.AsTime().Format(time.RFC3339)
		}
	}

	countResp, err := nsClient.CountWorkflow(ctx, &workflowservice.CountWorkflowExecutionsRequest{
		Namespace: namespace,
		Query:     fmt.Sprintf("WorkflowId = '%s'", workflowID),
	})
	if err == nil && countResp.Count > 1 {
		stats.CANCount = countResp.Count - 1
	}

	c.JSON(http.StatusOK, stats)
}
