package helpers

import (
	"context"
	"fmt"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/db/plugins"
	queueclient "github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/client"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/signal"
)

type EnqueueOrgSignalParams struct {
	OrgID  string
	Signal signal.Signal

	IdempotencyKey string
}

func (h *Helpers) EnqueueOrgSignal(ctx context.Context, params EnqueueOrgSignalParams) error {
	queue, err := h.ensureOrgQueue(ctx, params.OrgID)
	if err != nil {
		return err
	}

	if _, err := h.queueClient.EnqueueSignal(ctx, &queueclient.EnqueueSignalRequest{
		QueueID:        queue.ID,
		Signal:         params.Signal,
		OwnerID:        params.OrgID,
		OwnerType:      plugins.TableName(h.db, app.Org{}),
		IdempotencyKey: params.IdempotencyKey,
	}); err != nil {
		return fmt.Errorf("unable to enqueue %s signal: %w", params.Signal.Type(), err)
	}

	return nil
}
