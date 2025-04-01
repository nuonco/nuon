package worker

import (
	"go.temporal.io/sdk/workflow"
	"go.uber.org/zap"

	"github.com/powertoolsdev/mono/pkg/temporal/temporalzap"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app/general/signals"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app/general/worker/activities"
)

// @temporal-gen workflow
// @execution-timeout 10m
// @task-timeout 30s
func (w *Workflows) TerminateEventLoops(ctx workflow.Context, _ signals.RequestSignal) error {
	l := temporalzap.GetWorkflowLogger(ctx)
	namespaces := []string{
		"runners",
		"releases",
		"installs",
		"apps",
		"actions",
		"orgs",
		"components",
	}
	for _, ns := range namespaces {
		l.Info("terminating event loops", zap.String("namespace", ns))
		if _, err := activities.AwaitTerminateNamespaceWorkflowsByNamespace(ctx, ns); err != nil {
			l.Error("unable to terminate event loops", zap.String("namespace", ns), zap.Error(err))
		}
	}

	return nil
}
