package state

import (
	"context"

	"go.uber.org/zap"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/installs/signals/v2/state/stateregenerate"
	queueclient "github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/client"
	pkgstate "github.com/nuonco/nuon/services/ctl-api/internal/pkg/state"
)

type HintStateManagerRequest struct {
	InstallID       string
	HintType        pkgstate.HintType
	EntityID        string
	TriggeredByID   string
	TriggeredByType app.InstallStateGenerateSource
}

// @temporal-gen-v2 activity
// @start-to-close-timeout 30s
func (a *Activities) HintStateManager(ctx context.Context, req *HintStateManagerRequest) error {
	queueID, err := a.getStateManagerQueueID(ctx, req.InstallID)
	if err != nil {
		a.l.Warn("unable to get state-manager queue", zap.String("install_id", req.InstallID), zap.Error(err))
		return nil
	}

	targets := pkgstate.TargetsForHint(req.HintType, req.EntityID)
	if _, err := a.queueClient.EnqueueSignal(ctx, &queueclient.EnqueueSignalRequest{
		QueueID: queueID,
		Signal: &stateregenerate.Signal{
			InstallID:       req.InstallID,
			Targets:         targets,
			ForceAll:        true,
			TriggeredByID:   req.TriggeredByID,
			TriggeredByType: req.TriggeredByType,
		},
		OwnerID:   req.InstallID,
		OwnerType: "installs",
	}); err != nil {
		a.l.Warn("unable to hint state manager",
			zap.String("install_id", req.InstallID),
			zap.String("hint_type", string(req.HintType)),
			zap.Error(err))
	}
	return nil
}
