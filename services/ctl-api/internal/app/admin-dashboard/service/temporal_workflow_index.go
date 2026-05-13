package service

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	commonpb "go.temporal.io/api/common/v1"
	workflowpb "go.temporal.io/api/workflow/v1"
	"go.temporal.io/api/workflowservice/v1"
	tclient "go.temporal.io/sdk/client"
	"go.temporal.io/sdk/converter"
	"go.uber.org/zap"
)

// WorkflowIndexEntry is a single workflow in the index response.
type WorkflowIndexEntry struct {
	WorkflowID       string            `json:"workflow_id"`
	RunID            string            `json:"run_id"`
	WorkflowType     string            `json:"workflow_type"`
	Namespace        string            `json:"namespace"`
	Status           string            `json:"status"`
	StartTime        string            `json:"start_time"`
	HistoryLength    int64             `json:"history_length"`
	HistorySizeBytes int64             `json:"history_size_bytes"`
	CANCount         int64             `json:"can_count"`
	Memo             map[string]string `json:"memo,omitempty"`
	IsQueue          bool              `json:"is_queue"`
	QueueID          string            `json:"queue_id,omitempty"`
	Link             string            `json:"link,omitempty"`
}

const descConcurrency = 10

// TemporalWorkflowIndex streams workflow entries for a namespace as newline-delimited JSON.
// It lists all running workflows, then describes each one in parallel to get accurate
// history length and size.
func (s *service) TemporalWorkflowIndex(c *gin.Context) {
	namespace := c.Query("namespace")
	if namespace == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "namespace is required"})
		return
	}

	nsClient, err := s.temporalClient.GetNamespaceClient(namespace)
	if err != nil {
		s.l.Error("failed to get namespace client", zap.Error(err), zap.String("namespace", namespace))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to connect to namespace"})
		return
	}

	c.Writer.Header().Set("Content-Type", "application/x-ndjson")
	c.Writer.Header().Set("Transfer-Encoding", "chunked")
	c.Writer.Header().Set("Cache-Control", "no-cache")
	c.Writer.WriteHeader(http.StatusOK)

	ctx := c.Request.Context()
	var nextPageToken []byte
	totalSent := 0

	// Channel for entries ready to write, semaphore for concurrency.
	entryCh := make(chan *WorkflowIndexEntry, descConcurrency*2)
	var wg sync.WaitGroup

	// Writer goroutine: serialize entries to the response as they arrive.
	writeDone := make(chan struct{})
	go func() {
		defer close(writeDone)
		for entry := range entryCh {
			line, _ := json.Marshal(entry)
			line = append(line, '\n')
			_, _ = c.Writer.Write(line)
			c.Writer.Flush()
			totalSent++
		}
	}()

	// Semaphore to limit concurrent describe calls.
	sem := make(chan struct{}, descConcurrency)

	for {
		resp, err := nsClient.ListWorkflow(ctx, &workflowservice.ListWorkflowExecutionsRequest{
			Namespace:     namespace,
			PageSize:      100,
			NextPageToken: nextPageToken,
			Query:         "ExecutionStatus='Running'",
		})
		if err != nil {
			s.l.Error("failed to list workflows", zap.Error(err), zap.String("namespace", namespace))
			break
		}

		for _, wf := range resp.Executions {
			wg.Add(1)
			sem <- struct{}{} // acquire

			go func() {
				defer wg.Done()
				defer func() { <-sem }() // release

				entry := s.describeAndBuildEntry(ctx, nsClient, namespace, wf)
				entryCh <- entry
			}()
		}

		nextPageToken = resp.NextPageToken
		if len(nextPageToken) == 0 {
			break
		}
	}

	wg.Wait()
	close(entryCh)
	<-writeDone

	// Send a final summary line.
	summary, _ := json.Marshal(map[string]any{
		"_type": "summary",
		"total": totalSent,
	})
	summary = append(summary, '\n')
	_, _ = c.Writer.Write(summary)
	c.Writer.Flush()
}

// describeAndBuildEntry calls DescribeWorkflowExecution to get accurate history info.
func (s *service) describeAndBuildEntry(
	ctx context.Context,
	nsClient tclient.Client,
	namespace string,
	wf *workflowpb.WorkflowExecutionInfo,
) *WorkflowIndexEntry {
	entry := &WorkflowIndexEntry{
		WorkflowID:   wf.Execution.WorkflowId,
		RunID:        wf.Execution.RunId,
		WorkflowType: wf.Type.Name,
		Namespace:    namespace,
		Status:       cleanStatus(wf.Status.String()),
	}

	if wf.StartTime != nil {
		entry.StartTime = wf.StartTime.AsTime().Format(time.RFC3339)
	}

	// Use DescribeWorkflowExecution for accurate history length and size.
	desc, err := nsClient.DescribeWorkflowExecution(ctx, wf.Execution.WorkflowId, wf.Execution.RunId)
	if err != nil {
		s.l.Warn("failed to describe workflow",
			zap.String("workflow_id", wf.Execution.WorkflowId),
			zap.Error(err))
		// Fall back to list data (may be 0).
		entry.HistoryLength = wf.HistoryLength
	} else if desc.WorkflowExecutionInfo != nil {
		entry.HistoryLength = desc.WorkflowExecutionInfo.HistoryLength
		entry.HistorySizeBytes = desc.WorkflowExecutionInfo.HistorySizeBytes
		if desc.WorkflowExecutionInfo.StartTime != nil {
			entry.StartTime = desc.WorkflowExecutionInfo.StartTime.AsTime().Format(time.RFC3339)
		}
	}

	// Count total executions for this workflow ID to derive CAN count.
	countResp, err := nsClient.CountWorkflow(ctx, &workflowservice.CountWorkflowExecutionsRequest{
		Namespace: namespace,
		Query:     fmt.Sprintf("WorkflowId = '%s'", wf.Execution.WorkflowId),
	})
	if err == nil && countResp.Count > 1 {
		entry.CANCount = countResp.Count - 1 // subtract current running execution
	}

	// Parse memo.
	entry.Memo = decodeMemo(wf.Memo)
	if entry.Memo["type"] == "queue" || strings.HasPrefix(entry.WorkflowType, "Queue") {
		entry.IsQueue = true
		entry.QueueID = entry.Memo["id"]
		if entry.QueueID != "" {
			entry.Link = fmt.Sprintf("/queues/%s", entry.QueueID)
		}
	}

	return entry
}

// TemporalWorkflowNamespaces returns the list of known namespaces.
func (s *service) TemporalWorkflowNamespaces(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"namespaces":      temporalWorkerNamespaces,
		"temporal_ui_url": s.cfg.TemporalUIURL,
	})
}

// decodeMemo extracts string values from a Temporal workflow Memo.
func decodeMemo(memo *commonpb.Memo) map[string]string {
	result := make(map[string]string)
	if memo == nil {
		return result
	}
	dc := converter.GetDefaultDataConverter()
	for k, payload := range memo.Fields {
		var val string
		if err := dc.FromPayload(payload, &val); err != nil {
			var numVal float64
			if err2 := dc.FromPayload(payload, &numVal); err2 == nil {
				val = fmt.Sprintf("%v", numVal)
			} else {
				val = "(unreadable)"
			}
		}
		result[k] = val
	}
	return result
}

// cleanStatus strips the WORKFLOW_EXECUTION_STATUS_ prefix from protobuf status strings.
func cleanStatus(s string) string {
	s = strings.TrimPrefix(s, "WORKFLOW_EXECUTION_STATUS_")
	return strings.ToLower(strings.ReplaceAll(s, "_", "-"))
}
