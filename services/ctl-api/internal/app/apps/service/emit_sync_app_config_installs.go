package service

import (
	"context"

	"go.uber.org/zap"

	"github.com/nuonco/nuon/services/ctl-api/internal/app/apps/signals/syncappconfiginstalls"
	queueclient "github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/client"
)

func (s *service) emitSyncAppConfigInstallsSignal(ctx context.Context, appID, appConfigID string) {
	q, err := s.queueClient.GetQueueByOwner(ctx, appID, "apps")
	if err != nil {
		s.l.Warn("failed to get app queue for sync-app-config-installs signal",
			zap.String("app_id", appID),
			zap.Error(err))
		return
	}

	if _, err := s.queueClient.EnqueueSignal(ctx, &queueclient.EnqueueSignalRequest{
		QueueID: q.ID,
		Signal: &syncappconfiginstalls.Signal{
			AppID:          appID,
			NewAppConfigID: appConfigID,
		},
	}); err != nil {
		s.l.Warn("failed to enqueue sync-app-config-installs signal",
			zap.String("app_id", appID),
			zap.Error(err))
	}
}
