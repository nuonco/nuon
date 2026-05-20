package promotion

import (
	"fmt"

	"go.temporal.io/sdk/workflow"
	"go.uber.org/zap"

	generalactivities "github.com/nuonco/nuon/services/ctl-api/internal/app/general/worker/activities"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/log"
	queueclient "github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/client"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/signal"
)

const SignalType signal.SignalType = "general-promotion"

var _ signal.Signal = (*Signal)(nil)

type Signal struct {
	Tag string `json:"tag"`
}

func (s *Signal) Type() signal.SignalType { return SignalType }

func (s *Signal) Validate(_ workflow.Context) error {
	return nil
}

func (s *Signal) Execute(ctx workflow.Context) error {
	l, _ := log.WorkflowLogger(ctx)

	// Mark all active runner processes for shutdown in a single query.
	processResp, err := generalactivities.AwaitMarkActiveProcessesForShutdown(ctx, generalactivities.MarkActiveProcessesForShutdownRequest{})
	if err != nil {
		return fmt.Errorf("unable to mark processes for shutdown: %w", err)
	}
	if l != nil {
		l.Info("marked processes for shutdown", zap.Int64("rows_affected", processResp.RowsAffected))
	}

	// Request continue-as-new on all queues so they restart with the new
	// code version.
	queueResp, err := queueclient.AwaitRequestCANAll(ctx, &queueclient.RequestCANAllRequest{})
	if err != nil {
		return fmt.Errorf("unable to request CAN on all queues: %w", err)
	}
	if l != nil {
		l.Info("requested CAN on queues", zap.Int64("rows_affected", queueResp.RowsAffected))
	}

	// Idempotent + additive — chained off promote so new label-tagged
	// orgs land without a separate POST.
	autoLinkResp, err := generalactivities.AwaitEnsureSlackAutoLinks(ctx, generalactivities.EnsureSlackAutoLinksRequest{})
	if err != nil {
		return fmt.Errorf("ensure slack auto links: %w", err)
	}
	if l != nil {
		l.Info("reconciled slack auto-links",
			zap.Int("orgs_considered", autoLinkResp.OrgsConsidered),
			zap.Int("links_created", autoLinkResp.LinksCreated),
			zap.Int("subs_seeded", autoLinkResp.SubsSeeded),
		)
	}

	// Start (or replace) the enqueuer sweep and metrics cron workflows
	// as top-level Temporal cron workflows via the client.
	cronResp, err := generalactivities.AwaitEnsureCronWorkflows(ctx, generalactivities.EnsureCronWorkflowsRequest{})
	if err != nil {
		return fmt.Errorf("ensure cron workflows: %w", err)
	}
	if l != nil {
		l.Info("ensured cron workflows",
			zap.Bool("sweep_started", cronResp.SweepStarted),
			zap.Bool("metrics_started", cronResp.MetricsStarted),
		)
	}

	return nil
}
