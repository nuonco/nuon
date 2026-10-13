package activities

import (
	"context"
	"time"

	"github.com/nuonco/nuon/pkg/types/state"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/installs/helpers"
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
	started := time.Now()
	defer func() { a.stateMetrics.Record(ctx, "save", started, err) }()
	inputs, err := a.helpers.GetInstallInputsState(ctx, a.db, req.InstallID, false, false)
	if err != nil {
		return nil, err
	}
	snapshot := *req.State
	snapshot.Inputs = inputs
	helpers.MapLegacyFields(&snapshot)
	return a.helpers.PublishInstallState(ctx, &app.InstallState{
		InstallID: req.InstallID, TriggeredByID: req.TriggeredByID, TriggeredByType: req.TriggeredByType,
		State: &snapshot, GeneratedBy: req.GeneratedBy,
	}, req.RefreshedPartials)
}
