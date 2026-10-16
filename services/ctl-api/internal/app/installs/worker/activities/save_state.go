package activities

import (
	"context"
	"time"

	"github.com/nuonco/nuon/pkg/types/state"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/db/routing"
	pkgstate "github.com/nuonco/nuon/services/ctl-api/internal/pkg/state"
)

type SaveStateRequest struct {
	State *state.State `validate:"required"`

	InstallID       string                         `validate:"required"`
	TriggeredByID   string                         `validate:"required"`
	TriggeredByType string                         `validate:"required"`
	GeneratedBy     app.InstallStateGenerateSource `validate:"required"`

	RefreshedPartials []pkgstate.PartialName
	RefreshStartedAt  time.Time
}

// @temporal-gen-v2 activity
func (a *Activities) SaveState(ctx context.Context, req *SaveStateRequest) (result *app.InstallState, err error) {
	regenerate := &pkgstate.ExecuteRegenerationRequest{
		InstallID: req.InstallID, TriggeredByID: req.TriggeredByID, TriggeredByType: req.TriggeredByType,
		ForceAll: len(req.RefreshedPartials) == 0,
	}
	for _, partial := range req.RefreshedPartials {
		regenerate.Targets = append(regenerate.Targets, pkgstate.PartialTarget{Name: partial})
	}
	if _, err := a.helpers.RegenerateInstallState(ctx, regenerate, false); err != nil {
		return nil, err
	}

	ctx = routing.WithoutReplica(ctx)
	var head app.InstallStateHead
	if err := a.db.WithContext(ctx).Where(app.InstallStateHead{InstallID: req.InstallID}).First(&head).Error; err != nil {
		return nil, err
	}
	result = &app.InstallState{}
	if err := a.db.WithContext(ctx).Where(app.InstallState{ID: head.CurrentStateID, InstallID: req.InstallID}).First(result).Error; err != nil {
		return nil, err
	}
	return result, nil
}
