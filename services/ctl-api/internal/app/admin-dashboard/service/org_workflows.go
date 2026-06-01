package service

import (
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"go.temporal.io/api/workflowservice/v1"
	"go.uber.org/zap"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	orgshelpers "github.com/nuonco/nuon/services/ctl-api/internal/app/orgs/helpers"
	terminateworkflows "github.com/nuonco/nuon/services/ctl-api/internal/app/orgs/signals/terminate_workflows"
	queueclient "github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/client"
)

type orgWorkflowEntry struct {
	WorkflowID   string            `json:"workflow_id"`
	RunID        string            `json:"run_id"`
	WorkflowType string            `json:"workflow_type"`
	Namespace    string            `json:"namespace"`
	Status       string            `json:"status"`
	StartTime    string            `json:"start_time,omitempty"`
	Memo         map[string]string `json:"memo,omitempty"`
}

// OrgWorkflows lists all running Temporal workflows that belong to an org (via org-id memo).
func (s *service) OrgWorkflows(c *gin.Context) {
	orgID := c.Param("id")
	ctx := c.Request.Context()

	var allEntries []orgWorkflowEntry
	var mu sync.Mutex

	for _, ns := range temporalWorkerNamespaces {
		nsClient, err := s.temporalClient.GetNamespaceClient(ns)
		if err != nil {
			s.l.Warn("failed to get namespace client", zap.Error(err), zap.String("namespace", ns))
			continue
		}

		var nextPageToken []byte
		for {
			resp, err := nsClient.ListWorkflow(ctx, &workflowservice.ListWorkflowExecutionsRequest{
				Namespace:     ns,
				PageSize:      100,
				NextPageToken: nextPageToken,
				Query:         "ExecutionStatus='Running'",
			})
			if err != nil {
				s.l.Warn("failed to list workflows", zap.Error(err), zap.String("namespace", ns))
				break
			}

			for _, wf := range resp.Executions {
				memo := decodeMemo(wf.Memo)
				if memo["org-id"] != orgID && memo["owner-id"] != orgID {
					continue
				}

				entry := orgWorkflowEntry{
					WorkflowID:   wf.Execution.WorkflowId,
					RunID:        wf.Execution.RunId,
					WorkflowType: wf.Type.Name,
					Namespace:    ns,
					Status:       cleanStatus(wf.Status.String()),
					Memo:         memo,
				}
				if wf.StartTime != nil {
					entry.StartTime = wf.StartTime.AsTime().Format(time.RFC3339)
				}

				mu.Lock()
				allEntries = append(allEntries, entry)
				mu.Unlock()
			}

			nextPageToken = resp.NextPageToken
			if len(nextPageToken) == 0 {
				break
			}
		}
	}

	c.JSON(http.StatusOK, gin.H{"workflows": allEntries})
}

// TerminateOrgWorkflows enqueues a signal to terminate all running Temporal workflows
// belonging to an org. This runs asynchronously in the org's signal queue.
func (s *service) TerminateOrgWorkflows(c *gin.Context) {
	orgID := c.Param("id")
	ctx := c.Request.Context()

	// Find the org's signals queue.
	var queue app.Queue
	if res := s.db.WithContext(ctx).
		Where(app.Queue{OwnerID: orgID, Name: orgshelpers.OrgSignalsQueueName}).
		First(&queue); res.Error != nil {
		s.l.Error("unable to find org signals queue", zap.String("org_id", orgID), zap.Error(res.Error))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to find org signals queue"})
		return
	}

	// Enqueue the terminate-workflows signal.
	resp, err := s.queueClient.EnqueueSignal(ctx, &queueclient.EnqueueSignalRequest{
		QueueID: queue.ID,
		Signal:  &terminateworkflows.Signal{OrgID: orgID},
	})
	if err != nil {
		s.l.Error("unable to enqueue terminate-workflows signal", zap.String("org_id", orgID), zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to enqueue terminate signal: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":    "enqueued",
		"signal_id": resp.ID,
		"message":   "Terminate workflows signal enqueued. Workflows will be terminated in the background.",
	})
}
