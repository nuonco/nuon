package service

import (
	"context"
	"fmt"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/installs/helpers"
	forgetinstall "github.com/nuonco/nuon/services/ctl-api/internal/app/orgs/signals/forget_install"
	dbgenerics "github.com/nuonco/nuon/services/ctl-api/internal/pkg/db/generics"
	queueclient "github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/client"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/signal"
)

func (s *service) getInstallQueueID(ctx context.Context, installID, queueName string) (string, error) {
	var queue app.Queue
	if res := s.db.WithContext(ctx).Where("owner_id = ? AND name = ?", installID, queueName).First(&queue); res.Error != nil {
		return "", fmt.Errorf("unable to get install queue %s: %w", queueName, res.Error)
	}
	return queue.ID, nil
}

func (s *service) getInstallWorkflowsQueueID(ctx context.Context, installID string) (string, error) {
	return s.getInstallQueueID(ctx, installID, helpers.InstallWorkflowsQueueName)
}

func (s *service) getInstallSignalsQueueID(ctx context.Context, installID string) (string, error) {
	return s.getInstallQueueID(ctx, installID, helpers.InstallSignalsQueueName)
}

func (s *service) getInstallStateManagerQueueID(ctx context.Context, installID string) (string, error) {
	return s.getInstallQueueID(ctx, installID, helpers.InstallStateManagerQueueName)
}

func (s *service) enqueueInstallSignal(ctx context.Context, queueID string, sig signal.Signal, ownerID, ownerType string) error {
	_, err := s.queueClient.EnqueueSignal(ctx, &queueclient.EnqueueSignalRequest{
		QueueID:   queueID,
		Signal:    sig,
		OwnerID:   ownerID,
		OwnerType: ownerType,
	})
	return err
}

func (s *service) enqueueOrgForgetInstallSignal(ctx context.Context, orgID, installID string) error {
	sig := &forgetinstall.Signal{
		OrgID:     orgID,
		InstallID: installID,
	}

	err := s.runnersHelpers.EnqueueOrgSignal(ctx, orgID, sig)
	if err == nil || !dbgenerics.IsGormErrRecordNotFound(err) {
		return err
	}

	if err := s.orgsHelpers.EnsureOrgQueue(ctx, orgID); err != nil {
		return err
	}
	return s.runnersHelpers.EnqueueOrgSignal(ctx, orgID, sig)
}
