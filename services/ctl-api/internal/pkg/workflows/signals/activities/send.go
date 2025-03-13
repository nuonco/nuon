package activities

import (
	"context"

	appssignals "github.com/powertoolsdev/mono/services/ctl-api/internal/app/apps/signals"
	componentssignals "github.com/powertoolsdev/mono/services/ctl-api/internal/app/components/signals"
	generalsignals "github.com/powertoolsdev/mono/services/ctl-api/internal/app/general/signals"
	installssignals "github.com/powertoolsdev/mono/services/ctl-api/internal/app/installs/signals"
	orgssignals "github.com/powertoolsdev/mono/services/ctl-api/internal/app/orgs/signals"
	releasessignals "github.com/powertoolsdev/mono/services/ctl-api/internal/app/releases/signals"
	runnersignals "github.com/powertoolsdev/mono/services/ctl-api/internal/app/runners/signals"
)

type SendSignalRequest[T any] struct {
	ID string `validate:"required"`

	Signal T `validate:"required"`
}

// @temporal-gen activity
func (a *Activities) PkgSignalsSendRunnersSignal(ctx context.Context, req *SendSignalRequest[*runnersignals.Signal]) error {
	a.evClient.Send(ctx, req.ID, req.Signal)
	return nil
}

// @temporal-gen activity
func (a *Activities) PkgSignalsSendComponentsSignal(ctx context.Context, req *SendSignalRequest[*componentssignals.Signal]) error {
	a.evClient.Send(ctx, req.ID, req.Signal)
	return nil
}

// @temporal-gen activity
func (a *Activities) PkgSignalsSendInstallsSignal(ctx context.Context, req *SendSignalRequest[*installssignals.Signal]) error {
	a.evClient.Send(ctx, req.ID, req.Signal)
	return nil
}

// @temporal-gen activity
func (a *Activities) PkgSignalsSendAppsSignal(ctx context.Context, req *SendSignalRequest[*appssignals.Signal]) error {
	a.evClient.Send(ctx, req.ID, req.Signal)
	return nil
}

// @temporal-gen activity
func (a *Activities) PkgSignalsSendOrgsSignal(ctx context.Context, req *SendSignalRequest[*orgssignals.Signal]) error {
	a.evClient.Send(ctx, req.ID, req.Signal)
	return nil
}

// @temporal-gen activity
func (a *Activities) PkgSignalsSendReleasesSignal(ctx context.Context, req *SendSignalRequest[*releasessignals.Signal]) error {
	a.evClient.Send(ctx, req.ID, req.Signal)
	return nil
}

// @temporal-gen activity
func (a *Activities) PkgSignalsSendGeneralSignal(ctx context.Context, req *SendSignalRequest[*generalsignals.Signal]) error {
	a.evClient.Send(ctx, req.ID, req.Signal)
	return nil
}
