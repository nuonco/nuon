package helpers

import (
	"context"
	"fmt"

	"go.temporal.io/sdk/workflow"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	queueclient "github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/client"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/signal"
)

type webhookSubscriptionSignal struct {
	VCSConnectionID string `json:"vcs_connection_id"`
}

func (s *webhookSubscriptionSignal) Type() signal.SignalType           { return "vcs_webhook_subscription" }
func (s *webhookSubscriptionSignal) Validate(_ workflow.Context) error { return nil }
func (s *webhookSubscriptionSignal) Execute(_ workflow.Context) error  { return nil }

func (h *Helpers) EnqueueWebhookSubscriptionSignal(ctx context.Context, vcsConn *app.VCSConnection) error {
	queue, err := h.queueClient.GetQueueByOwner(ctx, vcsConn.ID, "vcs_connections")
	if err != nil {
		return fmt.Errorf("unable to find queue for vcs connection: %w", err)
	}

	_, err = h.queueClient.EnqueueSignal(ctx, &queueclient.EnqueueSignalRequest{
		QueueID: queue.ID,
		Signal: &webhookSubscriptionSignal{
			VCSConnectionID: vcsConn.ID,
		},
	})
	if err != nil {
		return fmt.Errorf("unable to enqueue webhook subscription signal: %w", err)
	}

	return nil
}
