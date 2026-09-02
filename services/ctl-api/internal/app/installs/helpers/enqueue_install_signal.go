package helpers

import (
	"context"
	"fmt"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/flow/signals/executeflow"
	queueclient "github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/client"
	queuesignal "github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/signal"
)

// EnqueueInstallSignal enqueues a v2 signal onto the named queue for an install.
func (h *Helpers) EnqueueInstallSignal(ctx context.Context, installID, queueName string, sig queuesignal.Signal) error {
	return h.enqueueInstallSignal(ctx, installID, queueName, sig, "", "")
}

// EnqueueInstallWorkflow starts an install workflow. The queue signal must be owned
// by the workflow: approve/cancel/retry/pause look the handler up by owner_id.
func (h *Helpers) EnqueueInstallWorkflow(ctx context.Context, installID, workflowID string) error {
	return h.enqueueInstallSignal(ctx, installID, InstallWorkflowsQueueName,
		&executeflow.Signal{WorkflowID: workflowID},
		workflowID, (&app.Workflow{}).TableName())
}

func (h *Helpers) enqueueInstallSignal(ctx context.Context, installID, queueName string, sig queuesignal.Signal, ownerID, ownerType string) error {
	var q app.Queue
	if res := h.db.WithContext(ctx).
		Where(app.Queue{OwnerID: installID, Name: queueName}).
		First(&q); res.Error != nil {
		return fmt.Errorf("unable to find %s queue for install %s: %w", queueName, installID, res.Error)
	}

	_, err := h.queueClient.EnqueueSignal(ctx, &queueclient.EnqueueSignalRequest{
		QueueID:   q.ID,
		Signal:    sig,
		OwnerID:   ownerID,
		OwnerType: ownerType,
	})
	return err
}
