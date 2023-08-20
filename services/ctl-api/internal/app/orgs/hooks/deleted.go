package hooks

import (
	"context"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/app/orgs/worker"
	"go.uber.org/zap"
)

func (o *Hooks) Deleted(ctx context.Context, orgID string) {
	o.l.Info("org deleted", zap.String("org", orgID))

	signal := worker.Signal{
		DryRun:    o.cfg.EnableWorkersDryRun,
		Operation: worker.OperationDeprovision,
	}
	err := o.client.SignalWorkflowInNamespace(ctx,
		defaultNamespace,
		worker.EventLoopWorkflowID(orgID),
		"",
		orgID,
		signal,
	)
	o.l.Info("event workflow signaled", zap.Error(err))
}
