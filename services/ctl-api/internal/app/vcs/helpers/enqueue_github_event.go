package helpers

import (
	"context"
	"fmt"

	"go.temporal.io/sdk/workflow"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	queueclient "github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/client"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/signal"
)

// githubEventSignal is a minimal signal type that matches the
// connection_event.Signal type string. We define it here to avoid
// an import cycle.
type githubEventSignal struct {
	GithubEventID string `json:"github_event_id"`
}

func (s *githubEventSignal) Type() signal.SignalType           { return "github_event" }
func (s *githubEventSignal) Validate(_ workflow.Context) error { return nil }
func (s *githubEventSignal) Execute(_ workflow.Context) error  { return nil }

// EnqueueGithubEvent enqueues a github_event signal to the webhook subscription's queue.
// The signal will find all VCS connections for the subscription's github install ID
// and fan out to matching app branches.
func (h *Helpers) EnqueueGithubEvent(ctx context.Context, sub *app.VCSWebhookSubscription, eventID string) error {
	queue, err := h.queueClient.GetQueueByOwner(ctx, sub.ID, "vcs_webhook_subscriptions")
	if err != nil {
		return fmt.Errorf("unable to find queue for webhook subscription: %w", err)
	}

	_, err = h.queueClient.EnqueueSignal(ctx, &queueclient.EnqueueSignalRequest{
		QueueID: queue.ID,
		Signal: &githubEventSignal{
			GithubEventID: eventID,
		},
	})
	if err != nil {
		return fmt.Errorf("unable to enqueue github event signal: %w", err)
	}

	return nil
}
