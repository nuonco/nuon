package service

import (
	"context"
	"fmt"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	queueclient "github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/client"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/signal"
)

// useInstallQueues returns true when both AppBranches and Queues features are enabled.
func (s *service) useInstallQueues(ctx context.Context) (bool, error) {
	return s.featuresClient.AllFeaturesEnabled(ctx, app.OrgFeatureAppBranches, app.OrgFeatureQueues)
}

// getInstallQueueID returns the queue ID for the given install.
func (s *service) getInstallQueueID(ctx context.Context, installID string) (string, error) {
	var queue app.Queue
	if res := s.db.WithContext(ctx).Where("owner_id = ?", installID).First(&queue); res.Error != nil {
		return "", fmt.Errorf("unable to get install queue: %w", res.Error)
	}
	return queue.ID, nil
}

// enqueueInstallSignal enqueues a v2 signal to the given install queue.
func (s *service) enqueueInstallSignal(ctx context.Context, queueID string, sig signal.Signal) error {
	_, err := s.queueClient.EnqueueSignal(ctx, &queueclient.EnqueueSignalRequest{
		QueueID: queueID,
		Signal:  sig,
	})
	return err
}
