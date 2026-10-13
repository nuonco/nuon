package activities

import (
	"context"
	"time"

	"github.com/nuonco/nuon/pkg/types/state"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/installs/helpers"
	pkgstate "github.com/nuonco/nuon/services/ctl-api/internal/pkg/state"
	"go.temporal.io/sdk/temporal"
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
	return nil, temporal.NewNonRetryableApplicationError("legacy state generation must restart with a revision baseline", helpers.InstallStateConflict, nil)
}
