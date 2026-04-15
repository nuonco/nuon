package service

import (
	"context"
	"fmt"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/installs/helpers"
	queueclient "github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/client"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/signal"
)

// getInstallQueueID returns the queue ID for the given install by queue name.
func (s *service) getInstallQueueID(ctx context.Context, installID, queueName string) (string, error) {
	var queue app.Queue
	if res := s.db.WithContext(ctx).Where("owner_id = ? AND name = ?", installID, queueName).First(&queue); res.Error != nil {
		return "", fmt.Errorf("unable to get install queue %s: %w", queueName, res.Error)
	}
	return queue.ID, nil
}

// getInstallWorkflowsQueueID returns the install-workflows queue ID.
func (s *service) getInstallWorkflowsQueueID(ctx context.Context, installID string) (string, error) {
	return s.getInstallQueueID(ctx, installID, helpers.InstallWorkflowsQueueName)
}

// enqueueInstallSignal enqueues a v2 signal to the given install queue.
func (s *service) enqueueInstallSignal(ctx context.Context, queueID string, sig signal.Signal, ownerID, ownerType string) error {
	_, err := s.queueClient.EnqueueSignal(ctx, &queueclient.EnqueueSignalRequest{
		QueueID:   queueID,
		Signal:    sig,
		OwnerID:   ownerID,
		OwnerType: ownerType,
	})
	return err
}
