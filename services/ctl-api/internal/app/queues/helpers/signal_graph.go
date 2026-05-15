package helpers

import (
	"context"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

// BuildSignalGraphNode recursively builds a signal graph tree.
// orgID is used to scope all DB queries to the caller's org.
func (h *Helpers) BuildSignalGraphNode(ctx context.Context, signal *app.QueueSignal, depth, maxDepth int, orgID string) SignalGraphNode {
	node := SignalGraphNode{
		Signal: signal,
	}

	if depth >= maxDepth {
		return node
	}

	if signal.Workflow.Namespace != "" && signal.Workflow.ID != "" {
		wfInfo := h.GetWorkflowInfo(ctx, signal.Workflow.Namespace, signal.Workflow.ID)
		if wfInfo != nil {
			node.WorkflowInfo = wfInfo

			seen := map[string]bool{}
			for _, as := range wfInfo.AwaitedSignals {
				if as.Signal != nil && !seen[as.Signal.ID] {
					if orgID == "" || (as.Signal.OrgID != nil && *as.Signal.OrgID == orgID) {
						seen[as.Signal.ID] = true
						child := h.BuildSignalGraphNode(ctx, as.Signal, depth+1, maxDepth, orgID)
						child.Relationship = "awaited"
						node.Children = append(node.Children, child)
					}
				} else if as.QueueSignalID != "" && !seen[as.QueueSignalID] {
					seen[as.QueueSignalID] = true
					var childSignal app.QueueSignal
					q := h.db.WithContext(ctx).Where(app.QueueSignal{ID: as.QueueSignalID})
					if orgID != "" {
						q = q.Where("org_id = ?", orgID)
					}
					if err := q.First(&childSignal).Error; err == nil {
						child := h.BuildSignalGraphNode(ctx, &childSignal, depth+1, maxDepth, orgID)
						child.Relationship = "awaited"
						node.Children = append(node.Children, child)
					}
				}
			}

			for _, es := range wfInfo.EnqueuedSignals {
				if es.Signal != nil && !seen[es.Signal.ID] {
					if orgID == "" || (es.Signal.OrgID != nil && *es.Signal.OrgID == orgID) {
						seen[es.Signal.ID] = true
						child := h.BuildSignalGraphNode(ctx, es.Signal, depth+1, maxDepth, orgID)
						child.Relationship = "enqueued"
						node.Children = append(node.Children, child)
					}
				} else if es.QueueSignalID != "" && !seen[es.QueueSignalID] {
					seen[es.QueueSignalID] = true
					var childSignal app.QueueSignal
					q := h.db.WithContext(ctx).Where(app.QueueSignal{ID: es.QueueSignalID})
					if orgID != "" {
						q = q.Where("org_id = ?", orgID)
					}
					if err := q.First(&childSignal).Error; err == nil {
						child := h.BuildSignalGraphNode(ctx, &childSignal, depth+1, maxDepth, orgID)
						child.Relationship = "enqueued"
						node.Children = append(node.Children, child)
					}
				}
			}
		}
	}

	return node
}
