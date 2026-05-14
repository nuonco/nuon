package service

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/admin-dashboard/service/views"
)

// SignalGraphNode represents one signal in the recursive graph.
type SignalGraphNode struct {
	Signal       *app.QueueSignal    `json:"signal"`
	WorkflowInfo *views.WorkflowInfo `json:"workflow_info,omitempty"`
	Children     []SignalGraphNode   `json:"children,omitempty"`
	Relationship string              `json:"relationship,omitempty"` // "awaited" or "enqueued"
}

// SignalGraph returns a recursive graph of a signal and everything it awaits.
// Pass ?depth=N to control recursion depth (default 1 for lazy loading).
func (s *service) SignalGraph(c *gin.Context) {
	queueID := c.Param("id")
	signalID := c.Param("signal_id")

	var signal app.QueueSignal
	if res := s.readDB().WithContext(c.Request.Context()).
		Where("id = ? AND queue_id = ?", signalID, queueID).
		First(&signal); res.Error != nil {
		// Try without queue_id filter (for lazy expand of child signals)
		if res2 := s.readDB().WithContext(c.Request.Context()).
			Where("id = ?", signalID).
			First(&signal); res2.Error != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Signal not found"})
			return
		}
	}

	maxDepth := 1
	if d := c.Query("depth"); d != "" {
		if n, err := strconv.Atoi(d); err == nil && n > 0 && n <= 10 {
			maxDepth = n
		}
	}

	node := s.buildSignalGraphNode(c, &signal, 0, maxDepth)

	c.JSON(http.StatusOK, gin.H{
		"graph":           node,
		"temporal_ui_url": s.cfg.TemporalUIURL,
	})
}

func (s *service) buildSignalGraphNode(c *gin.Context, signal *app.QueueSignal, depth, maxDepth int) SignalGraphNode {
	node := SignalGraphNode{
		Signal: signal,
	}

	if depth >= maxDepth {
		return node
	}

	// Get workflow info for this signal
	if signal.Workflow.Namespace != "" && signal.Workflow.ID != "" {
		wfInfo := s.getWorkflowInfo(c, signal.Workflow.Namespace, signal.Workflow.ID)
		if wfInfo != nil {
			node.WorkflowInfo = wfInfo

			// For each awaited signal, recursively build the graph
			seen := map[string]bool{}
			for _, as := range wfInfo.AwaitedSignals {
				if as.Signal != nil && !seen[as.Signal.ID] {
					seen[as.Signal.ID] = true
					child := s.buildSignalGraphNode(c, as.Signal, depth+1, maxDepth)
					child.Relationship = "awaited"
					node.Children = append(node.Children, child)
				} else if as.QueueSignalID != "" && !seen[as.QueueSignalID] {
					seen[as.QueueSignalID] = true
					var childSignal app.QueueSignal
					if err := s.readDB().WithContext(c.Request.Context()).
						Where("id = ?", as.QueueSignalID).
						First(&childSignal).Error; err == nil {
						child := s.buildSignalGraphNode(c, &childSignal, depth+1, maxDepth)
						child.Relationship = "awaited"
						node.Children = append(node.Children, child)
					}
				}
			}

			// For each enqueued signal, also recursively build the graph
			for _, es := range wfInfo.EnqueuedSignals {
				if es.Signal != nil && !seen[es.Signal.ID] {
					seen[es.Signal.ID] = true
					child := s.buildSignalGraphNode(c, es.Signal, depth+1, maxDepth)
					child.Relationship = "enqueued"
					node.Children = append(node.Children, child)
				} else if es.QueueSignalID != "" && !seen[es.QueueSignalID] {
					seen[es.QueueSignalID] = true
					var childSignal app.QueueSignal
					if err := s.readDB().WithContext(c.Request.Context()).
						Where("id = ?", es.QueueSignalID).
						First(&childSignal).Error; err == nil {
						child := s.buildSignalGraphNode(c, &childSignal, depth+1, maxDepth)
						child.Relationship = "enqueued"
						node.Children = append(node.Children, child)
					}
				}
			}
		}
	}

	return node
}
