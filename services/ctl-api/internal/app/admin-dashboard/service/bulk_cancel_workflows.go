package service

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"github.com/nuonco/nuon/pkg/generics"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	bulkcancelworkflows "github.com/nuonco/nuon/services/ctl-api/internal/app/general/signals/bulk_cancel_workflows"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/cctx"
	queueclient "github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/client"
)

const bulkCancelMaxWorkflows = 5000

// BulkCancelWorkflowsRequest takes explicit IDs or a filter; IDs win when both are set.
type BulkCancelWorkflowsRequest struct {
	WorkflowIDs []string `json:"workflow_ids"`

	Search        string `json:"search"`
	Type          string `json:"type"`
	Status        string `json:"status"`
	CreatedAfter  string `json:"created_after"`
	CreatedBefore string `json:"created_before"`
}

func (s *service) BulkCancelWorkflows(c *gin.Context) {
	ctx := c.Request.Context()

	var req BulkCancelWorkflowsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "unable to parse request: " + err.Error()})
		return
	}

	workflowIDs, err := s.resolveBulkCancelWorkflowIDs(c, req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if len(workflowIDs) == 0 {
		c.JSON(http.StatusOK, gin.H{
			"status":  "noop",
			"count":   0,
			"message": "No workflows matched — nothing to cancel.",
		})
		return
	}

	if len(workflowIDs) > bulkCancelMaxWorkflows {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": fmt.Sprintf("%d workflows matched, which is over the %d limit — narrow the filter first", len(workflowIDs), bulkCancelMaxWorkflows),
		})
		return
	}

	reason := "admin bulk cancel"
	if acct, _ := cctx.AccountFromGinContext(c); acct != nil && acct.Email != "" {
		reason = fmt.Sprintf("admin bulk cancel by %s", acct.Email)
	}

	queue, err := s.generalHelpers.EnsureGeneralQueue(ctx)
	if err != nil {
		s.l.Error("unable to ensure general queue", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to ensure general queue: " + err.Error()})
		return
	}

	resp, err := s.queueClient.EnqueueSignal(ctx, &queueclient.EnqueueSignalRequest{
		QueueID: queue.ID,
		Signal: &bulkcancelworkflows.Signal{
			WorkflowIDs: workflowIDs,
			Reason:      reason,
		},
	})
	if err != nil {
		s.l.Error("unable to enqueue bulk cancel signal", zap.Int("count", len(workflowIDs)), zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to enqueue bulk cancel signal: " + err.Error()})
		return
	}

	s.l.Info("enqueued bulk workflow cancellation",
		zap.String("signal_id", resp.ID),
		zap.String("queue_id", queue.ID),
		zap.String("reason", reason),
		zap.Int("count", len(workflowIDs)))

	c.JSON(http.StatusOK, gin.H{
		"status":    "enqueued",
		"count":     len(workflowIDs),
		"queue_id":  queue.ID,
		"signal_id": resp.ID,
		"message":   fmt.Sprintf("Cancelling %d workflows in the background.", len(workflowIDs)),
	})
}

func (s *service) resolveBulkCancelWorkflowIDs(c *gin.Context, req BulkCancelWorkflowsRequest) ([]string, error) {
	if len(req.WorkflowIDs) > 0 {
		return req.WorkflowIDs, nil
	}

	// cancelling by filter is a blind write, so restrict it to statuses the cancel path acts on
	if !generics.SliceContains(app.Status(req.Status), bulkCancelableStatuses()) {
		return nil, fmt.Errorf("cancelling by filter requires a cancelable status (one of %v)", bulkCancelableStatuses())
	}

	filter := workflowFilter{
		Search: req.Search,
		Type:   req.Type,
		Status: req.Status,
	}

	after, err := parseTimeQuery(req.CreatedAfter)
	if err != nil {
		return nil, fmt.Errorf("invalid created_after: %w", err)
	}
	filter.CreatedAfter = after

	before, err := parseTimeQuery(req.CreatedBefore)
	if err != nil {
		return nil, fmt.Errorf("invalid created_before: %w", err)
	}
	filter.CreatedBefore = before

	var workflowIDs []string
	if err := s.workflowQuery(c.Request.Context(), filter).
		Order("created_at DESC").
		Limit(bulkCancelMaxWorkflows+1).
		Pluck("id", &workflowIDs).Error; err != nil {
		return nil, fmt.Errorf("unable to list workflows to cancel: %w", err)
	}

	return workflowIDs, nil
}

type statusOption struct {
	Value app.Status `json:"value"`
	Label string     `json:"label"`
}

// bulkCancelableStatusOptions are the only statuses a filter-driven bulk cancel may target.
var bulkCancelableStatusOptions = []statusOption{
	{Value: app.StatusFailedPendingRetry, Label: "Awaiting retry"},
	{Value: app.StatusInProgress, Label: "In progress"},
	{Value: app.StatusPending, Label: "Pending"},
	{Value: app.AwaitingApproval, Label: "Awaiting approval"},
}

func bulkCancelableStatuses() []app.Status {
	statuses := make([]app.Status, 0, len(bulkCancelableStatusOptions))
	for _, opt := range bulkCancelableStatusOptions {
		statuses = append(statuses, opt.Value)
	}
	return statuses
}

// WorkflowFilterOptions keeps the admin dropdowns from drifting as types are added.
func (s *service) WorkflowFilterOptions(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"types":               app.AllWorkflowTypes(),
		"cancelable_statuses": bulkCancelableStatusOptions,
		"per_page_options":    workflowsPerPageOptions,
	})
}
