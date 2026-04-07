package helpers

import (
	"context"
	"fmt"

	githubevent "github.com/nuonco/nuon/services/ctl-api/internal/app/vcs/signals/v2/github_event"
	queueclient "github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/client"
)

// EnqueueGithubEventSignal looks up the queue for the given VCS connection
// and enqueues a github_event signal to process the webhook event.
func (h *Helpers) EnqueueGithubEventSignal(ctx context.Context, vcsConnectionID, webhookEventID string) error {
	q, err := h.queueClient.GetQueueByOwner(ctx, vcsConnectionID, "vcs_connections")
	if err != nil {
		return fmt.Errorf("unable to get vcs connection queue: %w", err)
	}

	if _, err := h.queueClient.EnqueueSignal(ctx, &queueclient.EnqueueSignalRequest{
		QueueID: q.ID,
		Signal: &githubevent.Signal{
			VCSConnectionID: vcsConnectionID,
			WebhookEventID:  webhookEventID,
		},
	}); err != nil {
		return fmt.Errorf("unable to enqueue github event signal: %w", err)
	}

	return nil
}
