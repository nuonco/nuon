package service

import (
	"context"

	"go.uber.org/zap"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/apps/signals/appconfigsynced"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/cctx"
	queueclient "github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/client"
)

func (s *service) emitAppConfigSyncedSignal(ctx context.Context, orgID, appID, appBranchID string) {
	var currentApp app.App
	if err := s.db.WithContext(ctx).Where(app.App{ID: appID}).First(&currentApp).Error; err != nil {
		s.l.Warn("failed to look up app for config synced signal", zap.Error(err))
		return
	}

	actorEmail := ""
	if acct, err := cctx.AccountFromContext(ctx); err == nil && acct.Email != "" {
		actorEmail = acct.Email
	}

	branchName := ""
	if appBranchID != "" {
		var branch app.AppBranch
		if err := s.db.WithContext(ctx).Where(app.AppBranch{ID: appBranchID}).First(&branch).Error; err == nil {
			branchName = branch.Name
		}
	}

	q, err := s.queueClient.GetQueueByOwner(ctx, appID, "apps")
	if err != nil {
		s.l.Warn("failed to get app queue for config synced signal", zap.Error(err))
		return
	}

	if _, err := s.queueClient.EnqueueSignal(ctx, &queueclient.EnqueueSignalRequest{
		QueueID: q.ID,
		Signal: &appconfigsynced.Signal{
			AppID:       appID,
			AppBranchID: appBranchID,
			AppName:     currentApp.Name,
			BranchName:  branchName,
			ActorEmail:  actorEmail,
		},
	}); err != nil {
		s.l.Warn("failed to enqueue app-config-synced signal", zap.Error(err), zap.String("app_id", appID))
	}
}
