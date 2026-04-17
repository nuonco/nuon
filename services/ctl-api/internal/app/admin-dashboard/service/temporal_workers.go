package service

import (
	"context"
	"fmt"
	"time"

	"github.com/a-h/templ"
	"github.com/gin-gonic/gin"
	enumspb "go.temporal.io/api/enums/v1"
	"go.uber.org/zap"
	"golang.org/x/sync/errgroup"

	pkgworkflows "github.com/nuonco/nuon/pkg/workflows"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/admin-dashboard/service/views"
)

// Known Temporal namespaces used by ctl-api workers.
var temporalWorkerNamespaces = []string{
	"general",
	"installs",
	"runners",
	"orgs",
	"components",
	"apps",
	"actions",
	"vcs",
	"onboardings",
}

// TemporalWorkers renders the temporal workers overview page.
func (s *service) TemporalWorkers(c *gin.Context) {
	ctx := c.Request.Context()

	namespacePollers, err := s.getTemporalWorkers(ctx)
	if err != nil {
		s.l.Error("failed to get temporal workers", zap.Error(err))
	}

	component := views.TemporalWorkers(namespacePollers, s.cfg.TemporalUIURL)
	templ.Handler(component).ServeHTTP(c.Writer, c.Request)
}

// TemporalWorkersTable returns just the table fragment for HTMX polling.
func (s *service) TemporalWorkersTable(c *gin.Context) {
	ctx := c.Request.Context()

	namespacePollers, err := s.getTemporalWorkers(ctx)
	if err != nil {
		s.l.Error("failed to get temporal workers for table", zap.Error(err))
	}

	component := views.TemporalWorkersTable(namespacePollers)
	templ.Handler(component).ServeHTTP(c.Writer, c.Request)
}

// TemporalWorkerDetail renders the detail page for a specific namespace.
func (s *service) TemporalWorkerDetail(c *gin.Context) {
	ctx := c.Request.Context()
	namespace := c.Param("namespace")

	info, err := s.getNamespaceWorkerInfo(ctx, namespace)
	if err != nil {
		s.l.Error("failed to get worker detail", zap.String("namespace", namespace), zap.Error(err))
		info = &views.NamespaceWorkerInfo{
			Namespace: namespace,
			Error:     err.Error(),
		}
	}

	component := views.TemporalWorkerDetailPage(info, s.cfg.TemporalUIURL)
	templ.Handler(component).ServeHTTP(c.Writer, c.Request)
}

// getTemporalWorkers fetches poller info for all known namespaces in parallel.
func (s *service) getTemporalWorkers(ctx context.Context) ([]*views.NamespaceWorkerInfo, error) {
	results := make([]*views.NamespaceWorkerInfo, len(temporalWorkerNamespaces))

	g, gCtx := errgroup.WithContext(ctx)
	for i, ns := range temporalWorkerNamespaces {
		g.Go(func() error {
			info, err := s.getNamespaceWorkerInfo(gCtx, ns)
			if err != nil {
				// Don't fail the whole page if one namespace is unreachable.
				results[i] = &views.NamespaceWorkerInfo{
					Namespace: ns,
					Error:     err.Error(),
				}
				return nil
			}
			results[i] = info
			return nil
		})
	}

	if err := g.Wait(); err != nil {
		return results, err
	}

	return results, nil
}

// getNamespaceWorkerInfo fetches worker/poller details for a single namespace.
func (s *service) getNamespaceWorkerInfo(ctx context.Context, namespace string) (*views.NamespaceWorkerInfo, error) {
	nsClient, err := s.temporalClient.GetNamespaceClient(namespace)
	if err != nil {
		return nil, fmt.Errorf("unable to get namespace client for %s: %w", namespace, err)
	}

	info := &views.NamespaceWorkerInfo{
		Namespace: namespace,
		TaskQueue: pkgworkflows.APITaskQueue,
	}

	// Describe workflow task queue pollers.
	wfResp, err := nsClient.DescribeTaskQueue(ctx, pkgworkflows.APITaskQueue, enumspb.TASK_QUEUE_TYPE_WORKFLOW)
	if err != nil {
		return nil, fmt.Errorf("unable to describe workflow task queue for %s: %w", namespace, err)
	}

	for _, p := range wfResp.GetPollers() {
		var lastAccess time.Time
		if p.GetLastAccessTime() != nil {
			lastAccess = p.GetLastAccessTime().AsTime()
		}
		info.WorkflowPollers = append(info.WorkflowPollers, views.PollerDetail{
			Identity:       p.GetIdentity(),
			LastAccessTime: lastAccess,
			RatePerSecond:  p.GetRatePerSecond(),
		})
	}

	if wfResp.GetStats() != nil {
		info.WorkflowStats = &views.TaskQueueStatsInfo{
			ApproximateBacklogCount: wfResp.GetStats().GetApproximateBacklogCount(),
			TasksAddRate:            wfResp.GetStats().GetTasksAddRate(),
			TasksDispatchRate:       wfResp.GetStats().GetTasksDispatchRate(),
		}
		if wfResp.GetStats().GetApproximateBacklogAge() != nil {
			info.WorkflowStats.ApproximateBacklogAge = wfResp.GetStats().GetApproximateBacklogAge().AsDuration()
		}
	}

	// Describe activity task queue pollers.
	actResp, err := nsClient.DescribeTaskQueue(ctx, pkgworkflows.APITaskQueue, enumspb.TASK_QUEUE_TYPE_ACTIVITY)
	if err != nil {
		// Activity queue might not have pollers; don't fail entirely.
		s.l.Warn("failed to describe activity task queue", zap.String("namespace", namespace), zap.Error(err))
		return info, nil
	}

	for _, p := range actResp.GetPollers() {
		var lastAccess time.Time
		if p.GetLastAccessTime() != nil {
			lastAccess = p.GetLastAccessTime().AsTime()
		}
		info.ActivityPollers = append(info.ActivityPollers, views.PollerDetail{
			Identity:       p.GetIdentity(),
			LastAccessTime: lastAccess,
			RatePerSecond:  p.GetRatePerSecond(),
		})
	}

	if actResp.GetStats() != nil {
		info.ActivityStats = &views.TaskQueueStatsInfo{
			ApproximateBacklogCount: actResp.GetStats().GetApproximateBacklogCount(),
			TasksAddRate:            actResp.GetStats().GetTasksAddRate(),
			TasksDispatchRate:       actResp.GetStats().GetTasksDispatchRate(),
		}
		if actResp.GetStats().GetApproximateBacklogAge() != nil {
			info.ActivityStats.ApproximateBacklogAge = actResp.GetStats().GetApproximateBacklogAge().AsDuration()
		}
	}

	return info, nil
}
