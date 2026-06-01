package helpers

import (
	"context"
	"fmt"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/db/plugins"
	queueclient "github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/client"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/signal"
)

// EnqueueRunnerSignal enqueues a v2 signal to the runner-signals queue for the
// given runner. The queue must already exist (created during runner provisioning).
func (h *Helpers) EnqueueRunnerSignal(ctx context.Context, runnerID string, sig signal.Signal) error {
	q, err := h.queueClient.GetQueueByOwnerAndName(ctx, runnerID, "runners", runnerSignalsQueueName)
	if err != nil {
		return fmt.Errorf("unable to find runner-signals queue for runner %s: %w", runnerID, err)
	}

	if _, err := h.queueClient.EnqueueSignal(ctx, &queueclient.EnqueueSignalRequest{
		QueueID:   q.ID,
		OwnerID:   runnerID,
		OwnerType: plugins.TableName(h.db, app.Runner{}),
		Signal:    sig,
	}); err != nil {
		return fmt.Errorf("unable to enqueue runner signal %s: %w", sig.Type(), err)
	}

	return nil
}

const orgSignalsQueueName = "org-signals"

// EnqueueOrgSignal enqueues a v2 signal to the org-signals queue for the given org.
func (h *Helpers) EnqueueOrgSignal(ctx context.Context, orgID string, sig signal.Signal) error {
	q, err := h.queueClient.GetQueueByOwnerAndName(ctx, orgID, "orgs", orgSignalsQueueName)
	if err != nil {
		return fmt.Errorf("unable to find org-signals queue for org %s: %w", orgID, err)
	}

	if _, err := h.queueClient.EnqueueSignal(ctx, &queueclient.EnqueueSignalRequest{
		QueueID:   q.ID,
		OwnerID:   orgID,
		OwnerType: plugins.TableName(h.db, app.Org{}),
		Signal:    sig,
	}); err != nil {
		return fmt.Errorf("unable to enqueue org signal %s: %w", sig.Type(), err)
	}

	return nil
}
