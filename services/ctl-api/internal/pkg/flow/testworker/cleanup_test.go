package testworker

import (
	"context"
	"time"

	"go.temporal.io/api/workflowservice/v1"
	"go.uber.org/zap"

	temporalclient "github.com/nuonco/nuon/pkg/temporal/client"
)

// cleanupStaleWorkflows terminates workflows left Running in the test
// namespace by previous test runs. When a test binary exits, its queue and
// handler workflows lose their only poller; the next run's worker inherits
// that backlog, which starves the live tests and blows their poll budgets.
// This runs as an fx.Invoke so it completes before the worker's OnStart hook
// begins polling.
//
// The namespace is registered here first: the frontend's namespace registry
// negatively caches a name it failed to resolve for the cache refresh
// interval, so listing a not-yet-registered namespace would leave every
// StartWorkflow/poll in that window failing with NamespaceNotFound even after
// the worker registers it.
func cleanupStaleWorkflows(tc temporalclient.Client, l *zap.Logger) {
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	if err := registerTestNamespace(ctx, tc); err != nil {
		l.Warn("stale workflow cleanup: unable to register namespace", zap.Error(err))
		return
	}

	c, err := tc.GetNamespaceClient(defaultNamespace)
	if err != nil {
		l.Warn("stale workflow cleanup: unable to get namespace client", zap.Error(err))
		return
	}

	var terminated int
	var nextPageToken []byte
	for {
		resp, err := c.ListWorkflow(ctx, &workflowservice.ListWorkflowExecutionsRequest{
			Namespace:     defaultNamespace,
			Query:         `ExecutionStatus='Running'`,
			NextPageToken: nextPageToken,
		})
		if err != nil {
			l.Warn("stale workflow cleanup: list failed", zap.Error(err))
			return
		}

		for _, ex := range resp.Executions {
			// Terminate errors are expected for workflows that finished
			// between list and terminate (visibility lag) — ignore them.
			if err := c.TerminateWorkflow(ctx,
				ex.Execution.WorkflowId, ex.Execution.RunId,
				"stale test workflow cleanup"); err == nil {
				terminated++
			}
		}

		nextPageToken = resp.NextPageToken
		if len(nextPageToken) == 0 {
			break
		}
	}

	if terminated > 0 {
		l.Info("terminated stale test workflows from previous runs",
			zap.Int("count", terminated))
	}
}
