package helpers

import (
	"context"
	"fmt"

	pkgworkflows "github.com/nuonco/nuon/pkg/workflows"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/db/plugins"
	queueclient "github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/client"
	emitterclient "github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/emitter/client"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/queuenames"
	queuesignal "github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/signal"
)

const (
	OrgHealthcheckCronsQueueName = queuenames.OrgHealthcheckCronsQueueName

	OrgRunnerHealthcheckSweepEmitter  = "org-runner-healthcheck-sweep"
	OrgProcessHealthcheckSweepEmitter = "org-process-healthcheck-sweep"
)

// EnsureOrgHealthcheckSweeps creates the org-healthcheck-crons queue and its two
// sweep cron emitters if missing. No-op unless the org has the
// org-healthcheck-sweeps feature. Idempotent.
func (h *Helpers) EnsureOrgHealthcheckSweeps(ctx context.Context, orgID string) error {
	sweeps, err := h.featuresClient.OrgHealthcheckSweepsEnabled(ctx, orgID)
	if err != nil {
		return fmt.Errorf("unable to evaluate org healthcheck sweeps flag: %w", err)
	}
	if !sweeps {
		return nil
	}

	spec, ok := queuenames.SpecByName(queuenames.OwnerOrgs, queuenames.OrgHealthcheckCronsQueueName)
	if !ok {
		return fmt.Errorf("org healthcheck crons queue is not registered")
	}
	q, err := h.queueClient.Create(ctx, &queueclient.CreateQueueRequest{
		OrgID:       &orgID,
		OwnerID:     orgID,
		OwnerType:   plugins.TableName(h.db, app.Org{}),
		Namespace:   pkgworkflows.RunnerHealthcheckCronsNamespace,
		Name:        spec.Name,
		MaxInFlight: spec.MaxInFlight,
		MaxDepth:    spec.MaxDepth,
	})
	if err != nil {
		return fmt.Errorf("unable to ensure org-healthcheck-crons queue: %w", err)
	}

	emitters, err := h.emitterClient.GetEmittersByQueueID(ctx, q.ID)
	if err != nil {
		return fmt.Errorf("unable to list emitters for org-healthcheck-crons queue: %w", err)
	}
	have := make(map[string]bool, len(emitters))
	for _, em := range emitters {
		have[em.Name] = true
	}

	sweepEmitters := []emitterclient.CreateEmitterRequest{
		{
			Name:         OrgRunnerHealthcheckSweepEmitter,
			Description:  "Org-wide runner health check sweep",
			CronSchedule: RunnerHealthcheckSchedule(h.cfg.Env),
			SignalType:   "org_runner_healthcheck_sweep",
		},
		{
			Name:         OrgProcessHealthcheckSweepEmitter,
			Description:  "Org-wide runner process health check sweep",
			CronSchedule: ProcessHealthcheckSchedule,
			SignalType:   "org_process_healthcheck_sweep",
		},
	}

	for _, req := range sweepEmitters {
		if have[req.Name] {
			continue
		}
		req.QueueID = q.ID
		req.Mode = app.QueueEmitterModeCron
		req.JitterWindow = runnerHealthcheckJitterWindow
		req.SignalTemplate = queuesignal.NewRaw(req.SignalType, map[string]any{
			"org_id": orgID,
		})
		if _, err := h.emitterClient.CreateEmitter(ctx, &req); err != nil {
			return fmt.Errorf("unable to create %s emitter: %w", req.Name, err)
		}
	}

	return nil
}
