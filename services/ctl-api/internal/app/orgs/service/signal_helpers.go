package service

import (
	"context"
	"fmt"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/orgs/helpers"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/db/plugins"
	queueclient "github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/client"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/signal"
)

// getOrgQueueID returns the queue ID for the given org by queue name.
func (s *service) getOrgQueueID(ctx context.Context, orgID, queueName string) (string, error) {
	var queue app.Queue
	if res := s.db.WithContext(ctx).Where("owner_id = ? AND name = ?", orgID, queueName).First(&queue); res.Error != nil {
		return "", fmt.Errorf("unable to get org queue %s: %w", queueName, res.Error)
	}
	return queue.ID, nil
}

// getOrgSignalsQueueID returns the org-signals queue ID.
func (s *service) getOrgSignalsQueueID(ctx context.Context, orgID string) (string, error) {
	return s.getOrgQueueID(ctx, orgID, helpers.OrgSignalsQueueName)
}

// enqueueOrgSignal enqueues a v2 signal to the given org queue.
func (s *service) enqueueOrgSignal(ctx context.Context, queueID string, sig signal.Signal, orgID string) error {
	_, err := s.queueClient.EnqueueSignal(ctx, &queueclient.EnqueueSignalRequest{
		QueueID:   queueID,
		Signal:    sig,
		OwnerID:   orgID,
		OwnerType: plugins.TableName(s.db, app.Org{}),
	})
	return err
}
